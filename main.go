package main

import (
	"db2rest/api"
	"db2rest/conf"
	"syscall"
	"os"
	"log"
	"flag"
	_ "github.com/lib/pq"
)

func main() {
	var file string
    flag.StringVar(&file, "c", "test.toml", "config file")
	flag.Parse()
	
	if file == "" {
		flag.Usage()
		os.Exit(1)
	}

	config, err := conf.LoadFile(file)
	if err != nil {
		log.Printf("error: %v\n", err)
		os.Exit(1)
	}

	server := api.NewServer(config)
	if err  := server.Run(syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL); err != nil {
		log.Printf("error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

