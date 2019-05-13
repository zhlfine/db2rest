package api

import (
	"fmt"
	"strings"
)

type Param struct {
	name 	string
	validators 	[]Validator
}

func (p *Param) ParseValidators(config ...string) error {
	p.validators = make([]Validator, len(config))
	for i, s := range config {
		parts := strings.SplitN(s, ":", 2)

		switch parts[0] {
		case "required":
			p.validators[i] = &requiredValidator{}
		case "pattern":
			p.validators[i] = &patternValidator{}
		default:
			return fmt.Errorf("invalid config: %s", s)
		}

		var part2 string
		if len(parts) > 1 {
			part2 = parts[1]
		}

		if err := p.validators[i].Init(part2); err != nil {
			return err
		}
	}
	return nil
}

func (p *Param) Validate(c *Context) error {
	v := c.Param(p.name)
	for _, validator := range p.validators {
		if err := validator.Validate(p, v); err != nil {
			return err
		}
	}
	return nil
}
