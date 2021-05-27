package paulzzh

import (
	"testing"
)

func TestPaulzzhGetImage302(t *testing.T) {
	p := Paulzzh{
		Size:  "wap",
		Proxy: 1,
	}
	img, err := p.GetImage("")
	if err != nil {
		t.Fatal(err)
	}
	if len(img) != 1 {
		t.Fatal("failed to get image from paulzzh")
	}
}

func TestPaulzzhGetImageJSON(t *testing.T) {
	p := Paulzzh{
		Type:  "json",
		Site:  "yandere",
		Proxy: 1,
	}
	img, err := p.GetImage("")
	if err != nil {
		t.Fatal(err)
	}
	if len(img) != 1 {
		t.Fatal("failed to get image from paulzzh")
	}
}
