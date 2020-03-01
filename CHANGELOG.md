# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)

## [Unreleased] - DATE
### Added
### Changed
### Deprecated
### Removed
### Fixed
### Security

## [1.2.0] - 02/29/20
### Added
- auto add trailing slash to path to clone to if not supplied by user
- add token to https clone urls
- add concurrency flag to limit goroutines while cloning
### Changed
### Deprecated
### Removed
### Fixed
### Security

## [1.1.10] - 01/19/20
### Added
- perserve dir flag for gitlab
### Changed
### Deprecated
### Removed
### Fixed
### Security

## [1.1.9] - 08/30/19
### Added
- color flag to toggle colorful output
- scm base url flag for gitlab
- ghorgignore to ignore specific repos
- skip archived repos flag
- how to fix ulmits in readme
- dedicated backup flag
### Changed
### Deprecated
### Removed
### Fixed
### Security

## [1.1.8] - 08/03/19
### Added
- Config output for clone
### Changed
### Deprecated
### Removed
### Fixed
### Security

## [1.1.7] - 08/03/19
### Added
### Changed
- version for homebrew
### Deprecated
### Removed
### Fixed
### Security

## [1.1.6] - 08/03/19
### Added
### Changed
### Deprecated
### Removed
### Fixed
- setting all envs from conf
### Security

## [1.1.5] - 08/03/19
### Added
- Tests for configs
### Changed
- Error messages
### Deprecated
### Removed
### Fixed
### Security

## [1.1.4] - 08/03/19
### Added
### Changed
### Deprecated
### Removed
### Fixed
- token verification
### Security

## [1.1.3] - 08/03/19
### Added
### Changed
### Deprecated
### Removed
### Fixed
- flag documentation
### Security


## [1.1.2] - 08/03/19
### Added
### Changed
- readme
### Deprecated
### Removed
### Fixed
- flag documentation
### Security

## [1.1.1] - 08/02/19
### Added
### Changed
- readme
### Deprecated
### Removed
### Fixed
- flags for certain commands
### Security

## [1.0.10] - 07/28/19
### Added
- gitlab support
- bitbucket support
### Changed
- readme
### Deprecated
### Removed
### Fixed
- ghorg conf file env's being overwritten
### Security

## [1.0.9] - 07/25/19
### Added
- viper/cobra for more robust cli commands and flags
### Changed
- readme
- .ghorg to $HOME/ghorg/conf.yaml
### Deprecated
### Removed
### Fixed
### Security

## [1.0.8] - 12/8/18
### Added
- changelog
- when no org is found default to search for username instead
- clone protocol to .ghorg to allow for https or ssh cloning
### Changed
- readme
### Deprecated
### Removed
### Fixed
### Security

## [1.0.0] - 05.26.2018
### Added
- initial version of `ghorg`
