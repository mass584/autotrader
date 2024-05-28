package main

import (
	"io"
	"os"
	"time"

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
	if len(args) < 2 {
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
		pair, err := entity.ExchangePairString(args[2])
		if err != nil {
			log.Fatal().Msgf("%v", err)
			os.Exit(1)
		}
		service.ScrapingTradesFromCoincheck(db, pair)
	case "order_price":
		service.DetermineOrderPrice()
	case "trade_signal":
		at := time.Date(2023, 3, 1, 10, 0, 0, 0, time.Local)
		service.CalculateTradeSignalOnCoincheck(db, entity.BTC_JPY, at)
	case "watch":
		service.WatchPostionOnCoincheck(db)
	default:
		log.Fatal().Msg("Invalid execution mode.")
		os.Exit(1)
	}

	os.Exit(0)
}
