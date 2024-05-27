package service_test

import (
	"testing"
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/repository/database"
	"github.com/mass584/autotrader/service"
)

type Trades []struct {
	Price  float64
	Volume float64
	Time   time.Time
}

func buildTradeCollection(trades Trades) entity.TradeCollection {
	baseTradeID := int(time.Now().UnixMilli())

	var tradeCollection entity.TradeCollection

	for idx, trade := range trades {
		tradeCollection = append(tradeCollection,
			entity.Trade{
				ExchangePlace: entity.Coincheck,
				ExchangePair:  entity.BTC_TO_JPY,
				TradeID:       baseTradeID + idx,
				Price:         trade.Price,
				Volume:        trade.Volume,
				Time:          trade.Time,
			},
		)
	}

	return tradeCollection
}

func TestCalculateTradeSignalOnCoincheck_TrendFollow(t *testing.T) {
	type args struct {
		signalAt        time.Time
		tradeCollection entity.TradeCollection
	}

	tests := []struct {
		name string
		args args
		want service.Decision
	}{
		{
			name: "短期移動平均が長期移動平均を下回った時にトレンドフォローが売りシグナルを指し示すこと",
			args: args{
				signalAt: time.Date(2024, 6, 1, 10, 0, 0, 0, time.Local),
				tradeCollection: buildTradeCollection(
					Trades{
						{
							Price:  1.0,
							Volume: 1.0,
							Time:   time.Date(2024, 5, 22, 12, 0, 0, 0, time.Local), // 10日前
						},
						{
							Price:  2.0,
							Volume: 1.0,
							Time:   time.Date(2024, 4, 12, 12, 0, 0, 0, time.Local), // 50日前
						},
					},
				),
			},
			want: service.Sell,
		},
		{
			name: "短期移動平均が長期移動平均を下回った時にトレンドフォローが売りシグナルを指し示すこと 境界値ケース",
			args: args{
				signalAt: time.Date(2024, 6, 1, 10, 0, 0, 0, time.Local),
				tradeCollection: buildTradeCollection(
					Trades{
						{
							Price:  1.0,
							Volume: 1.0,
							Time:   time.Date(2024, 5, 22, 12, 0, 0, 0, time.Local), // 10日前
						},
						{
							Price:  2.0,
							Volume: 1.0,
							Time:   time.Date(2024, 5, 21, 12, 0, 0, 0, time.Local), // 11日前
						},
						{
							Price:  2.0,
							Volume: 1.0,
							Time:   time.Date(2024, 4, 12, 12, 0, 0, 0, time.Local), // 50日前
						},
						{
							Price:  1.0,
							Volume: 1.0,
							Time:   time.Date(2024, 4, 11, 12, 0, 0, 0, time.Local), // 51日前
						},
					},
				),
			},
			want: service.Sell,
		},
		{
			name: "短期移動平均が長期移動平均を上回った時にトレンドフォローが買いシグナルを指し示すこと",
			args: args{
				signalAt: time.Date(2024, 6, 1, 10, 0, 0, 0, time.Local),
				tradeCollection: buildTradeCollection(
					Trades{
						{
							Price:  2.0,
							Volume: 1.0,
							Time:   time.Date(2024, 5, 22, 12, 0, 0, 0, time.Local), // 10日前
						},
						{
							Price:  1.0,
							Volume: 1.0,
							Time:   time.Date(2024, 4, 12, 12, 0, 0, 0, time.Local), // 50日前
						},
					},
				),
			},
			want: service.Buy,
		},
		{
			name: "短期移動平均が長期移動平均を上回った時にトレンドフォローが買いシグナルを指し示すこと 境界値ケース",
			args: args{
				signalAt: time.Date(2024, 6, 1, 10, 0, 0, 0, time.Local),
				tradeCollection: buildTradeCollection(
					Trades{
						{
							Price:  2.0,
							Volume: 1.0,
							Time:   time.Date(2024, 5, 22, 12, 0, 0, 0, time.Local), // 10日前
						},
						{
							Price:  1.0,
							Volume: 1.0,
							Time:   time.Date(2024, 5, 21, 12, 0, 0, 0, time.Local), // 11日前
						},
						{
							Price:  1.0,
							Volume: 1.0,
							Time:   time.Date(2024, 4, 12, 12, 0, 0, 0, time.Local), // 50日前
						},
						{
							Price:  2.0,
							Volume: 1.0,
							Time:   time.Date(2024, 4, 11, 12, 0, 0, 0, time.Local), // 51日前
						},
					},
				),
			},
			want: service.Buy,
		},
		{
			name: "短期移動平均の計算対象となる取引が存在しない場合はホールドシグナルを指し示すこと",
			args: args{
				signalAt: time.Date(2024, 6, 1, 10, 0, 0, 0, time.Local),
				tradeCollection: buildTradeCollection(
					Trades{
						{
							Price:  1.0,
							Volume: 1.0,
							Time:   time.Date(2024, 5, 21, 12, 0, 0, 0, time.Local), // 11日前
						},
					},
				),
			},
			want: service.Hold,
		},
		{
			name: "長期移動平均の計算対象となる取引が存在しない場合はホールドシグナルを指し示すこと",
			args: args{
				signalAt: time.Date(2024, 6, 1, 10, 0, 0, 0, time.Local),
				tradeCollection: buildTradeCollection(
					Trades{
						{
							Price:  1.0,
							Volume: 1.0,
							Time:   time.Date(2024, 4, 11, 12, 0, 0, 0, time.Local), // 51日前
						},
					},
				),
			},
			want: service.Hold,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テストデータの保存
			db.Where("1 = 1").Delete(&entity.Trade{})
			database.SaveTrades(db, tt.args.tradeCollection)
			defer func() {
				db.Where("1 = 1").Delete(&entity.Trade{})
			}()

			result, _ := service.CalculateTradeSignalOnCoincheck(db, entity.BTC_TO_JPY, tt.args.signalAt)
			if result != tt.want {
				t.Errorf("result = %v, want = %v", result, tt.want)
			}
		})
	}
}

