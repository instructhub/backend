package utils

import (
	"log"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var LangList = []string{"en", "es", "zh-tw", "zh-cn"}

type LangLocal string

// 定義可用的語言常數
const (
	English  LangLocal = "en"
	Spanish  LangLocal = "es"
	ChineseTW LangLocal = "zh-tw"
	ChineseCN LangLocal = "zh-cn"
)

var langLocalValidator validator.Func = func(fl validator.FieldLevel) bool {
    lang := fl.Field().String()
    for _, validLang := range LangList {
        if lang == validLang {
            return true
        }
    }
    return false
}

func InitVaildator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("lang", langLocalValidator)
	} else {
		log.Fatalf("error register vaildatioon")
	}
}
