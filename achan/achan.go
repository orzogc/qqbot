package achan

import (
	"github.com/orzogc/qqbot/achan/ownthink"
	"github.com/orzogc/qqbot/achan/tian"
	"github.com/orzogc/qqbot/achan/turing"
)

var (
	_ AI = (*tian.Query)(nil)
	_ AI = (*turing.Request)(nil)
	_ AI = (*ownthink.Request)(nil)
)

type AI interface {
	Chat(text string, id string) (string, error)
}
