package cli

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ArcadiaPower/axolotl/sdk/vault"
	"github.com/alecthomas/kingpin"
	"github.com/c-bata/go-prompt"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

type Axolotl struct {
	Debug              bool
	autoGimmeAwsCreds  bool
	defaultRegion      string
	awsCredentialsFile *vault.CredentialsFile
	profiles           map[string]string
}

func (a *Axolotl) AwsCredentialsFile() (*vault.CredentialsFile, error) {
	if a.awsCredentialsFile == nil {
		var err error
		a.awsCredentialsFile, err = vault.LoadCredentialsFromEnv()
		if err != nil {
			return nil, err
		}
	}

	return a.awsCredentialsFile, nil
}

// MustGetAWSProfileNames returns a list of AWS profile names
// based on the contents of the local credentials file
func (a *Axolotl) MustGetAWSProfileNames() []string {
	creds, err := a.AwsCredentialsFile()
	if err != nil {
		log.Fatalf("Error loading AWS credentials: %s", err.Error())
	}

	// filter out DEFAULT profile
	profileNames := []string{}
	for _, profile := range creds.ProfileNames() {
		if profile != "DEFAULT" {
			profileNames = append(profileNames, profile)
		}
	}

	return profileNames
}

// MustGetGACProfileNames returns a list of gimme-aws-creds profile names
func (a *Axolotl) MustGetGACProfileNames() []string {
	// Verify .okta_aws_login_config exists
	if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".okta_aws_login_config")); os.IsNotExist(err) {
		log.Fatalf(os.ExpandEnv("unable to locate .okta_aws_login_config in ${HOME}, please create it: %s\n\n\thttps://github.com/Nike-Inc/gimme-aws-creds#configuration"), err.Error())
	}

	// Parse .okta_aws_login_config for profile names
	f, err := os.Open(filepath.Join(os.Getenv("HOME"), ".okta_aws_login_config"))
	if err != nil {
		log.Fatalf("unable to open .okta_aws_login_config: %s", err.Error())
	}
	defer f.Close()

	profileNames := []string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			profileNames = append(profileNames, strings.Trim(line, "[]"))
		}
	}

	return profileNames
}

// AuthVerify checks if the user is authenticated and if not authenticates
// with gimme-aws-creds
func AuthVerify(enabled bool, profile Profile) error {
	if !enabled {
		return nil
	}

	// Check if aws cli is installed
	if _, err := exec.LookPath("aws"); err != nil {
		return fmt.Errorf("unable to locate `aws` in PATH, please install it: %w", err)
	}

	if canAuth(profile) {
		return nil
	}

	// If we are not authenticated, we will run gimme-aws-creds
	return AuthGimmeAwsCreds(profile)
}

// canAuth checks if the user is authenticated with the given profile
// TODO: This is really ugly and adds a performance hit for the aws cli call
// but there isn't a better way to verify credentials UNLESS this PR is merged
// to gimme-aws-creds: https://github.com/Nike-Inc/gimme-aws-creds/pull/300
func canAuth(profile Profile) bool {
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
	os.Setenv("AWS_PROFILE", profile.AWS)

	// Check if we are authenticated by running aws sts get-caller-identity
	// If we are not authenticated, we will get an error
	// If we are authenticated, we will get a json response
	// We will ignore the json response
	// restore environment
	cmd := exec.Command("aws", "sts", "get-caller-identity")
	err := cmd.Run()

	os.Clearenv()
	for _, e := range origEnv {
		os.Setenv(strings.Split(e, "=")[0], strings.Split(e, "=")[1])
	}

	if err == nil {
		return true
	}
	return false
}

