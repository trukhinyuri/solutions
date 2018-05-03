package middleware

import (
	"net/textproto"

	"git.containerum.net/ch/solutions/pkg/sErrors"
	"github.com/containerum/cherry/adaptors/gonic"
	headers "github.com/containerum/utils/httputil"
	"github.com/gin-gonic/gin"
)

//RequireAdminRole checks if user is admin
func RequireAdminRole(ctx *gin.Context) {
	if ctx.GetHeader(textproto.CanonicalMIMEHeaderKey(headers.UserRoleXHeader)) != "admin" {
		gonic.Gonic(sErrors.ErrAdminRequired(), ctx)
		return
	}
}
