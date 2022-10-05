package cli

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/99designs/aws-vault/v6/vault"
	"github.com/alecthomas/kingpin"
	"github.com/c-bata/go-prompt"
	"github.com/spf13/viper"
)

type Awswitch struct {
	Debug             bool
	autoGimmeAwsCreds bool
	awsConfigFile     *vault.ConfigFile
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

// AuthVerify checks if the user is authenticated and if not authenticates
// with gimme-aws-creds
func AuthVerify(enabled bool, profileName string) error {
	if !enabled {
		return nil
	}

	// TODO: This is really ugly and adds a performance hit for the aws cli call
	// but there isn't a better way to verify credentials UNLESS this PR is merged
	// to gimme-aws-creds: https://github.com/Nike-Inc/gimme-aws-creds/pull/300

	// Check if aws cli is installed
	if _, err := exec.LookPath("aws"); err != nil {
		return fmt.Errorf("unable to locate `aws` in PATH, please install it: %w", err)
	}

	// Temporarily set AWS_PROFILE to the profile we want to check
	// so that we can use the aws cli to check if we are authenticated
	// with the profile
	origEnv := os.Environ()
	os.Clearenv()
	for _, e := range origEnv {
		if !strings.HasPrefix(e, "AWS_PROFILE=") {
			os.Setenv(strings.Split(e, "=")[0], strings.Split(e, "=")[1])
		}
	}
	os.Setenv("AWS_PROFILE", profileName)

	// Check if we are authenticated by running aws sts get-caller-identity
	// If we are not authenticated, we will get an error
	// If we are authenticated, we will get a json response
	// We will ignore the json response
	cmd := exec.Command("aws", "sts", "get-caller-identity")
	err := cmd.Run()

	// restore environment
	os.Clearenv()
	for _, e := range origEnv {
		os.Setenv(strings.Split(e, "=")[0], strings.Split(e, "=")[1])
	}

	if err == nil {
		return nil
	}

	// If we are not authenticated, we will run gimme-aws-creds
	return AuthGimmeAwsCreds()
}

// AuthGimmeAwsCreds authenticates with gimme-aws-creds
func AuthGimmeAwsCreds() error {
	// Check if gimme-aws-creds is installed
	if _, err := exec.LookPath("gimme-aws-creds"); err != nil {
		return fmt.Errorf("unable to locate `gimme-aws-creds` in PATH, please install it: %w\n\n\thttps://github.com/Nike-Inc/gimme-aws-creds#installation", err)
	}

	// Verify .okta_aws_login_config exists
	if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".okta_aws_login_config")); os.IsNotExist(err) {
		return fmt.Errorf(os.ExpandEnv("unable to locate .okta_aws_login_config in ${HOME}, please create it: %w\n\n\thttps://github.com/Nike-Inc/gimme-aws-creds#configuration"), err)
	}

	// execute gimme-aws-creds
	cmd := exec.Command("gimme-aws-creds")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("unable to execute `gimme-aws-creds`: %w", err)
	}

	return nil
}

// ConfigureGlobals sets up the global flags and returns the global config
func ConfigureGlobals(app *kingpin.Application) *Awswitch {
	// Load config from file
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("fatal error config file: %v", err)
	}

	viper.GetBool("autoGimmeAwsCreds")

	a := &Awswitch{
		autoGimmeAwsCreds: viper.GetBool("autoGimmeAwsCreds"),
	}

	app.Flag("debug", "Show debugging output").
		BoolVar(&a.Debug)

	app.PreAction(func(c *kingpin.ParseContext) error {
		if !a.Debug {
			log.SetOutput(io.Discard)
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

// profileCompleter returns a list of profile names
func (a *Awswitch) profileCompleter() func(d prompt.Document) []prompt.Suggest {
	return func(d prompt.Document) []prompt.Suggest {
		s := []prompt.Suggest{}
		for _, p := range a.MustGetProfileNames() {
			s = append(s, prompt.Suggest{Text: p})
		}
		return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
	}
}
