package islandwind233

import (
	"io"
	"net/http"

	"github.com/orzogc/qqbot/setu/setu_utils"
)

const (
	CosplayURL = "https://islandwind233.pro/ZY/API/Cos/GetCos.php"
	CosplayID  = "islandwind233_cosplay"
)

type Cosplay struct{}

func (c *Cosplay) GetImage() ([][]byte, error) {
	req, err := http.NewRequest(http.MethodGet, CosplayURL, nil)
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
