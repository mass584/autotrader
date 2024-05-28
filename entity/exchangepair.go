package entity

type ExchangePair int

// DBに永続化されるので順番を変えないこと
//
//go:generate enumer -type=ExchangePair
const (
	BTC_JPY ExchangePair = iota + 1
	ETH_JPY
	ETH_BTC
	ETC_JPY
	XRP_JPY
	BCH_BTC
	MONA_JPY
)
