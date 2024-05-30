package service_test

import (
	"testing"
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/helper"
	"github.com/mass584/autotrader/service"
	"github.com/pkg/errors"
)

func TestTrendFollowingSignal(t *testing.T) {
	type args struct {
		signalAt        time.Time
		tradeCollection entity.TradeCollection
	}

	type want struct {
		value service.Decision
		error error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "短期移動平均が長期移動平均を下回った時にトレンドフォローが売りシグナルを指し示すこと",
			args: args{
				signalAt: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC),
				tradeCollection: helper.BuildTradeCollectionHelper(
					helper.Trades{
						{
							Price:  1.0,
							Volume: 1.0,
							Time:   time.Date(2024, 5, 22, 12, 0, 0, 0, time.UTC), // 10日前
						},
						{
							Price:  2.0,
							Volume: 1.0,
							Time:   time.Date(2024, 4, 12, 12, 0, 0, 0, time.UTC), // 50日前
						},
					},
				),
			},
			want: want{
				value: service.Sell,
				error: nil,
			},
		},
		{
			name: "短期移動平均が長期移動平均を下回った時にトレンドフォローが売りシグナルを指し示すこと 境界値ケース",
			args: args{
				signalAt: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC),
				tradeCollection: helper.BuildTradeCollectionHelper(
					helper.Trades{
						{
							Price:  1.0,
							Volume: 1.0,
							Time:   time.Date(2024, 5, 22, 12, 0, 0, 0, time.UTC), // 10日前
						},
						{
							Price:  2.0,
							Volume: 1.0,
							Time:   time.Date(2024, 5, 21, 12, 0, 0, 0, time.UTC), // 11日前
						},
						{
							Price:  2.0,
							Volume: 1.0,
							Time:   time.Date(2024, 4, 12, 12, 0, 0, 0, time.UTC), // 50日前
						},
						{
							Price:  1.0,
							Volume: 1.0,
							Time:   time.Date(2024, 4, 11, 12, 0, 0, 0, time.UTC), // 51日前
						},
					},
				),
			},
			want: want{
				value: service.Sell,
				error: nil,
			},
		},
		{
			name: "短期移動平均が長期移動平均を上回った時にトレンドフォローが買いシグナルを指し示すこと",
			args: args{
				signalAt: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC),
				tradeCollection: helper.BuildTradeCollectionHelper(
					helper.Trades{
						{
							Price:  2.0,
							Volume: 1.0,
							Time:   time.Date(2024, 5, 22, 12, 0, 0, 0, time.UTC), // 10日前
						},
						{
							Price:  1.0,
							Volume: 1.0,
							Time:   time.Date(2024, 4, 12, 12, 0, 0, 0, time.UTC), // 50日前
						},
					},
				),
			},
			want: want{
				value: service.Buy,
				error: nil,
			},
		},
		{
			name: "短期移動平均が長期移動平均を上回った時にトレンドフォローが買いシグナルを指し示すこと 境界値ケース",
			args: args{
				signalAt: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC),
				tradeCollection: helper.BuildTradeCollectionHelper(
					helper.Trades{
						{
							Price:  2.0,
							Volume: 1.0,
							Time:   time.Date(2024, 5, 22, 12, 0, 0, 0, time.UTC), // 10日前
						},
						{
							Price:  1.0,
							Volume: 1.0,
							Time:   time.Date(2024, 5, 21, 12, 0, 0, 0, time.UTC), // 11日前
						},
						{
							Price:  1.0,
							Volume: 1.0,
							Time:   time.Date(2024, 4, 12, 12, 0, 0, 0, time.UTC), // 50日前
						},
						{
							Price:  2.0,
							Volume: 1.0,
							Time:   time.Date(2024, 4, 11, 12, 0, 0, 0, time.UTC), // 51日前
						},
					},
				),
			},
			want: want{
				value: service.Buy,
				error: nil,
			},
		},
		{
			name: "短期移動平均の計算対象となる取引が存在しない場合はホールドシグナルを指し示すこと",
			args: args{
				signalAt: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC),
				tradeCollection: helper.BuildTradeCollectionHelper(
					helper.Trades{
						{
							Price:  1.0,
							Volume: 1.0,
							Time:   time.Date(2024, 5, 21, 12, 0, 0, 0, time.UTC), // 11日前
						},
					},
				),
			},
			want: want{
				value: service.Hold,
				error: service.ErrNoTradesInPeriod,
			},
		},
		{
			name: "長期移動平均の計算対象となる取引が存在しない場合はホールドシグナルを指し示すこと",
			args: args{
				signalAt: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC),
				tradeCollection: helper.BuildTradeCollectionHelper(
					helper.Trades{
						{
							Price:  1.0,
							Volume: 1.0,
							Time:   time.Date(2024, 4, 11, 12, 0, 0, 0, time.UTC), // 51日前
						},
					},
				),
			},
			want: want{
				value: service.Hold,
				error: service.ErrNoTradesInPeriod,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テストデータの保存
			helper.InsertTradeCollectionHelper(db, tt.args.tradeCollection)
			// テストデータの集計
			fromTime := tt.args.tradeCollection[len(tt.args.tradeCollection)-1].Time
			fromDate := time.Date(fromTime.Year(), fromTime.Month(), fromTime.Day(), 0, 0, 0, 0, time.UTC)
			toTime := tt.args.signalAt
			toDate := time.Date(toTime.Year(), toTime.Month(), toTime.Day(), 0, 0, 0, 0, time.UTC)
			helper.AggregateHelper(db, entity.Coincheck, entity.BTC_JPY, fromDate, toDate)
			defer func() {
				helper.DatabaseCleaner(db)
			}()

			result, err := service.TestTrendFollowingSignal(db, entity.Coincheck, entity.BTC_JPY, tt.args.signalAt)
			if result != tt.want.value {
				t.Errorf("result = %v, want = %v", result, tt.want.value)
			}
			if !errors.Is(err, tt.want.error) {
				t.Errorf("result = %v, want = %v", err, tt.want.error)
			}
		})
	}
}

