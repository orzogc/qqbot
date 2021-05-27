package search_utils

// 搜索结果
type SearchResult struct {
	URL   string `json:"url"`   // 搜索结果的链接
	Title string `json:"title"` // 搜索结果的标题
}

func ToString(result []SearchResult) (text string) {
	for _, r := range result {
		text += r.URL + "\n" + r.Title + "\n"
	}

	return text
}
