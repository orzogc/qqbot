package ownthink

import "testing"

func TestOwnthinkChat(t *testing.T) {
	o := &Ownthink{}
	reply, err := o.ChatWith("你好", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", reply)
}
