package islandwind233

import (
	"io"
	"net/http"

	"github.com/orzogc/qqbot/setu/setu_utils"
)

const (
	AnimeURL = "https://islandwind233.pro/ZY/API/GetImgApi.php"
	AnimeID  = "islandwind233_anime"
)

type Anime struct{}

func (a *Anime) GetImage() ([][]byte, error) {
	req, err := http.NewRequest(http.MethodGet, AnimeURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := setu_utils.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return [][]byte{body}, nil
}
