package main

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/encoding"
	"github.com/go-kratos/kratos/v2/encoding/form"
)

// Represents a REST API Controller
type Controller interface {
	SetupRoutes(
		version int,
		router *gin.RouterGroup,
	)
}

func BindQuery(vars url.Values, target interface{}) error {
	return encoding.GetCodec(form.Name).Unmarshal([]byte(vars.Encode()), target)
}

// BindForm bind form parameters to target.
func BindForm(req *http.Request, target interface{}) error {
	if err := req.ParseForm(); err != nil {
		return err
	}
	return encoding.GetCodec(form.Name).Unmarshal([]byte(req.Form.Encode()), target)
}

// BindVars bind path variables to target.
func BindReqVars(ctx *gin.Context, target interface{}) error {
	vars := make(url.Values, len(ctx.Params))
	for _, v := range ctx.Params {
		vars[v.Key] = []string{v.Value}
	}
	return BindQuery(vars, target)
}

// BindQuery bind query parameters to target.
func BindReqQuery(ctx *gin.Context, target interface{}) error {
	return BindQuery(ctx.Request.URL.Query(), target)
}
