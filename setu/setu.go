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

// 图片接口
type Setu interface {
	// 获取图片，keyword为搜索关键字，可以不支持搜索，可返回多个图片
	GetImage(keyword string) ([][]byte, error)
}
