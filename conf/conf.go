package conf

import (
	"db2rest/vexpr"
	"path/filepath"
	"os"
	"fmt"
	"log"
	"io/ioutil"
	"encoding/json"
	"github.com/BurntSushi/toml"
)

type Conf struct {
	data interface{}
}

func LoadEnv(env string) (*Conf, error) {
	file := os.Getenv(env)
	if file == "" {
		return nil, fmt.Errorf("env %s not specified", env)
	}
	return LoadFile(file)
}

func LoadFile(file string) (*Conf, error) {
	ext := filepath.Ext(file)
	switch ext {
	case ".toml":
		return LoadTOML(file)
	case ".json":
		return LoadJSON(file)
	default:
		return nil, fmt.Errorf("unsupported config format %s", file)
	}
}

func LoadTOML(file string) (*Conf, error) {
	c := make(map[string]interface{})
	log.Printf("load config %s\n", file)
	if _, err := toml.DecodeFile(file, &c); err != nil {
		return nil, err
	}
	return &Conf{data: c}, nil
}

func LoadJSON(file string) (*Conf, error) {
	c := make(map[string]interface{})
	log.Printf("load config %s\n", file)

	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &c); err != nil {
		return nil, err
	}
	return &Conf{data: c}, nil
}

func (c *Conf) Get(name string) (*Conf, error) {
	v, err := vexpr.Get(c.data, name)
	if err != nil {
		log.Printf("fail to evaluate value of %s: %v", name, err)
		return nil, err
	}
	return &Conf{data: v}, nil
}

func (c *Conf) GetString(name, def string) string {
	v, err := vexpr.GetString(c.data, name, def)
	if err != nil {
		log.Printf("fail to evaluate value of %s: %v", name, err)
		return def
	}
	if v == "" {
		return def
	}
	return v
}

func (c *Conf) GetInt(name string, def int) int {
	v, err := vexpr.GetInt(c.data, name, def)
	if err != nil {
		log.Printf("fail to evaluate value of %s: %v", name, err)
		return def
	}
	return v
}

func (c *Conf) GetLong(name string, def int64) int64 {
	v, err := vexpr.GetLong(c.data, name, def)
	if err != nil {
		log.Printf("fail to evaluate value of %s: %v", name, err)
		return def
	}
	return v
}

func (c *Conf) GetBool(name string, def bool) bool {
	v, err := vexpr.GetBool(c.data, name, def)
	if err != nil {
		log.Printf("fail to evaluate value of %s: %v, use default %v", name, err, def)
		return def
	}
	return v
}

func (c *Conf) GetFloat(name string, def float32) float32 {
	v, err := vexpr.GetFloat(c.data, name, def)
	if err != nil {
		log.Printf("fail to evaluate value of %s: %v", name, err)
		return def
	}
	return v
}

func (c *Conf) GetDouble(name string, def float64) float64 {
	v, err := vexpr.GetDouble(c.data, name, def)
	if err != nil {
		log.Printf("fail to evaluate value of %s: %v", name, err)
		return def
	}
	return v
}

func (c *Conf) Len(name string) (int, error) {
	return vexpr.Len(c.data, name)
}

func (c *Conf) Iterator(name string) (*Iterator, error) {
	i, err := vexpr.Len(c.data, name)
	if err != nil {
		return nil, err
	}
	return &Iterator{conf: c, name: name, len: i, index: 0}, nil
}

type Iterator struct {
	conf	*Conf
	name	string
	len 	int
	index 	int
}

func (i *Iterator) HasNext() bool {
	return i.index < i.len
}

func (i *Iterator) Next() (*Conf, error) {
	expr := fmt.Sprintf("%s[%d]", i.name, i.index)
	v, err := i.conf.Get(expr)
	i.index++
	return v, err
}


