// MIT License

// Copyright (c) 2019 tanzy2018

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
	en_trans "gopkg.in/go-playground/validator.v9/translations/en"
	fr_trans "gopkg.in/go-playground/validator.v9/translations/fr"
	id_trans "gopkg.in/go-playground/validator.v9/translations/id"
	ja_trans "gopkg.in/go-playground/validator.v9/translations/ja"
	nl_trans "gopkg.in/go-playground/validator.v9/translations/nl"
	ptBR_trans "gopkg.in/go-playground/validator.v9/translations/pt_BR"
	tr_trans "gopkg.in/go-playground/validator.v9/translations/tr"
	zh_trans "gopkg.in/go-playground/validator.v9/translations/zh"
	zhTW_trans "gopkg.in/go-playground/validator.v9/translations/zh_tw"
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
		return en_trans.RegisterDefaultTranslations(validate, trans)
	case "zh":
		return zh_trans.RegisterDefaultTranslations(validate, trans)
	case "zh_tw":
		return zhTW_trans.RegisterDefaultTranslations(validate, trans)
	case "fr":
		return fr_trans.RegisterDefaultTranslations(validate, trans)
	case "ja":
		return ja_trans.RegisterDefaultTranslations(validate, trans)
	case "id":
		return id_trans.RegisterDefaultTranslations(validate, trans)
	case "nl":
		return nl_trans.RegisterDefaultTranslations(validate, trans)
	case "pt_BR":
		return ptBR_trans.RegisterDefaultTranslations(validate, trans)
	case "tr":
		return tr_trans.RegisterDefaultTranslations(validate, trans)
	// default :en
	default:
		return en_trans.RegisterDefaultTranslations(validate, trans)
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
