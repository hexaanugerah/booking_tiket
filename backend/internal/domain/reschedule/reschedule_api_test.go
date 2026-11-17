package reschedule_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KawahPutih/backend/internal/domain/booking"
	reschedule "github.com/KawahPutih/backend/internal/domain/reschedule"
	"github.com/KawahPutih/backend/internal/domain/ticketing"
	"github.com/KawahPutih/backend/internal/domain/tickettype"
	"github.com/KawahPutih/backend/pkg/jwtutil"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type sentMail struct {
	To      string
	Subject string
	Body    string
}

type captureMailer struct {
	sent []sentMail
}

func (m *captureMailer) Send(ctx context.Context, to, subject, body string) error {
	m.sent = append(m.sent, sentMail{To: to, Subject: subject, Body: body})
	return nil
}

type envelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Meta    struct {
		Total int64 `json:"total"`
	} `json:"meta"`
}

type reschedulePayload struct {
	ID           string `json:"id"`
	BookingID    string `json:"booking_id"`
	BookingCode  string `json:"booking_code"`
	VisitorName  string `json:"visitor_name"`
	Email        string `json:"email"`
	OldVisitDate string `json:"old_visit_date"`
	NewVisitDate string `json:"new_visit_date"`
	Reason       string `json:"reason"`
	Status       string `json:"status"`
	AdminNote    string `json:"admin_note"`
	ProcessedAt  string `json:"processed_at"`
}

type testEnv struct {
	t        *testing.T
	db       *gorm.DB
	router   *gin.Engine
	mailer   *captureMailer
	userTok  string
	adminTok string
}

func newTestEnv(t *testing.T) *testEnv {
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

	require.NoError(t, db.AutoMigrate(
		&booking.Booking{},
		&booking.BookingItem{},
		&tickettype.TicketType{},
		&ticketing.Ticket{},
		&reschedule.Reschedule{},
	))

	m := &captureMailer{}
	svc := reschedule.NewService(reschedule.NewRepository(db), m)
	h := reschedule.NewHandler(svc)

	jwt := jwtutil.NewManager("test-access-secret", "test-refresh-secret", time.Hour, 24*time.Hour, "test")
	router := gin.New()
	reschedule.RegisterRoutes(router.Group("/api/v1"), h, jwt)

	userTok, _, err := jwt.GenerateAccessToken("user-1", "user@example.com", []string{"visitor"})
	require.NoError(t, err)
	adminTok, _, err := jwt.GenerateAccessToken("admin-1", "admin@example.com", []string{"ADMIN"})
	require.NoError(t, err)

	return &testEnv{t: t, db: db, router: router, mailer: m, userTok: userTok, adminTok: adminTok}
}

func (e *testEnv) call(method, path, token string, body any) (*httptest.ResponseRecorder, envelope) {
	e.t.Helper()
	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(e.t, err)
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, req)

	var env envelope
	require.NoError(e.t, json.Unmarshal(rec.Body.Bytes(), &env), rec.Body.String())
	return rec, env
}

func (e *testEnv) submit(bookingID, newDate string) (*httptest.ResponseRecorder, envelope) {
	return e.call(http.MethodPost, "/api/v1/bookings/"+bookingID+"/reschedule", e.userTok, map[string]string{
		"new_visit_date": newDate,
		"reason":         "Ada acara keluarga yang tidak bisa ditinggalkan.",
	})
}

func (e *testEnv) decide(id string, status, note string) (*httptest.ResponseRecorder, envelope) {
	return e.call(http.MethodPatch, "/api/v1/admin/reschedules/"+id+"/status", e.adminTok, map[string]string{
		"status":     status,
		"admin_note": note,
	})
}

func (e *testEnv) seedTicket(id string, capacity int, price int64) {
	e.t.Helper()
	require.NoError(e.t, e.db.Create(&tickettype.TicketType{
		Base:     tickettype.Base{ID: id},
		Code:     "CODE-" + id,
		Name:     "Tiket " + id,
		Category: "ENTRANCE",
		Price:    price,
		Capacity: capacity,
		IsActive: true,
	}).Error)
}

