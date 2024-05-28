package entity

type ExchangePlace int

// DBに永続化されるので順番を変えないこと
//
//go:generate enumer -type=ExchangePlace
const (
	Bitflyer ExchangePlace = iota + 1
	Coincheck
)
