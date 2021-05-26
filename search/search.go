package search

import (
	"github.com/orzogc/qqbot/search/google"
	"github.com/orzogc/qqbot/search/search_utils"
)

var _ Search = (*google.Google)(nil)

// 搜索接口
type Search interface {
	// 搜索text
	Search(text string) ([]search_utils.SearchResult, error)
}