func TestCalculateTradeSignalOnCoincheck_MeanReversion(t *testing.T) {
	type args struct {
		signalAt        time.Time
		tradeCollection entity.TradeCollection
	}

	tests := []struct {
		name string
		args args
		want service.Decision
	}{
		{
			name: "現在価格が移動平均を下回った時にミーンリバージョンが買いシグナルを指し示すこと",
			args: args{
				signalAt: time.Date(2024, 6, 1, 10, 0, 0, 0, time.Local),
				tradeCollection: buildTradeCollection(
					Trades{
						{
							Price:  1.0,
							Volume: 1.0,
							Time:   time.Date(2024, 6, 1, 9, 55, 0, 0, time.Local),
						},
						{
							Price:  2.0,
							Volume: 1.0,
							Time:   time.Date(2024, 6, 1, 9, 53, 0, 0, time.Local),
						},
					},
				),
			},
			want: service.Buy,
		},
		{
			name: "現在価格が移動平均を上回った時にミーンリバージョンが売りシグナルを指し示すこと",
			args: args{
				signalAt: time.Date(2024, 6, 1, 10, 0, 0, 0, time.Local),
				tradeCollection: buildTradeCollection(
					Trades{
						{
							Price:  2.0,
							Volume: 1.0,
							Time:   time.Date(2024, 6, 1, 9, 55, 0, 0, time.Local), // 10日前
						},
						{
							Price:  1.0,
							Volume: 1.0,
							Time:   time.Date(2024, 6, 1, 9, 53, 0, 0, time.Local), // 50日前
						},
					},
				),
			},
			want: service.Sell,
		},
		{
			name: "移動平均の計算対象となる取引が存在しない場合はホールドシグナルを指し示すこと",
			args: args{
				signalAt: time.Date(2024, 6, 1, 10, 0, 0, 0, time.Local),
				tradeCollection: buildTradeCollection(
					Trades{
						{
							Price:  1.0,
							Volume: 1.0,
							Time:   time.Date(2024, 6, 1, 9, 30, 0, 0, time.Local),
						},
					},
				),
			},
			want: service.Hold,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テストデータの保存
			db.Where("1 = 1").Delete(&entity.Trade{})
			database.SaveTrades(db, tt.args.tradeCollection)
			defer func() {
				db.Where("1 = 1").Delete(&entity.Trade{})
			}()

			_, result := service.CalculateTradeSignalOnCoincheck(db, entity.BTC_TO_JPY, tt.args.signalAt)
			if result != tt.want {
				t.Errorf("result = %v, want = %v", result, tt.want)
			}
		})
	}
}
