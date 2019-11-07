package groute_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/levigross/grequests"
	"github.com/stretchr/testify/assert"
	. "github.com/tanzy2018/groute"
)

const (
	baseTestURL = "http://localhost"
)

var (
	testEngin     *gin.Engine
	benchMarkOnce sync.Once
)

type routerTestObj struct{}

func (r routerTestObj) runWithAdd(group string, route interface{}, middleWare ...gin.HandlerFunc) {
	r.add(group, route, middleWare...)
	r.run()
}

// always for the benchmark unit test.
func (r routerTestObj) runWithAddOnce(group string, route interface{}, middleWare ...gin.HandlerFunc) {
	benchMarkOnce.Do(func() {
		os.Setenv("TEST_TYPE", "benchmark")
		r.add(group, route, middleWare...)
		r.run()
	})
}

func (r routerTestObj) run() {
	if testEngin == nil {
		panic("gin.Engine is not initialized")
	}
	go testEngin.Run()
}

func (r routerTestObj) add(group string, route interface{}, middleware ...gin.HandlerFunc) {
	var engin *gin.Engine
	// defualt gin mode
	gin.SetMode(gin.TestMode)
	if testEngin != nil {
		engin = testEngin
	} else {
		if mode := os.Getenv("TEST_TYPE"); mode == "benchmark" {
			engin = gin.New()
		} else {
			engin = gin.Default()
		}
		engin.HandleMethodNotAllowed = true
		engin.NoMethod(func(c *gin.Context) {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"state": 0, "code": http.StatusForbidden, "msg": "Method Is Not Allowed!"})
			return
		})
		engin.NoRoute(func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{"state": 0, "code": http.StatusBadRequest, "msg": "Endpoint Is Not Found!"})
			return
		})
		testEngin = engin
	}
	api := NewRouter(
		WithRouter(engin.Group(group)),
		WithMiddlerware(middleware...),
		WithVaidatorV9("zh"),
	)
	api.Add(route)
}

type TestRouter struct{}

func (tr *TestRouter) Demo1() Interface {
	return NewInterface(
		Interface{
			Path:   "/router-demo1",
			Param:  nil,
			Method: "GET",
		},
		func(c *Context) {
			c.GinContext.String(http.StatusOK, "request path:/router-demo1")
			return
		},
	)
}

func (tr *TestRouter) Demo2() Interface {
	return NewInterface(
		Interface{
			Path:   "/router-demo2",
			Param:  nil,
			Method: "GET",
		},
		func(c *Context) {
			c.GinContext.String(http.StatusOK, "request path:/router-demo2")
			return
		},
	)
}
func TestAdd(t *testing.T) {
	// Add struct
	var r routerTestObj
	r.runWithAdd("/", &TestRouter{})
	// Add Interface
	hd := NewInterface(
		Interface{
			Path:  "/router-demo3",
			Param: nil,

			Method: "GET",
		},
		func(c *Context) {
			c.GinContext.String(http.StatusOK, "request path:/router-demo3")
			return
		},
	)
	r.runWithAdd("/", hd)
	testURLs := []map[string]string{
		map[string]string{
			"url": "/router-demo1",
			"rsp": "request path:/router-demo1",
		},
		map[string]string{
			"url": "/router-demo2",
			"rsp": "request path:/router-demo2",
		},
		map[string]string{
			"url": "/router-demo3",
			"rsp": "request path:/router-demo3",
		},
	}

	for _, url := range testURLs {
		// Ask for the request
		rspGet, err := grequests.Get(baseTestURL+url["url"], nil)
		if err != nil {
			t.Fatal(url["url"], ":", err)
		}
		defer rspGet.Close()
		assert.Equal(t, true, rspGet.Ok)
		assert.Equal(t, url["rsp"], rspGet.String())
	}
}

