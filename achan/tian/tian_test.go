package tian

import "testing"

func TestChat(t *testing.T) {
	q := Tian{
		Key: "key",
	}
	reply, err := q.Chat("你好", "abc")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", reply)
}
