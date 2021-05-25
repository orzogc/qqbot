package ownthink

import "testing"

func TestChat(t *testing.T) {
	req := Request{}
	reply, err := req.Chat("你好", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", reply)
}
