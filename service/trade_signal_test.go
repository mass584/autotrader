package service_test

import (
	"testing"
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/helper"
	"github.com/mass584/autotrader/service"
)

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
				tradeCollection: helper.BuildTradeCollectionHelper(
					helper.Trades{
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
				tradeCollection: helper.BuildTradeCollectionHelper(
					helper.Trades{
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
				tradeCollection: helper.BuildTradeCollectionHelper(
					helper.Trades{
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
				tradeCollection: helper.BuildTradeCollectionHelper(
					helper.Trades{
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
				tradeCollection: helper.BuildTradeCollectionHelper(
					helper.Trades{
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
				tradeCollection: helper.BuildTradeCollectionHelper(
					helper.Trades{
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
			helper.InsertTradeCollectionHelper(db, tt.args.tradeCollection)
			// テストデータの集計
			fromTime := tt.args.tradeCollection[len(tt.args.tradeCollection)-1].Time
			fromDate := time.Date(fromTime.Year(), fromTime.Month(), fromTime.Day(), 0, 0, 0, 0, time.Local)
			toTime := tt.args.signalAt
			toDate := time.Date(toTime.Year(), toTime.Month(), toTime.Day(), 0, 0, 0, 0, time.Local)
			helper.AggregateHelper(db, entity.Coincheck, entity.BTC_JPY, fromDate, toDate)
			defer func() {
				helper.DatabaseCleaner(db)
			}()

			result, _ := service.TestCalculateTradeSignalOnCoincheck(db, entity.BTC_JPY, tt.args.signalAt)
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
				tradeCollection: helper.BuildTradeCollectionHelper(
					helper.Trades{
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
				tradeCollection: helper.BuildTradeCollectionHelper(
					helper.Trades{
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
				tradeCollection: helper.BuildTradeCollectionHelper(
					helper.Trades{
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
			helper.InsertTradeCollectionHelper(db, tt.args.tradeCollection)
			defer func() {
				helper.DatabaseCleaner(db)
			}()

			_, result := service.TestCalculateTradeSignalOnCoincheck(db, entity.BTC_JPY, tt.args.signalAt)
			if result != tt.want {
				t.Errorf("result = %v, want = %v", result, tt.want)
			}
		})
	}
}