func TestRouterMethod(t *testing.T) {
	var r routerTestObj
	type Params struct {
		Name string `form:"name" json:"name"`
	}

	// Add Interface
	hd := func(method string) Interface {
		return NewInterface(
			Interface{
				Path:   "/router-method",
				Param:  Params{},
				Method: method,
			},
			func(c *Context) {
				params := c.Param.(*Params)
				str := fmt.Sprintf("path:/router-method;method:%s;name:%s",
					c.GinContext.Request.Method, params.Name)
				c.GinContext.String(http.StatusOK, str)
				return
			},
		)
	}
	r.runWithAdd("/", hd("GET"))
	r.runWithAdd("/", hd("POST"))
	r.runWithAdd("/", hd("DELETE"))
	r.runWithAdd("/", hd("PUT"))
	r.runWithAdd("/", hd("HEAD"))
	r.runWithAdd("/", hd("PATCH"))

	roQuery := &grequests.RequestOptions{
		Params: map[string]string{
			"name": "Lin",
		},
	}

	roJSON := &grequests.RequestOptions{
		JSON: map[string]string{
			"name": "Lin",
		},
	}

	// GET
	res, err := grequests.Get(baseTestURL+"/router-method", roQuery)
	if err != nil {
		t.Fatal("grequests.Get:", err)
	}
	defer res.Close()
	assert.Equal(t, true, res.Ok)
	assert.Equal(t, "path:/router-method;method:GET;name:Lin", res.String())

	// POST
	res1, err := grequests.Post(baseTestURL+"/router-method", roJSON)
	if err != nil {
		t.Fatal("grequests.Get:", err)
	}
	defer res1.Close()
	assert.Equal(t, true, res1.Ok)
	assert.Equal(t, "path:/router-method;method:POST;name:Lin", res1.String())

	// PUT
	res2, err := grequests.Put(baseTestURL+"/router-method", roJSON)
	if err != nil {
		t.Fatal("grequests.Get:", err)
	}
	defer res2.Close()
	assert.Equal(t, true, res2.Ok)
	assert.Equal(t, "path:/router-method;method:PUT;name:Lin", res2.String())

	// DELETE
	res3, err := grequests.Delete(baseTestURL+"/router-method", roJSON)
	if err != nil {
		t.Fatal("grequests.Get:", err)
	}
	defer res3.Close()
	assert.Equal(t, true, res3.Ok)
	assert.Equal(t, "path:/router-method;method:DELETE;name:Lin", res3.String())

}

func TestRouterParamValidate(t *testing.T) {
	type Hit struct {
		Lin       string `form:"lin" json:"lin" binding:"required" err-required:"lin is required"`
		Tan       string `form:"tan" json:"tan" binding:"required" err-required:"tan is required"`
		Heartbeat int32  `form:"heartbeat" json:"heartbeat" binding:"min=1" err-min:"heartbeat must be greater than 0"`
	}

	hd := func(method string) Interface {
		return NewInterface(
			Interface{
				Path:   "/params-validate",
				Param:  Hit{},
				Method: method,
			},
			func(c *Context) {
				c.GinContext.String(http.StatusOK, "params validate test")
				return
			},
		)
	}
	var r routerTestObj
	r.runWithAdd("/", hd("GET"))
	r.runWithAdd("/", hd("POST"))
	ro := &grequests.RequestOptions{
		Params: map[string]string{
			"tan":       "",
			"lin":       "",
			"heartbeat": "-1",
		},
		JSON: map[string]interface{}{
			"tan":       "",
			"lin":       "",
			"heartbeat": 1,
		},
	}
	// get
	rspGet, err := grequests.Get(baseTestURL+"/params-validate", ro)
	if err != nil {
		t.Fatal("/params-validate:", err)
	}
	defer rspGet.Close()
	assert.Equal(t, true, rspGet.Ok)
	expected := map[string]interface{}{
		"code": float64(402),
		"msg": []interface{}{
			"lin is required",
			"tan is required",
			"heartbeat must be greater than 0"},
		"state": float64(0),
	}
	expectedMap := map[string]interface{}{
		"code": float64(402),
		"msg": map[string]interface{}{
			"lin":       "lin is required",
			"tan":       "tan is required",
			"heartbeat": "heartbeat must be greater than 0"},
		"state": float64(0),
	}
	var atual map[string]interface{}
	json.Unmarshal(rspGet.Bytes(), &atual)
	assert.NotEqual(t, nil, atual["msg"])
	assert.Equal(t, expected["code"], atual["code"])
	assert.Equal(t, expected["state"], atual["state"])
	switch atual["msg"].(type) {
	case []interface{}:
		atualMsg, expectedMsg := atual["msg"].([]interface{}), expected["msg"].([]interface{})
		sort.SliceStable(expectedMsg, func(i, j int) bool {
			return expectedMsg[i].(string) < expectedMsg[j].(string)
		})
		sort.SliceStable(atualMsg, func(i, j int) bool {
			return atualMsg[i].(string) < atualMsg[j].(string)
		})
		assert.Equal(t, expectedMsg, atualMsg)
	case map[string]interface{}:
		assert.Equal(t, expectedMap, atual)
	}
}
func TestRouterMiddleware(t *testing.T) {
	hd := NewInterface(
		Interface{
			Path:  "/middleware",
			Param: nil,

			Method: "GET",
		},
		func(c *Context) {
			c.GinContext.String(http.StatusOK, "test midddleware")
			return
		},
	)
	middleware := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.String(200, "request has been abort")
			c.Abort()
			return
			//c.Next()
		}
	}
	var r routerTestObj
	r.runWithAdd("/", hd, middleware())
	// get
	rspGet, err := grequests.Get(baseTestURL+"/middleware", nil)
	if err != nil {
		t.Fatal("/middleware:", err)
	}
	defer rspGet.Close()

	assert.Equal(t, true, rspGet.Ok)
	assert.Equal(t, "request has been abort", rspGet.String())
}

