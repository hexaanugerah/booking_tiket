package seed

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/KawahPutih/backend/internal/domain/article"
	"github.com/KawahPutih/backend/internal/domain/booking"
	"github.com/KawahPutih/backend/internal/domain/destinationinfo"
	"github.com/KawahPutih/backend/internal/domain/faq"
	"github.com/KawahPutih/backend/internal/domain/facility"
	tourismpackage "github.com/KawahPutih/backend/internal/domain/package"
	"github.com/KawahPutih/backend/internal/domain/payments"
	"github.com/KawahPutih/backend/internal/domain/review"
	"github.com/KawahPutih/backend/internal/domain/role"
	"github.com/KawahPutih/backend/internal/domain/ticketing"
	"github.com/KawahPutih/backend/internal/domain/tickettype"
	"github.com/KawahPutih/backend/internal/domain/visitorrule"
)

func Run(ctx context.Context, db *gorm.DB, log *zap.Logger) error {
	if err := seedRoles(ctx, db); err != nil {
		return fmt.Errorf("seed roles: %w", err)
	}
	if err := seedUsers(ctx, db); err != nil {
		return fmt.Errorf("seed users: %w", err)
	}
	if err := seedTicketTypes(ctx, db); err != nil {
		return fmt.Errorf("seed ticket types: %w", err)
	}
	if err := seedPackages(ctx, db); err != nil {
		return fmt.Errorf("seed packages: %w", err)
	}
	if err := seedDestinationInfo(ctx, db); err != nil {
		return fmt.Errorf("seed destination info: %w", err)
	}
	if err := seedFAQs(ctx, db); err != nil {
		return fmt.Errorf("seed faqs: %w", err)
	}
	if err := seedFacilities(ctx, db); err != nil {
		return fmt.Errorf("seed facilities: %w", err)
	}
	if err := seedArticles(ctx, db); err != nil {
		return fmt.Errorf("seed articles: %w", err)
	}
	if err := seedVisitorRules(ctx, db); err != nil {
		return fmt.Errorf("seed visitor rules: %w", err)
	}
	if err := seedDemoBookings(ctx, db); err != nil {
		return fmt.Errorf("seed demo bookings: %w", err)
	}
	// Seeder galeri sengaja TIDAK dipasang: galeri harus mulai kosong.
	if err := seedPaymentDestinations(ctx, db); err != nil {
		return fmt.Errorf("seed payment destinations: %w", err)
	}
	log.Info("seed data ready")
	return nil
}