func (e *testEnv) seedBooking(id, code, visitor, email, visitDate string, qty int, ticketID string, price int64) {
	e.t.Helper()
	require.NoError(e.t, e.db.Create(&booking.Booking{
		Base:         booking.Base{ID: id},
		BookingCode:  code,
		UserID:       "user-1",
		VisitorName:  visitor,
		Email:        email,
		Phone:        "08123456789",
		VisitDate:    visitDate,
		Status:       booking.StatusPaid,
		TotalTickets: qty,
		Subtotal:     price * int64(qty),
		Total:        price * int64(qty),
	}).Error)
	require.NoError(e.t, e.db.Create(&booking.BookingItem{
		Base:       booking.Base{ID: "item-" + id},
		BookingID:  id,
		TicketID:   ticketID,
		TicketName: "Tiket " + ticketID,
		Price:      price,
		Quantity:   qty,
	}).Error)
	require.NoError(e.t, e.db.Create(&ticketing.Ticket{
		Base:        ticketing.Base{ID: "ticket-" + id},
		BookingID:   id,
		BookingCode: code,
		VisitorName: visitor,
		Email:       email,
		VisitDate:   visitDate,
		TicketName:  "Tiket " + ticketID,
		Quantity:    qty,
		Total:       price * int64(qty),
		Status:      ticketing.TicketValid,
		GeneratedAt: time.Now(),
		QRData:      "qr-" + id,
	}).Error)
}

func (e *testEnv) bookingVisitDate(id string) string {
	e.t.Helper()
	var row struct{ VisitDate string }
	require.NoError(e.t, e.db.Table("bookings").Select("visit_date").Where("id = ?", id).Take(&row).Error)
	return row.VisitDate
}

func (e *testEnv) ticketVisitDate(bookingID string) string {
	e.t.Helper()
	var row struct{ VisitDate string }
	require.NoError(e.t, e.db.Table("tickets").Select("visit_date").Where("booking_id = ?", bookingID).Take(&row).Error)
	return row.VisitDate
}

func dateAfter(days int) string {
	return time.Now().AddDate(0, 0, days).Format("2006-01-02")
}

func decodeData(t *testing.T, env envelope) reschedulePayload {
	t.Helper()
	var payload reschedulePayload
	require.NoError(t, json.Unmarshal(env.Data, &payload))
	return payload
}

