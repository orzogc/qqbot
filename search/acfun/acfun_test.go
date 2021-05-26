package acfun

import "testing"

func TestAcFunVideoSearch(t *testing.T) {
	acfun := &AcFunVideo{}
	result, err := acfun.Search("ac娘")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) == 0 {
		t.Fatal("failed to get acfun complex search result")
	}
	t.Logf("%+v", result)
}

func TestAcFunArticleSearch(t *testing.T) {
	acfun := &AcFunArticle{}
	result, err := acfun.Search("ac娘")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) == 0 {
		t.Fatal("failed to get acfun article search result")
	}
	t.Logf("%+v", result)
}
