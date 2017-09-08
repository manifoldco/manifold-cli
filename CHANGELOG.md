# CHANGELOG

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

### Added

- Add `projects create` command to create new projects
- Add `projects update` command to update an existing project
- Group commands into categories for better help output
- Add `projects add` to add a resource to a project

### Fixed

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
