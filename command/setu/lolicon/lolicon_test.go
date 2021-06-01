package lolicon

import (
	"testing"
)

func TestLolicon(t *testing.T) {
	l := &Lolicon{}
	resp, err := l.Lolicon()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", resp)
	if len(resp.Data) != 1 {
		t.Fatal("failed to get image from lolicon")
	}
}

func TestLoliconGetImage(t *testing.T) {
	l := &Lolicon{
		Proxy: "disable",
	}
	img, err := l.GetImage("")
	if err != nil {
		t.Fatal(err)
	}
	if len(img.Images) != 1 {
		t.Fatal("failed to get image from lolicon")
	}
}
