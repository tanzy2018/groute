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
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	validatorv8 "gopkg.in/go-playground/validator.v8"
	validator "gopkg.in/go-playground/validator.v9"
)

// ErrHandle - handle the validator error.
type ErrHandle func(ctx *Context, err interface{})

// Option - optional config the router.
type Option func(*Options)

// Options - option config to initailize the router.
type Options struct {
	router         gin.IRouter
	middleware     []gin.HandlerFunc
	errHandle      ErrHandle
	errTagPrefix   string
	clientContext  context.Context
	useValidatorV9 bool
}

// WithRouter - set the route.
func WithRouter(router gin.IRouter) Option {
	return func(opts *Options) {
		opts.router = router
	}
}

// WithMiddlerware - set the global middleware for this router.
func WithMiddlerware(middleware ...gin.HandlerFunc) Option {
	return func(opts *Options) {
		opts.middleware = append(opts.middleware, middleware...)
	}
}

// WithErrHandle - set error handle for validator.
func WithErrHandle(errHandle ErrHandle) Option {
	return func(opts *Options) {
		opts.errHandle = errHandle
	}
}

// WithErrMsgTagPrefix - set the errprefix tag which define the error hints.
func WithErrMsgTagPrefix(prefix string) Option {
	return func(opts *Options) {
		opts.errTagPrefix = prefix + "-"
	}
}

// WithClientContext - set context used for call the backend services.
func WithClientContext(ctx context.Context) Option {
	return func(opts *Options) {
		opts.clientContext = ctx
	}
}

// WithVaidatorV9 - set validator v9
// supported locale:en,fr,id,ja,nl,pt_BR,tr,zh,zh_tw;default en
func WithVaidatorV9(locale string) Option {
	defaultLocale = locale
	binding.Validator = new(defaultValidator)
	return func(opts *Options) {
		opts.useValidatorV9 = true
	}
}

func getTagByContentType(contentType string) string {
	switch contentType {
	case gin.MIMEJSON:
		return "json"
	case gin.MIMEXML, gin.MIMEXML2:
		return "xml"
	case "", gin.MIMEMultipartPOSTForm, gin.MIMEPOSTForm:
		return "form"
	case gin.MIMEHTML:
		return "html"
	case gin.MIMEPlain:
		return "plain"
	case gin.MIMEYAML:
		return "yaml"
	}
	return ""
}

// DefaulErrHandle -  handler error when validator throw exception.
func defaulErrHandle(c *Context, err interface{}) {
	var msg interface{}
	switch err.(type) {
	case string:
		msg = err.(string)
	case error:
		msg = err.(error).Error()
	case []interface{}:
		msg = err.([]interface{})
	case map[string]string:
		msg = err.(map[string]string)
	case []string:
		msg = err.([]string)
	}
	var code interface{}
	// default code :request params failed to exam.
	code = 402
	if c.ErrCode != nil {
		code = c.ErrCode
	}
	c.GinContext.JSON(http.StatusOK, gin.H{
		"state": 0,
		"code":  code,
		"msg":   msg,
	})
	c.GinContext.Abort()
}

// Router global router manager
type Router struct {
	*Options
}

// NewRouter create a new router
func NewRouter(options ...Option) Router {
	opts := &Options{}
	for _, op := range options {
		op(opts)
	}
	if opts.router == nil {
		panic("gin router must be set and not be nil")
	}

	if opts.errHandle == nil {
		opts.errHandle = defaulErrHandle
	}

	if opts.errTagPrefix == "" {
		opts.errTagPrefix = "err-"
	}

	return Router{
		opts,
	}
}

