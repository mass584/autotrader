package main

import (
	"flag"
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
	"gorm.io/gorm/logger"
)

func main() {
	logfile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal().Msgf("%v", err)
		os.Exit(1)
	}
	defer logfile.Close()

	multiWriter := io.MultiWriter(logfile, os.Stdout)
	log.Logger = zerolog.New(multiWriter).With().Timestamp().Logger()

	modePtr := flag.String("mode", "", "実行モード")
	pairPtr := flag.String("pair", "BTC_JPY", "取引ペア")
	flag.Parse()

	pair, err := entity.ExchangePairString(*pairPtr)
	if err != nil {
		log.Fatal().Msgf("%v", err)
		os.Exit(1)
	}

	config, err := config.NewConfig()
	if err != nil {
		log.Fatal().Msg("Invalid config.")
		os.Exit(1)
	}

	db, err := gorm.Open(mysql.Open(config.DatabaseURL()), &gorm.Config{
		// 一旦サイレントにする。本当はzerologを渡したいがインターフェイスが合わなかった。
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal().Msg("Failed to connect database.")
		os.Exit(1)
	}

	switch *modePtr {
	case "scraping":
		service.ScrapingTradesFromCoincheck(db, pair)
	case "aggregation":
		service.AggregationAll(db, entity.Coincheck, pair)
	case "watch":
		service.WatchPostionOnCoincheck(db)
	case "optimize":
		at := time.Date(2023, 10, 1, 10, 0, 0, 0, time.Local)
		service.WatchPostionOnCoincheckForOptimize(db, at)
	default:
		log.Fatal().Msg("Invalid execution mode.")
		os.Exit(1)
	}

	os.Exit(0)
}
