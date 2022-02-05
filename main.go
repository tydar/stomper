package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/spf13/viper"
)

func main() {
	viper.SetDefault("Port", 32801)
	viper.SetDefault("Hostname", "localhost")
	viper.SetDefault("TCPDeadline", 0)
	viper.SetDefault("LogPath", "./stomper.log")
	viper.SetDefault("LogToFile", true)
	viper.SetDefault("LogToStdout", false)
	viper.SetDefault("SendWorkers", 1)
	viper.SetDefault("MetricsServer", false)
	viper.SetDefault("MetricsAddress", ":8080")

	// for now, we'll set one default queue to be /queue/main
	// and topics will be created as a string array from the config file
	viper.SetDefault("Topics", []string{"/queue/main"})

	viper.SetConfigName("stomper_config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/stomper/")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("stomper")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("CONFIG: No stomper_config.yaml file found in invocation dir or /etc/stomper/, using defaults.")
		} else {
			log.Fatal(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	// even if LogToStdout is false, if we have no log file we'll default to logging to STDOUT for now
	lfPath := viper.GetString("LogPath")
	logToFile := viper.GetBool("LogToFile")
	logToStdout := viper.GetBool("LogToStdout")
	if logToFile {
		lf, err := os.OpenFile(lfPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("CONFIG: Error opening file %s, logging to STDOUT\n", lfPath)
		} else {
			if logToStdout {
				// if LogToStdout key set and we've got a valid log file loc
				// we need a multiwriter to log to both
				log.SetOutput(io.MultiWriter(lf, os.Stdout))
			} else {
				//otherwise, just log to the file
				log.SetOutput(lf)
			}
		}
	}

	comms := make(chan CnxMgrMsg)
	cm := NewConnectionManager(viper.GetString("hostname"), viper.GetInt("port"), comms, viper.GetDuration("tcpdeadline"))

	topics := viper.GetStringSlice("topics")
	stQueues := make(map[string][][]Frame)
	for i := range topics {
		_, prs := stQueues[topics[i]]
		if prs {
			log.Printf("DUPLICATE_TOPIC: duplicate topics %s defined in config\n", topics[i])
		} else {
			log.Printf("CREATING_TOPIC: %s\n", topics[i])
			stQueues[topics[i]] = make([][]Frame, 0)
		}
	}
	st := &MemoryStore{
		Queues: stQueues,
	}

	e := NewEngine(st, cm, comms, viper.GetInt("SendWorkers"), viper.GetBool("MetricsServer"), viper.GetString("MetricsAddress"))
	err = e.Start()
	if err != nil {
		log.Fatal(err)
	}
}
