package okexv5

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/common/crypto"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/okgroup"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
	"github.com/thrasher-corp/gocryptotrader/log"
)

const (
	okExRateInterval = time.Second
	okExRequestRate  = 6
	okExAPIPath      = "api/"
	okExAPIURL       = "https://www.okex.com/" + okExAPIPath
	okExAPIVersion   = "v5/"
	okExExchangeName = "OKEX"
	// OkExWebsocketURL WebsocketURL
	OkExWebsocketURL = "wss://real.okex.com:8443/ws/v5"
	// API subsections
	okGroupSpotSubsection    = ""
	okGroupFuturesSubsection = ""
	okGroupSwapSubsection    = ""
	okGroupETTSubsection     = ""
	okGroupMarginSubsection  = ""
	// v5 api endpoint
	publicInstruments = "public/instruments"
	accountPositions  = "account/positions"

	// Futures based endpoints
	okGroupFuturePosition = "position"
	okGroupFutureLeverage = "leverage"
	okGroupFutureOrder    = "order"
	okGroupFutureHolds    = "holds"
	okGroupIndices        = "index"
	okGroupRate           = "rate"
	okGroupEsimtatedPrice = "estimated_price"
	okGroupOpenInterest   = "open_interest"
	// Perpetual swap based endpoints
	okGroupSettings              = "settings"
	okGroupDepth                 = "depth"
	okGroupFundingTime           = "funding_time"
	okGroupHistoricalFundingRate = "historical_funding_rate"
	okGroupSwapInstruments       = "instruments"
	// ETT endpoints
	okGroupConstituents    = "constituents"
	okGroupDefinePrice     = "define-price"
	okGroupPerpSwapRates   = "instruments/%s/historical_funding_rate?"
	okGroupPerpTickers     = "instruments/ticker"
	okGroupMarginPairData  = "accounts/%s/availability"
	okGroupMarginPairsData = "accounts/availability"
	okGroupSpotPairs       = "instruments"
)

// OKEX bases all account, spot and margin methods off okgroup implementation
type OKEX struct {
	okgroup.OKGroup
}

// GetSwapMarkets gets perpetual swap markets
func (o *OKEX) GetSwapMarkets() ([]okgroup.SwapInstrumentsData, error) {
	var resp []okgroup.SwapInstrumentsData
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection,
		okGroupSwapInstruments,
		nil, &resp, false)
}

// GetAllMarginRates gets interest rates for all margin currencies on OKEX
func (o *OKEX) GetAllMarginRates() ([]okgroup.MarginCurrencyData, error) {
	var resp []okgroup.MarginCurrencyData
	var result []map[string]interface{}
	var tempResp okgroup.MarginCurrencyData
	tempResp.Data = make(map[string]okgroup.MarginData)
	err := o.SendHTTPRequest(exchange.RestSpot, http.MethodGet,
		okGroupMarginSubsection,
		okGroupMarginPairsData,
		nil,
		&result,
		true)
	if err != nil {
		return resp, err
	}
	for i := range result {
		for k, v := range result[i] {
			if strings.Contains(k, "currency:") {
				var byteData []byte
				var marginData okgroup.MarginData
				currencyString := strings.Replace(k, "currency:", "", 1)
				byteData, err = json.Marshal(v)
				if err != nil {
					return resp, err
				}
				err = json.Unmarshal(byteData, &marginData)
				if err != nil {
					return resp, err
				}
				tempResp.Data[currencyString] = marginData
			}
			var strData string
			var ok bool
			strData, ok = result[i]["instrument_id"].(string)
			if !ok {
				return resp, errors.New("type conversion failed for instrument_id")
			}
			tempResp.InstrumentID = strData
			strData, ok = result[i]["product_id"].(string)
			if !ok {
				return resp, errors.New("type conversion failed for product_id")
			}
			tempResp.ProductID = strData
			resp = append(resp, tempResp)
		}
	}
	return resp, nil
}

