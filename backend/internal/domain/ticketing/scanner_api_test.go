package ticketing_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KawahPutih/backend/internal/domain/ticketing"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Uji alur scan dari halaman petugas: kamera mengirim hasil decode QR apa adanya
// sebagai ticket_id ke POST /api/v1/petugas/scanner/validate. Payload di sini
// sama dengan qr_data e-tiket (KPT-...) yang tercetak di PDF.

const scannedQR = "KPT-9F8E7D6C5B4A39281706"

type scanEnvelope struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Success bool   `json:"success"`
		Status  string `json:"status"`
		Message string `json:"message"`
		Ticket  struct {
			ID          string `json:"id"`
			BookingCode string `json:"booking_code"`
			VisitorName string `json:"visitor_name"`
			Status      string `json:"status"`
			QRData      string `json:"qr_data"`
		} `json:"ticket"`
	} `json:"data"`
}

func newScanEnv(t *testing.T) (*gin.Engine, *gorm.DB, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	require.NoError(t, db.AutoMigrate(&ticketing.Ticket{}, &ticketing.ScanRecord{}))

	svc := ticketing.NewService(ticketing.NewRepository(db), nil, zap.NewNop(), nil)
	h := ticketing.NewHandler(ticketing.NewUsecase(svc))

	jwt := jwtutil.NewManager("test-access-secret", "test-refresh-secret", time.Hour, 24*time.Hour, "test")
	router := gin.New()
	ticketing.RegisterRoutes(router.Group("/api/v1"), h, jwt)

	token, _, err := jwt.GenerateAccessToken("petugas-1", "petugas@example.com", []string{"PETUGAS"})
	require.NoError(t, err)

	require.NoError(t, db.Create(&ticketing.Ticket{
		Base:        ticketing.Base{ID: "ticket-1"},
		BookingID:   "b1e2d3c4-5f60-4718-8293-a4b5c6d7e8f9",
		BookingCode: "KP-20261119-7f3c9a",
		VisitorName: "Budi Santoso",
		Email:       "budi@example.com",
		VisitDate:   time.Now().Format("2006-01-02"),
		TicketName:  "Tiket Dewasa x2",
		Quantity:    2,
		Total:       100000,
		Status:      ticketing.TicketValid,
		GeneratedAt: time.Now(),
		QRData:      scannedQR,
	}).Error)

	return router, db, token
}

func postScan(t *testing.T, router *gin.Engine, token, ref string) (*httptest.ResponseRecorder, scanEnvelope) {
	t.Helper()
	body, err := json.Marshal(map[string]string{
		"ticket_id": ref,
		"gate":      "Gerbang 1",
		"petugas":   "Petugas Demo",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/petugas/scanner/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var env scanEnvelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env), rec.Body.String())
	return rec, env
}

func TestScannerValidateQRPayloadFromTicketPDF(t *testing.T) {
	router, db, token := newScanEnv(t)

	// Pindai pertama: QR teks dari PDF harus valid.
	rec, env := postScan(t, router, token, scannedQR)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.True(t, env.Data.Success)
	require.Equal(t, "VALID", env.Data.Status)
	require.Equal(t, "KP-20261119-7f3c9a", env.Data.Ticket.BookingCode)
	require.Equal(t, scannedQR, env.Data.Ticket.QRData)
	require.Contains(t, env.Data.Message, "valid")

	// Pindai kedua: tiket sudah terpakai.
	rec, env = postScan(t, router, token, scannedQR)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.False(t, env.Data.Success)
	require.Equal(t, "USED", env.Data.Status)

	// QR yang tidak dikenal ditolak, bukan error server.
	rec, env = postScan(t, router, token, "KPT-TIDAKADA")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.False(t, env.Data.Success)
	require.Equal(t, "INVALID", env.Data.Status)

	// Tanpa token petugas, endpoint menolak.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/petugas/scanner/validate", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	recUnauth := httptest.NewRecorder()
	router.ServeHTTP(recUnauth, req)
	require.Equal(t, http.StatusUnauthorized, recUnauth.Code)

	// Hanya tiket yang ditemukan yang dicatat sebagai scan record:
	// satu VALID dan satu USED. QR tak dikenal hanya masuk audit log.
	var scans int64
	require.NoError(t, db.Model(&ticketing.ScanRecord{}).Count(&scans).Error)
	require.EqualValues(t, 2, scans)
}
