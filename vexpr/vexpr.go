package vexpr

import (
	"strconv"
	"regexp"
	"reflect"
	"strings"
	"fmt"
)

var val_expr = regexp.MustCompile(`([^\[]+)(?:\[(\d*)\])?`)

func Get(data interface{}, expr string) (interface{}, error) {
	parts := strings.Split(expr, ",")
	for _, part := range parts {
		v, err := extract(data, part)
		if err != nil {
			return nil, err
		}
		if v != nil {
			return v, nil
		}
	}
	return nil, nil
}

func Len(data interface{}, expr string) (int, error) {
	v, err := Get(data, expr)
	if err != nil {
		return 0, err
	}
	if v == nil {
		return 0, nil
	}

	vv := reflect.ValueOf(v)
	if vv.Kind() != reflect.Array && vv.Kind() != reflect.Slice {
		return 0, fmt.Errorf("%s is not array (%v)", expr, vv.Type())
	}
	return vv.Len(), nil
}

func GetString(data interface{}, expr string, def string) (string, error) {
	v, err := Get(data, expr)
	if err != nil {
		return "", err
	}
	if v == nil {
		return def, nil
	}
	return fmt.Sprint(v), nil
}

func GetInt(data interface{}, expr string, def int) (int, error) {
	v, err := GetString(data, expr, "")
	if err != nil {
		return 0, err
	}
	if v == "" {
		return def, nil
	}

	i, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func GetLong(data interface{}, expr string, def int64) (int64, error) {
	v, err := GetString(data, expr, "")
	if err != nil {
		return 0, err
	}
	if v == "" {
		return def, nil
	}

	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func GetBool(data interface{}, expr string, def bool) (bool, error) {
	v, err := GetString(data, expr, "")
	if err != nil {
		return false, err
	}
	if v == "" {
		return def, nil
	}

	i, err := strconv.ParseBool(v)
	if err != nil {
		return false, err
	}
	return i, nil
}

func GetFloat(data interface{}, expr string, def float32) (float32, error) {
	v, err := GetString(data, expr, "")
	if err != nil {
		return 0, err
	}
	if v == "" {
		return def, nil
	}

	i, err := strconv.ParseFloat(v, 32)
	if err != nil {
		return 0, err
	}
	return float32(i), nil
}

func GetDouble(data interface{}, expr string, def float64) (float64, error) {
	v, err := GetString(data, expr, "")
	if err != nil {
		return 0, err
	}
	if v == "" {
		return def, nil
	}

	i, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func extract(data interface {}, expr string) (interface{}, error) {
	if data == nil || expr == "" {
		return nil, nil
	}

	ss := strings.SplitN(expr, ".", 2)
	f := val_expr.FindStringSubmatch(ss[0])

	rv := reflect.ValueOf(data)
	var vv reflect.Value

	if f[1] == "_" {
		vv = rv
	} else {
		switch rv.Kind() {
		case reflect.Map:
			mv := reflect.ValueOf(f[1])
			vv = rv.MapIndex(mv)
		case reflect.Struct:
			vv = rv.FieldByName(f[1])
		}
	}

	if !vv.IsValid() {
		return nil, nil
	}

	if f[2] != "" {
		vv = reflect.ValueOf(vv.Interface())
		if vv.Kind() != reflect.Array && vv.Kind() != reflect.Slice {
			return nil, fmt.Errorf("%s is not array (%v)", f[1], vv.Type())
		}
		i, _ := strconv.Atoi(f[2])
		if i >= vv.Len() {
			return nil, fmt.Errorf("index %d out of range %d", i, vv.Len())
		}
		vv = vv.Index(i)
	}

	if len(ss) > 1 {
		return extract(vv.Interface(), ss[1])
	} else {
		return convert(vv.Interface()), nil
	}
}

func convert(data interface {}) interface{} {
	if data == nil {
		return nil
	}

	rv := reflect.ValueOf(data)
	switch rv.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		if rv.Len() == 0 {
			return nil
		}
	}

	return data
}