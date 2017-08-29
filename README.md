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

#### Homebrew (OS X)

Homebrew can be installed via [brew.sh](http://brew.sh)

```
$ brew install manifoldco/brew/manifold-cli
```

#### Zip Archives (OS X, Linux, Windows)

Bare zip archives per release version are available on https://releases.manifold.co

For instructions on Windows, [click here](./.github/WINDOWS.md).


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


## License

Manifold's manifold-cli is released under the [BSD 3-Clause License](./LICENSE.md).
