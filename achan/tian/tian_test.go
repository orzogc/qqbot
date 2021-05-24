package tian

import "testing"

func TestChat(t *testing.T) {
	q := Query{
		Key:      "key",
		Question: "你好",
		UniqueID: "abc",
	}
	reply, err := q.Chat()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", reply)
}
