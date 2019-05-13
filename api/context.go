package api

import (
	"db2rest/vexpr"
	"log"
	"net/http"
)

type Context struct {
	api 		*Endpoint
	request 	*http.Request
	response	http.ResponseWriter
	values		map[string]interface{}
}

func (ctx *Context) Respond(statusCode int, contentType string, body []byte) {
	ctx.response.Header().Add("Content-Type", contentType)
	ctx.response.WriteHeader(statusCode)
	ctx.response.Write(body)
}

func (ctx *Context) RespondJson(statusCode int, body []byte) {
	ctx.Respond(statusCode, "application/json", body)
}

func (ctx *Context) RespondJsonString(statusCode int, body string) {
	ctx.Respond(statusCode, "application/json", []byte(body))
}

func (ctx *Context) RespondError(statusCode int, err error) {
	s := `{"error":"` + err.Error() + `"}`
	ctx.Respond(statusCode, "application/json", []byte(s))
}

func (ctx *Context) RespondNotFound() {
	ctx.Respond(404, "application/json", []byte(`{"found": false}`))
}

func (ctx *Context) Header(name string) string {
	return ctx.request.Header.Get(name)
}

func (ctx *Context) Param(name string) string {
	v, err := vexpr.GetString(ctx.values, name, "")
	if err != nil {
		log.Printf("fail to evaluate value of %s: %v", name, err)
	}
	return v
}

func (ctx *Context) Bool(name string) bool {
	v, err := vexpr.GetBool(ctx.values, name, false)
	if err != nil {
		log.Printf("fail to evaluate value of %s: %v", name, err)
	}
	return v
}

func (c *Context) OrElse(def, v string) string {
	if v == "" {
		return def
	}
	return v
}

func (c *Context) Quote(str string) string {
	if str == "" {
		return "null"
	}
	return "'" + str + "'"
}