package islandwind233

import (
	"github.com/orzogc/qqbot/setu/setu_utils"
)

const (
	CosplayURL = "https://islandwind233.pro/ZY/API/Cos/GetCos.php"
	CosplayID  = "islandwind233_cosplay"
)

type Cosplay struct{}

func (c *Cosplay) GetImage() ([][]byte, error) {
	body, err := setu_utils.Get(CosplayURL, nil)
	if err != nil {
		return nil, err
	}

	return [][]byte{body}, nil
}
