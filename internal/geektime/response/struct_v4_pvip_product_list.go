package response

// V4PvipProductListResponse ...
type V4PvipProductListResponse struct {
	Code int `json:"code"`
	Data struct {
		Page struct {
			More  bool `json:"more"`
			Total int  `json:"total"`
		} `json:"page"`
		List []struct {
			ProductID   int    `json:"product_id"`
			ProductType string `json:"product_type"`
			Score       int    `json:"score"`
		} `json:"list"`
		Products []V4PvipProductItem `json:"products"`
	} `json:"data"`
}

// V4PvipProductItem ...
type V4PvipProductItem struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Unit     string `json:"unit"`
	IsFinish bool   `json:"is_finish"`
	IsVideo  bool   `json:"is_video"`
	IsColumn bool   `json:"is_column"`
	Author   struct {
		Name string `json:"name"`
	} `json:"author"`
}

// V3ProductInfosResponse ...
type V3ProductInfosResponse struct {
	Code int `json:"code"`
	Data struct {
		Infos []struct {
			ID       int    `json:"id"`
			Type     string `json:"type"`
			Title    string `json:"title"`
			Subtitle string `json:"subtitle"`
			Unit     string `json:"unit"`
			IsFinish bool   `json:"is_finish"`
			IsVideo  bool   `json:"is_video"`
			IsColumn bool   `json:"is_column"`
			Author   struct {
				Name string `json:"name"`
			} `json:"author"`
		} `json:"infos"`
	} `json:"data"`
}
