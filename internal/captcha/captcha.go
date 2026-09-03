package captcha

import (
	"crypto/rand"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/big"
	"net/http"
	"sync"
	"time"
)

type Store struct {
	mu     sync.Mutex
	values map[string]entry
}
type entry struct {
	code    string
	expires time.Time
}

func NewStore() *Store { return &Store{values: make(map[string]entry)} }
func (s *Store) Issue(w http.ResponseWriter, r *http.Request) error {
	code, err := randomCode(5)
	if err != nil {
		return err
	}
	cookie, err := r.Cookie("qesyweb_captcha")
	if err != nil || cookie.Value == "" {
		id, tokenErr := randomCode(32)
		if tokenErr != nil {
			return tokenErr
		}
		cookie = &http.Cookie{Name: "qesyweb_captcha", Value: id}
	}
	http.SetCookie(w, &http.Cookie{Name: "qesyweb_captcha", Value: cookie.Value, Path: "/admin", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: 300})
	s.mu.Lock()
	s.values[cookie.Value] = entry{code: code, expires: time.Now().Add(5 * time.Minute)}
	s.mu.Unlock()
	picture := image.NewRGBA(image.Rect(0, 0, 160, 48))
	draw.Draw(picture, picture.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	for x := 0; x < 160; x++ {
		for y := 0; y < 48; y++ {
			if (x*y)%29 == 0 {
				picture.Set(x, y, color.RGBA{R: 80, G: 120, B: 180, A: 255})
			}
		}
	}
	drawCode(picture, code)
	w.Header().Set("Content-Type", "image/png")
	return png.Encode(w, picture)
}
func (s *Store) Verify(id, code string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.values[id]
	if ok {
		delete(s.values, id)
	}
	return ok && time.Now().Before(value.expires) && value.code == code
}
func randomCode(n int) (string, error) {
	letters := "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	result := make([]byte, n)
	for i := range result {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		result[i] = letters[index.Int64()]
	}
	return string(result), nil
}

var glyphs = map[byte][]string{
	'2': {"11110", "00001", "00010", "00100", "01000", "10000", "11111"},
	'3': {"11110", "00001", "00110", "00001", "00001", "00001", "11110"},
	'4': {"00010", "00110", "01010", "10010", "11111", "00010", "00010"},
	'5': {"11111", "10000", "11110", "00001", "00001", "00001", "11110"},
	'6': {"00110", "01000", "10000", "11110", "10001", "10001", "01110"},
	'7': {"11111", "00001", "00010", "00100", "01000", "01000", "01000"},
	'8': {"01110", "10001", "10001", "01110", "10001", "10001", "01110"},
	'9': {"01110", "10001", "10001", "01111", "00001", "00010", "11100"},
	'A': {"01110", "10001", "10001", "11111", "10001", "10001", "10001"},
	'B': {"11110", "10001", "10001", "11110", "10001", "10001", "11110"},
	'C': {"01111", "10000", "10000", "10000", "10000", "10000", "01111"},
	'D': {"11110", "10001", "10001", "10001", "10001", "10001", "11110"},
	'E': {"11111", "10000", "10000", "11110", "10000", "10000", "11111"},
	'F': {"11111", "10000", "10000", "11110", "10000", "10000", "10000"},
	'G': {"01110", "10001", "10000", "10111", "10001", "10001", "01110"},
	'H': {"10001", "10001", "10001", "11111", "10001", "10001", "10001"},
	'J': {"00111", "00010", "00010", "00010", "10010", "10010", "01100"},
	'K': {"10001", "10010", "10100", "11000", "10100", "10010", "10001"},
	'L': {"10000", "10000", "10000", "10000", "10000", "10000", "11111"},
	'M': {"10001", "11011", "10101", "10101", "10001", "10001", "10001"},
	'N': {"10001", "11001", "10101", "10011", "10001", "10001", "10001"},
	'P': {"11110", "10001", "10001", "11110", "10000", "10000", "10000"},
	'Q': {"01110", "10001", "10001", "10001", "10101", "10010", "01101"},
	'R': {"11110", "10001", "10001", "11110", "10100", "10010", "10001"},
	'S': {"01111", "10000", "10000", "01110", "00001", "00001", "11110"},
	'T': {"11111", "00100", "00100", "00100", "00100", "00100", "00100"},
	'U': {"10001", "10001", "10001", "10001", "10001", "10001", "01110"},
	'V': {"10001", "10001", "10001", "10001", "10001", "01010", "00100"},
	'W': {"10001", "10001", "10001", "10101", "10101", "11011", "10001"},
	'X': {"10001", "10001", "01010", "00100", "01010", "10001", "10001"},
	'Y': {"10001", "10001", "01010", "00100", "00100", "00100", "00100"},
	'Z': {"11111", "00001", "00010", "00100", "01000", "10000", "11111"},
}

func drawCode(dst *image.RGBA, code string) {
	const scale, glyphWidth, gap = 4, 5, 3
	startX := (160 - len(code)*(glyphWidth*scale+gap)) / 2
	for index := range []byte(code) {
		glyph := glyphs[code[index]]
		if glyph == nil {
			continue
		}
		x0 := startX + index*(glyphWidth*scale+gap)
		for y, row := range glyph {
			for x, pixel := range row {
				if pixel != '1' {
					continue
				}
				for dy := 0; dy < scale; dy++ {
					for dx := 0; dx < scale; dx++ {
						dst.Set(x0+x*scale+dx, 10+y*scale+dy, color.RGBA{R: 30, G: 70, B: 120, A: 255})
					}
				}
			}
		}
	}
}
