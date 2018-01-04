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

## Configuration Files

Manifold CLI has two files that help you codify some of your manifold settings and stacks, `.manifold.yml` for context configuration and `stack.yml` for stack configuration.  Neither are required, but can augment / simplify your use of Manifold depending on your use cases.  

### .manifold.yml

This hidden file can be created by a `manifold init` and keeps track of context, specifically:
- `project`
- `team` 

By doing so you no longer have to pass this in for an arg, nor worry about switching context via the `manifold switch` command.  It can be overwritten via `-t` for team and `-p` for project however.  

This file can be checked into your source control as it contains no secrets, but keep in mind for open source projects it won't be useful to others as only you have access  to your team and project.

### stack.yml

This file can initialized via `manifold stack init` and keeps track of your stack, i.e what services your code base requires to exist in order to function. Example being a db, logger, etc.  You can manually craft this file, or you can use any of the following commands to create, modify and apply the stack file:

This file can be checked into your source control as it contains no secrets.  The applying of a stack is linked to the project and team of which the command is run, so it is safe to use the same or similar stack.yml files in multiple repositories.

- `manifold stack init` - Prompts for a project name and create an empty stack.yml file.  Optional arguments are:
    - `(-t TEAM)`
    - `(-p PROJECT)`
    - `(-g)`generates the yaml file, interactively iterating over the already existing services, prompting if the user would like them added to the `stack.yml` file or not.
- `manifold stack add OPTIONAL_RESOURCE_NAME` - Add services to the yaml file, allows for specifying a service by name, or select from a list.  Acts similar to the `manifold create` command, however the `add` command does no actual provisioning, instead its results in an update to your `stack.yml` file.  Optional arguments are:
  - `( -product PRODUCT_NAME )`
  - `( -plan PLAN_NAME )`
  - `( -region REGION_NAME )`
  - `( -title RESOURCE_TITLE )`
- `manifold stack remove` - remove a service from the yaml file, can be specified as an argument, if non given, user is asked to select from a list.  `manifold stack remove OPTIONAL_RESOURCE_NAME`. 
- `manifold stack plan` - reads the yaml file, compares it against what the context (team / project) currently has and presents the changes the yaml file would create
- `manidold stack apply` - Applies the stack defined in the yaml file interactively to the current context.  Will prompt to confirm project name and every change to be applied.  Optional arguments are:
  - `(-t TEAM)`
  - `(-p PROJECT)`
  - `(--yes)` auto say yes to all prompts, VERY DANGEROUS


Example `stack.yml` file:
```
project: my-new-project
resources:
  email:
    title: Email Service
    product: mailgun
    plan: musket
    region: global
  logger:
    title: Logging Service
    product: logdna
    plan: zepto
    region: global
  my-first-database:
    title: My First Database
    product: jawsdb-maria
    plan: leopard
    region: aws-us-east1

```

## License

Manifold's manifold-cli is released under the [BSD 3-Clause License](./LICENSE.md).