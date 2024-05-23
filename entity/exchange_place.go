package entity

type ExchangePlace int

// DBに永続化されるので順番を変えないこと
const (
	Bitflyer ExchangePlace = iota + 1
	Coincheck
)
