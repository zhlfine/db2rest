package api

import (
	"db2rest/db"
	"db2rest/conf"
	"net/url"
	"strings"
	"errors"
	"encoding/json"
	"io/ioutil"
	"text/template"
	"log"
	"fmt"
	"net/http"
)

type Endpoint struct {
	conf 			*conf.Conf
	db 				*db.Client
	url				string
	method			string
	params			[]*Param
	paramDefaults	map[string]string
	fieldMap		map[string]string
	csvMap			map[string]string
	tpl				*template.Template
	sqltype			string
	output 			string
	converter 		string
	converter_csv	string
	fun1			func(*Context) (db.Output, error)
	fun2			func(db.Output)
}

func NewEndpoint(conf *conf.Conf, db *db.Client) (*Endpoint, error) {
	e := &Endpoint{conf: conf, db: db}
	e.url = conf.GetString("url", "")
	e.method = conf.GetString("method", "GET")
	e.output = conf.GetString("output_type", "list")
	e.converter = conf.GetString("output_converter", "lowercamel")
	e.converter_csv = conf.GetString("output_converter_csv", "screamingsnake")
	e.sqltype = conf.GetString("sql_type", "query")
	if e.url == ""{
		return nil, errors.New("api url is not set")
	}
	if err := e.InitParams(); err != nil 		{return nil, err}
	if err := e.InitParamDefaults(); err != nil {return nil, err}
	if err := e.InitFieldMap(); err != nil 		{return nil, err}
	if err := e.InitCsvMap(); err != nil 		{return nil, err}
	if err := e.InitTemplate(); err != nil 		{return nil, err}
	if err := e.InitFunc(); err != nil 			{return nil, err}
	return e, nil
}

func (e *Endpoint) Handle(resp http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s\n", req.Method, req.RequestURI)
	ctx, err := e.Context(resp, req)
	if err != nil {
		ctx.RespondError(400, err)
		return
	}
	output, err := e.fun1(ctx)
	if err != nil {
		ctx.RespondError(500, err)
		return
	}
	e.fun2(output)
}

func (e *Endpoint) InitParams() error {
	e.params = make([]*Param, 0)
	it, err := e.conf.Iterator("params")
	if err != nil {
		return err
	}
	for it.HasNext() {
		i, err := it.Next()
		if err != nil {
			return err
		}
		s := i.GetString("_", "")
		fields := strings.Fields(s)
		p := &Param{name: fields[0]}
		if err := p.ParseValidators(fields[1:]...); err != nil {
			return err
		}
		e.params = append(e.params, p)
	}
	return nil
}

func (e *Endpoint) InitParamDefaults() error {
	e.paramDefaults = make(map[string]string)
	s := e.conf.GetString("param_defaults", "")
	if s != "" {
		m, err := url.ParseQuery(s)
		if err != nil {
			return err
		}
		for i := range m {
			e.paramDefaults[i] = m[i][0]
		}
	}
	return nil
}

func (e *Endpoint) InitFieldMap() error {
	e.fieldMap = make(map[string]string)
	it, err := e.conf.Iterator("output_map")
	if err != nil {
		return err
	}
	for it.HasNext() {
		i, err := it.Next()
		if err != nil {
			return err
		}
		s := i.GetString("_", "")
		if s != "" {
			parts := strings.SplitN(s, ":", 2)
			e.fieldMap[parts[0]] = strings.TrimSpace(parts[1])
		}
	}
	return nil
}

func (e *Endpoint) InitCsvMap() error {
	e.csvMap = make(map[string]string)
	it, err := e.conf.Iterator("output_map_csv")
	if err != nil {
		return err
	}
	for it.HasNext() {
		i, err := it.Next()
		if err != nil {
			return err
		}
		s := i.GetString("_", "")
		if s != "" {
			parts := strings.SplitN(s, ":", 2)
			e.csvMap[parts[0]] = strings.TrimSpace(parts[1])
		}
	}
	return nil
}

func (e *Endpoint) InitTemplate() error {
	name := fmt.Sprintf("%s%s", e.method, e.url)
	sql := e.conf.GetString("sql", "")
	if sql == "" {
		return errors.New("sql is not configured")
	}
	tpl, err := template.New(name).Parse(sql)
	if err != nil {
		return  err
	}
	e.tpl = tpl
	return nil
}

func (e *Endpoint) InitFunc() error {
	switch e.sqltype {
	case "query":
		e.fun1 = e.QueryOutput
		e.fun2 = e.db.Query
	case "update":
		e.fun1 = e.UpdateOutput
		e.fun2 = e.db.Exec
	default:
		return fmt.Errorf("invalid sql type %s", e.sqltype)
	}
	return nil
}

func (e *Endpoint) Context(resp http.ResponseWriter, req *http.Request) (*Context, error) {
	ctx := &Context{api: e, request: req, response: resp}
	if err := e.Parse(ctx); err != nil {
		return ctx, err
	}
	if err := e.Validate(ctx); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (e *Endpoint) Validate(ctx *Context) error {
	for _, p := range e.params {
		if err := p.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (e *Endpoint) Parse(ctx *Context) error {
	ctx.values = make(map[string]interface{})
	for p := range e.paramDefaults {
		ctx.values[p] = e.paramDefaults[p]
	}
	body, err := ioutil.ReadAll(ctx.request.Body)
	if err != nil {
		return err
	}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &ctx.values); err != nil {
			return err
		}
	}
	query, err := url.ParseQuery(ctx.request.URL.RawQuery)
	if err != nil {
		return err
	}
	for i := range query {
		ctx.values[i] = query[i][0]
	}
	return nil
}

func (e *Endpoint) SQL(ctx *Context) (sql string, err error) {
	var buf strings.Builder
	if err = e.tpl.Execute(&buf, ctx); err != nil {
		return
	}
	sql = e.db.FormatSQL(buf.String())
	return
}

func (e *Endpoint) QueryOutput(ctx *Context) (db.Output, error) {
	out := e.output
	if ctx.Bool("csv") {
		out = "csv"
	}
	switch out {
	case "list":
		return NewListOutput(ctx), nil
	case "single":
		return NewSingleOutput(ctx), nil
	case "csv":
		return NewCsvOutput(ctx), nil
	default:
		return nil, fmt.Errorf("invalid output type %s", out)
	}
}

func (e *Endpoint) UpdateOutput(ctx *Context) (db.Output, error) {
	return &ExecOutput{ctx: ctx}, nil
}

