package zuanbot

import "github.com/orzogc/qqbot/qqbot_utils"

const (
	ZuanbotID  = "zuanbot" // ID
	ZuanbotURL = "https://zuanbot.com/api.php?level=min&lang=zh_cn"
)

// zuanbot
type Zuanbot struct{}

// 获取祖安语句，实现Zuan接口
func (z *Zuanbot) GetText() (string, error) {
	resp, err := qqbot_utils.Get(ZuanbotURL, nil)
	if err != nil {
		return "", nil
	}

	return string(resp), nil
}
