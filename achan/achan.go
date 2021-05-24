package achan

import (
	"github.com/orzogc/qqbot/achan/tian"
	"github.com/orzogc/qqbot/achan/turing"
)

var (
	_ AI = (*tian.Query)(nil)
	_ AI = (*turing.Request)(nil)
)

type AI interface {
	Chat() (string, error)
}