// GetMarginRates gets interest rates for margin currencies
func (o *OKEX) GetMarginRates(instrumentID currency.Pair) (okgroup.MarginCurrencyData, error) {
	var resp okgroup.MarginCurrencyData
	resp.Data = make(map[string]okgroup.MarginData)
	var result []map[string]interface{}
	err := o.SendHTTPRequest(exchange.RestSpot, http.MethodGet,
		okGroupMarginSubsection,
		fmt.Sprintf(okGroupMarginPairData, instrumentID),
		nil,
		&result,
		true)
	if err != nil {
		return resp, err
	}
	for i := range result {
		for k, v := range result[i] {
			var byteData []byte
			var marginData okgroup.MarginData
			byteData, err = json.Marshal(v)
			if err != nil {
				return resp, err
			}
			if strings.Contains(k, instrumentID.Base.String()) {
				err = json.Unmarshal(byteData, &marginData)
				if err != nil {
					return resp, err
				}
				resp.Data[instrumentID.Base.String()] = marginData
			} else if strings.Contains(k, instrumentID.Quote.String()) {
				err = json.Unmarshal(byteData, &marginData)
				if err != nil {
					return resp, err
				}
				resp.Data[instrumentID.Quote.String()] = marginData
			}
		}
		var strData string
		var ok bool
		strData, ok = result[i]["instrument_id"].(string)
		if !ok {
			return resp, errors.New("type conversion failed for instrument_id")
		}
		resp.InstrumentID = strData
		strData, ok = result[i]["product_id"].(string)
		if !ok {
			return resp, errors.New("type conversion failed for product_id")
		}
		resp.ProductID = strData
	}
	return resp, nil
}

// GetSpotMarkets gets perpetual swap markets' data
func (o *OKEX) GetSpotMarkets() ([]okgroup.TradingPairData, error) {
	var resp []okgroup.TradingPairData
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSpotSubsection, okGroupSpotPairs, nil, &resp, false)
}

// GetFundingRate gets funding rate of a given currency
func (o *OKEX) GetFundingRate(marketName, limit string) ([]okgroup.PerpSwapFundingRates, error) {
	params := url.Values{}
	params.Set("limit", limit)
	var resp []okgroup.PerpSwapFundingRates
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection,
		fmt.Sprintf(okGroupPerpSwapRates, marketName)+params.Encode(),
		nil, &resp, false)
}

// GetPerpSwapMarkets gets perpetual swap markets' data
func (o *OKEX) GetPerpSwapMarkets() ([]okgroup.TickerData, error) {
	var resp []okgroup.TickerData
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection,
		okGroupPerpTickers,
		nil, &resp, false)
}

// GetFuturesPostions Get the information of all holding positions in futures trading.
// Due to high energy consumption, you are advised to capture data with the "Futures Account of a Currency" API instead.
func (o *OKEX) GetFuturesPostions() (resp okgroup.GetFuturesPositionsResponse, _ error) {
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupFuturesSubsection, okGroupFuturePosition, nil, &resp, true)
}

// GetFuturesPostionsForCurrency Get the information of holding positions of a contract.
func (o *OKEX) GetFuturesPostionsForCurrency(instrumentID string) (resp okgroup.GetFuturesPositionsForCurrencyResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v", instrumentID, okGroupFuturePosition)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupFuturesSubsection, requestURL, nil, &resp, true)
}

// GetFuturesAccountOfAllCurrencies Get the futures account info of all token.
// Due to high energy consumption, you are advised to capture data with the "Futures Account of a Currency" API instead.
func (o *OKEX) GetFuturesAccountOfAllCurrencies() (resp okgroup.FuturesAccountForAllCurrenciesResponse, _ error) {
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupFuturesSubsection, okgroup.OKGroupAccounts, nil, &resp, true)
}

// GetFuturesAccountOfACurrency Get the futures account info of a token.
func (o *OKEX) GetFuturesAccountOfACurrency(instrumentID string) (resp okgroup.FuturesCurrencyData, _ error) {
	requestURL := fmt.Sprintf("%v/%v", okgroup.OKGroupAccounts, instrumentID)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupFuturesSubsection, requestURL, nil, &resp, true)
}

// GetFuturesLeverage Get the leverage of the futures account
func (o *OKEX) GetFuturesLeverage(instrumentID string) (resp okgroup.GetFuturesLeverageResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupAccounts, instrumentID, okGroupFutureLeverage)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupFuturesSubsection, requestURL, nil, &resp, true)
}

// SetFuturesLeverage Adjusting the leverage for futures account。
// Cross margin request requirements:  {"leverage":"10"}
// Fixed margin request requirements: {"instrument_id":"BTC-USD-180213","direction":"long","leverage":"10"}
func (o *OKEX) SetFuturesLeverage(request okgroup.SetFuturesLeverageRequest) (resp okgroup.SetFuturesLeverageResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupAccounts, request.Currency, okGroupFutureLeverage)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodPost, okGroupFuturesSubsection, requestURL, request, &resp, true)
}

// GetFuturesBillDetails Shows the account’s historical coin in flow and out flow.
// All paginated requests return the latest information (newest) as the first page sorted by newest (in chronological time) first.
func (o *OKEX) GetFuturesBillDetails(request okgroup.GetSpotBillDetailsForCurrencyRequest) (resp []okgroup.GetSpotBillDetailsForCurrencyResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v%v", okgroup.OKGroupAccounts, request.Currency, okgroup.OKGroupLedger, okgroup.FormatParameters(request))
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupFuturesSubsection, requestURL, nil, &resp, true)
}