func TestRouterAsynchronous(t *testing.T) {
	nameHd := func(pass bool) ErrHandleFunc {
		return func(c *Context) error {
			if !pass {
				// debug
				log.Println("name err")
				c.ErrCode = 200
				return errors.New("name err")
			}

			fn := func(ctx context.Context) error {
				select {
				case <-ctx.Done():
				case <-time.After(time.Millisecond*100 + time.Millisecond*time.Duration(rand.Intn(1000))):
					c.Lock()
					if c.Extra == nil {
						c.Extra = make(map[string]interface{})
					}
					c.Extra["name"] = "Tan"
					c.Unlock()
					// debug
					log.Println("name:Tan")
				}
				return nil
			}
			return fn(c.ClientContext)
		}
	}

	ageHd := func(pass bool) ErrHandleFunc {
		return func(c *Context) error {
			if !pass {
				c.ErrCode = 201
				// debug
				log.Println("age err")
				return errors.New("age err")
			}

			fn := func(ctx context.Context) error {
				select {
				case <-ctx.Done():
				case <-time.After(time.Millisecond*100 + time.Millisecond*time.Duration(rand.Intn(1000))):
					c.Lock()
					if c.Extra == nil {
						c.Extra = make(map[string]interface{})
					}
					c.Extra["age"] = 27
					c.Unlock()
					// debug
					log.Println("age:27")
				}
				return nil
			}
			return fn(c.ClientContext)
		}
	}

	addrHd := func(pass bool) ErrHandleFunc {
		return func(c *Context) error {
			if !pass {
				// debug
				log.Println("addr err")
				c.ErrCode = 202
				return errors.New("addr err")
			}

			fn := func(ctx context.Context) error {
				select {
				case <-ctx.Done():
				case <-time.After(time.Millisecond*100 + time.Millisecond*time.Duration(rand.Intn(1000))):
					c.Lock()
					if c.Extra == nil {
						c.Extra = make(map[string]interface{})
					}
					c.Extra["address"] = "Black Street No.1"
					c.Unlock()
					// debug
					log.Println("address:Black Street No.1")
				}
				return nil
			}
			return fn(c.ClientContext)
		}
	}

	hd := func(path string, asyncHdlf ...ErrHandleFunc) Interface {
		return NewInterface(
			Interface{
				Path:            path,
				Param:           nil,
				Method:          "GET",
				AsyncHandleFunc: asyncHdlf,
			},
			func(c *Context) {
				c.GinContext.JSON(http.StatusOK, gin.H{
					"msg":  "asynchronous handle test",
					"data": c.Extra,
				})
				return
			},
		)
	}
	var r routerTestObj
	eles := []map[string]interface{}{
		map[string]interface{}{
			"path":     "/router-asynch-name-hd-failure",
			"hdlfs":    ErrHandleFuncChain{nameHd(false), ageHd(true), addrHd(true)},
			"expected": "{\"code\":200,\"msg\":\"name err\",\"state\":0}",
		},
		// map[string]interface{}{
		// 	"path":     "/router-asynch-age-hd-failure",
		// 	"hdlfs":    ErrHandleFuncChain{nameHd(true), ageHd(false), addrHd(true)},
		// 	"expected": "{\"code\":201,\"msg\":\"age err\",\"state\":0}",
		// },
		// map[string]interface{}{
		// 	"path":     "/router-asynch-addr-hd-failure",
		// 	"hdlfs":    ErrHandleFuncChain{nameHd(true), ageHd(true), addrHd(false)},
		// 	"expected": "{\"code\":202,\"msg\":\"addr err\",\"state\":0}",
		// },
		// map[string]interface{}{
		// 	"path":     "/router-asynch-all-success",
		// 	"hdlfs":    ErrHandleFuncChain{nameHd(true), ageHd(true), addrHd(true)},
		// 	"expected": "{\"data\":{\"address\":\"Black Street No.1\",\"age\":27,\"name\":\"Tan\"},\"msg\":\"asynchronous handle test\"}",
		// },
		// map[string]interface{}{
		// 	"path":     "/router-asynch-one-success",
		// 	"hdlfs":    ErrHandleFuncChain{nameHd(true)},
		// 	"expected": "{\"data\":{\"name\":\"Tan\"},\"msg\":\"asynchronous handle test\"}",
		// },
	}

	for _, ele := range eles {
		r.add("/", hd(ele["path"].(string), ele["hdlfs"].(ErrHandleFuncChain)...))
	}
	r.run()

	for _, ele := range eles {
		rsp, err := grequests.Get(baseTestURL+ele["path"].(string), nil)
		if err != nil {
			t.Fatalf("grequests.Get:%s", ele["path"].(string))
		}
		assert.Equal(t, true, rsp.Ok)
		assert.Equal(t, ele["expected"].(string), rsp.String())
		rsp.Close()
	}
}

