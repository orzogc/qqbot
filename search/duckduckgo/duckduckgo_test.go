package duckduckgo

import "testing"

func TestDuckDuckGoSearch(t *testing.T) {
	duckduckgo := New()
	result, err := duckduckgo.Search("abc")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) == 0 {
		t.Fatal("failed to get duckduckgo search result")
	}
	t.Logf("%+v", result)
}
