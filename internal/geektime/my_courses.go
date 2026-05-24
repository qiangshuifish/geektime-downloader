package geektime

import (
	"github.com/go-resty/resty/v2"
	"github.com/nicoxiang/geektime-downloader/internal/geektime/response"
)

// MyProduct represents a product in the user's subscription list
type MyProduct struct {
	ID       int
	Type     string
	Title    string
	Subtitle string
	Unit     string
	IsFinish bool
	IsVideo  bool
	Author   string
}

// MyProducts fetches all products in the user's subscription list using cursor-based pagination
func (c *Client) MyProducts() ([]MyProduct, error) {
	var allProducts []MyProduct
	prev := 0

	for {
		var res response.V4PvipProductListResponse
		r := c.newRequest(
			resty.MethodPost,
			DefaultBaseURL,
			V4PvipProductListPath,
			nil,
			map[string]interface{}{
				"tag_ids":       []int{},
				"product_type":  0,
				"product_form":  0,
				"pvip":          0,
				"prev":          prev,
				"size":          20,
				"sort":          8,
				"with_articles": true,
			},
			&res,
		)
		if _, err := do(r); err != nil {
			return nil, err
		}

		// Collect products from the response
		for _, p := range res.Data.Products {
			allProducts = append(allProducts, MyProduct{
				ID:       p.ID,
				Type:     p.Type,
				Title:    p.Title,
				Subtitle: p.Subtitle,
				Unit:     p.Unit,
				IsFinish: p.IsFinish,
				IsVideo:  p.IsVideo,
				Author:   p.Author.Name,
			})
		}

		if !res.Data.Page.More {
			break
		}

		// Use the score of the last item as cursor for next page
		if len(res.Data.List) > 0 {
			prev = res.Data.List[len(res.Data.List)-1].Score
		} else {
			break
		}
	}

	return allProducts, nil
}
