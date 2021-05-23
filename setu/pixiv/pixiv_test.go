package pixiv

import (
	"testing"
)

func TestPixiv(t *testing.T) {
	p := New("")
	p.Tags = "acfun"
	img, err := p.GetImage()
	if err != nil {
		t.Fatal(err)
	}
	if len(img) == 0 {
		t.Fatal("failed to get image from lolicon")
	}
}
