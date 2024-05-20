package coincheck

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
	BTC_JPY ExchangePairCode = "btc_jpy"
	ETC_JPY ExchangePairCode = "etc_jpy"
	NO_DEAL ExchangePairCode = ""
)

func getExchangePairCode(exchangePair entity.ExchangePair) ExchangePairCode {
	switch exchangePair {
	case entity.BTC_TO_JPY:
		return BTC_JPY
	case entity.ETH_TO_JPY:
		return NO_DEAL
	case entity.ETC_TO_JPY:
		return ETC_JPY
	default:
		return NO_DEAL
	}
}

type OrderBooksResponse struct {
	Asks [][]string `json:"asks"`
	Bids [][]string `json:"bids"`
}

func GetOrderBook(exchangePair entity.ExchangePair) entity.OrderBook {
	code := getExchangePairCode(exchangePair)
	if code == NO_DEAL {
		fmt.Println("Error: No deal")
		return entity.OrderBook{}
	}

	resp, err := http.Get("https://coincheck.com/api/order_books?pair=" + string(code))
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

	var mappedResp OrderBooksResponse
	err = json.Unmarshal(body, &mappedResp)
	if err != nil {
		fmt.Println("Error:", err)
		return entity.OrderBook{}
	}

	var orderBook entity.OrderBook

	for _, item := range mappedResp.Bids {
		price, err := strconv.ParseFloat(item[0], 64)
		if err != nil {
			fmt.Println("Error:", err)
		}
		volume, err := strconv.ParseFloat(item[1], 64)
		if err != nil {
			fmt.Println("Error:", err)
		}
		orderBook.Bids = append(orderBook.Bids, entity.Order{Price: price, Volume: volume})
	}
	for _, item := range mappedResp.Asks {
		price, err := strconv.ParseFloat(item[0], 64)
		if err != nil {
			fmt.Println("Error:", err)
		}
		volume, err := strconv.ParseFloat(item[1], 64)
		if err != nil {
			fmt.Println("Error:", err)
		}
		orderBook.Asks = append(orderBook.Asks, entity.Order{Price: price, Volume: volume})
	}

	return orderBook
}

type Order string

const (
	ASC  Order = "asc"
	DESC Order = "desc"
)

type Pagination struct {
	Limit         int   `json:"limit"`
	Order         Order `json:"order"`
	StartingAfter *int  `json:"starting_after"`
	EndingBefore  *int  `json:"ending_before"`
}

type OrderType string

const (
	BUY  OrderType = "buy"
	SELL OrderType = "sell"
)

type TradesResponse struct {
	Success    bool       `json:"success"`
	Pagination Pagination `json:"pagination"`
	Data       []struct {
		ID        int    `json:"id"`
		Amount    string `json:"amount"`
		Rate      string `json:"rate"`
		Pair      string `json:"pair"`
		OrderType string `json:"order_type"`
		CreatedAt string `json:"created_at"`
	} `json:"data"`
}

func GetRecentTrades(exchangePair entity.ExchangePair) entity.TradeCollection {
	code := getExchangePairCode(exchangePair)
	if code == NO_DEAL {
		fmt.Println("Error: No deal")
		return []entity.Trade{}
	}

	query := "pair=" + string(code) + "&limit=100"
	resp, err := http.Get("https://coincheck.com/api/trades?" + query)
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

	var mappedResp TradesResponse
	err = json.Unmarshal(body, &mappedResp)
	if err != nil {
		fmt.Println("Error:", err)
		return []entity.Trade{}
	}

	var recentTrades entity.TradeCollection

	for _, trade := range mappedResp.Data {
		time, err := time.Parse(time.RFC3339, trade.CreatedAt)
		if err != nil {
			fmt.Println("Error:", err)
		}

		price, err := strconv.ParseFloat(trade.Rate, 64)
		if err != nil {
			fmt.Println("Error:", err)
		}

		volume, err := strconv.ParseFloat(trade.Amount, 64)
		if err != nil {
			fmt.Println("Error:", err)
		}

		recentTrades = append(recentTrades, entity.Trade{Price: price, Volume: volume, Time: time})
	}

	return recentTrades
}
