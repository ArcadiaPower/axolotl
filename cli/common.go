package cli

import (
	"io/ioutil"
	"log"
	"strings"

	"github.com/99designs/aws-vault/v6/vault"
	"github.com/alecthomas/kingpin"
)

type Awswitch struct {
	Debug         bool
	awsConfigFile *vault.ConfigFile
}

func (a *Awswitch) AwsConfigFile() (*vault.ConfigFile, error) {
	if a.awsConfigFile == nil {
		var err error
		a.awsConfigFile, err = vault.LoadConfigFromEnv()
		if err != nil {
			return nil, err
		}
	}

	return a.awsConfigFile, nil
}

func (a *Awswitch) MustGetProfileNames() []string {
	config, err := a.AwsConfigFile()
	if err != nil {
		log.Fatalf("Error loading AWS config: %s", err.Error())
	}
	return config.ProfileNames()
}

// ConfigureGlobals sets up the global flags and returns the global config
func ConfigureGlobals(app *kingpin.Application) *Awswitch {
	a := &Awswitch{}

	app.Flag("debug", "Show debugging output").
		BoolVar(&a.Debug)

	app.PreAction(func(c *kingpin.ParseContext) error {
		if !a.Debug {
			log.SetOutput(ioutil.Discard)
		}

		log.Printf("awswitch %s", app.Model().Version)

		return nil
	})

	return a
}

// environ is a slice of environment variables in the form "key=value"
type environ []string

// Unset an environment variable by key
func (e *environ) Unset(key string) {
	for i := range *e {
		if strings.HasPrefix((*e)[i], key+"=") {
			(*e)[i] = (*e)[len(*e)-1]
			*e = (*e)[:len(*e)-1]
			break
		}
	}
}

// Set adds an environment variable, replacing any existing ones of the same key
func (e *environ) Set(key, val string) {
	e.Unset(key)
	*e = append(*e, key+"="+val)
}
