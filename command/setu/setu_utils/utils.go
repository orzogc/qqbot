package setu_utils

const (
	PixivURL         = "https://www.pixiv.net/"          // pixiv链接
	PixivArtworksURL = "https://www.pixiv.net/artworks/" // pixiv artworks链接
)

// 图片
type Image struct {
	Text   string   // 图片描述或链接之类的文字
	Images [][]byte // 多张图片
}
