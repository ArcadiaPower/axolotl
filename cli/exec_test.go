package cli_test

import (
	"testing"

	"github.com/alecthomas/kingpin"
	"github.com/joepurdy/awswitch/cli"
	"github.com/stretchr/testify/assert"
)

func TestExecCommand(t *testing.T) {
	app := kingpin.New("awswitch", "")
	cli.ConfigureExecCommand(app, &cli.Awswitch{})
	_, err := app.Parse([]string{
		"exec", "--profile", "test", "--region", "us-east-1", "--", "sh", "-c", "echo $AWS_PROFILE",
	})

	assert.NoError(t, err)
}