func TestRescheduleApprovalFlow(t *testing.T) {
	env := newTestEnv(t)
	oldDate := dateAfter(3)
	newDate := dateAfter(5)
	env.seedTicket("tt-1", 100, 50000)
	env.seedBooking("bk-1", "KP-TEST-1", "Budi Santoso", "budi@example.com", oldDate, 2, "tt-1", 50000)

	// Pengunjung biasa tidak boleh membuka endpoint admin.
	rec, _ := env.call(http.MethodGet, "/api/v1/admin/reschedules", env.userTok, nil)
	require.Equal(t, http.StatusForbidden, rec.Code)

	// Pengajuan tidak langsung mengubah tanggal booking.
	rec, envl := env.submit("bk-1", newDate)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	created := decodeData(t, envl)
	require.Equal(t, reschedule.StatusPending, created.Status)
	require.Equal(t, oldDate, created.OldVisitDate)
	require.Equal(t, newDate, created.NewVisitDate)
	require.Equal(t, oldDate, env.bookingVisitDate("bk-1"))

	// Pengajuan kedua saat masih pending ditolak.
	rec, envl = env.submit("bk-1", dateAfter(7))
	require.Equal(t, http.StatusConflict, rec.Code)
	require.Contains(t, envl.Message, "menunggu persetujuan")

	// Admin melihat pengajuan beserta data pengunjung.
	rec, envl = env.call(http.MethodGet, "/api/v1/admin/reschedules?status=PENDING", env.adminTok, nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.EqualValues(t, 1, envl.Meta.Total)
	var list []reschedulePayload
	require.NoError(t, json.Unmarshal(envl.Data, &list))
	require.Len(t, list, 1)
	require.Equal(t, "Budi Santoso", list[0].VisitorName)
	require.Equal(t, "budi@example.com", list[0].Email)
	require.Equal(t, "KP-TEST-1", list[0].BookingCode)

	// Setujui: booking dan e-tiket ikut berubah, email terkirim.
	rec, envl = env.decide(created.ID, "APPROVED", "Kuota tersedia.")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	approved := decodeData(t, envl)
	require.Equal(t, reschedule.StatusApproved, approved.Status)
	require.NotEmpty(t, approved.ProcessedAt)
	require.Equal(t, newDate, env.bookingVisitDate("bk-1"))
	require.Equal(t, newDate, env.ticketVisitDate("bk-1"))
	require.Len(t, env.mailer.sent, 1)
	require.Equal(t, "budi@example.com", env.mailer.sent[0].To)
	require.Contains(t, env.mailer.sent[0].Subject, "KP-TEST-1")
	require.Contains(t, env.mailer.sent[0].Body, newDate)

	// Pengajuan yang sudah diproses tidak dapat diproses ulang.
	rec, _ = env.decide(created.ID, "APPROVED", "")
	require.Equal(t, http.StatusConflict, rec.Code)

	// Pengajuan kedua ditolak: tanggal booking tidak berubah dan jatah tidak
	// terpakai. Pengajuan yang masih pending memblokir pengajuan berikutnya.
	secondDate := dateAfter(9)
	rec, envl = env.submit("bk-1", secondDate)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	second := decodeData(t, envl)

	rec, envl = env.submit("bk-1", dateAfter(11))
	require.Equal(t, http.StatusConflict, rec.Code)
	require.Contains(t, envl.Message, "menunggu persetujuan")

	rec, envl = env.decide(second.ID, "REJECTED", "Tanggal baru penuh.")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	rejected := decodeData(t, envl)
	require.Equal(t, reschedule.StatusRejected, rejected.Status)
	require.Equal(t, newDate, env.bookingVisitDate("bk-1"))

	rec, envl = env.submit("bk-1", dateAfter(11))
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	third := decodeData(t, envl)

	rec, envl = env.decide(third.ID, "APPROVED", "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, dateAfter(11), env.bookingVisitDate("bk-1"))

	// Dua jadwal ulang sudah disetujui: pengajuan berikutnya ditolak.
	rec, envl = env.submit("bk-1", dateAfter(13))
	require.Equal(t, http.StatusConflict, rec.Code)
	require.Contains(t, envl.Message, "Batas maksimal 2 kali")
}

func TestRescheduleApprovalRechecksCapacity(t *testing.T) {
	env := newTestEnv(t)
	env.seedTicket("tt-1", 3, 50000)
	env.seedBooking("bk-1", "KP-TEST-2", "Siti Aminah", "siti@example.com", dateAfter(3), 2, "tt-1", 50000)

	newDate := dateAfter(5)
	rec, envl := env.submit("bk-1", newDate)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	created := decodeData(t, envl)

	// Booking lain mengisi kuota tanggal baru setelah pengajuan dibuat.
	env.seedBooking("bk-2", "KP-TEST-3", "Andi Wijaya", "andi@example.com", newDate, 2, "tt-1", 50000)

	rec, envl = env.decide(created.ID, "APPROVED", "")
	require.Equal(t, http.StatusConflict, rec.Code)
	require.Contains(t, envl.Message, "tidak tersedia")
	require.Equal(t, dateAfter(3), env.bookingVisitDate("bk-1"))
	require.Empty(t, env.mailer.sent)
}

func TestRescheduleApprovalFailsWhenBookingDateChanged(t *testing.T) {
	env := newTestEnv(t)
	env.seedTicket("tt-1", 100, 50000)
	env.seedBooking("bk-1", "KP-TEST-4", "Rina Lestari", "rina@example.com", dateAfter(3), 1, "tt-1", 50000)

	rec, envl := env.submit("bk-1", dateAfter(5))
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	created := decodeData(t, envl)

	// Tanggal booking berubah di luar alur (mis. diubah admin lain).
	require.NoError(t, env.db.Table("bookings").Where("id = ?", "bk-1").
		Update("visit_date", dateAfter(4)).Error)

	rec, envl = env.decide(created.ID, "APPROVED", "")
	require.Equal(t, http.StatusConflict, rec.Code)
	require.Contains(t, envl.Message, "sudah berubah")
	require.Equal(t, dateAfter(4), env.bookingVisitDate("bk-1"))
	require.Empty(t, env.mailer.sent)
}

func TestRescheduleRejectsInvalidPayload(t *testing.T) {
	env := newTestEnv(t)
	env.seedTicket("tt-1", 100, 50000)
	env.seedBooking("bk-1", "KP-TEST-5", "Dewi Kartika", "dewi@example.com", dateAfter(3), 1, "tt-1", 50000)

	rec, _ := env.call(http.MethodPost, "/api/v1/bookings/bk-1/reschedule", env.userTok, map[string]string{
		"new_visit_date": dateAfter(5),
		"reason":         "pendek",
	})
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	rec, _ = env.call(http.MethodPost, "/api/v1/bookings/bk-1/reschedule", env.userTok, map[string]string{
		"new_visit_date": "",
		"reason":         "Alasan yang cukup panjang untuk lolos validasi.",
	})
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)

	rec, _ = env.call(http.MethodGet, "/api/v1/bookings/bk-1/reschedules", "invalid-token", nil)
	require.Equal(t, http.StatusUnauthorized, rec.Code)

}
