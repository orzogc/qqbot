package setu

import (
	"github.com/orzogc/qqbot/setu/islandwind233"
	"github.com/orzogc/qqbot/setu/lolicon"
)

var (
	_ Setu = (*lolicon.Query)(nil)
	_ Setu = (*lolicon.Response)(nil)
	_ Setu = (*islandwind233.Anime)(nil)
	_ Setu = (*islandwind233.Cosplay)(nil)
)

type Setu interface {
	GetImage() ([][]byte, error)
}