// PlaceFuturesOrder OKEx futures trading only supports limit orders.
// You can place an order only if you have enough funds. Once your order is placed, the amount will be put on hold in the order lifecycle.
// The assets and amount on hold depends on the order's specific type and parameters.
func (o *OKEX) PlaceFuturesOrder(request okgroup.PlaceFuturesOrderRequest) (resp okgroup.PlaceFuturesOrderResponse, _ error) {
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodPost, okGroupFuturesSubsection, okGroupFutureOrder, request, &resp, true)
}

// PlaceFuturesOrderBatch Batch contract placing order operation.
func (o *OKEX) PlaceFuturesOrderBatch(request okgroup.PlaceFuturesOrderBatchRequest) (resp okgroup.PlaceFuturesOrderBatchResponse, _ error) {
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodPost, okGroupFuturesSubsection, okgroup.OKGroupOrders, request, &resp, true)
}

// CancelFuturesOrder Cancelling an unfilled order.
func (o *OKEX) CancelFuturesOrder(request okgroup.CancelFuturesOrderRequest) (resp okgroup.CancelFuturesOrderResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupCancelOrder, request.InstrumentID, request.OrderID)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodPost, okGroupFuturesSubsection, requestURL, request, &resp, true)
}

// CancelFuturesOrderBatch With best effort, cancel all open orders.
func (o *OKEX) CancelFuturesOrderBatch(request okgroup.CancelMultipleSpotOrdersRequest) (resp okgroup.CancelMultipleSpotOrdersResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v", okgroup.OKGroupCancelBatchOrders, request.InstrumentID)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodPost, okGroupFuturesSubsection, requestURL, request, &resp, true)
}

// GetFuturesOrderList List your orders. Cursor pagination is used.
// All paginated requests return the latest information (newest) as the first page sorted by newest (in chronological time) first.
func (o *OKEX) GetFuturesOrderList(request okgroup.GetFuturesOrdersListRequest) (resp okgroup.GetFuturesOrderListResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v%v", okgroup.OKGroupOrders, request.InstrumentID, okgroup.FormatParameters(request))
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupFuturesSubsection, requestURL, nil, &resp, true)
}

// GetFuturesOrderDetails Get order details by order ID.
func (o *OKEX) GetFuturesOrderDetails(request okgroup.GetFuturesOrderDetailsRequest) (resp okgroup.GetFuturesOrderDetailsResponseData, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupOrders, request.InstrumentID, request.OrderID)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupFuturesSubsection, requestURL, nil, &resp, true)
}

// GetFuturesTransactionDetails  Get details of the recent filled orders. Cursor pagination is used.
// All paginated requests return the latest information (newest) as the first page sorted by newest (in chronological time) first.
func (o *OKEX) GetFuturesTransactionDetails(request okgroup.GetFuturesTransactionDetailsRequest) (resp []okgroup.GetFuturesTransactionDetailsResponse, _ error) {
	requestURL := fmt.Sprintf("%v%v", okgroup.OKGroupGetSpotTransactionDetails, okgroup.FormatParameters(request))
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupFuturesSubsection, requestURL, nil, &resp, true)
}

// GetFuturesContractInformation Get market data. This endpoint provides the snapshots of market data and can be used without verifications.
func (o *OKEX) GetFuturesContractInformation() (resp []okgroup.GetFuturesContractInformationResponse, _ error) {
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupFuturesSubsection, okgroup.OKGroupInstruments, nil, &resp, false)
}

// GetAllFuturesTokenInfo Get the last traded price, best bid/ask price, 24 hour trading volume and more info of all contracts.
func (o *OKEX) GetAllFuturesTokenInfo() (resp []okgroup.GetFuturesTokenInfoResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v", okgroup.OKGroupInstruments, okgroup.OKGroupTicker)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupFuturesSubsection, requestURL, nil, &resp, false)
}

// GetFuturesTokenInfoForCurrency Get the last traded price, best bid/ask price, 24 hour trading volume and more info of a contract.
func (o *OKEX) GetFuturesTokenInfoForCurrency(instrumentID string) (resp okgroup.GetFuturesTokenInfoResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupInstruments, instrumentID, okgroup.OKGroupTicker)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupFuturesSubsection, requestURL, nil, &resp, false)
}

