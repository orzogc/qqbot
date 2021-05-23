package setu

import (
	"github.com/orzogc/qqbot/setu/islandwind233"
	"github.com/orzogc/qqbot/setu/lolicon"
	"github.com/orzogc/qqbot/setu/paulzzh"
	"github.com/orzogc/qqbot/setu/pixiv"
)

var (
	_ Setu = (*lolicon.Query)(nil)
	_ Setu = (*lolicon.Response)(nil)
	_ Setu = (*islandwind233.Anime)(nil)
	_ Setu = (*islandwind233.Cosplay)(nil)
	_ Setu = (*paulzzh.Query)(nil)
	_ Setu = (*paulzzh.Response)(nil)
	_ Setu = (*pixiv.Pixiv)(nil)
)

type Setu interface {
	GetImage() ([][]byte, error)
}
