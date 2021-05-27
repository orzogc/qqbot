package pixiv

import (
	"testing"
)

func TestPixivGetImage(t *testing.T) {
	p := New("")
	img, err := p.GetImage("acfun")
	if err != nil {
		t.Fatal(err)
	}
	if len(img) == 0 {
		t.Fatal("failed to get image from pixiv")
	}
}
