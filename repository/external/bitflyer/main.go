package bitflyer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/pkg/errors"
)

type ExchangePairCode string

var ErrIDIsTooOld = errors.New("ID is too old")

const (
	BTC_JPY ExchangePairCode = "BTC_JPY"
	XRP_JPY ExchangePairCode = "XRP_JPY"
	ETH_JPY ExchangePairCode = "ETH_JPY"
	ETH_BTC ExchangePairCode = "ETH_BTC"
	BCH_BTC ExchangePairCode = "BCH_BTC"
	NO_DEAL ExchangePairCode = ""
)

func getBitflyerExchangePairCode(exchangePair entity.ExchangePair) ExchangePairCode {
	switch exchangePair {
	case entity.BTC_JPY:
		return BTC_JPY
	case entity.ETH_JPY:
		return ETH_JPY
	case entity.ETC_JPY:
		return NO_DEAL
	case entity.XRP_JPY:
		return XRP_JPY
	case entity.BCH_BTC:
		return BCH_BTC
	case entity.ETH_BTC:
		return ETH_BTC
	default:
		return NO_DEAL
	}
}

type BoardResponse struct {
	MidPrice float64 `json:"mid_price"`
	Bids     []struct {
		Price float64 `json:"price"`
		Size  float64 `json:"size"`
	} `json:"bids"`
	Asks []struct {
		Price float64 `json:"price"`
		Size  float64 `json:"size"`
	} `json:"asks"`
}

func GetOrderBook(exchangePair entity.ExchangePair) entity.OrderBook {
	code := getBitflyerExchangePairCode(exchangePair)
	if code == NO_DEAL {
		fmt.Println("Error: No deal")
		return entity.OrderBook{}
	}

	resp, err := http.Get("https://api.bitflyer.com/v1/board?product_code=" + string(code))
	if err != nil {
		fmt.Println("Error:", err)
		return entity.OrderBook{}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return entity.OrderBook{}
	}

	var mappedResp BoardResponse
	err = json.Unmarshal(body, &mappedResp)
	if err != nil {
		fmt.Println("Error:", err)
		return entity.OrderBook{}
	}

	var orderBook entity.OrderBook

	for _, bids := range mappedResp.Bids {
		orderBook.Bids = append(orderBook.Bids, entity.Order{Price: bids.Price, Volume: bids.Size})
	}
	for _, asks := range mappedResp.Asks {
		orderBook.Asks = append(orderBook.Asks, entity.Order{Price: asks.Price, Volume: asks.Size})
	}

	return orderBook
}

type Side string

const (
	Default Side = ""
	Buy     Side = "BUY"
	Sell    Side = "SELL"
)

type ExecutionsResponse []struct {
	Id                         int     `json:"id"`
	Side                       Side    `json:"side"`
	Price                      float64 `json:"price"`
	Size                       float64 `json:"size"`
	ExecDate                   string  `json:"exec_date"`
	BuyChildOrderAcceptanceId  string  `json:"buy_child_order_acceptance_id"`
	SellChildOrderAcceptanceId string  `json:"sell_child_order_acceptance_id"`
}

type BitflyerBadRequestResponse struct {
	Status       int    `json:"status"`
	ErrorMessage string `json:"error_message"`
}

func GetTradesByLastID(exchangePair entity.ExchangePair, lastID int) (entity.TradeCollection, error) {
	code := getBitflyerExchangePairCode(exchangePair)
	if code == NO_DEAL {
		err := fmt.Errorf("Exchange pair %s is not supported by Bitflyer.", exchangePair.String())
		return nil, errors.WithStack(err)
	}

	query := "product_code=" + string(code) + "&before=" + strconv.Itoa(lastID+1) + "&count=500"
	resp, err := http.Get("https://api.bitflyer.com/v1/executions?" + query)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()

	handledErr := []int{http.StatusOK, http.StatusBadRequest}
	if !slices.Contains(handledErr, resp.StatusCode) {
		err := fmt.Errorf("Status code %d", resp.StatusCode)
		return nil, errors.WithStack(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if resp.StatusCode == http.StatusBadRequest {
		var mappedResp BitflyerBadRequestResponse
		err = json.Unmarshal(body, &mappedResp)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if mappedResp.Status == -156 {
			return nil, ErrIDIsTooOld
		} else {
			err = fmt.Errorf("%v", mappedResp)
			return nil, errors.WithStack(err)
		}
	}

	var mappedResp ExecutionsResponse
	err = json.Unmarshal(body, &mappedResp)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var recentTrades entity.TradeCollection
	for _, execution := range mappedResp {
		time, err := time.Parse(time.RFC3339, execution.ExecDate+"Z")
		if err != nil {
			return nil, errors.WithStack(err)
		}
		recentTrades = append(
			recentTrades,
			entity.Trade{
				ExchangePlace: entity.Bitflyer,
				ExchangePair:  exchangePair,
				TradeID:       execution.Id,
				Price:         execution.Price,
				Volume:        execution.Size,
				Time:          time,
			},
		)
	}

	return recentTrades, nil
}
