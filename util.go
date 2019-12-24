package main

import "strings"

import "regexp"

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
