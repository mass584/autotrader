package coincheck

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/mass584/autotrader/entity"
)

type ExchangePairCode string

const (
	BTC_JPY  ExchangePairCode = "btc_jpy"
	ETC_JPY  ExchangePairCode = "etc_jpy"
	MONA_JPY ExchangePairCode = "mona_jpy"
	NO_DEAL  ExchangePairCode = ""
)

func GetExchangePairCode(exchangePair entity.ExchangePair) ExchangePairCode {
	switch exchangePair {
	case entity.BTC_JPY:
		return BTC_JPY
	case entity.ETC_JPY:
		return ETC_JPY
	case entity.MONA_JPY:
		return MONA_JPY
	default:
		return NO_DEAL
	}
}

func GetOrderBook(exchangePair entity.ExchangePair) entity.OrderBook {
	code := GetExchangePairCode(exchangePair)
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
	if resp.StatusCode != http.StatusOK {
		// レートリミットに引っかかると403が返ってくる
		fmt.Println("Error:", err)
		return entity.OrderBook{}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return entity.OrderBook{}
	}

	var mappedResp struct {
		Asks [][]string `json:"asks"`
		Bids [][]string `json:"bids"`
	}
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

type OrderType string

const (
	BUY  OrderType = "buy"
	SELL OrderType = "sell"
)

func GetRecentTrades(exchangePair entity.ExchangePair) entity.TradeCollection {
	code := GetExchangePairCode(exchangePair)
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
	if resp.StatusCode != http.StatusOK {
		// レートリミットに引っかかると403が返ってくる
		fmt.Println("Error:", err)
		return entity.TradeCollection{}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return []entity.Trade{}
	}

	var mappedResp struct {
		Success    bool `json:"success"`
		Pagination struct {
			Limit         int   `json:"limit"`
			Order         Order `json:"order"`
			StartingAfter *int  `json:"starting_after"`
			EndingBefore  *int  `json:"ending_before"`
		} `json:"pagination"`
		Data []struct {
			ID        int       `json:"id"`
			Amount    string    `json:"amount"`
			Rate      string    `json:"rate"`
			Pair      string    `json:"pair"`
			OrderType OrderType `json:"order_type"`
			CreatedAt string    `json:"created_at"`
		} `json:"data"`
	}
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

		recentTrades = append(
			recentTrades,
			entity.Trade{
				ExchangePlace: entity.Coincheck,
				ExchangePair:  exchangePair,
				TradeID:       trade.ID,
				Price:         price,
				Volume:        volume,
				Time:          time,
			},
		)
	}

	return recentTrades
}

type AllTrades struct {
	ID    int
	Trade entity.Trade
}

func GetAllTradesByLastId(exchangePair entity.ExchangePair, lastId int) entity.TradeCollection {
	code := GetExchangePairCode(exchangePair)
	if code == NO_DEAL {
		fmt.Println("Error: No deal")
		return entity.TradeCollection{}
	}

	query := "pair=" + string(code) + "&last_id=" + strconv.Itoa(lastId+1)
	resp, err := http.Get("https://coincheck.com/ja/exchange/orders/completes?" + query)
	if err != nil {
		fmt.Println("Error:", err)
		return entity.TradeCollection{}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// レートリミットに引っかかると403が返ってくる
		fmt.Println("Error:", err)
		return entity.TradeCollection{}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return entity.TradeCollection{}
	}

	var mappedResp struct {
		Completes []struct {
			ID        int    `json:"id"`
			Amount    string `json:"amount"`
			Rate      string `json:"rate"`
			OrderType string `json:"order_type"`
			CreatedAt string `json:"created_at"`
		} `json:"completes"`
	}

	err = json.Unmarshal(body, &mappedResp)
	if err != nil {
		fmt.Println("Error:", err)
		return entity.TradeCollection{}
	}

	var trades entity.TradeCollection

	for _, complete := range mappedResp.Completes {

		time, err := time.Parse(time.RFC3339, complete.CreatedAt)
		if err != nil {
			fmt.Println("Error:", err)
		}

		price, err := strconv.ParseFloat(complete.Rate, 64)
		if err != nil {
			fmt.Println("Error:", err)
		}

		volume, err := strconv.ParseFloat(complete.Amount, 64)
		if err != nil {
			fmt.Println("Error:", err)
		}

		trades = append(
			trades, entity.Trade{
				ExchangePlace: entity.Coincheck,
				ExchangePair:  exchangePair,
				TradeID:       complete.ID,
				Price:         price,
				Volume:        volume,
				Time:          time,
			},
		)
	}

	return trades
}
