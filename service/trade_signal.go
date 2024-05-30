package service

import (
	"errors"
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/repository/database"
	"gorm.io/gorm"
)

type Decision string

const (
	Sell Decision = "SELL"
	Buy  Decision = "BUY"
	Hold Decision = "HOLD"
)

var (
	ErrAggregationIsNotFinished = errors.New("Aggregation is not finished")
	ErrNoTradesInPeriod         = errors.New("No trades in the period")
)

// 指定した期間で集計対象期間を利用できる場合、集計結果を参照する
// 集計結果が欠落している場合はエラーを返す
func calculateSimpleMovingAverage(
	db *gorm.DB,
	exchangePlace entity.ExchangePlace,
	exchangePair entity.ExchangePair,
	signalAt time.Time, // 期間の右端
	term time.Duration, // 期間の長さ
) (float64, error) {
	fromDatetime := signalAt.Add(-1 * term)
	tmp := fromDatetime.Add(24 * time.Hour) // 左端の24時間後
	fromDate := time.Date(tmp.Year(), tmp.Month(), tmp.Day(), 0, 0, 0, 0, time.UTC)

	toDatetime := signalAt
	toDate := time.Date(toDatetime.Year(), toDatetime.Month(), toDatetime.Day(), 0, 0, 0, 0, time.UTC)

	var aggregations []entity.TradeAggregation
	var trades []entity.Trade
	if toDate.After(fromDate) { // 集計結果が参照可能な場合
		aggregations = database.GetTradeAggregationsByDateRange(db, exchangePlace, exchangePair, fromDate, toDate)
		// 集計はUTCの0時を境界とした1日単位で行われているので、左右の中途半端な領域はオンデマンドで集計しなおす
		tradesLeft := database.GetTradesByTimeRange(db, exchangePlace, exchangePair, fromDatetime, fromDate)
		tradesRight := database.GetTradesByTimeRange(db, exchangePlace, exchangePair, toDate, toDatetime)
		trades = append(tradesLeft, tradesRight...)
	} else { // 集計結果が参照不可能な場合
		trades = database.GetTradesByTimeRange(db, exchangePlace, exchangePair, fromDatetime, toDatetime)
	}

	// 集計済みかどうか確認
	days := int(toDate.Sub(fromDate).Hours()/24) + 1
	if days < 0 {
		days = 0
	}

	if days != len(aggregations) {
		return 0, ErrAggregationIsNotFinished
	}

	var totalTransaction float64
	var totalCount int

	for _, aggregation := range aggregations {
		totalTransaction += aggregation.TotalTransaction
		totalCount += aggregation.TotalCount
	}

	for _, trade := range trades {
		totalTransaction += trade.Price * trade.Volume
		totalCount += 1
	}

	if totalCount == 0 {
		return 0, ErrNoTradesInPeriod
	}

	return totalTransaction / float64(totalCount), nil
}

// 一般的なパラメータとして、短期移動平均と長期移動平均の期間を10日と50日とする
func trendFollowingSignal(
	db *gorm.DB,
	exchangePlace entity.ExchangePlace,
	exchangePair entity.ExchangePair,
	signalAt time.Time,
) (Decision, error) {
	// 過去10日分の取引データを取得する
	shortSMA, err := calculateSimpleMovingAverage(db, exchangePlace, exchangePair, signalAt, 10*24*time.Hour)
	if err != nil {
		return Hold, err
	}
	// 過去50日分の取引データを取得する
	longSMA, err := calculateSimpleMovingAverage(db, exchangePlace, exchangePair, signalAt, 50*24*time.Hour)
	if err != nil {
		return Hold, err
	}

	if shortSMA > longSMA {
		return Buy, nil
	} else if shortSMA < longSMA {
		return Sell, nil
	}
	return Hold, nil
}

func TestTrendFollowingSignal(db *gorm.DB, exchangePlace entity.ExchangePlace, exchangePair entity.ExchangePair, signalAt time.Time) (Decision, error) {
	return trendFollowingSignal(db, exchangePlace, exchangePair, signalAt)
}

// どれくらいの期間での単純移動平均を取るかのパラメータチューニングが必要
func meanReversionSignal(
	db *gorm.DB,
	exchangePlace entity.ExchangePlace,
	exchangePair entity.ExchangePair,
	signalAt time.Time,
) (Decision, error) {
	sma, err := calculateSimpleMovingAverage(db, exchangePlace, exchangePair, signalAt, 10*time.Minute)
	if err != nil {
		return Hold, err
	}

	trade, error := database.GetTradeByLatestBefore(db, exchangePlace, exchangePair, signalAt)
	if error != nil {
		return Hold, err
	}

	currentPrice := trade.Price

	if currentPrice < sma {
		return Buy, nil
	} else if currentPrice > sma {
		return Sell, nil
	}
	return Hold, nil
}

func TestMeanReversionSignal(db *gorm.DB, exchangePlace entity.ExchangePlace, exchangePair entity.ExchangePair, signalAt time.Time) (Decision, error) {
	return meanReversionSignal(db, exchangePlace, exchangePair, signalAt)
}
