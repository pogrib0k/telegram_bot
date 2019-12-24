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

func (d dbLogger) BeforeQuery(q *pg.QueryEvent) {
}

func (d dbLogger) AfterQuery(q *pg.QueryEvent) {
	fmt.Println(q.FormattedQuery())
}

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

	db := pg.Connect(&pg.Options{
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

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

		if update.Message.IsCommand() {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			switch update.Message.Command() {
			case "ping":
				msg.Text = "pong"
			case "register":
				arg1, arg2 := CutTwoArguments(update.Message.CommandArguments())
				if arg1 == "" || arg2 == "" {
					msg.Text = `Неправильно введена команда.
Правильная форма: ` + "`/register <login> <password>`" + `.
Логин должен состоять из ` + "`A-Z, a-z, 0-9, _`" + `;
а пароль должен состоять из ` + "`A-Z, a-z, 0-9, _, *, -, #, $, %, !, @, ^`" + `.`
					msg.ParseMode = "Markdown"
					break
				}
				user1 := &User{
					Name:     arg1,
					Email:    "shit@gogle.net",
					Password: arg2,
				}
				err = db.Insert(user1)
				if err != nil {
					pgErr, ok := err.(pg.Error)
					if ok && pgErr.IntegrityViolation() {
						msg.Text = "А пользователь уже есть, для вас мест нет (ну по крайней мере под таким логином)."
					} else {
						msg.Text = "У вас тротлинг. Сочувствую."
					}
					break
				}
				msg.Text = fmt.Sprintf("Ваш логин - `%s`, ваш пароль - `%s`. Вас зарегистрировали, поэтому вы ЛОХ.", arg1, arg2)
				msg.ParseMode = "Markdown"
			case "login":
				arg1, arg2 := CutTwoArguments(update.Message.CommandArguments())
				if arg1 == "" || arg2 == "" {
					msg.Text = `Неправильно введена команда.
Правильная форма: ` + "`/login <login> <password>`" + `.
Логин должен состоять из ` + "`A-Z, a-z, 0-9, _`" + `;
а пароль должен состоять из ` + "`A-Z, a-z, 0-9, _, *, -, #, $, %, !, @, ^`" + `.`
					msg.ParseMode = "Markdown"
					break
				}
				user := new(User)
				err = db.Model(user).Where("name = ? AND password = ?", arg1, arg2).Select()
				if err != nil {
					if err == pg.ErrNoRows {
						audioMessage := tgbotapi.NewVoiceUpload(update.Message.Chat.ID, "tryAgain.mp3")
						audioMessage.Caption = "Неверный логин или пароль. Попробуй *ЕЩЁ РАЗ*."
						audioMessage.ParseMode = "Markdown"
						bot.Send(audioMessage)
					} else {
						msg.Text = "У вас тротлинг. Сочувствую."
					}
					break
				}
				log.Println(user)
				msg.Text = fmt.Sprintf("Вроде как всё правильно, поздравляю вы внутри.")
				msg.ParseMode = "Markdown"
			case "anime":
				msg.Text = fmt.Sprintf("Для пидоров. Вот к примеру таких как %s (@%s)", update.Message.From.FirstName, update.Message.From.UserName)
			case "soldat":
				msg.Text = "https://www.youtube.com/watch?v=POb02mjj2zE"
			case "bonk":
				msgSticker := tgbotapi.NewStickerShare(update.Message.Chat.ID, "CAADAgAD_QEAAvNWPxeLf-J5M600mhYE")
				bot.Send(msgSticker)
			default:
				msg.Text = "Шото ты хуйню ввёл. Попробуй ЕЩЁ раз!"
			}
			if len(msg.Text) == 0 {
				continue
			}
			_, err = bot.Send(msg)
			if err != nil {
				log.Println(err.Error())
			}
			continue
		}

		msg.ReplyToMessageID = update.Message.MessageID

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
