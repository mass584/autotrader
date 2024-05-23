package entity

type ExchangePair int

// DBに永続化されるので順番を変えないこと
const (
	BTC_TO_JPY ExchangePair = iota + 1
	ETH_TO_JPY
	ETH_TO_BTC
	ETC_TO_JPY
	XRP_TO_JPY
	BCH_TO_BTC
)
