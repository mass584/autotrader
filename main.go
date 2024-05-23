package main

import (
	"io"
	"os"

	"github.com/mass584/autotrader/config"
	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	logfile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal().Msg("Failed to open log file.")
		os.Exit(1)
	}
	defer logfile.Close()

	multiWriter := io.MultiWriter(logfile, os.Stdout)
	log.Logger = zerolog.New(multiWriter).With().Timestamp().Logger()

	args := os.Args
	if len(args) != 2 {
		log.Fatal().Msg("Invalid argument.")
		os.Exit(1)
	}

	config, err := config.NewConfig()
	if err != nil {
		log.Fatal().Msg("Invalid config.")
		os.Exit(1)
	}

	db, err := gorm.Open(mysql.Open(config.DatabaseURL()))
	if err != nil {
		log.Fatal().Msg("Failed to connect database.")
		os.Exit(1)
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
		log.Fatal().Msg("Invalid execution mode.")
		os.Exit(1)
	}

	os.Exit(0)
}
