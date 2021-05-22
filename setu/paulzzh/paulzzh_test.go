package paulzzh

import (
	"testing"
)

func TestGetImage302(t *testing.T) {
	query := Query{
		Size:  "wap",
		Proxy: 1,
	}
	img, err := query.GetImage()
	if err != nil {
		t.Fatal(err)
	}
	if len(img) != 1 {
		t.Fatal("failed to get image from paulzzh")
	}
}

func TestGetImageJSON(t *testing.T) {
	query := Query{
		Type:  "json",
		Site:  "yandere",
		Proxy: 1,
	}
	img, err := query.GetImage()
	if err != nil {
		t.Fatal(err)
	}
	if len(img) != 1 {
		t.Fatal("failed to get image from paulzzh")
	}
}