// GetFuturesFilledOrder Get the recent 300 transactions of all contracts. Pagination is not supported here.
// The whole book will be returned for one request. Websocket is recommended here.
func (o *OKEX) GetFuturesFilledOrder(request okgroup.GetFuturesFilledOrderRequest) (resp []okgroup.GetFuturesFilledOrdersResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v%v", okgroup.OKGroupInstruments, request.InstrumentID, okgroup.OKGroupTrades, okgroup.FormatParameters(request))
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupFuturesSubsection, requestURL, nil, &resp, false)
}

// GetFuturesHoldAmount Get the number of futures with hold.
func (o *OKEX) GetFuturesHoldAmount(instrumentID string) (resp okgroup.GetFuturesHoldAmountResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupAccounts, instrumentID, okGroupFutureHolds)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupFuturesSubsection, requestURL, nil, &resp, true)
}

// GetFuturesIndices Get Indices of tokens. This is a public endpoint, no identity verification is needed.
func (o *OKEX) GetFuturesIndices(instrumentID string) (resp okgroup.GetFuturesIndicesResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupInstruments, instrumentID, okGroupIndices)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupFuturesSubsection, requestURL, nil, &resp, false)
}

// GetFuturesExchangeRates Get the fiat exchange rates. This is a public endpoint, no identity verification is needed.
func (o *OKEX) GetFuturesExchangeRates() (resp okgroup.GetFuturesExchangeRatesResponse, _ error) {
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupFuturesSubsection, okGroupRate, nil, &resp, false)
}

// GetFuturesEstimatedDeliveryPrice the estimated delivery price. It is available 3 hours before delivery.
// This is a public endpoint, no identity verification is needed.
func (o *OKEX) GetFuturesEstimatedDeliveryPrice(instrumentID string) (resp okgroup.GetFuturesEstimatedDeliveryPriceResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupInstruments, instrumentID, okGroupEsimtatedPrice)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupFuturesSubsection, requestURL, nil, &resp, false)
}

// GetFuturesOpenInterests Get the open interest of a contract. This is a public endpoint, no identity verification is needed.
func (o *OKEX) GetFuturesOpenInterests(instrumentID string) (resp okgroup.GetFuturesOpenInterestsResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupInstruments, instrumentID, okGroupOpenInterest)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupFuturesSubsection, requestURL, nil, &resp, false)
}

// GetFuturesCurrentPriceLimit The maximum buying price and the minimum selling price of the contract.
// This is a public endpoint, no identity verification is needed.
func (o *OKEX) GetFuturesCurrentPriceLimit(instrumentID string) (resp okgroup.GetFuturesCurrentPriceLimitResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupInstruments, instrumentID, okgroup.OKGroupPriceLimit)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupFuturesSubsection, requestURL, nil, &resp, false)
}

// GetFuturesCurrentMarkPrice The maximum buying price and the minimum selling price of the contract.
// This is a public endpoint, no identity verification is needed.
func (o *OKEX) GetFuturesCurrentMarkPrice(instrumentID string) (resp okgroup.GetFuturesCurrentMarkPriceResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupInstruments, instrumentID, okgroup.OKGroupMarkPrice)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupFuturesSubsection, requestURL, nil, &resp, false)
}

// GetFuturesForceLiquidatedOrders Get force liquidated orders. This is a public endpoint, no identity verification is needed.
func (o *OKEX) GetFuturesForceLiquidatedOrders(request okgroup.GetFuturesForceLiquidatedOrdersRequest) (resp []okgroup.GetFuturesForceLiquidatedOrdersResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v%v", okgroup.OKGroupInstruments, request.InstrumentID, okgroup.OKGroupLiquidation, okgroup.FormatParameters(request))
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupFuturesSubsection, requestURL, nil, &resp, false)
}

// GetFuturesTagPrice Get the tag price. This is a public endpoint, no identity verification is needed.
func (o *OKEX) GetFuturesTagPrice(instrumentID string) (resp okgroup.GetFuturesTagPriceResponse, _ error) {
	// OKEX documentation is missing for this endpoint. Guessing "tag_price" for the URL results in 404
	return okgroup.GetFuturesTagPriceResponse{}, common.ErrNotYetImplemented
}

// GetSwapPostions Get the information of all holding positions in swap trading.
// Due to high energy consumption, you are advised to capture data with the "Swap Account of a Currency" API instead.
func (o *OKEX) GetSwapPostions() (resp []okgroup.GetSwapPostionsResponse, _ error) {
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection, okGroupFuturePosition, nil, &resp, true)
}

// GetSwapPostionsForContract Get the information of holding positions of a contract.
func (o *OKEX) GetSwapPostionsForContract(instrumentID string) (resp okgroup.GetSwapPostionsResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v", instrumentID, okGroupFuturePosition)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection, requestURL, nil, &resp, true)
}

