package islandwind233

import "github.com/orzogc/qqbot/qqbot_utils"

const (
	CosplayURL = "https://islandwind233.pro/ZY/API/Cos/GetCos.php"
	CosplayID  = "islandwind233_cosplay"
)

type Cosplay struct{}

func (c *Cosplay) GetCosplayImage() ([]byte, error) {
	body, err := qqbot_utils.Get(CosplayURL, nil)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *Cosplay) GetImage(keyword string) ([][]byte, error) {
	img, err := c.GetCosplayImage()
	if err != nil {
		return nil, err
	}

	return [][]byte{img}, nil
}
