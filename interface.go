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
func NewInterface(inter Interface, handler func(c *Context)) Interface {
	inter.Handle = handler
	return inter
}
