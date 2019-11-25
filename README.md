# groute

## A lightweight structual web framework based on gin framework and validator

### Simple example

```golang

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/tanzy2018/groute"
)

type Student struct{}

type Params struct {
	Id   int    `form:"id" binding:"required,min=10" err-required:"id is required" err-min:"id must be greater than 10 or equal to 10"`
	Name string `form:"name" binding:"required" err-required:"name is required"`
}

func (s *Student) Info() groute.Interface {
	return groute.NewInterface(
		groute.Interface{
			Param:  Params{},
			Method: "GET",
			Path:   "/info",
		},
		func(c *groute.Context) {
			params := c.Param.(*Params)
			c.GinContext.JSON(200, gin.H{
				"id":   params.Id,
				"name": params.Name,
			})
			return
		},
	)
}

func (s *Student) Score() groute.Interface {
	return groute.NewInterface(
		groute.Interface{
			Param:  Params{},
			Method: "GET",
			Path:   "/score",
		},
		func(c *groute.Context) {
			params := c.Param.(*Params)
			c.GinContext.JSON(200, gin.H{
				"id":    params.Id,
				"name":  params.Name,
				"score": 100,
			})
			return
		},
	)
}

func main() {
	engine := gin.Default()
	api := groute.NewRouter(
		groute.WithVaidatorV9("zh"),
		groute.WithRouter(engine.Group("/student")),
	)
	api.Add(&Student{})
	engine.Run()
}


```

### Run this simple example

``` html
go run main.go
```

### Output of simple example

```html

request:curl http://localhost:8080/student/score\?id\=9\&name\=lin
output: {"code":402,"msg":{"id":"id must be greater than 10 or equal to 10"},"state":0}

request:curl curl http://localhost:8080/student/score\?id\=91\&name\=lin
output:{"id":91,"name":"lin","score":100}

```
