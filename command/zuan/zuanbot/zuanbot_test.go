package zuanbot

import "testing"

func TestZuanbotZuan(t *testing.T) {
	zuanbot := &Zuanbot{}
	text, err := zuanbot.GetText()
	if err != nil {
		t.Fatal(err)
	}
	if text == "" {
		t.Fatal("failed to get zuanbot response")
	}
	t.Logf("%s", text)
}
