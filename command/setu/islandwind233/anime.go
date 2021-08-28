package islandwind233

import (
	"github.com/orzogc/qqbot/command/setu/setu_utils"
	"github.com/orzogc/qqbot/qqbot_utils"
)

const (
	AnimeURL = "https://iw233.cn/API/ghs.php" // 二次元图片接口
	AnimeID  = "islandwind233_anime"          // 二次元图片ID
)

// 二次元图片
type Anime struct{}

// 获取图片
func (a *Anime) GetAnimeImage() ([]byte, error) {
	body, err := qqbot_utils.Get(AnimeURL, nil)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// 获取图片，实现Setu接口，keyword没有用
func (a *Anime) GetImage(keyword string) (*setu_utils.Image, error) {
	img, err := a.GetAnimeImage()
	if err != nil {
		return nil, err
	}

	return &setu_utils.Image{Images: [][]byte{img}}, nil
}
