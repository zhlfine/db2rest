package api

import (
	"regexp"
	"errors"
	"fmt"
	"strings"
)

type Validator interface {
	Validate(p *Param, value string) error
	Init(config string) error
}

type requiredValidator struct {
	required	bool
}

func (vd *requiredValidator) Init(config string) error {
	if config == "" {
		vd.required = true
		return nil
	}

	switch strings.ToLower(config) {
	case "true", "t", "yes", "y", "on", "1":
		vd.required = true
	case "false", "f", "no", "n", "off", "0":
		vd.required = false
	default:
		return fmt.Errorf("invalid config: %s", config)
	}

	return nil
}

func (vd *requiredValidator) Validate(p *Param, v string) error {
	if vd.required && v == ""{
		return fmt.Errorf("param %s is required", p.name)
	}
	return nil
}

type patternValidator struct {
	regex	*regexp.Regexp
}

func (vd *patternValidator) Init(config string) (err error) {
	if config == "" {
		err = errors.New("invalid pattern config")
	} else {
		vd.regex, err = regexp.Compile(config)
	}
	return
}

func (vd *patternValidator) Validate(p *Param, v string) error {
	if v != "" && !vd.regex.MatchString(v){
		return fmt.Errorf("param %s is invalid", p.name)
	}
	return nil
}

