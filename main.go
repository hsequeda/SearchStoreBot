package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	bot *tgbotapi.BotAPI
)

func init() {
	var err error

	bot, err = tgbotapi.NewBotAPI("1258072778:AAG86IhHoDKRG-aKQgnEqoOJPLr3Migiuto")
	if err != nil {
		logrus.Error(err)
	}

	bot.Debug = true

	_, err = bot.SetWebhook(tgbotapi.NewWebhook("https://searchstorebot.herokuapp.com/"))
	if err != nil {
		logrus.Error(err)
	}

	info, err := bot.GetWebhookInfo()
	if err != nil {
		logrus.Error(err)
	}

	if info.LastErrorDate != 0 {
		logrus.Printf("[Telegram callback failed]%s", info.LastErrorMessage)
	}

	if err := InitDb(); err != nil {
		logrus.Error(err)
	}
}

func main() {
	logrus.Info("starting bot")
	updates := bot.ListenForWebhook("/InversorTelegramBot/")

	for update := range updates {
		if update.Message != nil {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Este bot no recive mensaje 😠"))
			continue
		}
		if update.InlineQuery != nil {
			if len(update.InlineQuery.Query) >= 4 {
				results, err := GetResultList(update.InlineQuery)
				if err != nil {
					continue
				}
				resp, err := bot.AnswerInlineQuery(tgbotapi.InlineConfig{
					InlineQueryID: update.InlineQuery.ID,
					Results:       results,
				})
				if err != nil {
					continue
				}
				logrus.Info(resp)

			}
		}

	}
}

func GetResultList(inlineQuery *tgbotapi.InlineQuery) ([]interface{}, error) {
	var resultList = make([]interface{}, 0)
	storeList, err := data.GetWhenMatchWithRawData(inlineQuery.Query)
	if err != nil {
		return nil, err
	}

	for i := range storeList {
		msgText := fmt.Sprintf(
			`
			Municipio: %s,
			Reparto: %s,
			Telefono: %s,
			Horario: ( %s - %s ),
			Direccion: %s,
			Localizacion: ( %f, %f ),
			Ubicacion: <a href="%s"></a>.
			`,
			storeList[i].Municipality, storeList[i].Department, storeList[i].Phone, storeList[i].Open, storeList[i].Close,
			storeList[i].Address, storeList[i].Geolocation.Latitude, storeList[i].Geolocation.Latitude, storeList[i].MapUrl)

		inlineQueryResult := tgbotapi.NewInlineQueryResultArticleHTML(uuid.New().String(), storeList[i].Name, msgText)

		resultList = append(resultList, inlineQueryResult)
	}
	return resultList, err
}
