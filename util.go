package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// UserCooldown структура для кулдауна
type UserCooldown struct {
	time int64
	got  bool
}

// BotUser данные пользывателя, который пишет боту
type BotUser struct {
	ID       int
	Login    string
	LoggedIn bool
}

var loggedUsers = make(map[int]*BotUser)

var cooldowns = make(map[string]map[int]*UserCooldown)

// Check делает check, ну шо го ты доволен?! А?!
func Check(path string) bool {
	if len(path) == 0 {
		return false
	}
	return strings.HasSuffix(path, ".webp")
}

// CheckText проверяет приколы
func CheckText(str string) int {

	str = strings.ToLower(str)
	if strings.Contains(str, "ping") {
		return 1
	}
	return 0
}

// CutTwoArguments делает хуйню
func CutTwoArguments(str string) (string, string) {

	rgxL := regexp.MustCompile(`[^\w]`)
	rgxP := regexp.MustCompile(`[^\w*\-#$%!@^]`)
	arr := strings.Fields(str)
	if len(arr) == 2 && rgxL.Find([]byte(arr[0])) == nil && rgxP.Find([]byte(arr[1])) == nil {
		return arr[0], arr[1]
	}
	return "", ""
}

// GetCooldown а ты подумай по названию
func GetCooldown(cmd string, ID int, cooldown int64) int {

	if cooldowns[cmd] == nil {
		cooldowns[cmd] = make(map[int]*UserCooldown)
	}
	if v, ok := cooldowns[cmd][ID]; ok == true && time.Now().Unix()-v.time < cooldown {
		return int(v.time + cooldown - time.Now().Unix())
	}
	return -1
}

// SetCooldown устанавливаем откат жопы для челика
func SetCooldown(cmd string, ID int) {
	cooldowns[cmd][ID] = &UserCooldown{
		time: time.Now().Unix(),
		got:  false,
	}
}

func CheckCooldown(cmd string, ID int, time int64) (string, bool) {
	if cooldown := GetCooldown(cmd, ID, time); cooldown > 0 {
		if !cooldowns[cmd][ID].got {
			cooldowns[cmd][ID].got = true
			return fmt.Sprintf("Ещё не время подожди где-то %d секунд.", cooldown), true
		}
		return "", true
	}
	return "", false
}
