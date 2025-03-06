package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/lavatee/ai_hr/internal/endpoint"
	"github.com/lavatee/ai_hr/internal/repository"
	"github.com/lavatee/ai_hr/internal/service"
	_ "github.com/lib/pq"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{PrettyPrint: true})
	bot, err := telego.NewBot("8073581597:AAEgvEFmgH7nIaT8kHSc6bj1r59rSEupUck", telego.WithDefaultDebugLogger())
	if err != nil {
		logrus.Fatalf("telegram bot error: %s", err.Error())
	}
	if err := InitConfig(); err != nil {
		logrus.Fatalf("config error: %s", err.Error())
	}
	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("env error: %s", err.Error())
	}
	db, err := repository.NewPostgresDB(viper.GetString("db.host"), viper.GetString("db.port"), viper.GetString("db.user"), os.Getenv("DB_PASSWORD"), viper.GetString("db.dbname"), viper.GetString("db.sslmode"))
	if err != nil {
		logrus.Fatalf("db error: %s", err.Error())
	}
	repo := repository.NewRepository(db)
	services := service.NewService(repo, viper.GetString("aiApi.model"), viper.GetString("aiApi.address"), viper.GetString("aiApi.key"))
	endp := endpoint.NewEndpoint(services)
	updates, err := bot.UpdatesViaLongPolling(nil)
	if err != nil {
		logrus.Fatalf("updates error: %s", err.Error())
	}
	bh, err := th.NewBotHandler(bot, updates)
	if err != nil {
		logrus.Fatalf("bot handler error: %s", err.Error())
	}
	defer bh.Stop()
	defer bot.StopLongPolling()

	bh.Handle(endp.HandleStart, th.CommandEqual("start"))
	bh.Handle(endp.HandleAnyMessage, th.AnyMessage())
	bh.HandleCallbackQuery(endp.HandleCallback)
	bh.Start()
}

func InitConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
