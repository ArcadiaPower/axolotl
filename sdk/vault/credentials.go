package vault

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

// CredentialsFile is an abstraction over what is in ~/.aws/credentials
type CredentialsFile struct {
	Path    string
	iniFile *ini.File
}

// ProfileSection is a profile section of the credentials file
type ProfileSection struct {
	Name string `ini:"-"`
	// These values are available in the credentials file, but we don't use them
	// AWSAccessKeyID     string `ini:"aws_access_key_id,omitempty"`
	// AWSSecretAccessKey string `ini:"aws_secret_access_key,omitempty"`
	// AWSSessionToken    string `ini:"aws_session_token,omitempty"`
	// AWSSecurityToken   string `ini:"aws_security_token,omitempty"`
}

func (c *CredentialsFile) parseFile() error {
	log.Printf("Parsing credentials file %s", c.Path)

	f, err := ini.LoadSources(ini.LoadOptions{
		AllowNestedValues:   true,
		InsensitiveSections: false,
		InsensitiveKeys:     true,
	}, c.Path)
	if err != nil {
		return fmt.Errorf("error parsing credentials file %s: %w", c.Path, err)
	}
	c.iniFile = f
	return nil
}

// ProfileSections returns all the profile sections in the credentials
func (c *CredentialsFile) ProfileSections() []ProfileSection {
	result := []ProfileSection{}

	if c.iniFile == nil {
		return result
	}

	for _, section := range c.iniFile.SectionStrings() {
		profile, _ := c.ProfileSection(section)
		result = append(result, profile)
	}

	return result
}

// ProfileSection returns the profile section with the matching name. If there isn't any,
// an empty profile with the provided name is returned, along with false.
func (c *CredentialsFile) ProfileSection(name string) (ProfileSection, bool) {
	profile := ProfileSection{
		Name: name,
	}
	if c.iniFile == nil {
		return profile, false
	}

	section, err := c.iniFile.GetSection(name)
	if err != nil {
		return profile, false
	}
	if err = section.MapTo(&profile); err != nil {
		panic(err)
	}
	return profile, true
}

// ProfileNames returns a slice of profile names from the AWS credentials
func (c *CredentialsFile) ProfileNames() []string {
	profileNames := []string{}
	for _, profile := range c.ProfileSections() {
		profileNames = append(profileNames, profile.Name)
	}
	return profileNames
}

// LoadCredentials loads and parses a credential file. No error is returned if the file doesn't exist
func LoadCredentials(path string) (*CredentialsFile, error) {
	credentials := &CredentialsFile{
		Path: path,
	}
	if _, err := os.Stat(path); err == nil {
		if parseErr := credentials.parseFile(); parseErr != nil {
			return nil, parseErr
		}
	}
	return credentials, nil
}

// LoadCredentialsFromEnv finds the credential file from the environment
func LoadCredentialsFromEnv() (*CredentialsFile, error) {
	file, err := credentialPath()
	if err != nil {
		return nil, err
	}

	log.Printf("Loading credential file %s", file)
	return LoadCredentials(file)
}

// credentialPath returns either $AWS_SHARED_CREDENTIALS_FILE or ~/.aws/credentials
func credentialPath() (string, error) {
	file := os.Getenv("AWS_SHARED_CREDENTIALS_FILE")
	if file == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		file = filepath.Join(home, "/.aws/credentials")
	} else {
		log.Printf("Using AWS_SHARED_CREDENTIALS_FILE value: %s", file)
	}
	return file, nil
}
