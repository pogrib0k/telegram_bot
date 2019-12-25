package main

import (
	"bytes"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-pg/pg"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"golang.org/x/image/webp"
)

type dbLogger struct{}

var db *pg.DB

func (d dbLogger) BeforeQuery(q *pg.QueryEvent) {
}

func (d dbLogger) AfterQuery(q *pg.QueryEvent) {
	fmt.Println(q.FormattedQuery())
}

// User данные про пользователя для запроса в БД
type User struct {
	ID       int64
	Name     string
	Email    string
	Password string
}

func (u User) String() string {
	return fmt.Sprintf("User<%d %s %s %s>", u.ID, u.Name, u.Email, u.Password)
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	db = pg.Connect(&pg.Options{
		User:     os.Getenv("PSQL_LOGIN"),
		Password: os.Getenv("PSQL_PASS"),
		Database: os.Getenv("PSQL_DB"),
		Addr:     os.Getenv("PSQL_HOST"),
		OnConnect: func(conn *pg.Conn) error {
			_, err := conn.Exec("set search_path=?", os.Getenv("PSQL_SCHEMA"))
			if err != nil {
				panic(err.Error())
			}
			return nil
		},
	})
	db.AddQueryHook(dbLogger{})
	defer db.Close()

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TG_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	time.Sleep(time.Second)
	updates.Clear()

	go SendAllChattables(bot)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.IsCommand() {
			HandleCommand(update, bot)
		}


		if update.Message.Sticker != nil {
			//fwd := tgbotapi.NewForward(update.Message.Chat.ID, update.Message.Chat.ID, update.Message.MessageID)

			FileID := update.Message.Sticker.FileID

			//log.Println(FileID)

			var fc tgbotapi.FileConfig

			fc.FileID = update.Message.Sticker.FileID
			log.Println(update.Message.Sticker.FileID)

			pathURL, err := bot.GetFileDirectURL(FileID)
			log.Println(pathURL)
			resp, err := http.Get(pathURL)

			if err != nil {
				log.Println(err.Error())
				continue
			}
			if !Check(pathURL) {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ты ебанат, порядок не тот!!!")
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
	}
}