// GetSwapAccountOfAllCurrency Get the perpetual swap account info of a token.
// Margin ratio set as 10,000 when users have no open position.
func (o *OKEX) GetSwapAccountOfAllCurrency() (resp okgroup.GetSwapAccountOfAllCurrencyResponse, _ error) {
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection, okgroup.OKGroupAccounts, nil, &resp, true)
}

// GetSwapAccountSettingsOfAContract Get leverage level and margin mode of a contract.
func (o *OKEX) GetSwapAccountSettingsOfAContract(instrumentID string) (resp okgroup.GetSwapAccountSettingsOfAContractResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupAccounts, instrumentID, okGroupSettings)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection, requestURL, nil, &resp, true)
}

// SetSwapLeverageLevelOfAContract Setting the leverage level of a contract
// TODO this returns invalid parameters, but matches spec. Unsure how to fix
func (o *OKEX) SetSwapLeverageLevelOfAContract(request okgroup.SetSwapLeverageLevelOfAContractRequest) (resp okgroup.SetSwapLeverageLevelOfAContractResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupAccounts, request.InstrumentID, okGroupFutureLeverage)
	request.InstrumentID = ""
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodPost, okGroupSwapSubsection, requestURL, request, &resp, true)
}

// GetSwapBillDetails Shows the account’s historical coin in flow and out flow.
// All paginated requests return the latest information (newest) as the first page sorted by newest (in chronological time) first.
func (o *OKEX) GetSwapBillDetails(request okgroup.GetSpotBillDetailsForCurrencyRequest) (resp []okgroup.GetSwapBillDetailsResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v%v", okgroup.OKGroupAccounts, request.Currency, okgroup.OKGroupLedger, okgroup.FormatParameters(request))
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection, requestURL, nil, &resp, true)
}

// PlaceSwapOrder OKEx perpetual swap trading only supports limit orders，USD as quote currency for orders.
// You can place an order only if you have enough funds. Once your order is placed, the amount will be put on hold in the order lifecycle.
// The assets and amount on hold depends on the order's specific type and parameters.
func (o *OKEX) PlaceSwapOrder(request okgroup.PlaceSwapOrderRequest) (resp okgroup.PlaceSwapOrderResponse, _ error) {
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodPost, okGroupSwapSubsection, okGroupFutureOrder, request, &resp, true)
}

// PlaceMultipleSwapOrders Batch contract placing order operation.
func (o *OKEX) PlaceMultipleSwapOrders(request okgroup.PlaceMultipleSwapOrdersRequest) (resp okgroup.PlaceMultipleSwapOrdersResponse, _ error) {
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodPost, okGroupSwapSubsection, okgroup.OKGroupOrders, request, &resp, true)
}

// CancelSwapOrder Cancelling an unfilled order
func (o *OKEX) CancelSwapOrder(request okgroup.CancelSwapOrderRequest) (resp okgroup.CancelSwapOrderResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupCancelOrder, request.InstrumentID, request.OrderID)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodPost, okGroupSwapSubsection, requestURL, nil, &resp, true)
}

// CancelMultipleSwapOrders With best effort, cancel all open orders.
func (o *OKEX) CancelMultipleSwapOrders(request okgroup.CancelMultipleSwapOrdersRequest) (resp okgroup.CancelMultipleSwapOrdersResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v", okgroup.OKGroupCancelBatchOrders, request.InstrumentID)
	request.InstrumentID = ""
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodPost, okGroupSwapSubsection, requestURL, request, &resp, true)
}

// GetSwapOrderList List your orders. Cursor pagination is used.
// All paginated requests return the latest information (newest) as the first page sorted by newest (in chronological time) first.
func (o *OKEX) GetSwapOrderList(request okgroup.GetSwapOrderListRequest) (resp okgroup.GetSwapOrderListResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v%v", okgroup.OKGroupOrders, request.InstrumentID, okgroup.FormatParameters(request))
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection, requestURL, nil, &resp, true)
}

// GetSwapOrderDetails Get order details by order ID.
func (o *OKEX) GetSwapOrderDetails(request okgroup.GetSwapOrderDetailsRequest) (resp okgroup.GetSwapOrderListResponseData, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupOrders, request.InstrumentID, request.OrderID)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection, requestURL, nil, &resp, true)
}

// GetSwapTransactionDetails Get details of the recent filled orders
func (o *OKEX) GetSwapTransactionDetails(request okgroup.GetSwapTransactionDetailsRequest) (resp []okgroup.GetSwapTransactionDetailsResponse, _ error) {
	requestURL := fmt.Sprintf("%v%v", okgroup.OKGroupGetSpotTransactionDetails, okgroup.FormatParameters(request))
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection, requestURL, nil, &resp, true)
}

