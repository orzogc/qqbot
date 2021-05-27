package moment_utils

// 动态
type Moment struct {
	URL    string   `json:"url"`    // 动态链接
	Author string   `json:"author"` // 动态作者
	Text   string   `json:"text"`   // 动态文本
	Images [][]byte `json:"images"` // 动态图片
}

func (m *Moment) ToString() string {
	return m.URL + "\n" + m.Author + "：" + m.Text + "\n"
}
