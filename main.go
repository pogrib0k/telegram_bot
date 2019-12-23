package main

import (
	"bytes"
	"image/png"
	"log"
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/image/webp"
)

// Check делает check, ну шо го ты доволен?! А?!
func Check(path string) bool {
	if len(path) == 0 {
		return false
	}
	return strings.HasSuffix(path, ".webp")
}

func CheckText(str string) int {

	str = strings.ToLower(str)
	if strings.Contains(str, "ping") {
		return 1
	}
	return 0
}

func main() {
	bot, err := tgbotapi.NewBotAPI("934013449:AAGU3PstPF0F5_HWIFkg3OWi1Ao4uKGguzY")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

		if CheckText(update.Message.Text) == 1 {

			msg = tgbotapi.NewMessage(update.Message.Chat.ID, "pong!")
			bot.Send(msg)
			continue
		}

		msg.ReplyToMessageID = update.Message.MessageID

		if update.Message.Sticker != nil {
			//fwd := tgbotapi.NewForward(update.Message.Chat.ID, update.Message.Chat.ID, update.Message.MessageID)

			FileID := update.Message.Sticker.FileID

			//log.Println(FileID)

			var fc tgbotapi.FileConfig

			fc.FileID = update.Message.Sticker.FileID

			pathURL, err := bot.GetFileDirectURL(FileID)
			log.Println(pathURL)
			resp, err := http.Get(pathURL)

			if err != nil {
				log.Println(err.Error())
				continue
			}
			if !Check(pathURL) {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Ты ебанат, порядок не тот!!!")
				bot.Send(msg)
				continue
			}
			defer resp.Body.Close()

			img0, err := webp.Decode(resp.Body)
			if err != nil {
				log.Println(err)
				continue
			}

			bufPNG := new(bytes.Buffer)
			err = png.Encode(bufPNG, img0)
			if err != nil {
				log.Println(err)
				continue
			}

			photoUpload := tgbotapi.NewPhotoUpload(update.Message.Chat.ID, tgbotapi.FileBytes{Name: "ban.png", Bytes: bufPNG.Bytes()})

			_, err = bot.Send(photoUpload)

			if err != nil {
				log.Println(err.Error())
			}

		}

		bot.Send(msg)
	}
}
