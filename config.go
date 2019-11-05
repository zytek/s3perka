package main

import (
	"context"
	"flag"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
)

func setup() Config {
	var config Config

	var configFile string

	flag.StringVar(&configFile, "c", "config.toml", "path to config file")
	flag.Parse()

	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal("Failed to read config file: ", err)
	}
	err = toml.Unmarshal(file, &config)
	if err != nil {
		log.Fatal("Failed to parse config file: ", err)
	}
	if config.Parallel < 1 {
		log.Println("debug: using default parallelism level 10")
		config.Parallel = 10
	}
	return config
}

func handleInterrupt(cancelFunc context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		log.Println("Received an interrupt, canceling copy...")
		log.Println("")
		cancelFunc()
		log.Println("Exiting")
		os.Exit(0)
	}()
}
