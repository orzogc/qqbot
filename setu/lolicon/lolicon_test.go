package lolicon

import (
	"testing"
)

func TestLolicon(t *testing.T) {
	q := &Query{}
	resp, err := q.Lolicon()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", resp)
	if len(resp.Data) != 1 {
		t.Fatal("failed to get image from lolicon")
	}
}

func TestGetImage(t *testing.T) {
	q := &Query{
		Proxy: "disable",
	}
	img, err := q.GetImage()
	if err != nil {
		t.Fatal(err)
	}
	if len(img) != 1 {
		t.Fatal("failed to get image from lolicon")
	}
}
