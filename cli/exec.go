package cli

import (
	"fmt"
	"os"
	"syscall"

	"github.com/alecthomas/kingpin"
	"github.com/c-bata/go-prompt"
	osexec "golang.org/x/sys/execabs"
)

type ExecCommandInput struct {
	ProfileName string
	Region      string
	Command     string
	Args        []string
	Verify      bool
}

func ConfigureExecCommand(app *kingpin.Application, a *Axolotl) {
	input := ExecCommandInput{
		Verify: a.autoGimmeAwsCreds,
	}

	app.Flag("profile", "The AWS profile to execute as").
		Short('p').
		HintAction(a.MustGetProfileNames).
		StringVar(&input.ProfileName)

	app.Flag("region", "The AWS region to execute to").
		Default("us-east-1").
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

		if input.ProfileName == "" {
			saveTermState()
			fmt.Println("Please select profile.")
			input.ProfileName = prompt.Input("> ", a.profileCompleter())
			restoreTermState()
		}

		return ExecCommand(input)
	})
}

func ExecCommand(input ExecCommandInput) error {
	env := environ(os.Environ())
	env.Set("AWS_DEFAULT_PROFILE", input.ProfileName)
	env.Set("AWS_PROFILE", input.ProfileName)
	env.Set("AWS_REGION", input.Region)
	env.Set("AWS_AXOLOTL", "42")

	if err := AuthVerify(input.Verify, input.ProfileName); err != nil {
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