// GetSwapContractInformation Get market data.
func (o *OKEX) GetSwapContractInformation() (resp []okgroup.GetSwapContractInformationResponse, _ error) {
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection, okgroup.OKGroupInstruments, nil, &resp, false)
}

// GetAllSwapTokensInformation Get the last traded price, best bid/ask price, 24 hour trading volume and more info of all contracts.
func (o *OKEX) GetAllSwapTokensInformation() (resp []okgroup.GetAllSwapTokensInformationResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v", okgroup.OKGroupInstruments, okgroup.OKGroupTicker)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection, requestURL, nil, &resp, false)
}

// GetSwapTokensInformationForCurrency Get the last traded price, best bid/ask price, 24 hour trading volume and more info of all contracts.
func (o *OKEX) GetSwapTokensInformationForCurrency(instrumentID string) (resp okgroup.GetAllSwapTokensInformationResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupInstruments, instrumentID, okgroup.OKGroupTicker)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection, requestURL, nil, &resp, false)
}

// GetSwapFilledOrdersData Get details of the recent filled orders
func (o *OKEX) GetSwapFilledOrdersData(request *okgroup.GetSwapFilledOrdersDataRequest) (resp []okgroup.GetSwapFilledOrdersDataResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v%v", okgroup.OKGroupInstruments, request.InstrumentID, okgroup.OKGroupTrades, okgroup.FormatParameters(request))
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection, requestURL, nil, &resp, false)
}

// GetSwapIndices Get Indices of tokens.
func (o *OKEX) GetSwapIndices(instrumentID string) (resp okgroup.GetSwapIndecesResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupInstruments, instrumentID, okGroupIndices)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection, requestURL, nil, &resp, false)
}

// GetSwapExchangeRates Get the fiat exchange rates.
func (o *OKEX) GetSwapExchangeRates() (resp okgroup.GetSwapExchangeRatesResponse, _ error) {
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection, okGroupRate, nil, &resp, false)
}

// GetSwapOpenInterest Get the open interest of a contract.
func (o *OKEX) GetSwapOpenInterest(instrumentID string) (resp okgroup.GetSwapExchangeRatesResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupInstruments, instrumentID, okGroupOpenInterest)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection, requestURL, nil, &resp, false)
}

// GetSwapCurrentPriceLimits Get the open interest of a contract.
func (o *OKEX) GetSwapCurrentPriceLimits(instrumentID string) (resp okgroup.GetSwapCurrentPriceLimitsResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupInstruments, instrumentID, okgroup.OKGroupPriceLimit)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection, requestURL, nil, &resp, false)
}

// GetSwapForceLiquidatedOrders Get force liquidated orders.
func (o *OKEX) GetSwapForceLiquidatedOrders(request okgroup.GetSwapForceLiquidatedOrdersRequest) (resp []okgroup.GetSwapForceLiquidatedOrdersResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v%v", okgroup.OKGroupInstruments, request.InstrumentID, okgroup.OKGroupLiquidation, okgroup.FormatParameters(request))
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection, requestURL, nil, &resp, false)
}

// GetSwapOnHoldAmountForOpenOrders Get On Hold Amount for Open Orders.
func (o *OKEX) GetSwapOnHoldAmountForOpenOrders(instrumentID string) (resp okgroup.GetSwapOnHoldAmountForOpenOrdersResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupAccounts, instrumentID, okGroupFutureHolds)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection, requestURL, nil, &resp, true)
}

// GetSwapNextSettlementTime Get the time of next settlement.
func (o *OKEX) GetSwapNextSettlementTime(instrumentID string) (resp okgroup.GetSwapNextSettlementTimeResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupInstruments, instrumentID, okGroupFundingTime)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection, requestURL, nil, &resp, false)
}

// GetSwapMarkPrice Get the time of next settlement.
func (o *OKEX) GetSwapMarkPrice(instrumentID string) (resp okgroup.GetSwapMarkPriceResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupInstruments, instrumentID, okgroup.OKGroupMarkPrice)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection, requestURL, nil, &resp, false)
}

// GetSwapFundingRateHistory Get Funding Rate History.
func (o *OKEX) GetSwapFundingRateHistory(request okgroup.GetSwapFundingRateHistoryRequest) (resp []okgroup.GetSwapFundingRateHistoryResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v%v", okgroup.OKGroupInstruments, request.InstrumentID, okGroupHistoricalFundingRate, okgroup.FormatParameters(request))
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupSwapSubsection, requestURL, nil, &resp, false)
}

