# axolotl

![_Axolotl_](https://i.imgur.com/wcOZg4d.jpg)

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/ArcadiaPower/axolotl?style=for-the-badge)![GitHub release (latest by date)](https://img.shields.io/github/v/release/ArcadiaPower/axolotl?style=for-the-badge)![GitHub](https://img.shields.io/github/license/ArcadiaPower/axolotl?style=for-the-badge)

axolotl (`ax`) is an opinionated CLI that minimally emulates the behavior of the `aws-vault exec` command to run ad-hoc commands or switch to a subshell of a specific AWS profiles on the fly.

Additionally, credentials are obtained automatically using `gimme-aws-creds` to make a simple one command workflow for switching AWS profiles and credentials. This behavior can be disabled by running `ax --no-verify` and re-enabled by `ax --verify`.

## Prerequisites

If you're installing direct from source instead of with homebrew and you want the default automatic credential acquisition to work you'll need to install [gimme-aws-creds] yourself. This is automatically taken care of when installing with homebrew.

## Installation

This is a Go CLI and as such can be installed the standard Go way if you have a working Go installation. A homebrew package is automatically provided for tagged releases if you don't have or want Go installed on your computer.

Install with `go install`
```bash
go install github.com/ArcadiaPower/axolotl@latest
```

__OR__

Install with homebrew
```bash
brew tap ArcadiaPower/tap
brew install ArcadiaPower/tap/axolotl
```

Note: Installing with homebrew has the added benefit of automatically installing `gimme-aws-creds` as a dependency if it wasn't already installed.

## Configuration

The configuration file is created automatically at `$HOME/.config/ax/config.yaml` if it doesn't already exist. 
- autogimmeawscreds - This enables automatic credential verification and acquisition with `gimme-aws-creds`, the default is true.
- defaultregion - This sets the default AWS region that will be used, the default is `us-east-1`.

## Usage

To switch to a named profile and the default AWS Region of `us-east-1`:
```bash
ax --profile example-staging
```

To switch to a named profile and a specific AWS Region:
```bash
ax --profile example-staging --region us-west-2
```

Same as the last example, but use short flags:
```bash
ax -p example-staging -r us-west-2
```

Execute a single command using a named profile:
```bash
ax -p example-staging -- aws sts get-caller-identity
```

Change the default region to `us-west-2` and execute a single command using a named profile without having to specify the region:
```bash
ax --default-region us-west-2

ax -p example-staging -- aws sts get-caller-identity
```

### Profile Completion

If you run `ax` without passing any arguments the tool provides autocomplete and tabcomplete functionality based on the profile names in your local `~/.aws/credentials` file or the file specified by the `$AWS_SHARED_CREDENTIALS_FILE` environment variable if set.

## Credit and Why Yet Another Tool

This tool exists thanks to the inspiration of far greater utilities, specifically [aws-vault], [saml2aws], and [gimme-aws-creds]. It's born out of a need for a workflow to authenticate many AWS accounts via Okta SSO and solves a specific niche that the existing tools didn't quite cover. 

I wanted the simplicity of the `aws-vault exec` command with the requirement for Okta based SAML authentication which wasn't an option because the authors of `aws-vault` recommend other tools like `saml2aws` for obtaining credentials through a SAML provider: https://github.com/99designs/aws-vault/issues/235

`gimme-aws-creds` was a better fit than `saml2aws` for obtaining the credentials since it allows getting credentials for all profiles rather than one by one. This tool simply recreates a minimal version of `aws-vault exec` with `gimme-aws-creds` as the mechanism for obtaining credentials.

## License

ax is released under the [MIT License](https://opensource.org/licenses/MIT)

[aws-vault]: https://github.com/99designs/aws-vault
[saml2aws]: https://github.com/Versent/saml2aws
[gimme-aws-creds]: https://github.com/Nike-Inc/gimme-aws-creds