func TestRouterSynchronous(t *testing.T) {
	nameHd := func(pass bool) ErrHandleFunc {
		return func(c *Context) error {
			if !pass {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
				c.ErrCode = 200
				return errors.New("name err")
			}
			c.Lock()
			if c.Extra == nil {
				c.Extra = make(map[string]interface{})
			}
			c.Extra["name"] = "Tan"
			c.Unlock()
			return nil
		}
	}

	ageHd := func(pass bool) ErrHandleFunc {
		return func(c *Context) error {
			if !pass {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
				c.ErrCode = 201
				return errors.New("age err")
			}
			c.Lock()
			if c.Extra == nil {
				c.Extra = make(map[string]interface{})
			}
			c.Extra["age"] = 27
			c.Unlock()
			return nil
		}
	}

	speekoutHd := func(pass bool, dependenceKey string, key string) ErrHandleFunc {
		return func(c *Context) error {
			if !pass {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
				c.ErrCode = 201
				return fmt.Errorf("speak out %s err", key)
			}

			if c.Extra == nil {
				c.ErrCode = 202
				return errors.New("no message to speak")
			}

			v, ok := c.Extra[dependenceKey]
			if !ok {
				c.ErrCode = 203
				return fmt.Errorf("can't find %s to speak", key)
			}

			c.Lock()
			c.Extra[key] = fmt.Sprintf("speak out %v", v)
			c.Unlock()
			return nil
		}
	}

	hd := func(path string, asyncHdlf ErrHandleFuncChain, syncHdfl ErrHandleFuncChain) Interface {
		return NewInterface(
			Interface{
				Path:            path,
				Param:           nil,
				Method:          "GET",
				AsyncHandleFunc: asyncHdlf,
				SyncHandleFunc:  syncHdfl,
			},
			func(c *Context) {
				c.GinContext.JSON(http.StatusOK, gin.H{
					"msg":  "asynchronous handle test",
					"data": c.Extra,
				})
				return
			},
		)
	}

	eles := []map[string]interface{}{
		map[string]interface{}{
			"path":     "/router-synch-name-hd-failure",
			"async":    ErrHandleFuncChain{nameHd(false), ageHd(true)},
			"sync":     ErrHandleFuncChain{speekoutHd(true, "name", "speak-name"), speekoutHd(true, "age", "speak-age")},
			"expected": "{\"code\":200,\"msg\":\"name err\",\"state\":0}",
		},
		map[string]interface{}{
			"path":     "/router-synch-age-hd-failure",
			"async":    ErrHandleFuncChain{nameHd(true), ageHd(false)},
			"sync":     ErrHandleFuncChain{speekoutHd(true, "name", "speak-name"), speekoutHd(true, "age", "speak-age")},
			"expected": "{\"code\":201,\"msg\":\"age err\",\"state\":0}",
		},
		map[string]interface{}{
			"path":     "/router-synch-speak-name-failure",
			"async":    ErrHandleFuncChain{nameHd(true), ageHd(true)},
			"sync":     ErrHandleFuncChain{speekoutHd(false, "name", "speak-name"), speekoutHd(true, "age", "speak-age")},
			"expected": "{\"code\":201,\"msg\":\"speak out speak-name err\",\"state\":0}",
		},
		map[string]interface{}{
			"path":     "/router-synch-speak-age-failure",
			"async":    ErrHandleFuncChain{nameHd(true), ageHd(true)},
			"sync":     ErrHandleFuncChain{speekoutHd(true, "name", "speak-name"), speekoutHd(false, "age", "speak-age")},
			"expected": "{\"code\":201,\"msg\":\"speak out speak-age err\",\"state\":0}",
		},
		map[string]interface{}{
			"path":     "/router-synch-no-message-failure",
			"async":    ErrHandleFuncChain{},
			"sync":     ErrHandleFuncChain{speekoutHd(true, "name", "speak-name"), speekoutHd(true, "age", "speak-age")},
			"expected": "{\"code\":202,\"msg\":\"no message to speak\",\"state\":0}",
		},

		map[string]interface{}{
			"path":     "/router-synch-no-name-failure",
			"async":    ErrHandleFuncChain{ageHd(true)},
			"sync":     ErrHandleFuncChain{speekoutHd(true, "name", "speak-name"), speekoutHd(true, "age", "speak-age")},
			"expected": "{\"code\":203,\"msg\":\"can't find speak-name to speak\",\"state\":0}",
		},
		map[string]interface{}{
			"path":     "/router-synch-no-age-failure",
			"async":    ErrHandleFuncChain{nameHd(true)},
			"sync":     ErrHandleFuncChain{speekoutHd(true, "name", "speak-name"), speekoutHd(true, "age", "speak-age")},
			"expected": "{\"code\":203,\"msg\":\"can't find speak-age to speak\",\"state\":0}",
		},

		map[string]interface{}{
			"path":     "/router-sync-success",
			"async":    ErrHandleFuncChain{nameHd(true), ageHd(true)},
			"sync":     ErrHandleFuncChain{speekoutHd(true, "name", "speak-name"), speekoutHd(true, "age", "speak-age")},
			"expected": "{\"data\":{\"age\":27,\"name\":\"Tan\",\"speak-age\":\"speak out 27\",\"speak-name\":\"speak out Tan\"},\"msg\":\"asynchronous handle test\"}",
		},
	}

	var r routerTestObj
	for _, ele := range eles {
		r.add("/", hd(
			ele["path"].(string),
			ele["async"].(ErrHandleFuncChain),
			ele["sync"].(ErrHandleFuncChain)))
	}
	r.run()

	for _, ele := range eles {
		rsp, err := grequests.Get(baseTestURL+ele["path"].(string), nil)
		if err != nil {
			t.Fatalf("grequests.Get:%s", ele["path"].(string))
		}
		assert.Equal(t, true, rsp.Ok)
		assert.Equal(t, ele["expected"].(string), rsp.String())
		rsp.Close()
	}

}

