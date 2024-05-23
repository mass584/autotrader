package main

import (
	"fmt"
	"os"

	"github.com/mass584/autotrader/config"
	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/service"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	args := os.Args
	if len(args) != 2 {
		fmt.Printf("Invalid argument.")
		return
	}

	config, err := config.NewConfig()
	if err != nil {
		fmt.Printf("Invalid config.")
		return
	}

	db, err := gorm.Open(mysql.Open(config.DatabaseURL()))
	if err != nil {
		fmt.Printf("Failed to connect database.")
		return
	}

	mode := args[1]
	switch mode {
	case "scraping":
		service.ScrapingTradesFromCoincheck(db, entity.BTC_TO_JPY)
	case "order_price":
		service.DetermineOrderPrice()
	case "trade_signal":
		service.CalculateTradeSignal()
	default:
		fmt.Printf("Invalid execution mode.")
	}
}
