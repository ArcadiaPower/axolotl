# Developers, Developers, Developers

This document is intended to help developers get started with the project. It is not intended to be a comprehensive guide to the project, but rather a guide to getting started.

## Prerequisites

- [Go](https://golang.org/doc/install) 
  - The project is currently built with Go 1.19.X, but should work with newer (and to some extent older) versions of Go without trouble.
- [gimme-aws-creds](https://github.com/Nike-Inc/gimme-aws-creds)

Currently the project is only tested on macOS and Linux. It may work on Windows by virtue of the Go std library doing the heavy lifting for cross-platform support, but it is not officially supported and may require some additional work. If you're primarily a Windows developer and would like to help make this project work on Windows, please reach out to us.

## Installation for local development testing

There's a simple Makefile in the root of the project to help with local testing. By running `make install` you can install the binary to your `$HOME/.bin` directory. If you have this directory added to your `$PATH` you can then run the local build of `ax` from anywhere on your system.

> **Note** In addition to placing a copy of the binary at `$HOME/.bin/ax`, the Makefile also previously built the binary in the root of the repo. So if you don't have or want `$HOME/.bin` in your `$PATH`, you can still run the local build of `ax` by running `./ax` from the root of the repo.

To verify the installed binary is the local build, run `ax --version` and you should see a version number matching the latest release appended with `-dirty`. ex: `v1.x.x-dirty`

## Cutting a new release

The release pipeline is automated via GitHub Actions and GoReleaser. To cut a new release, simply create a new tag in the format `v1.x.x` off the latest commit to `main` and push it to the remote. The release pipeline will automatically build and publish the release to GitHub.

Example:

```bash
git tag v1.0.22
git push origin v1.0.22
```

## Questions?

If you have any questions, please make use of GitHub issues. There's no such thing as a bad question, so don't be afraid to ask!
