package islandwind233

import (
	"testing"
)

func TestAnimeGetImage(t *testing.T) {
	a := &Anime{}
	img, err := a.GetImage("")
	if err != nil {
		t.Fatal(err)
	}
	if len(img.Images) != 1 {
		t.Fatal("failed to get anime image from islandwind233")
	}
}

func TestCosplayGetImage(t *testing.T) {
	c := &Cosplay{}
	img, err := c.GetImage("")
	if err != nil {
		t.Fatal(err)
	}
	if len(img.Images) != 1 {
		t.Fatal("failed to get cosplay image from islandwind233")
	}
}