// AuthGimmeAwsCreds authenticates with gimme-aws-creds
func AuthGimmeAwsCreds(profile Profile) error {
	// Check if gimme-aws-creds is installed
	if _, err := exec.LookPath("gimme-aws-creds"); err != nil {
		return fmt.Errorf("unable to locate `gimme-aws-creds` in PATH, please install it: %w\n\n\thttps://github.com/Nike-Inc/gimme-aws-creds#installation", err)
	}

	// Verify .okta_aws_login_config exists
	if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".okta_aws_login_config")); os.IsNotExist(err) {
		return fmt.Errorf(os.ExpandEnv("unable to locate .okta_aws_login_config in ${HOME}, please create it: %w\n\n\thttps://github.com/Nike-Inc/gimme-aws-creds#configuration"), err)
	}

	// execute gimme-aws-creds
	cmd := exec.Command("gimme-aws-creds", "--profile", profile.GimmeAWSCreds)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("unable to execute `gimme-aws-creds`: %w", err)
	}

	// Verify we are authenticated to AWS now that we have run gimme-aws-creds
	if canAuth(profile) {
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		home = "~"
	}

	// TODO: It might be worthwhile to try and handle this case better if enough
	// people run into it. For now, let's just return an error.
	return fmt.Errorf("unable to authenticate to AWS with profile %s. Check %s/.aws/credentials to ensure this profile exists", profile.AWS, home)
}

// ConfigureGlobals sets up the global flags and returns the global config
func ConfigureGlobals(app *kingpin.Application) *Axolotl {
	// Load config from file
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("fatal error config file: %v", err)
	}

	viper.GetBool("autoGimmeAwsCreds")
	viper.GetString("defaultRegion")
	viper.GetStringMapString("profiles")

	a := &Axolotl{
		autoGimmeAwsCreds: viper.GetBool("autoGimmeAwsCreds"),
		defaultRegion:     viper.GetString("defaultRegion"),
		profiles:          viper.GetStringMapString("profiles"),
	}

	var (
		noVerify      bool
		verify        bool
		defaultRegion string
	)

	app.Flag("debug", "Show debugging output").
		BoolVar(&a.Debug)

	app.Flag("verify", "Enable automatic credentials with `gimme-aws-creds`").
		BoolVar(&verify)

	app.Flag("no-verify", "Disable automatic credentials with `gimme-aws-creds`").
		BoolVar(&noVerify)

	app.Flag("default-region", "Set default AWS region").
		StringVar(&defaultRegion)

	app.PreAction(func(c *kingpin.ParseContext) error {
		if !a.Debug {
			log.SetOutput(io.Discard)
		}

		log.Printf("Axolotl %s", app.Model().Version)

		if noVerify {
			viper.Set("autoGimmeAwsCreds", false)
			if err := viper.WriteConfig(); err != nil {
				log.Fatalf("error writing config: %s", err.Error())
			}
			fmt.Println("Disabled automatic credentials with gimme-aws-creds")
			os.Exit(0)
		}

		if verify {
			viper.Set("autoGimmeAwsCreds", true)
			if err := viper.WriteConfig(); err != nil {
				log.Fatalf("error writing config: %s", err.Error())
			}
			fmt.Println("Enabled automatic credentials with gimme-aws-creds")
			os.Exit(0)
		}

		if defaultRegion != "" {
			viper.Set("defaultRegion", defaultRegion)
			if err := viper.WriteConfig(); err != nil {
				log.Fatalf("error writing config: %s", err.Error())
			}
			fmt.Println("Set default region to", defaultRegion)
			os.Exit(0)
		}

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

// go-prompt has a bug that hijacks the terminal state breaking signal handling.
// This function is a workaround to restore the terminal state until the bug is fixed.
// See:
// - https://github.com/c-bata/go-prompt/issues/233
// - https://github.com/c-bata/go-prompt/pull/239
var termState *term.State

func saveTermState() {
	oldState, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		return
	}
	termState = oldState
}

func restoreTermState() {
	if termState != nil {
		term.Restore(int(os.Stdin.Fd()), termState)
	}
}

// awsProfileCompleter returns a list of AWS profile names
func (a *Axolotl) awsProfileCompleter() func(d prompt.Document) []prompt.Suggest {
	return func(d prompt.Document) []prompt.Suggest {
		s := []prompt.Suggest{}
		for _, p := range a.MustGetAWSProfileNames() {
			s = append(s, prompt.Suggest{Text: p})
		}
		return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
	}
}

// gacProfileCompleter returns a list of gimme-aws-creds profile names
func (a *Axolotl) gacProfileCompleter() func(d prompt.Document) []prompt.Suggest {
	return func(d prompt.Document) []prompt.Suggest {
		s := []prompt.Suggest{}
		for _, p := range a.MustGetGACProfileNames() {
			s = append(s, prompt.Suggest{Text: p})
		}
		return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
	}
}
