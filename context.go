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
	"sync"

	"github.com/gin-gonic/gin"
)

// Context request context,implementing the *gin.Context
type Context struct {
	sync.Mutex
	// GinContext - reuse the gin Context.
	GinContext *gin.Context
	// ClientContext - context used to call the backend service.
	ClientContext context.Context
	// Param - store the requested body data.
	Param interface{}
	// ErrCode - custome http code when return the error hints.
	ErrCode interface{}
	// Extra - data read from special middleware will be set here.
	Extra map[string]interface{}
	// ErrHandle - handle error hints when validate failed.
	ErrHandle ErrHandle
}
