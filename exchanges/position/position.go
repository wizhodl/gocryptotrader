package position

type PositionSide string

const PositionSideLong PositionSide = "Long"
const PositionSideShort PositionSide = "Short"

type Position struct {
	FutureSymbol     string
	Qty              float64
	EntryPrice       float64
	MarkPrice        float64
	Leverage         float64
	MaxQty           float64
	Side             PositionSide
	LiquidationPrice float64
	UnrealisedPnl    float64
	RealisedPnl      float64
}
