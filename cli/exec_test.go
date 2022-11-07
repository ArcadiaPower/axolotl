package cli_test

import (
	"testing"

	"github.com/ArcadiaPower/awswitch/cli"
	"github.com/alecthomas/kingpin"
	"github.com/stretchr/testify/assert"
)

func TestExecCommand(t *testing.T) {
	app := kingpin.New("awswitch", "")
	a := cli.ConfigureGlobals(app)
	cli.ConfigureExecCommand(app, a)
	_, err := app.Parse([]string{
		"--profile", "test", "--region", "us-east-1", "--", "sh", "-c", "echo $AWS_PROFILE",
	})

	assert.NoError(t, err)
}
