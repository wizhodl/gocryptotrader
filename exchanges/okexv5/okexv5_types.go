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
	Fee          string `json:"fee"`
	FeeCcy       string `json:"feeCcy"`
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

type MarginMode string

const (
	Isolated MarginMode = "isolated"
	Cross    MarginMode = "cross"
)

type SetLeverageRequest struct {
	InstrumentID string     `json:"instId"`
	Lever        string     `json:"lever"`
	MgnMode      MarginMode `json:"mgnMode"`
}

type SetLeverageResponse struct {
	InstrumentID string `json:"instId"`
	Lever        string `json:"lever"`
}

type MarketTicker struct {
	InstType  string `json:"instType"`
	InstId    string `json:"instId"`
	Last      string `json:"last"`
	LastSz    string `json:"lastSz"`
	AskPx     string `json:"askPx"`
	AskSz     string `json:"askSz"`
	BidPx     string `json:"bidPx"`
	BidSz     string `json:"bidSz"`
	Open24h   string `json:"open24h"`
	High24h   string `json:"high24h"`
	Low24h    string `json:"low24h"`
	VolCcy24h string `json:"volCcy24h"`
	Ts        string `json:"ts"`
}

type FundingRateHistory struct {
	InstrumentID string  `json:"instId"`
	FundingRate  float64 `json:"fundingRate,string"`
	FundingTime  int64   `json:"fundingTime,string"`
}

type AccountConfig struct {
	UserID  string `json:"uid"`
	AcctLv  string `json:"acctLv"`  // 账户层级 1：简单交易模式，2：单币种保证金模式，3：跨币种保证金模式
	PosMode string `json:"posMode"` // 持仓方式 long_short_mode：双向持仓 net_mode：单向持仓 仅适用交割/永续
}

type PosMode string

const (
	LongShortMode PosMode = "long_short_mode"
	NetMode       PosMode = "net_mode"
)
