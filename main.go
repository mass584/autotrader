package main

import (
	"flag"
	"io"
	"os"

	"github.com/mass584/autotrader/config"
	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
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
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = zerolog.New(multiWriter).With().Timestamp().Logger()

	modePtr := flag.String("mode", "scraping", "実行モード")
	placePtr := flag.String("place", "Bitflyer", "取引ペア")
	pairPtr := flag.String("pair", "BTC_JPY", "取引ペア")
	flag.Parse()

	place, err := entity.ExchangePlaceString(*placePtr)
	if err != nil {
		log.Error().Caller().Err(err).Send()
		os.Exit(1)
	}

	pair, err := entity.ExchangePairString(*pairPtr)
	if err != nil {
		log.Error().Caller().Err(err).Send()
		os.Exit(1)
	}

	config, err := config.NewConfig()
	if err != nil {
		log.Error().Caller().Err(err).Send()
		os.Exit(1)
	}

	db, err := gorm.Open(mysql.Open(config.DatabaseURL()), &gorm.Config{
		// 一旦サイレントにする。本当はzerologを渡したいがインターフェイスが合わなかった。
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Error().Caller().Err(err).Send()
		os.Exit(1)
	}

	switch *modePtr {
	case "scraping":
		service.ScrapingTrades(db, place, pair)
	case "aggregation":
		err := service.AggregationAll(db, place, pair)
		if err != nil {
			log.Error().Stack().Err(err).Send()
			os.Exit(1)
		}
	case "watch":
		service.WatchPostion(db, place, pair)
	case "watch_simulation":
		service.WatchPostionSimulation(db, place, pair)
	default:
		log.Error().Msg("Invalid execution mode.")
		os.Exit(1)
	}

	os.Exit(0)
}
