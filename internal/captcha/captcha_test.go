package captcha

import (
	"image/png"
	"net/http/httptest"
	"testing"
)

func TestIssueReturnsPNGAndCookie(t *testing.T) {
	store := NewStore()
	request := httptest.NewRequest("GET", "/admin/captcha", nil)
	response := httptest.NewRecorder()
	if err := store.Issue(response, request); err != nil {
		t.Fatal(err)
	}
	if response.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("content type = %q", response.Header().Get("Content-Type"))
	}
	decoded, err := png.Decode(response.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	inkPixels := 0
	for y := 0; y < decoded.Bounds().Dy(); y++ {
		for x := 0; x < decoded.Bounds().Dx(); x++ {
			r, g, b, _ := decoded.At(x, y).RGBA()
			if r != 0xffff || g != 0xffff || b != 0xffff {
				inkPixels++
			}
		}
	}
	if inkPixels < 100 {
		t.Fatalf("captcha has too few visible pixels: %d", inkPixels)
	}
	if response.Header().Get("Set-Cookie") == "" {
		t.Fatal("captcha cookie missing")
	}
}
