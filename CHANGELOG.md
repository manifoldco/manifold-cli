# CHANGELOG

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## Unreleased

### Changed

- Reference to go-manifold updated to 0.9.1 release for verification code change

## [0.12.0] - 2017-11-30

### Added
- GitHub login flow

- Get available roles from the server

### Fixed

- Change verification code validation to support numeric codes

## [0.11.0] - 2017-11-27

### Added

- `--role` flag to `teams invite`
- `teams join` using an invitation token
- `create` cmd has search for product, plan and region
- `transfer` allows you to transfer resources from a team to a user and vice versa

### Fixed

- User display name accepts unicode letters

## [0.10.1] - 2017-11-15

### Fixed

- `os/user.Current` when cgo is disabled

## [0.10.0] - 2017-11-03

### Added
- Client-side name generation to `manifold create`
- `manifold alias` can rename configuration keys on a per-resource basis

## [0.9.3] - 2017-10-30

### Fixed
- Travis builds not picking up new versions

## [0.9.2] - 2017-10-30

### Added
- Support for creating and viewing resources from unlisted plans

### Fixed
- Resize will no longer crash due to uninitialized Analytics
- Ask for resource name/title after resource settings has been selected
- Disable spinner animation when piping output

## [0.9.1] - 2017-10-27

### Fixed
- Install script version has been updated

## [0.9.0] - 2017-10-26

### Added
- Improved visuals for select boxes

### Fixed
- Project create failure due to Name/Title change
- Set-role should error before prompts
- Context would panic due to missing session
- Change logic for inferring object title from name

## [0.8.1] - 2017-10-24

### Added
- `manifold teams members` can list members of a team
- `manifold teams invite` now supports Role selection
- `manifold teams set-role` can update the role of an existing team member
- `manifold teams remove` will remove a team member or revoke an invite
- `manifold context` command to display current account and context
- `manifold tokens add` will create a new API token
- `MANIFOLD_API_TOKEN` as means for authenticating

### Fixed
- Update success and failure output for `manifold config set`
- Prevent error during resize attempt of custom resource
- Amend output for NewUsageError to pull from framework help
- `teams create` can be used non-interatively
- Terminology has changed: Name becomes Title, Label becomes Name
- Remove -t from title flag to prevent panic

## [0.7.1] - 2017-10-12

### Fixed

- Homebrew install

## [0.7.0] - 2017-10-06

### Added

- release linux and darwin as `tar.gz`

### Fixed

- `config unset` accepts `--team` flag
- `config unset` fails if no key is passed
- `config unset` fails with a friendlier message

## [0.6.0] - 2017-10-03

### Added

- `manifold projects delete` can delete projects, provided they contain
  no resources
- Refactor loading of different API clients into `api` package
- `delete` supports project flag
- `sso` command for getting resource single sign-on link
- build for high sierra in brew
- `projects add` and `projects remove` event
- `services providers` to list providers
- `services products` to list products
- `services plans` to list plans
- install.sh script to download latest release

### Fixed

- `project add` fails when there is no project
- `project remove` should not list resources without project
- colors breaking alignment of tabwriter

## [0.5.1] - 2017-09-15

### Added

- Chained e-mail verification step to the end of the signup command

### Fixed

- `resize` message output

## [0.5.0] - 2017-09-14

### Added

- Group commands into categories for better help output
- Update sorting function to use `sort.Slice`
- Sort plan by price and name during resource creation
- `projects add` adds a resource to a project
- `projects create` creates new projects
- `projects list` lists projects
- `projects remove` removes a resource from a project
- `projects update` updates an existing project
- `verify $EMAIL_CODE` verifys users e-mail with the CLI
- `run` supports --project/-p
- `init` saves project context instead of app context
- `list` improved UI
- Add `reszie` for resizing a resource

### Fixed

- Fetching resources or operations with nil team returns personal account
- Operations not assigning team and project id
- SelectResource uses project label instead of app name
- `export` accepts project label instead of app name
- `export` lists resources without projects
- `list` accepts project flag instead of app name
- `list` now adheres to team context for operations
- `update` accepts project label instead of app name
- `view` shows project label instead of app name
- `view` shows resource still in provision

### Removed

- `apps` commands, replaced with `projects` commands
- `rename` command in favour of `update`
- Fix `--me` when in a team context

## [0.4.0] - 2017-08-31

### Added

- Add `billing redeem` command to redeem coupon codes
- Add `teams list` command to list all teams and # of members
- Conflict response when redeeming coupons
- Add `view` command to view resource details

### Fixed

- Optional params for redeeming coupons
- Updated spinners
- Update resource selects to be label-focused
- Make plugin execution architecture aware
- Output adjusted for team selector
- Ensure columns accept long names in Teams list

## [0.3.0] - 2017-08-29

### Added

- Allow creating custom resources. Custom resources have no backing product.
  Instead, they hold custom user provided config.
- Introduce `config set` and `config unset` for adding/updating/removing custom
  config on custom resources.
- Rename app command to apps
- Add 'teams create' to create a Team
- Add 'teams update' to update a Team
- Add 'teams leave' to leave a Team
- Add 'teams invite' to invite to a Team
- Add 'switch' to switch Teams, so interact with a team's resources
- Add --team and --me flags for operating under your account, or a team's
- Add --team and --me flags for billing
- Add `billing redeem` command to redeem coupon codes
- Rewrote README, adding installation instructions

### Fixed

- Add a newline after signup output, so it won't mess up the prompt.
- Credit Card number can be between 12 and 19 digits and cvv either 3 or 4.
- Plugin config initializes map before setting if it doesn't exist
- Plugin config defines the .manifold.yml path prior to saving if not defined

## [0.2.6] - 2017-08-16

### Fixed

- Fix release process tag selection.

## [0.2.5] - 2017-08-16

### Added

- Add delete command to delete resources.
- Add rename command to rename resources.

### Fixed

- Fix help output to display `manifold` instead of `manifold-cli`.

## [0.2.4] - 2017-08-15

### Added

- Add the update command, to modify and resize resources.

## [0.2.3] - 2017-08-10

### Fixed

- *Acutally* correct botched release for proper version. For real. Hopefully.

## [0.2.2] - 2017-08-10

### Fixed

- *Acutally* correct botched release for proper version

## [0.2.1] - 2017-08-10

### Fixed

- Correct botched release for proper version

## [0.2.0] - 2017-08-10

### Added

- Add signup command to signup with Manifold.
- Add init command to configure a directory with defaults.
- Add billing command to add/update billing profile.

### Fixed

- The built command line tool is now called `manifold`.
- Include a user agent header in requests.

## [0.1.0] - 2017-07-19

### Added

- The start of time with the brand new `manifold-cli` tool!
- Intoducing the ability to login and out of a session through `manifold-cli
  login` and `manifold-cli logout`.
- Enabling a user to login using `MANIFOLD_EMAIL` and `MANIFOLD_PASSWORD`
  through any command.
- Allowing a user to export all credentials or only those for a specific app
  through `manifold-cli export`.
- Allowing a user to start a process by having Manifold inject the credentials
  directly into the process at startup through the `manifold-cli run` command.
- Enabling a user to provision a resource using the `manifold-cli create`
  command with a wizard or via a script using flags and arguments.