func TestMeanReversionSignal(t *testing.T) {
	type args struct {
		signalAt        time.Time
		tradeCollection entity.TradeCollection
	}

	type want struct {
		value service.Decision
		error error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "現在価格が移動平均を下回った時にミーンリバージョンが買いシグナルを指し示すこと",
			args: args{
				signalAt: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC),
				tradeCollection: helper.BuildTradeCollectionHelper(
					helper.Trades{
						{
							Price:  1.0,
							Volume: 1.0,
							Time:   time.Date(2024, 6, 1, 9, 55, 0, 0, time.UTC),
						},
						{
							Price:  2.0,
							Volume: 1.0,
							Time:   time.Date(2024, 6, 1, 9, 53, 0, 0, time.UTC),
						},
					},
				),
			},
			want: want{
				value: service.Buy,
				error: nil,
			},
		},
		{
			name: "現在価格が移動平均を上回った時にミーンリバージョンが売りシグナルを指し示すこと",
			args: args{
				signalAt: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC),
				tradeCollection: helper.BuildTradeCollectionHelper(
					helper.Trades{
						{
							Price:  2.0,
							Volume: 1.0,
							Time:   time.Date(2024, 6, 1, 9, 55, 0, 0, time.UTC), // 10日前
						},
						{
							Price:  1.0,
							Volume: 1.0,
							Time:   time.Date(2024, 6, 1, 9, 53, 0, 0, time.UTC), // 50日前
						},
					},
				),
			},
			want: want{
				value: service.Sell,
				error: nil,
			},
		},
		{
			name: "移動平均の計算対象となる取引が存在しない場合はホールドシグナルを指し示すこと",
			args: args{
				signalAt: time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC),
				tradeCollection: helper.BuildTradeCollectionHelper(
					helper.Trades{
						{
							Price:  1.0,
							Volume: 1.0,
							Time:   time.Date(2024, 6, 1, 9, 30, 0, 0, time.UTC),
						},
					},
				),
			},
			want: want{
				value: service.Hold,
				error: service.ErrNoTradesInPeriod,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テストデータの保存
			helper.InsertTradeCollectionHelper(db, tt.args.tradeCollection)
			defer func() {
				helper.DatabaseCleaner(db)
			}()

			result, err := service.TestMeanReversionSignal(db, entity.Coincheck, entity.BTC_JPY, tt.args.signalAt)
			if result != tt.want.value {
				t.Errorf("result = %v, want = %v", result, tt.want.value)
			}
			if !errors.Is(err, tt.want.error) {
				t.Errorf("result = %v, want = %v", err, tt.want.error)
			}
		})
	}
}
