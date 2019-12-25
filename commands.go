package main

import (
	"fmt"
	"log"

	"github.com/go-pg/pg"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Chattables канал для сообщений с команд
var Chattables = make(chan tgbotapi.Chattable)

// PingHandler handles command Ping
func PingHandler(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	Chattables <- tgbotapi.NewMessage(update.Message.Chat.ID, "pong")
}

// RegisterHandler handles command register
func RegisterHandler(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	arg1, arg2 := CutTwoArguments(update.Message.CommandArguments())
	if arg1 == "" || arg2 == "" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, `Неправильно введена команда.
Правильная форма: `+"`/register <login> <password>`"+`.
Логин должен состоять из `+"`A-Z, a-z, 0-9, _`"+`;
а пароль должен состоять из `+"`A-Z, a-z, 0-9, _, *, -, #, $, %, !, @, ^`"+`.`)
		msg.ParseMode = "Markdown"
		Chattables <- msg
		return
	}

	user1 := &User{
		Name:     arg1,
		Email:    "shit@gogle.net",
		Password: arg2,
	}

	err := db.Insert(user1)
	if err != nil {
		pgErr, ok := err.(pg.Error)
		if ok && pgErr.IntegrityViolation() {
			Chattables <- tgbotapi.NewMessage(update.Message.Chat.ID, "А пользователь уже есть, для вас мест нет (ну по крайней мере под таким логином).")
		} else {
			Chattables <- tgbotapi.NewMessage(update.Message.Chat.ID, "У вас тротлинг. Сочувствую.")
		}
		return
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ваш логин - `%s`, ваш пароль - `%s`. Вас зарегистрировали, поэтому вы ЛОХ.", arg1, arg2))
	msg.ParseMode = "Markdown"
	Chattables <- msg
}

// LoginHandler handles command login
func LoginHandler(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	arg1, arg2 := CutTwoArguments(update.Message.CommandArguments())
	if arg1 == "" || arg2 == "" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, `Неправильно введена команда.
Правильная форма: `+"`/login <login> <password>`"+`.
Логин должен состоять из `+"`A-Z, a-z, 0-9, _`"+`;
а пароль должен состоять из `+"`A-Z, a-z, 0-9, _, *, -, #, $, %, !, @, ^`"+`.`)
		msg.ParseMode = "Markdown"
		Chattables <- msg
		return
	}
	user := new(User)
	err := db.Model(user).Where("name = ? AND password = ?", arg1, arg2).Select()
	if err != nil {
		if err == pg.ErrNoRows {
			audioMessage := tgbotapi.NewVoiceUpload(update.Message.Chat.ID, "tryAgain.mp3")
			audioMessage.Caption = "Неверный логин или пароль. Попробуй *ЕЩЁ РАЗ*."
			audioMessage.ParseMode = "Markdown"
			Chattables <- audioMessage
		} else {
			Chattables <- tgbotapi.NewMessage(update.Message.Chat.ID, "У вас тротлинг. Сочувствую.")
		}
		return
	}
	log.Println(user)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Вроде как всё правильно, поздравляю вы внутри."))
	msg.ParseMode = "Markdown"
	Chattables <- msg
}

//AnimeHandler handles command anime
func AnimeHandler(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	Chattables <- tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Для пидоров. Вот к примеру таких как %s (@%s)", update.Message.From.FirstName, update.Message.From.UserName))
}

//SoldatHandler handles command soldat
func SoldatHandler(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	// if v, ok := loggedUsers[update.Message.From.ID]; !ok || v == nil || !v.LoggedIn {
	// 	msg.Text = "Я не знаю, кто вы. Залогинтесь, пожалуйста."
	// 	SetCooldown(update.Message.Command(), update.Message.From.ID)
	// 	return
	// }
	Chattables <- tgbotapi.NewMessage(update.Message.Chat.ID, "https://www.youtube.com/watch?v=POb02mjj2zE")
}

//BonkHandler handles command bonk
func BonkHandler(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	msgSticker := tgbotapi.NewStickerShare(update.Message.Chat.ID, "CAADAgAD_QEAAvNWPxeLf-J5M600mhYE")
	Chattables <- msgSticker
}

// HandleCommand executes command handlers
func HandleCommand(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	switch update.Message.Command() {
	case "ping":
		if txt, cd := CheckCooldown(update.Message.Command(), update.Message.From.ID, 10); cd {
			if len(txt) > 0 {
				Chattables <- tgbotapi.NewMessage(update.Message.Chat.ID, txt)
			}
			break
		}
		PingHandler(update, bot)
		SetCooldown(update.Message.Command(), update.Message.From.ID)

	case "register":
		if txt, cd := CheckCooldown(update.Message.Command(), update.Message.From.ID, 10); cd {
			if len(txt) > 0 {
				Chattables <- tgbotapi.NewMessage(update.Message.Chat.ID, txt)
			}
			break
		}
		RegisterHandler(update, bot)
		SetCooldown(update.Message.Command(), update.Message.From.ID)

	case "login":
		if txt, cd := CheckCooldown(update.Message.Command(), update.Message.From.ID, 10); cd {
			if len(txt) > 0 {
				Chattables <- tgbotapi.NewMessage(update.Message.Chat.ID, txt)
			}
			break
		}
		LoginHandler(update, bot)
		SetCooldown(update.Message.Command(), update.Message.From.ID)

	case "anime":
		if txt, cd := CheckCooldown(update.Message.Command(), update.Message.From.ID, 10); cd {
			if len(txt) > 0 {
				Chattables <- tgbotapi.NewMessage(update.Message.Chat.ID, txt)
			}
			break
		}
		AnimeHandler(update, bot)
		SetCooldown(update.Message.Command(), update.Message.From.ID)

	case "soldat":
		if txt, cd := CheckCooldown(update.Message.Command(), update.Message.From.ID, 10); cd {
			if len(txt) > 0 {
				Chattables <- tgbotapi.NewMessage(update.Message.Chat.ID, txt)
			}
			break
		}
		SoldatHandler(update, bot)
		SetCooldown(update.Message.Command(), update.Message.From.ID)

	case "bonk":
		if txt, cd := CheckCooldown(update.Message.Command(), update.Message.From.ID, 10); cd {
			if len(txt) > 0 {
				Chattables <- tgbotapi.NewMessage(update.Message.Chat.ID, txt)
			}
			break
		}
		BonkHandler(update, bot)
		SetCooldown(update.Message.Command(), update.Message.From.ID)
	}

}

// SendAllChattables send all command messages
func SendAllChattables(bot *tgbotapi.BotAPI) {
	for msg := range Chattables {
		if _, err := bot.Send(msg); err != nil {
			log.Println(err.Error())
		}
	}
}