// add to router
func (r *Router) addInterface(inter Interface) {
	if inter.Method == "" {
		inter.Method = "POST"
	}
	hdlf := func(c *gin.Context) {

		req := &Context{
			GinContext: c,
		}
		if inter.ErrHandle != nil {
			req.ErrHandle = inter.ErrHandle
		} else {
			req.ErrHandle = r.errHandle
		}

		if inter.Param != nil {
			req.Param = reflect.New(reflect.TypeOf(inter.Param)).Interface()
		}

		if inter.Param != nil {
			if err := c.ShouldBind(req.Param); err != nil {
				var errMap map[string]string
				// handle validator v9
				if v, ok := err.(validator.ValidationErrors); ok && r.useValidatorV9 {
					pType := reflect.TypeOf(inter.Param)
					errMap = make(map[string]string, len(v))
					tagType := getTagByContentType(c.GetHeader("Content-Type"))
					if tagType == "" {
						req.ErrHandle(req, fmt.Sprintf("unsupported Content-Type:%s", c.GetHeader("Content-Type")))
						return
					}
					for _, e := range v {
						structField, ok := pType.FieldByName(e.Field())
						if !ok {
							continue
						}

						errmsg := structField.Tag.Get(r.errTagPrefix + e.Tag())
						if errmsg == "" {
							errmsg = e.Translate(translator)
						}
						fieldTag := fieldTagName(tagType, structField)
						errMap[fieldTag] = errmsg
					}
				}
				// handle validator v8
				if v, ok := err.(validatorv8.ValidationErrors); ok && !r.useValidatorV9 {
					pType := reflect.TypeOf(inter.Param)
					errMap = make(map[string]string, len(v))
					tagType := getTagByContentType(c.GetHeader("Content-Type"))
					if tagType == "" {
						req.ErrHandle(req, fmt.Sprintf("unsupported Content-Type:%s", c.GetHeader("Content-Type")))
						return
					}
					for _, e := range v {

						structField, ok := pType.FieldByName(e.Field)
						if !ok {
							continue
						}

						errmsg := structField.Tag.Get(r.errTagPrefix + e.Tag)
						if errmsg == "" {
							errmsg = fmt.Sprintf(
								"param '%s' with value '%v' failed on the validation tag '%s'",
								fieldTagName(tagType, structField),
								e.Value,
								e.Tag,
							)
						}
						fieldTag := fieldTagName(tagType, structField)
						errMap[fieldTag] = errmsg
					}
				}
				if len(errMap) != 0 {
					req.ErrHandle(req, errMap)
					return
				}
			}
		}

		// handle the asynchronous middleware
		if r.clientContext != nil {
			req.ClientContext = r.clientContext
		} else {
			req.ClientContext = req.GinContext
		}

		if len(inter.AsyncHandleFunc) == 1 {
			if err := inter.AsyncHandleFunc[0](req); err != nil {
				req.ErrHandle(req, err)
				return
			}
		}

		if len(inter.AsyncHandleFunc) > 1 {
			ctx, cancel := context.WithCancel(req.GinContext)
			handleLen := len(inter.AsyncHandleFunc)
			errChan := make(chan error, 1)
			for _, fn := range inter.AsyncHandleFunc {
				fn := fn
				wrap(ctx, func(ctx context.Context) {
					select {
					case <-ctx.Done():
					case errChan <- fn(req):
					}
				})
			}
			if err := <-firstError(errChan, handleLen, cancel); err != nil {
				req.ErrHandle(req, err)
				return
			}
		}
		// handle the synchronous middleware
		if len(inter.SyncHandleFunc) > 0 {
			for _, fn := range inter.SyncHandleFunc {
				if err := fn(req); err != nil {
					req.ErrHandle(req, err)
					return
				}
			}
		}
		inter.Handle(req)
	}

	method := r.router.POST
	switch strings.ToLower(inter.Method) {
	case "get":
		method = r.router.GET
	case "post":
		method = r.router.POST
	case "put":
		method = r.router.PUT
	case "delete":
		method = r.router.DELETE
	case "patch":
		method = r.router.PATCH
	case "head":
		method = r.router.HEAD
	default:
	}

	if len(r.middleware) > 0 {
		hdlfs := append(r.middleware, hdlf)
		method(inter.Path, hdlfs...)
	} else {
		method(inter.Path, hdlf)
	}
}

func fieldTagName(tagType string, field reflect.StructField) string {
	sl := strings.Split(field.Tag.Get(tagType), ",")
	if len(sl) > 0 && sl[0] != "" {
		return sl[0]
	}
	return ""
}

// auto register all the exported function to the route
func (r *Router) addStruct(in interface{}) {
	t := reflect.TypeOf(in)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		err := fmt.Errorf("the given interface [realType:%T,baseType:%s] is not a pointer of struct", in, t.Kind())
		panic(err)
	}
	val := reflect.ValueOf(in)
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		if m.Type.NumOut() == 1 && m.Type.Out(0) == reflect.TypeOf(Interface{}) {
			r.addInterface(val.Method(m.Index).Call(nil)[0].Interface().(Interface))
		}
	}
}

// Add add route
func (r *Router) Add(in interface{}) {
	switch in.(type) {
	case Interface:
		r.addInterface(in.(Interface))
	default:
		r.addStruct(in)
	}
}

func wrap(ctx context.Context, f func(context.Context)) {
	go f(ctx)
}

func firstError(in chan error, rounds int, cancel func()) <-chan error {
	out := make(chan error, 1)
	i := 0
	for err := range in {
		if err != nil {
			cancel()
			out <- err
			close(out)
			return out
		}
		i++
		if i >= rounds {
			break
		}
	}
	out <- nil
	close(out)
	return out
}
