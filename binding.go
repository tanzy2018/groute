package groute

import (
	"reflect"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales"
	english "github.com/go-playground/locales/en"
	french "github.com/go-playground/locales/fr"
	indonesia "github.com/go-playground/locales/id"
	japanese "github.com/go-playground/locales/ja"
	brazilianBR "github.com/go-playground/locales/pt_BR"
	turkish "github.com/go-playground/locales/tr"
	zhongwen "github.com/go-playground/locales/zh"
	zhongwenTW "github.com/go-playground/locales/zh_Hant_TW"
	ut "github.com/go-playground/universal-translator"
	"gopkg.in/go-playground/validator.v9"
	enlocale "gopkg.in/go-playground/validator.v9/translations/en"
	frlocale "gopkg.in/go-playground/validator.v9/translations/fr"
	idlocale "gopkg.in/go-playground/validator.v9/translations/id"
	jalocale "gopkg.in/go-playground/validator.v9/translations/ja"
	nllocale "gopkg.in/go-playground/validator.v9/translations/nl"
	ptBRlocale "gopkg.in/go-playground/validator.v9/translations/pt_BR"
	trlocale "gopkg.in/go-playground/validator.v9/translations/tr"
	zhlocale "gopkg.in/go-playground/validator.v9/translations/zh"
	zhTWlocale "gopkg.in/go-playground/validator.v9/translations/zh_tw"
)

type defaultValidator struct {
	once     sync.Once
	validate *validator.Validate
}

var _ binding.StructValidator = &defaultValidator{}
var defaultLocale = "en"
var translator ut.Translator

// ValidateStruct receives any kind of type, but only performed struct or pointer to struct type.
func (v *defaultValidator) ValidateStruct(obj interface{}) error {
	value := reflect.ValueOf(obj)
	valueType := value.Kind()
	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	if valueType == reflect.Struct {
		v.lazyinit()
		if err := v.validate.Struct(obj); err != nil {
			return err
		}
	}
	return nil
}

// Engine returns the underlying validator engine which powers the default
// Validator instance. This is useful if you want to register custom validations
// or struct level validations. See validator GoDoc for more info -
// https://godoc.org/gopkg.in/go-playground/validator.v9
func (v *defaultValidator) Engine() interface{} {
	v.lazyinit()
	return v.validate
}

func (v *defaultValidator) lazyinit() {
	v.once.Do(func() {
		v.validate = validator.New()
		if err := v.registerTranslations(
			defaultLocale,
			v.validate,
			v.getTrans(defaultLocale)); err != nil {
			panic(err)
		}
		v.validate.SetTagName("binding")
	})
}

func (v *defaultValidator) registerTranslations(
	locale string, validate *validator.Validate, trans ut.Translator) error {
	switch locale {
	case "en":
		return enlocale.RegisterDefaultTranslations(validate, trans)
	case "zh":
		return zhlocale.RegisterDefaultTranslations(validate, trans)
	case "zh_tw":
		return zhTWlocale.RegisterDefaultTranslations(validate, trans)
	case "fr":
		return frlocale.RegisterDefaultTranslations(validate, trans)
	case "ja":
		return jalocale.RegisterDefaultTranslations(validate, trans)
	case "id":
		return idlocale.RegisterDefaultTranslations(validate, trans)
	case "nl":
		return nllocale.RegisterDefaultTranslations(validate, trans)
	case "pt_BR":
		return ptBRlocale.RegisterDefaultTranslations(validate, trans)
	case "tr":
		return trlocale.RegisterDefaultTranslations(validate, trans)
	// default :en
	default:
		return enlocale.RegisterDefaultTranslations(validate, trans)
	}
}

func (v *defaultValidator) getTrans(locale string) ut.Translator {
	var localeInstance locales.Translator
	switch locale {
	case "en", "nl":
		localeInstance = english.New()
	case "zh":
		localeInstance = zhongwen.New()
	case "zh_tw":
		localeInstance = zhongwenTW.New()
	case "fr":
		localeInstance = french.New()
	case "ja":
		localeInstance = japanese.New()
	case "id":
		localeInstance = indonesia.New()
	case "pt_BR":
		localeInstance = brazilianBR.New()
	case "tr":
		localeInstance = turkish.New()
	default:
		localeInstance = english.New()
	}

	uni := ut.New(localeInstance, localeInstance)
	switch locale {
	case "zh_tw":
		locale = "zh"
	case "nl":
		locale = "en"
	}
	trans, _ := uni.GetTranslator(locale)
	translator = trans
	return trans
}
