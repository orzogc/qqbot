package pixiv

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/NateScarlet/pixiv/pkg/artwork"
	"github.com/NateScarlet/pixiv/pkg/client"
	"github.com/orzogc/qqbot/qqbot_utils"
	"github.com/orzogc/qqbot/setu/setu_utils"
)

const ID = "pixiv"

var (
	ErrorNoTag        = errors.New("pixiv图片搜索没有关键字")
	ErrorSearchFailed = errors.New("没找到关键字对应的pixiv图片")
)

type SearchOption struct {
	Page              uint                  `json:"page"`
	Order             artwork.Order         `json:"order"`
	ContentRating     artwork.ContentRating `json:"contentRating"`
	Mode              artwork.SearchMode    `json:"mode"`
	WidthLessThan     uint64                `json:"widthLessThan"`
	WidthGreaterThan  uint64                `json:"widthGreaterThan"`
	HeightLessThan    uint64                `json:"heightLessThan"`
	HeightGreaterThan uint64                `json:"heightGreaterThan"`
}

type Pixiv struct {
	ctx          context.Context
	SearchOption *SearchOption
}

func login(PHPSESSID string) context.Context {
	c := &client.Client{}
	c.SetDefaultHeader("User-Agent", client.DefaultUserAgent)
	if PHPSESSID != "" {
		c.SetPHPSESSID(PHPSESSID)
	}
	ctx := context.Background()

	return client.With(ctx, c)
}

func New(PHPSESSID string) *Pixiv {
	return &Pixiv{
		ctx:          login(PHPSESSID),
		SearchOption: new(SearchOption),
	}
}

func (p *Pixiv) GetImage(keyword string) ([][]byte, error) {
	if keyword == "" {
		return nil, ErrorNoTag
	}

	if p.SearchOption.Order != artwork.OrderDateDSC && p.SearchOption.Order != artwork.OrderDateASC {
		return nil, fmt.Errorf("Order必须为空或date，，现为%s", p.SearchOption.Order)
	}
	if p.SearchOption.ContentRating != artwork.ContentRatingAll && p.SearchOption.ContentRating != artwork.ContentRatingSafe && p.SearchOption.ContentRating != artwork.ContentRatingR18 {
		return nil, fmt.Errorf("ContentRating必须为空、safe或r18，，现为%s", p.SearchOption.ContentRating)
	}
	if p.SearchOption.Mode != artwork.SearchModeTag && p.SearchOption.Mode != artwork.SearchModePartialTag && p.SearchOption.Mode != artwork.SearchModeTitleOrCaption {
		return nil, fmt.Errorf("Mode必须为空、s_tag或s_tc，，现为%s", p.SearchOption.Mode)
	}

	result, err := artwork.Search(p.ctx, keyword, func(op *artwork.SearchOptions) {
		op.Page = int(p.SearchOption.Page)
		op.Order = p.SearchOption.Order
		op.ContentRating = p.SearchOption.ContentRating
		op.Mode = p.SearchOption.Mode
		op.WidthLessThan = int64(p.SearchOption.WidthLessThan)
		op.WidthGreaterThan = int64(p.SearchOption.WidthGreaterThan)
		op.HeightLessThan = int64(p.SearchOption.HeightLessThan)
		op.HeightGreaterThan = int64(p.SearchOption.HeightGreaterThan)
	})
	if err != nil {
		return nil, err
	}
	artworks := result.Artworks()
	rand.Seed(time.Now().UnixNano())
	if len(artworks) == 0 {
		return nil, fmt.Errorf("%w：%s", ErrorSearchFailed, keyword)
	}
	art := artworks[rand.Intn(len(artworks))]
	art.FetchPages(p.ctx)

	images := make([][]byte, len(art.Pages))
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i, a := range art.Pages {
		wg.Add(1)
		go func(i int, s string) {
			defer wg.Done()
			req, err := http.NewRequest(http.MethodGet, s, nil)
			if err != nil {
				return
			}
			req.Header.Set("Referer", setu_utils.PixivURL)
			resp, err := qqbot_utils.Client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return
			}
			mu.Lock()
			defer mu.Unlock()
			images[i] = body
		}(i, a.Image.Original)
	}
	wg.Wait()

	img := make([][]byte, 0, len(images))
	for _, i := range images {
		if len(i) != 0 {
			img = append(img, i)
		}
	}
	if len(img) == 0 {
		return nil, fmt.Errorf("获取pixiv图片失败")
	}

	return img, nil
}
