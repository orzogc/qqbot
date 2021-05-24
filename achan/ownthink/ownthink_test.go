package ownthink

import "testing"

func TestChat(t *testing.T) {
	req := Request{
		Spoken: "你好",
	}
	reply, err := req.Chat()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", reply)
}
