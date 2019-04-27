package main

import (
	"os"
	"flag"
	"log"
	"github.com/BurntSushi/toml"
)

type arrayflags []string

func (f *arrayflags) String() string {
    return "config files"
}

func (f *arrayflags) Set(value string) error {
    *f = append(*f, value)
    return nil
}

var files arrayflags = []string{`C:\calix\golang\db2ms\test.toml`}
var conf Config

func loadConfig() {
	flag.Var(&files, "c", "config file")
	flag.Parse()

	if len(files) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	for _, f := range files {
		log.Printf("load config %s\n", f)
		if _, err := toml.DecodeFile(f, &conf); err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	loadConfig()
	
	do()

	log.Println("done")
}

