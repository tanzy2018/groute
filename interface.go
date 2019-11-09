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

// HandleFunc - handle function.
type HandleFunc func(*Context)

// ErrHandleFunc -
type ErrHandleFunc func(*Context) error

// ErrHandleFuncChain - handle function chain.
type ErrHandleFuncChain []ErrHandleFunc

// Interface define the router interface
type Interface struct {
	// SyncHandleFunc - the special middleware to handle the request param in synchronous way
	// but only excutes after all the AsyncHandleFunc successing,because some
	// param may rely on result of the asynchronouse handlefuncs.
	SyncHandleFunc ErrHandleFuncChain
	// AsyncHandleFunc - the special middleware to handle the request param in asynchronous way.
	AsyncHandleFunc ErrHandleFuncChain
	// Path - starts with "/".
	Path string
	// Method - one of `POST,GET,DELETE,PUT,HEAD,PATCH`,case insensitive.
	Method string
	// Param - requrest params
	Param interface{}
	// Handle function that handles the business logic .
	Handle HandleFunc
	// ErrHandle
	ErrHandle ErrHandle
}

// NewInterface - create a new Interface instance.
func NewInterface(inter Interface, handle func(c *Context)) Interface {
	inter.Handle = handle
	return inter
}
