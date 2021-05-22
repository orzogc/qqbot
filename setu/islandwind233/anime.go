package islandwind233

import (
	"github.com/orzogc/qqbot/setu/setu_utils"
)

const (
	AnimeURL = "https://islandwind233.pro/ZY/API/GetImgApi.php"
	AnimeID  = "islandwind233_anime"
)

type Anime struct{}

func (a *Anime) GetImage() ([][]byte, error) {
	body, err := setu_utils.Get(AnimeURL, nil)
	if err != nil {
		return nil, err
	}

	return [][]byte{body}, nil
}