// seedPaymentDestinations mengisi daftar awal tujuan pembayaran.
// Untuk keperluan UJI COBA, nomor rekening/HP diisi angka ACAK — bukan kosong —
// supaya halaman pembayaran langsung menampilkan nomor tujuan. Nomor ini tetap
// BISA DIUBAH kapan saja: lewat endpoint admin (PUT /admin/payment-destinations/:id)
// atau langsung di tabel payment_destinations.
func seedPaymentDestinations(ctx context.Context, db *gorm.DB) error {
	var count int64
	if err := db.WithContext(ctx).Model(&payments.PaymentDestination{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		defaults := []payments.PaymentDestination{
			{MethodCode: payments.MethodQRIS, ProviderName: "QRIS (semua e-wallet & mobile banking)", Active: false, SortOrder: 1},
			{MethodCode: payments.MethodBCAVA, ProviderName: "BCA", AccountNumber: randomAccountNumber("8808", 8), AccountName: "Kawah Putih Official", Active: true, SortOrder: 10},
			{MethodCode: payments.MethodBRIVA, ProviderName: "BRI", AccountNumber: randomAccountNumber("88908", 7), AccountName: "Kawah Putih Official", Active: true, SortOrder: 11},
			{MethodCode: payments.MethodMandiriVA, ProviderName: "Mandiri", AccountNumber: randomAccountNumber("8908", 8), AccountName: "Kawah Putih Official", Active: true, SortOrder: 12},
			{MethodCode: payments.MethodBNIVA, ProviderName: "BNI", AccountNumber: randomAccountNumber("988", 9), AccountName: "Kawah Putih Official", Active: true, SortOrder: 13},
			{MethodCode: payments.MethodGoPay, ProviderName: "GoPay", AccountNumber: randomAccountNumber("08", 10), AccountName: "Kawah Putih Official", Active: true, SortOrder: 20},
			{MethodCode: payments.MethodDANA, ProviderName: "DANA", AccountNumber: randomAccountNumber("08", 10), AccountName: "Kawah Putih Official", Active: true, SortOrder: 21},
			{MethodCode: payments.MethodOVO, ProviderName: "OVO", AccountNumber: randomAccountNumber("08", 10), AccountName: "Kawah Putih Official", Active: true, SortOrder: 22},
			{MethodCode: payments.MethodShopeePay, ProviderName: "ShopeePay", AccountNumber: randomAccountNumber("08", 10), AccountName: "Kawah Putih Official", Active: true, SortOrder: 23},
			{MethodCode: payments.MethodAlfamart, ProviderName: "Alfamart", AccountNumber: randomAccountNumber("", 12), AccountName: "Kawah Putih Official", Active: true, SortOrder: 30},
			{MethodCode: payments.MethodIndomaret, ProviderName: "Indomaret", AccountNumber: randomAccountNumber("", 12), AccountName: "Kawah Putih Official", Active: true, SortOrder: 31},
		}
		for i := range defaults {
			if err := db.WithContext(ctx).Create(&defaults[i]).Error; err != nil {
				return err
			}
		}
		return nil
	}

	// Database yang sudah pernah di-seed dengan nomor kosong tetap diisi nomor
	// acak agar uji coba bisa jalan tanpa harus menghapus data. QRIS tidak
	// menyimpan nomor (menampilkan QR dari provider), jadi dilewati.
	var rows []payments.PaymentDestination
	if err := db.WithContext(ctx).
		Where("method_code <> ? AND (account_number = '' OR account_number IS NULL)", payments.MethodQRIS).
		Find(&rows).Error; err != nil {
		return err
	}
	for i := range rows {
		rows[i].AccountNumber = randomAccountNumber("", 12)
		if rows[i].AccountName == "" {
			rows[i].AccountName = "Kawah Putih Official"
		}
		if err := db.WithContext(ctx).Model(&rows[i]).Updates(map[string]interface{}{
			"account_number": rows[i].AccountNumber,
			"account_name":   rows[i].AccountName,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

// randomAccountNumber membuat nomor placeholder acak (untuk uji coba):
// prefix di depan (mis. "8808" untuk VA BCA) + sisanya digit 0-9 acak.
func randomAccountNumber(prefix string, digits int) string {
	b := make([]byte, 0, len(prefix)+digits)
	b = append(b, prefix...)
	for i := 0; i < digits; i++ {
		b = append(b, byte('0'+rand.Intn(10)))
	}
	return string(b)
}

func seedRoles(ctx context.Context, db *gorm.DB) error {
	roles := []role.Role{
		{Name: role.RoleAdmin, Description: "Full management access"},
		{Name: role.RoleManager, Description: "Monitoring, analytics, reports"},
		{Name: role.RolePetugas, Description: "Ticket verification at the gate"},
		{Name: role.RolePengunjung, Description: "Regular visitor account"},
		// ROLE-20: role super_admin wajib ada agar endpoint /super-admin/*
		// (manajemen penetapan role) punya pemilik yang sah.
		{Name: role.RoleSuperAdmin, Description: "Super administrator: manages role assignment"},
	}
	for _, r := range roles {
		var count int64
		if err := db.WithContext(ctx).Model(&role.Role{}).Where("name = ?", r.Name).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := db.WithContext(ctx).Create(&r).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

type staffAccount struct {
	name        string
	email       string
	phone       string
	password    string
	passwordEnv string
	roleName    string
	gate        string
	shift       string
}

// userRow menyisipkan langsung ke users dan mengembalikan id yang dibuat.
type userRow struct {
	ID           string `gorm:"type:char(36);primaryKey"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Name         string
	Email        string
	Phone        string
	PasswordHash string
	IsActive     bool
	IsVerified   bool
	Gate         string
	Shift        string
}

func (userRow) TableName() string { return "users" }

const staffEmailDomain = "kawahputih.id"

func staffAccountsFromEnv() []staffAccount {
	configured := []staffAccount{
		{name: "Administrator Kawah Putih", email: "admin@" + staffEmailDomain, phone: "081200000002", passwordEnv: "SEED_ADMIN_PASSWORD", roleName: role.RoleAdmin},
		{name: "Manajer Kawah Putih", email: "manager@" + staffEmailDomain, phone: "081200000003", passwordEnv: "SEED_MANAGER_PASSWORD", roleName: role.RoleManager},
		{name: "Petugas Kawah Putih", email: "petugas@" + staffEmailDomain, phone: "081200000004", passwordEnv: "SEED_PETUGAS_PASSWORD", roleName: role.RolePetugas, gate: "Gerbang 1", shift: "Pagi (07:00 - 15:00)"},
		// ROLE-20: akun super_admin untuk mengelola penetapan role lewat /super-admin/*.
		{name: "Super Admin Kawah Putih", email: "superadmin@" + staffEmailDomain, phone: "081200000001", passwordEnv: "SEED_SUPERADMIN_PASSWORD", roleName: role.RoleSuperAdmin},
	}
	accounts := make([]staffAccount, 0, len(configured))
	for _, account := range configured {
		account.password = strings.TrimSpace(os.Getenv(account.passwordEnv))
		if account.password != "" {
			accounts = append(accounts, account)
		}
	}
	return accounts
}

func seedUsers(ctx context.Context, db *gorm.DB) error {
	// Bersihkan akun admin legacy dari seed lama yang emailnya salah format
	// ("admin@kawahputih.idkawahputih.id") agar tidak mengganggu pembuatan akun yang benar.
	const brokenEmail = "admin@kawahputih.id" + staffEmailDomain
	if err := db.WithContext(ctx).Exec(
		`DELETE FROM user_roles WHERE user_id IN (SELECT id FROM users WHERE email = ?)`, brokenEmail,
	).Error; err != nil {
		return err
	}
	if err := db.WithContext(ctx).Exec(`DELETE FROM users WHERE email = ?`, brokenEmail).Error; err != nil {
		return err
	}

	for _, acc := range staffAccountsFromEnv() {
		var count int64
		if err := db.WithContext(ctx).Table("users").Where("email = ?", acc.email).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(acc.password), 12)
		if err != nil {
			return err
		}
		row := userRow{
			ID: uuid.NewString(), Name: acc.name, Email: acc.email, Phone: acc.phone,
			PasswordHash: string(hash), IsActive: true, IsVerified: true,
			Gate: acc.gate, Shift: acc.shift,
		}
		if err := db.WithContext(ctx).Create(&row).Error; err != nil {
			return err
		}
		// tautkan role
		var roleRow struct {
			ID string
		}
		if err := db.WithContext(ctx).Table("roles").Select("id").Where("name = ?", acc.roleName).Scan(&roleRow).Error; err != nil {
			return err
		}
		if roleRow.ID == "" {
			return fmt.Errorf("role %s missing", acc.roleName)
		}
		if err := db.WithContext(ctx).Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", row.ID, roleRow.ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedTicketTypes(ctx context.Context, db *gorm.DB) error {
	type t struct {
		Code, Name, Description, Category string
		Price                             int64
		Capacity                          int
	}
	rows := []t{
		// HTM (Masuk)
		{"HTM-WNI-REG", "HTM Wisatawan Nusantara", "HTM wisatawan Nusantara (reguler).", "ENTRANCE", 31000, 1000},
		{"HTM-WNA-REG", "HTM Wisatawan Mancanegara", "HTM wisatawan mancanegara (reguler).", "ENTRANCE", 90000, 500},
		{"HTM-WNI-SUNRISE", "HTM Wisatawan Nusantara (Sunrise)", "HTM wisatawan Nusantara paket sunrise. Termasuk Sunan Ibu & Kawah Putih.", "SUNRISE", 46000, 500},
		{"HTM-WNA-SUNRISE", "HTM Wisatawan Mancanegara (Sunrise)", "HTM wisatawan mancanegara paket sunrise. Termasuk Sunan Ibu & Kawah Putih.", "SUNRISE", 105000, 300},

		// LHKP Charge
		{"LHKP-REG", "Charge LHKP (Reguler)", "Lintas Hutan Kawah Putih (LHKP).", "TRANSPORT", 179000, 200},
		{"LHKP-SUNRISE", "Charge LHKP (Sunrise)", "Lintas Hutan Kawah Putih (LHKP) untuk paket sunrise.", "TRANSPORT", 179000, 200},

		// Ontang Anting (Kuda)
		{"ONTANG-ANTING", "Ontang Anting", "Layanan ontang anting (kuda tunggang).", "ACTIVITY", 32000, 150},
		{"ONTANG-ANTING-SUNRISE", "Ontang Anting (Sunrise)", "Layanan ontang anting (kuda tunggang) untuk paket sunrise.", "ACTIVITY", 38000, 150},

		// Kendaraan
		{"VEHICLE-RODA2", "Kendaraan Roda 2", "Biaya masuk kendaraan roda 2.", "TRANSPORT", 10000, 1000},
		{"VEHICLE-RODA4", "Kendaraan Roda 4", "Biaya masuk kendaraan roda 4.", "TRANSPORT", 18000, 800},
		{"VEHICLE-RODA6", "Kendaraan Roda 6", "Biaya masuk kendaraan roda 6.", "TRANSPORT", 48000, 300},

		// Layanan
		{"HELM-STORAGE", "Jasa Penitipan Helm", "Jasa penitipan helm.", "SERVICE", 3000, 500},
		{"JASA-FOTOGRAFER", "Jasa Fotografer", "Layanan jasa fotografer.", "SERVICE", 11000, 50},

		// Aktivitas
		{"KUDA-TUNGGANG", "Kuda Tunggang", "Aktivitas naik kuda tunggangan.", "ACTIVITY", 55000, 80},
		{"PANAHAN", "Panahan", "Aktivitas panahan.", "ACTIVITY", 22000, 100},
		{"SUNAN-IBU", "Sunan Ibu", "Akses ke area Sunan Ibu.", "ACTIVITY", 15000, 300},
		{"JEMBATAN-APUNG", "Jembatan Apung", "Akses jembatan apung.", "ACTIVITY", 17000, 200},
		{"SKYWALK-CANOPY", "Skywalk Canopy", "Akses Skywalk Canopy.", "ACTIVITY", 17000, 200},

		// Paket/Layanan premium
		{"PREWEDDING", "Prewedding", "Layanan prewedding di Kawah Putih.", "SERVICE", 550000, 20},
		{"SHOOTING", "Shooting", "Layanan shooting/produser.", "SERVICE", 3300000, 10},
	}
	for _, r := range rows {
		// Upsert berdasarkan kode: sisipkan bila belum ada, selain itu segarkan nama/harga agar DB lama ikut tarif baru, lalu aktifkan lagi agar bisa dipesan.
		var existing tickettype.TicketType
		err := db.WithContext(ctx).Where("code = ?", r.Code).First(&existing).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if err == gorm.ErrRecordNotFound {
			tt := tickettype.TicketType{
				Code: r.Code, Name: r.Name, Description: r.Description, Category: r.Category,
				Price: r.Price, Capacity: r.Capacity, IsActive: true,
			}
			if err := db.WithContext(ctx).Create(&tt).Error; err != nil {
				return err
			}
			continue
		}
		existing.Name = r.Name
		existing.Description = r.Description
		existing.Category = r.Category
		existing.Price = r.Price
		existing.Capacity = r.Capacity
		existing.IsActive = true
		if err := db.WithContext(ctx).Save(&existing).Error; err != nil {
			return err
		}
	}
	return nil
}

// (Seeder galeri dihapus agar galeri mulai tanpa data demo.
// Seeder konten lain dipulihkan: artikel, paket wisata, info destinasi, FAQ,
// fasilitas, dan aturan pengunjung.)

func seedArticles(ctx context.Context, db *gorm.DB) error {
	type a struct {
		Title, Slug, Excerpt, Content string
	}
	rows := []a{
		{"Sejarah Kawah Putih", "sejarah-kawah-putih", "Dari kawah Gunung Patuha yang legendaris hingga destinasi wisata ikonik Jawa Barat.", "Kawah Putih ditemukan oleh seorang ahli kimia berkebangsaan Jerman, Dr. Franz Wilhelm Junghuhn, pada tahun 1837."},
		{"Tips Berkunjung ke Kawah Putih", "tips-berkunjung-kawah-putih", "Panduan singkat: jam terbaik, cuaca, dan perlengkapan yang perlu dibawa.", "Datanglah pada pagi hari sekitar pukul 07.00–10.00 agar kabut tipis belum turun dan pemandangan kawah paling jernih."},
		{"Fasilitas di Kawasan Wisata", "fasilitas-kawasan-wisata", "Area parkir, kamar mandi, mushola, dan penginapan di sekitar kawah.", "Kawasan wisata dilengkapi area parkir luas, mushola, kamar mandi, warung makan, dan spotting area foto."},
	}
	for _, r := range rows {
		var count int64
		if err := db.WithContext(ctx).Table("articles").Where("slug = ?", r.Slug).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		art := article.Article{
			Title: r.Title, Slug: r.Slug, Excerpt: r.Excerpt, Content: r.Content,
			Status: article.StatusPublished, AuthorID: "", PublishedAt: new(int64),
		}
		*art.PublishedAt = 0
		if err := db.WithContext(ctx).Create(&art).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedPackages(ctx context.Context, db *gorm.DB) error {
	type p struct {
		Name, Slug, Description string
		PriceCents              int64
		DurationHours           int
		MaxCapacity             int
	}
	rows := []p{
		{"Kawah Putih Standard", "kawah-putih-standard", "Tiket masuk + shuttle PP + welcome drink.", 45000, 3, 500},
		{"Sunrise Tour", "sunrise-tour", "Akses pagi (sebelum jam buka umum) + caldera tour + guide.", 95000, 5, 120},
		{"Family Weekend", "family-weekend", "Paket keluarga akhir pekan: masuk 2 dewasa + 3 anak + boat.", 120000, 6, 80},
	}
	for _, r := range rows {
		var count int64
		if err := db.WithContext(ctx).Table("tourism_packages").Where("slug = ?", r.Slug).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		pkg := tourismpackage.Package{
			Name: r.Name, Slug: r.Slug, Description: r.Description,
			PriceCents: r.PriceCents, Currency: "IDR", DurationHours: r.DurationHours,
			MaxCapacity: r.MaxCapacity, IsActive: true,
		}
		if err := db.WithContext(ctx).Create(&pkg).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedDestinationInfo(ctx context.Context, db *gorm.DB) error {
	var count int64
	if err := db.WithContext(ctx).Table("destination_information").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	item := destinationinfo.DestinationInfo{
		Title:           "Informasi Kawah Putih",
		Description:     "Kawah Putih adalah danau kawah vulkanik di kawasan Gunung Patuha, Ciwidey. Airnya berwarna putih kehijauan yang dapat berubah mengikuti kondisi alam, dikelilingi kabut tipis dan pepohonan.",
		History:         "Kawah Putih ditemukan oleh ahli kimia berkebangsaan Jerman, Dr. Franz Wilhelm Junghuhn, pada tahun 1837 saat meneliti Gunung Patuha.",
		DestinationDesc: "Destinasi alam unggulan Jawa Barat dengan panorama kaldera, jalur pejalan kaki, serta titik foto di bibir kawah.",
		Location:        "Ciwidey, Kabupaten Bandung, Jawa Barat",
		OperatingHours:  "07.00 - 17.00 WIB (setiap hari)",
		Contact:         "0800-1-123456",
		AdditionalInfo:  "Disarankan datang pada pagi hari untuk mendapatkan suasana kabut terbaik. Gunakan pakaian hangat karena suhu cukup dingin.",
		Status:          destinationinfo.StatusPublished,
	}
	return db.WithContext(ctx).Create(&item).Error
}

func seedFAQs(ctx context.Context, db *gorm.DB) error {
	type f struct {
		Question, Answer, Category string
		SortOrder                  int
	}
	rows := []f{
		{"Bagaimana cara membeli tiket?", "Pilih menu \"Pesan Tiket\", tentukan tanggal kunjungan dan jumlah tiket, lalu lanjutkan ke pembayaran. E-tiket dikirim setelah pembayaran dikonfirmasi.", "Tiket", 1},
		{"Jam operasional Kawah Putih?", "Kawasan dibuka setiap hari pukul 07.00 - 17.00 WIB. Disarankan datang pagi hari untuk suasana kabut terbaik.", "Jam Operasional", 2},
		{"Apa saja fasilitas yang tersedia?", "Tersedia area parkir, pusat informasi, musala, toilet, kafetaria, dan jalur pejalan kaki menuju kawah.", "Fasilitas", 3},
		{"Apakah tiket bisa dipesan online?", "Ya. Seluruh tiket dapat dipesan secara online melalui website ini dan dibayar melalui QRIS, transfer bank, atau kartu.", "Pembayaran", 4},
		{"Bagaimana cara melihat e-tiket?", "E-tiket tersedia di halaman E-Tiket setelah pembayaran. Anda dapat mengunduh, mencetak, atau membagikannya.", "Tiket", 5},
	}
	for _, r := range rows {
		var count int64
		if err := db.WithContext(ctx).Table("faqs").Where("question = ?", r.Question).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		item := faq.FAQ{
			Question: r.Question, Answer: r.Answer, Category: r.Category,
			SortOrder: r.SortOrder, IsPublished: true,
		}
		if err := db.WithContext(ctx).Create(&item).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedFacilities(ctx context.Context, db *gorm.DB) error {
	type f struct {
		Name, Description, Icon, Location, OperatingHours string
		SortOrder                                         int
	}
	rows := []f{
		{"Area Parkir", "Area parkir luas untuk mobil pribadi dan bus wisata.", "Car", "Pintu Masuk Utama", "07.00 - 17.00", 1},
		{"Pusat Informasi", "Layanan informasi wisata dan pemandu kawasan.", "Info", "Area Penerimaan", "07.00 - 17.00", 2},
		{"Musala", "Tempat ibadah yang bersih untuk pengunjung.", "Landmark", "Area Fasilitas", "07.00 - 17.00", 3},
		{"Toilet & Kamar Bilas", "Toilet pria dan wanita beserta kamar bilas.", "DoorOpen", "Area Fasilitas", "07.00 - 17.00", 4},
		{"Kafetaria", "Makanan, minuman, dan camilan ringan.", "Coffee", "Area Wahana", "08.00 - 16.00", 5},
	}
	for _, r := range rows {
		var count int64
		if err := db.WithContext(ctx).Table("facilities").Where("name = ?", r.Name).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		item := facility.Facility{
			Name: r.Name, Description: r.Description, Icon: r.Icon, Location: r.Location,
			OperatingHours: r.OperatingHours, Status: facility.StatusActive,
			SortOrder: r.SortOrder, IsPublished: true,
		}
		if err := db.WithContext(ctx).Create(&item).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedVisitorRules(ctx context.Context, db *gorm.DB) error {
	type r struct {
		Title, Description, Category, Priority string
		SortOrder                              int
	}
	rows := []r{
		{"Tetap di Jalur yang Disediakan", "Demi keselamatan, pengunjung wajib berada di jalur pejalan kaki yang telah disediakan.", "Keselamatan", visitorrule.PriorityHigh, 1},
		{"Dilarang Berenang di Area Kawah", "Air kawah mengandung zat berbahaya dan area dilarang untuk aktivitas berenang.", "Larangan", visitorrule.PriorityCritical, 2},
		{"Jaga Kebersihan Kawasan", "Buanglah sampah pada tempatnya dan bawa turun sampah pribadi Anda.", "Kebersihan", visitorrule.PriorityNormal, 3},
		{"Patuhi Arahan Petugas", "Ikuti arahan petugas di lapangan untuk keselamatan dan kenyamanan bersama.", "Informasi Umum", visitorrule.PriorityNormal, 4},
	}
	for _, r := range rows {
		var count int64
		if err := db.WithContext(ctx).Table("visitor_rules").Where("title = ?", r.Title).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		item := visitorrule.VisitorRule{
			Title: r.Title, Description: r.Description, Category: r.Category,
			Priority: r.Priority, SortOrder: r.SortOrder, IsPublished: true,
		}
		if err := db.WithContext(ctx).Create(&item).Error; err != nil {
			return err
		}
	}
	return nil
}

type demoVisitor struct {
	Name, Email, Phone string
}

type demoBookingSpec struct {
	Code        string
	Visitor     int
	TicketCode  string
	Quantity    int
	OffsetDays  int
	VisitTime   string
	CreatedHour int
	CreatedMin  int
	Method      string
	Rating      int
	Comment     string
}

// demoSeedEnabled menyalakan data demo booking/ulasan secara default di luar
// production. Set SEED_DEMO_BOOKINGS=false untuk mematikan, atau true untuk
// tetap membuatnya di production.
func demoSeedEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("SEED_DEMO_BOOKINGS"))) {
	case "true", "1":
		return true
	case "false", "0":
		return false
	}
	return !strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production")
}

// seedDemoBookings mengisi riwayat booking dua minggu terakhir beserta
// pembayaran, e-tiket, dan satu ulasan per booking. Setiap booking memakai
// jenis tiket dan jumlah berbeda sehingga nominalnya bervariasi. Seeder ini
// idempoten: bila sudah ada booking berkode KP-DEMO-*, tidak ada yang dibuat.
func seedDemoBookings(ctx context.Context, db *gorm.DB) error {
	if !demoSeedEnabled() {
		return nil
	}

	var count int64
	if err := db.WithContext(ctx).Model(&booking.Booking{}).
		Where("booking_code LIKE ?", "KP-DEMO-%").
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	visitors := []demoVisitor{
		{"Bagas Pratama", "bagas.pratama@example.com", "081311220011"},
		{"Nabila Rahmawati", "nabila.rahma@example.com", "081322330022"},
		{"Rizky Maulana", "rizky.maulana@example.com", "081333440033"},
		{"Intan Permata Sari", "intan.permata@example.com", "081344550044"},
		{"Fajar Nugroho", "fajar.nugroho@example.com", "081355660055"},
		{"Salsabila Putri", "salsa.putri@example.com", "081366770066"},
		{"Dimas Aryo Wibowo", "dimas.aryo@example.com", "081377880077"},
		{"Melati Kusuma", "melati.kusuma@example.com", "081388990088"},
	}

	specs := []demoBookingSpec{
		{
			Code: "KP-DEMO-01", Visitor: 0, TicketCode: "HTM-WNI-REG", Quantity: 2,
			OffsetDays: 13, VisitTime: "07:20", CreatedHour: 19, CreatedMin: 42, Method: payments.MethodQRIS,
			Rating: 5,
			Comment: "Pemandangannya bagus banget, apalagi pas pagi masih ada kabutnya. Tukang foto di area kawah ramah dan harganya wajar. Cuma jalannya agak licin jadi hati-hati.",
		},
		{
			Code: "KP-DEMO-02", Visitor: 1, TicketCode: "HTM-WNA-REG", Quantity: 1,
			OffsetDays: 12, VisitTime: "08:45", CreatedHour: 20, CreatedMin: 15, Method: payments.MethodBCAVA,
			Rating: 4,
			Comment: "Tiket gampang dibeli online dan masuknya nggak ribet. Pas datang agak ramai jadi parkirnya agak jauh. Overall worth it buat liburan keluarga.",
		},
		{
			Code: "KP-DEMO-03", Visitor: 2, TicketCode: "HTM-WNI-SUNRISE", Quantity: 3,
			OffsetDays: 10, VisitTime: "09:10", CreatedHour: 9, CreatedMin: 5, Method: payments.MethodGoPay,
			Rating: 3,
			Comment: "Tempatnya oke, tapi pas saya datang kabutnya tebal jadi kawahnya nggak begitu kelihatan. Toilet bersih, warungnya juga lumayan lengkap.",
		},
		{
			Code: "KP-DEMO-04", Visitor: 3, TicketCode: "LHKP-REG", Quantity: 1,
			OffsetDays: 8, VisitTime: "07:05", CreatedHour: 16, CreatedMin: 48, Method: payments.MethodDANA,
			Rating: 5,
			Comment: "Anak-anak senang banget bisa naik kuda dan lihat kawah langsung. Petugasnya sabar jawab pertanyaan. Bakal balik lagi kalau ke Bandung.",
		},
		{
			Code: "KP-DEMO-05", Visitor: 4, TicketCode: "ONTANG-ANTING", Quantity: 2,
			OffsetDays: 7, VisitTime: "13:30", CreatedHour: 21, CreatedMin: 10, Method: payments.MethodQRIS,
			Rating: 4,
			Comment: "Suka sama suasananya, dingin dan tenang. Minusnya sinyal susah dan antre shuttle lumayan lama pas sore.",
		},
		{
			Code: "KP-DEMO-06", Visitor: 5, TicketCode: "PREWEDDING", Quantity: 1,
			OffsetDays: 5, VisitTime: "10:15", CreatedHour: 11, CreatedMin: 33, Method: payments.MethodMandiriVA,
			Rating: 2,
			Comment: "Sayang pas kesini hujan, jalanan jadi becek dan agak susah lewat. Mungkin lain kali coba lagi pas cuaca bagus.",
		},
		{
			Code: "KP-DEMO-07", Visitor: 6, TicketCode: "VEHICLE-RODA4", Quantity: 3,
			OffsetDays: 3, VisitTime: "15:00", CreatedHour: 6, CreatedMin: 58, Method: payments.MethodOVO,
			Rating: 4,
			Comment: "Danau kawahnya unik, warnanya benar-benar putih kehijauan. Datang jam 8 masih sepi jadi puas foto-foto.",
		},
		{
			Code: "KP-DEMO-08", Visitor: 7, TicketCode: "KUDA-TUNGGANG", Quantity: 2,
			OffsetDays: 1, VisitTime: "08:30", CreatedHour: 18, CreatedMin: 22, Method: payments.MethodShopeePay,
			Rating: 5,
			Comment: "Pelayanannya cepat, e-tiket langsung bisa dipakai scan. Kopi di kafetaria enak, cocok buat hangat-hangat setelah keliling.",
		},
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		loc := now.Location()

		userIDs := make([]string, len(visitors))
		for i, visitor := range visitors {
			var existing userRow
			err := tx.Where("email = ?", visitor.Email).Take(&existing).Error
			switch {
			case err == nil:
				userIDs[i] = existing.ID
				continue
			case !errors.Is(err, gorm.ErrRecordNotFound):
				return err
			}

			row := userRow{
				ID: uuid.NewString(), CreatedAt: now.AddDate(0, 0, -60), UpdatedAt: now.AddDate(0, 0, -60),
				Name: visitor.Name, Email: visitor.Email, Phone: visitor.Phone,
				IsActive: true, IsVerified: true,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			userIDs[i] = row.ID

			var roleRow struct{ ID string }
			if err := tx.Table("roles").Select("id").Where("name = ?", role.RolePengunjung).Scan(&roleRow).Error; err != nil {
				return err
			}
			if roleRow.ID != "" {
				if err := tx.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", row.ID, roleRow.ID).Error; err != nil {
					return err
				}
			}
		}

		for _, spec := range specs {
			var ticket tickettype.TicketType
			if err := tx.Where("code = ?", spec.TicketCode).First(&ticket).Error; err != nil {
				return fmt.Errorf("ticket type %s: %w", spec.TicketCode, err)
			}

			visitor := visitors[spec.Visitor]
			visitDay := now.AddDate(0, 0, -spec.OffsetDays)
			y, m, d := visitDay.Date()
			visitDate := time.Date(y, m, d, 0, 0, 0, 0, loc)
			visitDateStr := visitDate.Format("2006-01-02")
			createdAt := visitDate.AddDate(0, 0, -2).
				Add(time.Duration(spec.CreatedHour)*time.Hour + time.Duration(spec.CreatedMin)*time.Minute)
			paidAt := createdAt.Add(7 * time.Minute)

			hm, err := time.Parse("15:04", spec.VisitTime)
			if err != nil {
				return fmt.Errorf("visit time %s: %w", spec.VisitTime, err)
			}
			arrival := time.Date(y, m, d, hm.Hour(), hm.Minute(), 0, 0, loc)

			total := ticket.Price * int64(spec.Quantity)
			bookingID := uuid.NewString()
			b := booking.Booking{
				Base:        booking.Base{ID: bookingID, CreatedAt: createdAt, UpdatedAt: createdAt},
				BookingCode: spec.Code, UserID: userIDs[spec.Visitor],
				VisitorName: visitor.Name, Email: visitor.Email, Phone: visitor.Phone,
				VisitDate: visitDateStr,
				Note:      fmt.Sprintf("Rencana datang pukul %s WIB.", spec.VisitTime),
				Status:    booking.StatusUsed,
				Subtotal:  total, Total: total, TotalTickets: spec.Quantity,
				CheckedInAt: &arrival, CheckedInGate: "Gerbang 1",
			}
			if err := tx.Create(&b).Error; err != nil {
				return err
			}
			if err := tx.Create(&booking.BookingItem{
				Base:      booking.Base{ID: uuid.NewString(), CreatedAt: createdAt, UpdatedAt: createdAt},
				BookingID: bookingID, TicketID: ticket.ID, TicketName: ticket.Name,
				Price: ticket.Price, Quantity: spec.Quantity,
			}).Error; err != nil {
				return err
			}

			if err := tx.Create(&payments.Payment{
				Base:      payments.Base{ID: uuid.NewString(), CreatedAt: createdAt, UpdatedAt: createdAt},
				BookingID: bookingID, Provider: payments.ProviderMidtrans, Method: spec.Method,
				Reference: "DEMO-PAY-" + spec.Code, MidtransOrderID: "DEMO-ORDER-" + spec.Code, Amount: total,
				Status: payments.StatusPaid, SettlementStatus: payments.SettlementSettled,
				PaidAt: &paidAt, ExpiredAt: createdAt.Add(payments.PaymentWindow),
			}).Error; err != nil {
				return err
			}

			if err := tx.Create(&ticketing.Ticket{
				Base:        ticketing.Base{ID: uuid.NewString(), CreatedAt: paidAt, UpdatedAt: paidAt},
				BookingID:   bookingID, BookingCode: spec.Code,
				VisitorName: visitor.Name, Email: visitor.Email, VisitDate: visitDateStr,
				TicketName: ticket.Name, Quantity: spec.Quantity, Total: total,
				Status: ticketing.TicketUsed, GeneratedAt: paidAt,
				CheckedInAt: &arrival, CheckedInGate: "Gerbang 1",
				QRData: "DEMO-QR-" + spec.Code,
			}).Error; err != nil {
				return err
			}

			reviewedAt := visitDate.AddDate(0, 0, 1).
				Add(19*time.Hour + time.Duration((spec.OffsetDays*13)%40)*time.Minute)
			reviewBookingID := bookingID
			if err := tx.Create(&review.Review{
				Base:       review.Base{ID: uuid.NewString(), CreatedAt: reviewedAt, UpdatedAt: reviewedAt},
				UserID:     userIDs[spec.Visitor],
				BookingID:  &reviewBookingID,
				Rating:     spec.Rating,
				Comment:    spec.Comment,
				Status:     review.StatusApproved,
				IsApproved: true,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
