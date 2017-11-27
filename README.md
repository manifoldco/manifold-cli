# manifold-cli

Manage your cloud services like a developer.

[Homepage](https://manifold.co) |
[Twitter](https://twitter.com/manifoldco) |
[Code of Conduct](./CODE_OF_CONDUCT.md) |
[Contribution Guidelines](./.github/CONTRIBUTING.md)

[![Travis](https://img.shields.io/travis/manifoldco/manifold-cli/master.svg)](https://travis-ci.org/manifoldco/manifold-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/manifoldco/manifold-cli)](https://goreportcard.com/report/github.com/manifoldco/manifold-cli)
[![License](https://img.shields.io/badge/license-BSD-blue.svg)](./LICENSE.md)


## Installation

### Install script

To install or update manifold, you can use the install script using cURL:

```
curl -o- https://raw.githubusercontent.com/manifoldco/manifold-cli/master/install.sh | sh
```

You can customize the install directory, profile, and version using the
`MANIFOLD_DIR`, `PROFILE` and `MANIFOLD_VERSION` variables. Eg: `curl ... |
MANIFOLD_DIR=/usr/local sh` for a global install.

### Homebrew (OS X)

Homebrew can be installed via [brew.sh](http://brew.sh)

```
$ brew install manifoldco/brew/manifold-cli
```

### Zip Archives (OS X, Linux, Windows)

Bare zip archives per release version are available on https://releases.manifold.co

For instructions on Windows, [click here](./.github/WINDOWS.md).


## Autocomplete

If you have bash and bash-completion installed, you can enable autocomplete with:
```
curl -o- https://raw.githubusercontent.com/manifoldco/manifold-cli/master/autocomplete.sh | sh
```

## Quickstart

First you must create an account.

```
$ manifold signup
```

Then you can create your first resource.

```
$ manifold create
```

Followed by running your process with the appropriate ENV.

```
$ manifold run ./bin/server
```

## OAuth Identity

Manifold gives you the ability to login to using a third-party OAuth provider such as GitHub.

### GitHub Authentication

To use the GitHub login, you must create an OAuth App within GitHub. To do this, go to your
[developer settings](https://github.com/settings/developers) page, and click the "New OAuth App"
button. Use the following configuration settings:

| Name                       | Value                                  |
|----------------------------|----------------------------------------|
| Application Name           | Manifold CLI                           |
| Homepage URL               | https://manifold.co                    | 
| Application Description    | Manifold CLI GitHub Login              |
| Authorization callback URL | http://127.0.0.1:49152/github/callback |


The GitHub Client ID is compiled in using the `MANIFOLD_GITHUB_CLIENT_ID` environment variable.

## License

Manifold's manifold-cli is released under the [BSD 3-Clause License](./LICENSE.md).
