package okexv5

type OrderRequest struct {
	OrderID      string `url:"ordId"`  // [required] order ID
	InstrumentID string `url:"instId"` // [required]trading pair
}

type GetOrderResponse struct {
	AvgPrice     string `json:"avgPx,omitempty"` // 可能为空字符串，所以不能用 float64
	FilledSize   string `json:"accFillSz,omitempty"`
	InstrumentID string `json:"instId"`
	OrderID      string `json:"ordId"`
	Price        string `json:"px,omitempty"`
	Side         string `json:"side"`
	Size         string `json:"sz,omitempty"`
	State        string `json:"state"`
	Timestamp    int64  `json:"uTime,string"`
	Type         string `json:"ordType"`
	// "feeCcy":"",
	// "fee":"",
}
