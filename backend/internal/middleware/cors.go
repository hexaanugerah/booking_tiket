package middleware

import (
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// devOrigin cocok dengan origin pengembangan: localhost, loopback, dan alamat
// privat LAN (192.168.x.x, 10.x.x.x, 172.16-31.x.x) dengan port apa pun. Dipakai
// hanya saat allowDevOrigins aktif (APP_ENV bukan production) supaya aplikasi
// tetap bisa dibuka dari HP atau PC lain di jaringan yang sama.
var devOrigin = regexp.MustCompile(
	`^https?://(localhost|127\.0\.0\.1|\[::1\]|10(?:\.\d{1,3}){3}|192\.168(?:\.\d{1,3}){2}|172\.(?:1[6-9]|2\d|3[01])(?:\.\d{1,3}){2})(:\d+)?$`,
)

// RequestID menandai setiap request dengan ID unik, diteruskan di header respons dan tersedia bagi handler/logger untuk melacak satu request.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		c.Set("request_id", id)
		c.Writer.Header().Set("X-Request-ID", id)
		c.Next()
	}
}

// CORS mengizinkan origin frontend yang dikonfigurasi memanggil API dengan kredensial.
// FRONTEND_URL boleh memuat BEBERAPA origin sekaligus (dipisah koma), contoh:
//   FRONTEND_URL=http://localhost:3000,http://192.168.1.5:3000
// supaya aplikasi tetap bisa dibuka dari HP di jaringan yang sama. Origin yang
// cocok di-echo balik ke header Access-Control-Allow-Origin; origin asing tidak
// pernah diberi header izin (tetap tidak pernah tanda bintang bila ada kredensial).
//
// allowDevOrigins dipakai di luar production: `next dev` otomatis pindah ke port
// 3001/3002 kalau 3000 terpakai, dan saat halaman dibuka dari HP origin-nya
// berupa IP LAN (mis. http://192.168.1.5:3000). Tanpa ini request dari perangkat
// lain diblokir browser (muncul sebagai "Tidak dapat terhubung ke server").
func CORS(allowedOrigins []string, allowDevOrigins bool) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		if o = strings.TrimSpace(o); o != "" {
			allowed[o] = struct{}{}
		}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		_, isAllowed := allowed[origin]
		if !isAllowed && allowDevOrigins && devOrigin.MatchString(origin) {
			isAllowed = true
		}
		if isAllowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Add("Vary", "Origin")
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