func BenchmarkBlank(b *testing.B) {
	hd := NewInterface(
		Interface{
			Path:   "/benchmark-blank",
			Param:  nil,
			Method: "GET",
		},
		func(c *Context) {
		},
	)
	var r routerTestObj
	r.runWithAddOnce("/", hd)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(baseTestURL + "/benchmark-blank")
		if err == nil {
			resp.Body.Close()
		}
	}
}

func BenchmarkGinBlank(b *testing.B) {
	gin.SetMode(gin.TestMode)
	engin := gin.New()
	engin.GET("/benchmark-gin-blank", func(*gin.Context) {

	})
	benchMarkOnce.Do(func() {
		go engin.Run()
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(baseTestURL + "/benchmark-gin-blank")
		if err == nil {
			resp.Body.Close()
		}
	}
}

func TestBlank(t *testing.T) {
	hd := NewInterface(
		Interface{
			Path:   "/benchmark-blank",
			Param:  nil,
			Method: "GET",
		},
		func(c *Context) {
		},
	)
	var r routerTestObj
	r.runWithAddOnce("/", hd)
	now := time.Now()
	resp, err := http.Get(baseTestURL + "/benchmark-blank")
	if err == nil {
		resp.Body.Close()
	}
	assert.Equal(t, 0, time.Since(now).Microseconds())
}

func TestGinBlank(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engin := gin.New()
	engin.GET("/benchmark-gin-blank", func(*gin.Context) {

	})
	benchMarkOnce.Do(func() {
		go engin.Run()
	})
	now := time.Now()
	resp, err := http.Get(baseTestURL + "/benchmark-gin-blank")
	if err == nil {
		resp.Body.Close()
	}
	assert.Equal(t, 0, time.Since(now).Microseconds())
}
