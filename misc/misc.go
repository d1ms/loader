package misc

import (
	"encoding/json"
	"fmt"
	"os"
)

//Configuration is structure for cfg file
type Configuration struct {
	WorkersNumber int
	Cookie        string
	Metka         string
	Stages        string
	StartYear     int
	UpdateTimeout string
	Refferer      string
}

var (
	defaultConfig = Configuration{
		WorkersNumber: 1,
	}
	configPath = "config.json"
)

//ReadConfig is read config from json file "config.json"
func ReadConfig(configPath string) Configuration {

	file, err := os.Open(configPath)
	if err != nil {
		fmt.Println("Unable to open configuration file.")
		return defaultConfig
	}
	decoder := json.NewDecoder(file)
	var config Configuration
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println(err)
		return defaultConfig
	}
	return config
}
