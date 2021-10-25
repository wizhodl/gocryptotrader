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

type PlaceOrderRequest struct {
	InstrumentID string `json:"instId"`            // trading pair
	TdMode       string `json:"tdMode"`            // 交易模式 保证金模式：isolated：逐仓 ；cross：全仓 非保证金模式：cash：非保证金
	ClientOID    string `json:"clOrdId,omitempty"` // the order ID customized by yourself
	Side         string `json:"side"`              // buy or sell
	PosSide      string `json:"posSide,omitempty"` // 持仓方向 在双向持仓模式下必填，且仅可选择 long 或 short
	OrdType      string `json:"ordType"`           // market：市价单 limit：限价单 post_only：只做maker单 fok：全部成交或立即取消 ioc：立即成交并取消剩余optimal_limit_ioc：市价委托立即成交并取消剩余（仅适用交割、永续）
	Size         string `json:"sz"`
	Price        string `json:"px,omitempty"`         // price
	ReduceOnly   bool   `json:"reduceOnly,omitempty"` // 仅适用于币币杠杆订单
	SizeCurrency string `json:"tgtCcy,omitempty"`     // 市价单委托数量的类型 base_ccy: 交易货币 ；quote_ccy：计价货币 仅适用于币币订单
}

type PlaceOrderResponse struct {
	OrderID   string `json:"ordId"`
	ClientOid string `json:"clOrdId"`
	Msg       string `json:"sMsg"`
}

type Instrument struct {
	InstrumentID string `json:"instId"`
	Uly          string `json:"uly"`
	BaseCcy      string `json:"baseCcy"`
	QuoteCcy     string `json:"quoteCcy"`
	SettleCcy    string `json:"settleCcy"`
	CtVal        string `json:"ctVal"`
	CtMult       string `json:"ctMult"`
	CtValCcy     string `json:"ctValCcy"`
	ListTime     string `json:"listTime"`
	ExpTime      string `json:"expTime"`
	Lever        string `json:"lever"`
	TickSz       string `json:"tickSz"`
	LotSz        string `json:"lotSz"`
	MinSz        string `json:"minSz"`
	CtType       string `json:"ctType"`
	Alias        string `json:"alias"`
	State        string `json:"state"`
}

type CancelOrderRequest struct {
	InstrumentID string `json:"instId"`
	OrderID      string `json:"ordId,omitempty"`
	ClientOID    string `json:"clOrdId,omitempty"`
}

type CancelOrderResponse struct {
	OrderID   string `json:"ordId"`
	ClientOID string `json:"clOrdId"`
	SCode     string `json:"sCode"`
	SMsg      string `json:"sMsg"`
}

type GetTradingAccountResponse struct {
	Currency  string `json:"ccy"`
	Balance   string `json:"cashBal"`
	Available string `json:"availEq"`
	Frozen    string `json:"frozenBal"`
}

type GetPositionResponse struct {
	InstrumentType string `json:"instType"`
	InstrumentID   string `json:"instId"`
	MgnMode        string `json:"mgnMode"`
	PosId          string `json:"posId"`
	PosSide        string `json:"posSide"`
	Pos            string `json:"pos"`
	PosCcy         string `json:"posCcy"`
	AvailPos       string `json:"availPos"`
	AvgPx          string `json:"avgPx"`
	Lever          string `json:"lever"`
	LiqPx          string `json:"liqPx"`
	Last           string `json:"last"`
}

type SetLeverageRequest struct {
	InstrumentID string `json:"instId"`
	Lever        string `json:"lever"`
	MgnMode      string `json:"mgnMode"`
}

type SetLeverageResponse struct {
	InstrumentID string `json:"instId"`
	Lever        string `json:"lever"`
}
