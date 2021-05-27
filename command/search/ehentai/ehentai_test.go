package ehentai

import "testing"

func TestEHentaiSearch(t *testing.T) {
	ehentai := &EHentai{}
	result, err := ehentai.Search("female")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) == 0 {
		t.Fatal("failed to get e-hentai search result")
	}
	t.Logf("%+v", result)
}
