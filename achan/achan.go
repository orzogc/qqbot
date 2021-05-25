package achan

import (
	"github.com/orzogc/qqbot/achan/ownthink"
	"github.com/orzogc/qqbot/achan/tian"
	"github.com/orzogc/qqbot/achan/turing"
)

var (
	_ AI = (*tian.Tian)(nil)
	_ AI = (*turing.Turing)(nil)
	_ AI = (*ownthink.Ownthink)(nil)
)

type AI interface {
	Chat(text string, id string) (string, error)
}
