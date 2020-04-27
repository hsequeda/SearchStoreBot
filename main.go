package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strings"
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

	_, err = bot.SetWebhook(tgbotapi.NewWebhook("https://searchstorebot.herokuapp.com/" + bot.Token))
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
	port := os.Getenv("PORT")
	updates := bot.ListenForWebhook("/" + bot.Token)
	go http.ListenAndServe("0.0.0.0:"+port, nil)
	for update := range updates {
		switch {
		case update.Message != nil:
			_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Este bot no recibe mensajes 😡"))
			if err != nil {
				logrus.Warn(err)
			}
		case update.InlineQuery != nil:
			if len(update.InlineQuery.Query) >= 4 {
				logrus.Info(" QueryUpdate: ", update.InlineQuery)
				results, err := GetResultList(update.InlineQuery)
				if err != nil {
					continue
				}

				_, err = bot.AnswerInlineQuery(tgbotapi.InlineConfig{
					InlineQueryID: update.InlineQuery.ID,
					Results:       results,
				})
				if err != nil {
					continue
				}
			}
		}
	}
}

func GetResultList(inlineQuery *tgbotapi.InlineQuery) ([]interface{}, error) {
	var resultList = make([]interface{}, 0)
	replacer := strings.NewReplacer(" ", "", "\n", "", "\t", "", "\f", "", "\r", "", "!", "", "?", "", "#", "",
		"$", "", "%", "", "&", "", "'", "", "\"", "", "(", "", ")", "", "*", "", "+", "", ",", "", "-", "", ".", "", "/",
		"", ":", "", ";", "", "<", "", "=", "", ">", "", "@", "", "[", "", "^", "", "_", "", "`", "", "{", "", "|", "",
		"}", "", "~", "", "]", "", "\\", "")
	rawData := replacer.Replace(strings.ToLower(inlineQuery.Query))
	storeList, err := data.GetWhenMatchWithRawData(rawData)
	if err != nil {
		return nil, err
	}

	for i := range storeList {
		msgText := fmt.Sprintf(
			`
			<b>🏬Tienda :%s</b>
			Municipio : %s
			Reparto: %s
			☎️Telefono: %s
			Horario: ( %s - %s )
			Direccion: %s
			Localizacion: ( %f, %f )
			<a href="%s">🗺Ver en Mapa</a>.
			`,
			storeList[i].Name, storeList[i].Municipality, storeList[i].Department, storeList[i].Phone, storeList[i].Open,
			storeList[i].Close, storeList[i].Address, storeList[i].Geolocation.Latitude, storeList[i].Geolocation.Longitude,
			storeList[i].MapUrl)
		inlineQueryResult := tgbotapi.NewInlineQueryResultArticleHTML(uuid.New().String(), storeList[i].Name, msgText)
		resultList = append(resultList, inlineQueryResult)
	}
	return resultList, err
}
