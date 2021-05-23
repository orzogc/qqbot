package pixiv

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/NateScarlet/pixiv/pkg/artwork"
	"github.com/NateScarlet/pixiv/pkg/client"
	"github.com/orzogc/qqbot/setu/setu_utils"
)

const ID = "pixiv"

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
	Tags         string
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
		ctx: login(PHPSESSID),
	}
}

func (p *Pixiv) GetImage() ([][]byte, error) {
	if p.SearchOption.Order != artwork.OrderDateDSC && p.SearchOption.Order != artwork.OrderDateASC {
		return nil, fmt.Errorf("Order必须为空或date，，现为%s", p.SearchOption.Order)
	}
	if p.SearchOption.ContentRating != artwork.ContentRatingAll && p.SearchOption.ContentRating != artwork.ContentRatingSafe && p.SearchOption.ContentRating != artwork.ContentRatingR18 {
		return nil, fmt.Errorf("ContentRating必须为空、safe或r18，，现为%s", p.SearchOption.ContentRating)
	}
	if p.SearchOption.Mode != artwork.SearchModeTag && p.SearchOption.Mode != artwork.SearchModePartialTag && p.SearchOption.Mode != artwork.SearchModeTitleOrCaption {
		return nil, fmt.Errorf("Mode必须为空、s_tag或s_tc，，现为%s", p.SearchOption.Mode)
	}

	result, err := artwork.Search(p.ctx, p.Tags, func(op *artwork.SearchOptions) {
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
	art := artworks[rand.Intn(len(artworks))]
	art.FetchPages(p.ctx)

	images := make([][]byte, 0, len(art.Pages))
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, a := range art.Pages {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			req, err := http.NewRequest(http.MethodGet, s, nil)
			if err != nil {
				return
			}
			req.Header.Set("Referer", setu_utils.PixivURL)
			resp, err := setu_utils.Client.Do(req)
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
			images = append(images, body)
		}(a.Image.Original)
	}
	wg.Wait()

	if len(images) == 0 {
		return nil, fmt.Errorf("获取pixiv图片失败")
	}

	return images, nil
}
