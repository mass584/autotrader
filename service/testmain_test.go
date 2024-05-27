package service_test

import (
	"io"
	"os"
	"testing"

	"github.com/mass584/autotrader/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func TestMain(m *testing.M) {
	// ログと標準出力の設定
	logfile, err := os.OpenFile("../log.test.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal().Msg("Failed to open log file.")
		os.Exit(1)
	}
	defer logfile.Close()

	multiWriter := io.MultiWriter(logfile, os.Stdout)
	log.Logger = zerolog.New(multiWriter).With().Timestamp().Logger().Level(zerolog.WarnLevel)

	// 環境変数の読み込み
	config, err := config.NewConfig()
	if err != nil {
		log.Fatal().Msg("Invalid config.")
		os.Exit(1)
	}
	config.DatabaseName = "autotrader_test"

	// データベースコネクションの作成
	db, err = gorm.Open(mysql.Open(config.DatabaseURL()))
	if err != nil {
		log.Fatal().Msg("Failed to connect database.")
		os.Exit(1)
	}

	m.Run()
}
