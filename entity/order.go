package entity

type Order struct {
	Price  float64
	Volume float64
}

type OrderBook struct {
	Asks []Order
	Bids []Order
}
