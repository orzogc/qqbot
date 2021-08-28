package islandwind233

import (
	"github.com/orzogc/qqbot/command/setu/setu_utils"
	"github.com/orzogc/qqbot/qqbot_utils"
)

const (
	CosplayURL = "https://iw233.cn/API/cos.php" // cosplay图片接口
	CosplayID  = "islandwind233_cosplay"        // cosplay图片ID
)

// cosplay图片
type Cosplay struct{}

// 获取图片
func (c *Cosplay) GetCosplayImage() ([]byte, error) {
	body, err := qqbot_utils.Get(CosplayURL, nil)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// 获取图片，实现Setu接口，keyword没有用
func (c *Cosplay) GetImage(keyword string) (*setu_utils.Image, error) {
	img, err := c.GetCosplayImage()
	if err != nil {
		return nil, err
	}

	return &setu_utils.Image{Images: [][]byte{img}}, nil
}
