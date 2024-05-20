package bitflyer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/mass584/auto-trade/entity"
)

type ExchangePairCode string

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
	case entity.BTC_TO_JPY:
		return BTC_JPY
	case entity.ETH_TO_JPY:
		return ETH_JPY
	case entity.ETC_TO_JPY:
		return NO_DEAL
	case entity.XRP_TO_JPY:
		return XRP_JPY
	case entity.BCH_TO_BTC:
		return BCH_BTC
	case entity.ETH_TO_BTC:
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

func GetRecentTrades(exchangePair entity.ExchangePair) entity.TradeCollection {
	code := getBitflyerExchangePairCode(exchangePair)
	if code == NO_DEAL {
		fmt.Println("Error: No deal")
		return []entity.Trade{}
	}

	query := "product_code=" + string(code) + "&count=100"
	resp, err := http.Get("https://api.bitflyer.com/v1/executions?" + query)
	if err != nil {
		fmt.Println("Error:", err)
		return []entity.Trade{}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return []entity.Trade{}
	}

	var mappedResp ExecutionsResponse
	err = json.Unmarshal(body, &mappedResp)
	if err != nil {
		fmt.Println("Error:", err)
		return []entity.Trade{}
	}

	var recentTrades entity.TradeCollection
	for _, execution := range mappedResp {
		id := "bitflyer-" + strconv.Itoa(execution.Id)

		time, err := time.Parse(time.RFC3339, execution.ExecDate+"Z")
		if err != nil {
			fmt.Println("Error:", err)
		}
		recentTrades = append(recentTrades, entity.Trade{ID: id, Price: execution.Price, Volume: execution.Size, Time: time})
	}

	return recentTrades
}
