package handler

import (
	"strconv"

	"kamogawa/config"
	"kamogawa/core"
	"kamogawa/identity"

	"github.com/gin-gonic/gin"
)

// TODO: branch if logged in
func Login(c *gin.Context) {
	var validationError identity.IdentityError
	a, _ := strconv.ParseInt(c.Query("error"), 10, 64)
	validationError = identity.IdentityError(a)
	core.HTMLWithGlobalState(c, "login.tmpl", gin.H{
		"Email":        c.Query("email"),
		"HasError":     validationError > 0,
		"ErrorMessage": validationError.String(),
		"IsDev":        config.Env == config.Dev,
	})
}
