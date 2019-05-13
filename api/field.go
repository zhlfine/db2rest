package api

import (
	"database/sql"
	"strings"
	"github.com/iancoleman/strcase"
)

type Field struct {
	column 		string
	dbType		string
	name		string
}

func NewField(ctx *Context, col *sql.ColumnType) *Field {
	field := &Field{column: col.Name(), dbType: col.DatabaseTypeName()}
	if s, ok := ctx.api.fieldMap[field.column]; ok {
		field.name = s
	} else {
		field.name = convert_name(field.column, ctx.api.converter)
	}
	return field
}

func NewCsvField(ctx *Context, col *sql.ColumnType) *Field {
	field := &Field{column: col.Name(), dbType: col.DatabaseTypeName()}
	if s, ok := ctx.api.csvMap[field.column]; ok {
		field.name = s
	} else {
		field.name = convert_name(field.column, ctx.api.converter_csv)
	}
	return field
}

func convert_name(name, converter string) string {
	switch converter {
	case "camel":
		return strcase.ToCamel(name)
	case "lowercamel":
		return strcase.ToLowerCamel(name)
	case "snake":
		return strcase.ToSnake(name)
	case "screamingsnake":
		return strcase.ToScreamingSnake(name)
	case "kebab":
		return strcase.ToKebab(name)
	case "screamingkebab":
		return strcase.ToScreamingKebab(name)
	case "dotdelimited":
		return strcase.ToDelimited(name, '.')
	case "dotscreamingdelimited":
		return strcase.ToScreamingDelimited(name, '.', true)
	default:
		return name
	}
}

func (f *Field) AppendJsonName(b *strings.Builder) {
	b.Write(json_quote)
	b.WriteString(f.name)
	b.Write(json_quote)
}

func (f *Field) AppendJsonValue(b *strings.Builder, value []byte) {
	switch f.dbType {
	case "DECIMAL", "INT", "BIGINT", "FLOAT8", "BOOL":
		b.Write(value)
	default:
		b.Write(json_quote)
		b.Write(value)
		b.Write(json_quote)
	}
}