package utils

import (
	"fmt"
	"regexp"

	"github.com/nyaruka/phonenumbers"
)

// ValidPhone 手机号合法性校验
func ValidPhone(code, phone string) bool {
	if code == "" || phone == "" {
		return false
	}

	fullNumber := fmt.Sprintf("+%s%s", code, phone)
	num, err := phonenumbers.Parse(fullNumber, "")
	if err != nil {
		return false
	}
	return phonenumbers.IsValidNumber(num)
}

// ValidEmail 邮箱合法性校验
func ValidEmail(email string) bool {
	if email == "" {
		return false
	}
	match, _ := regexp.MatchString("^[a-zA-Z0-9_.-]+@[a-zA-Z0-9-]+(\\.[a-zA-Z0-9-]+)*\\.[a-zA-Z0-9]{2,6}$", email)
	return match
}