// GetETT List the assets in ETT account. Get information such as balance, amount on hold/ available.
func (o *OKEX) GetETT() (resp []okgroup.GetETTResponse, _ error) {
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupETTSubsection, okgroup.OKGroupAccounts, nil, &resp, true)
}

// GetETTAccountInformationForCurrency Getting the balance, amount available/on hold of a token in ETT account.
func (o *OKEX) GetETTAccountInformationForCurrency(currency string) (resp okgroup.GetETTResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v", okgroup.OKGroupAccounts, currency)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupETTSubsection, requestURL, nil, &resp, true)
}

// GetETTBillsDetails Bills details. All paginated requests return the latest information (newest)
// as the first page sorted by newest (in chronological time) first
func (o *OKEX) GetETTBillsDetails(currency string) (resp []okgroup.GetETTBillsDetailsResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v/%v", okgroup.OKGroupAccounts, currency, okgroup.OKGroupLedger)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupETTSubsection, requestURL, nil, &resp, true)
}

// PlaceETTOrder You can place subscription or redemption orders under ETT trading.
// You can place an order only if you have enough funds. Once your order is placed,
// the amount will be put on hold in the order lifecycle.
// The assets and amount on hold depends on the order's specific type and parameters.
func (o *OKEX) PlaceETTOrder(request *okgroup.PlaceETTOrderRequest) (resp okgroup.PlaceETTOrderResponse, _ error) {
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodPost, okGroupETTSubsection, okgroup.OKGroupOrders, nil, &resp, true)
}

// CancelETTOrder Cancel an unfilled order.
func (o *OKEX) CancelETTOrder(orderID string) (resp okgroup.PlaceETTOrderResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v", okgroup.OKGroupOrders, orderID)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodDelete, okGroupETTSubsection, requestURL, nil, &resp, true)
}

// GetETTOrderList List your orders. Cursor pagination is used. All paginated requests return the latest information
// (newest) as the first page sorted by newest (in chronological time) first.
func (o *OKEX) GetETTOrderList(request okgroup.GetETTOrderListRequest) (resp []okgroup.GetETTOrderListResponse, _ error) {
	requestURL := fmt.Sprintf("%v%v", okgroup.OKGroupOrders, okgroup.FormatParameters(request))
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupETTSubsection, requestURL, nil, &resp, true)
}

// GetETTOrderDetails Get order details by order ID.
func (o *OKEX) GetETTOrderDetails(orderID string) (resp okgroup.GetETTOrderListResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v", okgroup.OKGroupOrders, orderID)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupETTSubsection, requestURL, nil, &resp, true)
}

// GetETTConstituents Get ETT Constituents.This is a public endpoint, no identity verification is needed.
func (o *OKEX) GetETTConstituents(ett string) (resp okgroup.GetETTConstituentsResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v", okGroupConstituents, ett)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupETTSubsection, requestURL, nil, &resp, false)
}

// GetETTSettlementPriceHistory Get ETT settlement price history. This is a public endpoint, no identity verification is needed.
func (o *OKEX) GetETTSettlementPriceHistory(ett string) (resp []okgroup.GetETTSettlementPriceHistoryResponse, _ error) {
	requestURL := fmt.Sprintf("%v/%v", okGroupDefinePrice, ett)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, okGroupETTSubsection, requestURL, nil, &resp, false)
}

