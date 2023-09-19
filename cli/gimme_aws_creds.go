package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/pkg/browser"
)

const (
	gimmeAWSCredsURL    = "https://github.com/Nike-Inc/gimme-aws-creds#installation"
	gimmeAWSCredsConfig = ".okta_aws_login_config"
)

func checkGimmeAwsCredsInstallation() error {
	if _, err := exec.LookPath("gimme-aws-creds"); err != nil {
		return fmt.Errorf("unable to locate `gimme-aws-creds` in PATH, please install it: %w\n\n\t%s", err, gimmeAWSCredsURL)
	}
	return nil
}

func verifyConfigFileExists() error {
	configFilePath := filepath.Join(os.Getenv("HOME"), gimmeAWSCredsConfig)
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return fmt.Errorf("unable to locate %s in ${HOME}, please create it: %w\n\n\t%s", gimmeAWSCredsConfig, err, gimmeAWSCredsURL)
	}
	return nil
}

func executeGimmeAwsCreds(profile Profile) error {
	cmd := exec.Command("gimme-aws-creds", "--profile", profile.GimmeAWSCreds)
	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()
	cmd.Stdin = os.Stdin

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("unable to start `gimme-aws-creds`: %w", err)
	}

	var wg sync.WaitGroup
	urlRegex := regexp.MustCompile(`https://[a-zA-Z0-9.-/=?_]+`)
	wg.Add(2) // Two goroutines for stdout and stderr

	readPipe := func(pipeName string, pipe io.Reader) {
		defer wg.Done()

		scanner := bufio.NewScanner(pipe)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line) // Echo the line back to stdout

			domains := []string{"oktapreview.com", "okta.com"}

			var hasOktaURL bool
			for _, domain := range domains {
				if strings.Contains(line, domain) {
					hasOktaURL = true
					break
				}
			}

			if hasOktaURL {
				urlMatch := urlRegex.FindString(line)
				if urlMatch != "" {
					if err := browser.OpenURL(urlMatch); err != nil {
						fmt.Printf("Failed to open URL: %s\n", err)
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Error reading %s: %s\n", pipeName, err)
		}
	}

	go readPipe("stdout", stdoutPipe)
	go readPipe("stderr", stderrPipe)

	wg.Wait()

	// Wait for command to finish
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error executing `gimme-aws-creds`: %w", err)
	}

	return nil
}
