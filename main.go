package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ArcadiaPower/awswitch/cli"
	"github.com/spf13/viper"

	"github.com/alecthomas/kingpin"
)

// Version is provided at compile time
var Version string

func init() {
	// initialize viper config
	configName := "config"
	configType := "yaml"
	configPath := os.ExpandEnv("${HOME}/.config/awswitch")

	viper.AddConfigPath(configPath)
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)

	// set default values
	viper.SetDefault("autoGimmeAwsCreds", true)

	configFile := filepath.Join(configPath, fmt.Sprintf("%s.%s", configName, configType))

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// create configPath directory
		if err := os.MkdirAll(configPath, 0755); err != nil {
			log.Fatalf("error creating config directory: %s", err.Error())
		}

		if err := viper.SafeWriteConfig(); err != nil {
			log.Fatalf("error writing config file: %s", err.Error())
		}
	}
}

func main() {
	app := kingpin.New("awswitch", "A helper utility for switching AWS profiles in subshells.")
	app.Version(Version)

	a := cli.ConfigureGlobals(app)
	cli.ConfigureExecCommand(app, a)

	kingpin.MustParse(app.Parse(os.Args[1:]))
}
