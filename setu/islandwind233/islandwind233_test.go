package islandwind233

import (
	"testing"
)

func TestAnime(t *testing.T) {
	a := &Anime{}
	img, err := a.GetImage("")
	if err != nil {
		t.Fatal(err)
	}
	if len(img) != 1 {
		t.Fatal("failed to get anime image from islandwind233")
	}
}

func TestCosplay(t *testing.T) {
	c := &Cosplay{}
	img, err := c.GetImage("")
	if err != nil {
		t.Fatal(err)
	}
	if len(img) != 1 {
		t.Fatal("failed to get cosplay image from islandwind233")
	}
}
