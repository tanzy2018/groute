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
