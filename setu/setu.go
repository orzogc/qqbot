package setu

import (
	"github.com/orzogc/qqbot/setu/islandwind233"
	"github.com/orzogc/qqbot/setu/lolicon"
	"github.com/orzogc/qqbot/setu/paulzzh"
	"github.com/orzogc/qqbot/setu/pixiv"
)

var (
	_ Setu = (*lolicon.Lolicon)(nil)
	_ Setu = (*islandwind233.Anime)(nil)
	_ Setu = (*islandwind233.Cosplay)(nil)
	_ Setu = (*paulzzh.Paulzzh)(nil)
	_ Setu = (*pixiv.Pixiv)(nil)
)

type Setu interface {
	GetImage(keyword string) ([][]byte, error)
}
