package api

import (
	"sync"
	"fmt"
	"strings"
	"encoding/csv"
	"database/sql"
)

var json_bracket_1	= []byte("[")
var json_bracket_2	= []byte("]")
var json_brace_1	= []byte("{")
var json_brace_2	= []byte("}")
var json_quote 		= []byte("\"")
var json_colon		= []byte(":")
var json_comma		= []byte(",")
var json_null		= []byte("null")

type ExecOutput struct {
	ctx 			*Context
	lastInsertId	int64
	rowsAffected	int64
}

func (o *ExecOutput) SQL() (string, error) {
	return o.ctx.api.SQL(o.ctx);
}

func (o *ExecOutput) Columns(cols []*sql.ColumnType) error {
	return nil;
}

func (o *ExecOutput) Row(row []*[]byte) error {
	return nil;
}

func (o *ExecOutput) Error(err error) {
	o.ctx.RespondError(500, err)
}

func (o *ExecOutput) Affected(lastInsertId, rowsAffected int64) {
	o.lastInsertId = lastInsertId
	o.rowsAffected = rowsAffected
}

func (o *ExecOutput) End() {
	if o.lastInsertId > 0 {
		s := fmt.Sprintf(`{"lastInsertId":%d,"rowsAffected":%d}`, o.lastInsertId, o.rowsAffected)
		o.ctx.RespondJson(200, []byte(s))
	} else {
		s := fmt.Sprintf(`{"rowsAffected":%d}`, o.rowsAffected)
		o.ctx.RespondJson(200, []byte(s))
	}
}

type ListOutput struct {
	ctx 	*Context
	fields	[]*Field
	once	bool
	buffer  *strings.Builder
}

func NewListOutput(ctx *Context) *ListOutput {
	return &ListOutput{ctx: ctx}
}

func (o *ListOutput) SQL() (string, error) {
	return o.ctx.api.SQL(o.ctx);
}

func (o *ListOutput) Columns(cols []*sql.ColumnType) error {
	o.fields = make([]*Field, len(cols))
	for i, col := range cols {
		o.fields[i] = NewField(o.ctx, col)
	}

	o.once = false
	o.buffer = new(strings.Builder)
	o.buffer.Write(json_bracket_1)

	return nil;
}

func (o *ListOutput) Row(row []*[]byte) error {
	if o.once {
		o.buffer.Write(json_comma)
	}
	o.once = true

	o.buffer.Write(json_brace_1)

	f := false
	for i := 0; i < len(row); i++ {
		field := o.fields[i]
		val := *row[i]
		if len(val) > 0 {
			if f {
				o.buffer.Write(json_comma)
			}
			f = true

			field.AppendJsonName(o.buffer)
			o.buffer.Write(json_colon)
			field.AppendJsonValue(o.buffer, val)
		}
	}
	o.buffer.Write(json_brace_2)

	return nil;
}

func (o *ListOutput) Affected(lastInsertId, rowsAffected int64) {

}

func (o *ListOutput) Error(err error) {
	o.ctx.RespondError(500, err)
}

func (o *ListOutput) End() {
	if o.once {
		o.buffer.Write(json_bracket_2)
		o.ctx.RespondJson(200, []byte(o.buffer.String()))
	} else {
		o.ctx.RespondNotFound()
	}
}

type SingleOutput struct {
	list	*ListOutput
	once	sync.Once
}

func NewSingleOutput(ctx *Context) *SingleOutput {
	return &SingleOutput{list: NewListOutput(ctx)}
}

func (o *SingleOutput) SQL() (string, error) {
	return o.list.ctx.api.SQL(o.list.ctx);
}

func (o *SingleOutput) Columns(cols []*sql.ColumnType) error {
	o.list.fields = make([]*Field, len(cols))
	for i, col := range cols {
		o.list.fields[i] = NewField(o.list.ctx, col)
	}

	o.list.once = false
	o.list.buffer = new(strings.Builder)
	return nil;
}

func (o *SingleOutput) Row(row []*[]byte) error {
	if !o.list.once {
		return o.list.Row(row)
	}
	return nil
}

func (o *SingleOutput) Affected(lastInsertId, rowsAffected int64) {
	
}

func (o *SingleOutput) Error(err error) {
	o.list.ctx.RespondError(500, err)
}

func (o *SingleOutput) End() {
	if o.list.once {
		o.list.ctx.RespondJson(200, []byte(o.list.buffer.String()))
	} else {
		o.list.ctx.RespondNotFound()
	}
}

type CsvOutput struct {
	ctx 	*Context
 	w 		*csv.Writer
}

func NewCsvOutput(ctx *Context) *CsvOutput {
	return &CsvOutput{ctx: ctx}
}

func (o *CsvOutput) Write(p []byte) (n int, err error) {
	return o.ctx.response.Write(p)
}

func (o *CsvOutput) SQL() (string, error) {
	return o.ctx.api.SQL(o.ctx);
}

func (o *CsvOutput) Columns(cols []*sql.ColumnType) error {
	fields := make([]string, len(cols))
	for i, col := range cols {
		fields[i] = NewCsvField(o.ctx, col).name
	}

	filename := o.ctx.Param("filename")
	if filename == "" {
		filename = o.ctx.request.URL.EscapedPath()
	}
	o.ctx.response.Header().Add("Content-Disposition", `attachment; filename="` + filename + `.csv"`)

	o.w = csv.NewWriter(o)
	if err := o.w.Write(fields); err != nil {
		return err
	}
	return nil;
}

func (o *CsvOutput) Row(row []*[]byte) error {
	fields := make([]string, len(row))
	for i := 0; i < len(row); i++ {
		fields[i] = string(*row[i])
	}
	if err := o.w.Write(fields); err != nil {
		return err
	}
	return nil;
}

func (o *CsvOutput) Affected(lastInsertId, rowsAffected int64) {
	
}

func (o *CsvOutput) Error(err error) {
	o.ctx.RespondError(500, err)
}

func (o *CsvOutput) End() {
	o.w.Flush()
	if err := o.w.Error(); err != nil {
		o.ctx.RespondError(500, err)
	}
}
