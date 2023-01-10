package cli

import (
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/alecthomas/kingpin"
	"github.com/c-bata/go-prompt"
	"github.com/spf13/viper"
	osexec "golang.org/x/sys/execabs"
)

// Profile represents a profile mapping between AWS and gimme-aws-creds profiles
type Profile struct {
	AWS           string // AWS profile name
	GimmeAWSCreds string // gimme-aws-creds profile name
}

type ExecCommandInput struct {
	Profile    Profile
	Region     string
	Command    string
	Args       []string
	Verify     bool
	AutoRegion bool
}

func ConfigureExecCommand(app *kingpin.Application, a *Axolotl) {
	input := ExecCommandInput{
		Verify: a.autoGimmeAwsCreds,
	}

	app.Flag("profile", "The AWS profile to execute as").
		Short('p').
		HintAction(a.MustGetAWSProfileNames).
		StringVar(&input.Profile.AWS)

	app.Flag("region", "The AWS region to execute to").
		Default(a.defaultRegion).
		Short('r').
		HintOptions("us-east-1", "us-west-2").
		StringVar(&input.Region)

	app.Arg("cmd", "The command to run, defaults to $SHELL").
		Default(os.Getenv("SHELL")).
		StringVar(&input.Command)

	app.Arg("args", "The arguments to pass to the command").
		StringsVar(&input.Args)

	app.Action(func(c *kingpin.ParseContext) error {
		if os.Getenv("AWS_AXOLOTL") != "" {
			return fmt.Errorf("ax sessions should be nested with care, unset AWS_AXOLOTL to force")
		}

		if input.Profile.AWS == "" {
			saveTermState()
			fmt.Println("Please select AWS profile.")
			input.Profile.AWS = prompt.Input("> ", a.awsProfileCompleter())
			restoreTermState()
		}

		var ok bool
		input.Profile.GimmeAWSCreds, ok = a.profiles[input.Profile.AWS]
		if !ok {
			gacProfiles := a.MustGetGACProfileNames()
			if len(gacProfiles) == 1 {
				input.Profile.GimmeAWSCreds = gacProfiles[0]
			} else {
				saveTermState()
				fmt.Println("Please select gimme-aws-creds profile.")
				input.Profile.GimmeAWSCreds = prompt.Input("> ", a.gacProfileCompleter())
				restoreTermState()
			}

			// save the mapping for next time
			a.profiles[input.Profile.AWS] = input.Profile.GimmeAWSCreds
			viper.Set("profiles", a.profiles)
			if err := viper.WriteConfig(); err != nil {
				log.Fatalf("error writing config file: %s", err.Error())
			}
		}

		return ExecCommand(input)
	})
}

func ExecCommand(input ExecCommandInput) error {
	env := environ(os.Environ())
	env.Set("AWS_DEFAULT_PROFILE", input.Profile.AWS)
	env.Set("AWS_PROFILE", input.Profile.AWS)
	env.Set("AWS_DEFAULT_REGION", input.Region)
	env.Set("AWS_REGION", input.Region)
	env.Set("AWS_AXOLOTL", "42")

	if err := AuthVerify(input.Verify, input.Profile); err != nil {
		return err
	}

	argv0, err := osexec.LookPath(input.Command)
	if err != nil {
		return fmt.Errorf("couldn't find the executable '%s': %w", input.Command, err)
	}

	argv := make([]string, 0, 1+len(input.Args))
	argv = append(argv, input.Command)
	argv = append(argv, input.Args...)

	if err := syscall.Exec(argv0, argv, env); err != nil {
		return fmt.Errorf("failed to execute '%s': %w", argv0, err)
	}

	return nil
}
