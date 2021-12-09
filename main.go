package main

import (
    "log"
    "fmt"

    "github.com/spf13/viper"
)

func main() {
    viper.SetDefault("Port", 2000)
    viper.SetDefault("Hostname", "localhost")
    viper.SetDefault("TCPDeadline", 10)


    viper.SetConfigName("stomper_config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("/etc/stomper/")
    viper.AddConfigPath(".")

    err := viper.ReadInConfig()
    if err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); ok {
            log.Println("CONFIG: No stomper_config.yaml file found in invocation dir or /etc/stomper/, using defaults.")
        } else {
            log.Fatal(fmt.Errorf("fatal error config file: %w", err))
        }
    }

	comms := make(chan CnxMgrMsg)
	cm := NewConnectionManager(viper.GetString("hostname"), viper.GetInt("port"), comms, viper.GetDuration("tcpdeadline"))
	st := &MemoryStore{
		Queues: map[string][]Frame{"/queue/test": make([]Frame, 0)},
	}
	e := NewEngine(st, cm, comms)
	err = e.Start()
	if err != nil {
		log.Fatal(err)
	}
}