// v5 不需要 requestType
func (o *OKEX) SendHTTPRequest(ep exchange.URL, httpMethod, requestType, requestPath string, data, result interface{}, authenticated bool) (err error) {
	if authenticated && !o.AllowAuthenticatedRequest() {
		return fmt.Errorf(exchange.WarningAuthenticatedRequestWithoutCredentialsSet,
			o.Name)
	}
	endpoint, err := o.API.Endpoints.GetURL(ep)
	if err != nil {
		return err
	}
	now := time.Now()
	utcTime := now.UTC().Format(time.RFC3339)
	payload := []byte("")

	if data != nil {
		payload, err = json.Marshal(data)
		if err != nil {
			return errors.New("sendHTTPRequest: Unable to JSON request")
		}

		if o.Verbose {
			log.Debugf(log.ExchangeSys, "Request JSON: %s\n", payload)
		}
	}

	path := endpoint + o.APIVersion + requestPath
	if o.Verbose {
		log.Debugf(log.ExchangeSys, "Sending %v request to %s \n", requestType, path)
	}

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	if authenticated {
		signPath := fmt.Sprintf("/%v%v%v", okExAPIPath, o.APIVersion, requestPath)
		hmac := crypto.GetHMAC(crypto.HashSHA256,
			[]byte(utcTime+httpMethod+signPath+string(payload)),
			[]byte(o.API.Credentials.Secret))
		headers["OK-ACCESS-KEY"] = o.API.Credentials.Key
		headers["OK-ACCESS-SIGN"] = crypto.Base64Encode(hmac)
		headers["OK-ACCESS-TIMESTAMP"] = utcTime
		headers["OK-ACCESS-PASSPHRASE"] = o.API.Credentials.ClientID
		if o.API.IsTestnet {
			headers["x-simulated-trading"] = "1"
		}
	}

	// Requests that have a 30+ second difference between the timestamp and the API service time will be considered expired or rejected
	ctx, cancel := context.WithDeadline(context.Background(), now.Add(30*time.Second))
	defer cancel()
	var intermediary json.RawMessage
	type response struct {
		Code string        `json:"code,omitempty"`
		Msg  string        `json:"msg,omitempty"`
		Data []interface{} `json:"data,omitempty"`
	}

	resp := response{}
	err = o.SendPayload(ctx, &request.Item{
		Method:        strings.ToUpper(httpMethod),
		Path:          path,
		Headers:       headers,
		Body:          bytes.NewBuffer(payload),
		Result:        &intermediary,
		AuthRequest:   authenticated,
		Verbose:       o.Verbose,
		HTTPDebugging: o.HTTPDebugging,
		HTTPRecording: o.HTTPRecording,
	})
	if err != nil {
		return err
	}

	err = json.Unmarshal(intermediary, &resp)
	if err == nil {
		if resp.Code != "0" {
			msg := resp.Msg
			if len(resp.Data) > 0 {
				type sMsgData struct {
					SCode string `json:"sCode"`
					SMsg  string `json:"sMsg"`
				}
				var msgData sMsgData
				dataJson, _ := json.Marshal(resp.Data[0])
				if err := json.Unmarshal(dataJson, &msgData); err == nil && msgData.SMsg != "" {
					msg = msgData.SMsg
				}
			}
			return fmt.Errorf("error[%v]: %v", resp.Code, msg)
		}
	}

	if len(resp.Data) > 0 {
		if isArrayData(requestPath) {
			dataJson, _ := json.Marshal(resp.Data)
			return json.Unmarshal(dataJson, result)
		} else {
			data := resp.Data[0]
			dataJson, _ := json.Marshal(data)
			return json.Unmarshal(dataJson, result)
		}
	}

	return nil
}

func isArrayData(path string) bool {
	if strings.HasPrefix(path, publicInstruments) {
		return true
	} else if strings.HasPrefix(path, accountPositions) {
		return true
	}
	return false
}

func (o *OKEX) GetOrder(request OrderRequest) (resp GetOrderResponse, _ error) {
	requestURL := fmt.Sprintf("trade/order%v", okgroup.FormatParameters(request))
	err := o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, "", requestURL, request, &resp, true)
	return resp, err
}

func (o *OKEX) PlaceOrder(request *PlaceOrderRequest) (resp PlaceOrderResponse, _ error) {
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodPost, "", "trade/order", request, &resp, true)
}

func (o *OKEX) GetInstruments(i asset.Item) (resp []Instrument, _ error) {
	var instType string
	switch i {
	case asset.Spot:
		instType = "SPOT"
	case asset.Futures:
		instType = "FUTURES"
	case asset.CoinMarginedFutures:
		instType = "FUTURES"
	case asset.PerpetualSwap:
		instType = "SWAP"
	}
	requestURL := fmt.Sprintf("%s?instType=%s", publicInstruments, instType)
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, "", requestURL, nil, &resp, false)
}

func (o *OKEX) cancelOrder(request CancelOrderRequest) (resp CancelOrderResponse, _ error) {
	requestURL := fmt.Sprintf("trade/cancel-order%s", okgroup.FormatParameters(request))
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodPost, "", requestURL, request, &resp, true)
}

func (o *OKEX) GetTradingAccounts() (resp []GetTradingAccountResponse, _ error) {
	type accountData struct {
		Details []GetTradingAccountResponse `json:"details"`
		TotalEq string                      `json:"totalEq"`
	}
	var data accountData
	err := o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, "", "account/balance", nil, &data, true)
	if data.TotalEq != "" {
		resp = data.Details
	}
	return resp, err
}

func (o *OKEX) GetAccountPositions() (resp []GetPositionResponse, _ error) {
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodGet, "", accountPositions, nil, &resp, true)
}

func (o *OKEX) SetLeverage(pair currency.Pair, leverage string, mgnMode string) (resp SetLeverageResponse, _ error) {
	request := SetLeverageRequest{
		InstrumentID: pair.String(),
		Lever:        leverage,
		MgnMode:      mgnMode,
	}
	return resp, o.SendHTTPRequest(exchange.RestSpot, http.MethodPost, "", "account/set-leverage", request, &resp, true)
}
