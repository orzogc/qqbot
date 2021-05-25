package ownthink

import "testing"

func TestChat(t *testing.T) {
	o := Ownthink{}
	reply, err := o.Chat("你好", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", reply)
}
