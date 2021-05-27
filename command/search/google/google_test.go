package google

import "testing"

func TestGoogleSearch(t *testing.T) {
	google := &Google{}
	result, err := google.Search("abc")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) == 0 {
		t.Fatal("failed to get google search result")
	}
	t.Logf("%+v", result)
}
