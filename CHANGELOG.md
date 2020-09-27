# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)

## [1.5.0] - unrealeased
### Added
- gitea support; thanks @6543
- skip forks flag; thanks @6543
### Changed
### Deprecated
### Removed
### Fixed
- github example
### Security

## [1.4.0] - 09/4/20
### Added
- GHORG_GITHUB_TOPICS to filter cloning repos matching specified topics; thanks @ryanaross
- GHORG_MATCH_PREFIX to filter cloning repos by prefix
- example commands directory
- base-url to github for self hosted github instances
- github client mocks
### Changed
### Deprecated
### Removed
- GHORG_GITLAB_DEFAULT_NAMESPACE
### Fixed
- Bug with trailing slash env vars; thanks @ryanaross
### Security

## [1.3.1] - 07/11/20
### Added
- ascii time output when users use ghorgignore
- ascii time output when users use output-dir flag
### Changed
- default GHORG_ABSOLUTE_PATH_TO_CLONE_TO to $HOME/Desktop/ghorg
### Deprecated
### Removed
### Fixed
### Security

## [1.3.0] - 07/11/20
### Added
- auto downcase of ghorg clone folder name; thanks @zamariola
- auto underscore of ghorg clone folder name
- vendoring of dependencies
- go modules
- easter egg
### Changed
- ghorg configuration location to $HOME/.config/ghorg or $XDG_CONFIG_HOME https://github.com/gabrie30/ghorg/issues/65; thanks @liljenstolpe
### Deprecated
### Removed
### Fixed
- version number to 1.3.0; thanks @alexcurtin
### Security
- reset remote to not include apikey https://github.com/gabrie30/ghorg/issues/64; thanks @mcinerney

## [1.2.2] - 04/06/20
### Added
### Changed
### Deprecated
### Removed
### Fixed
- GitLab client; thanks @awesomebytes
### Security

## [1.2.1] - 03/02/20
### Added
### Changed
### Deprecated
### Removed
### Fixed
- Gitlab https clone url to include oauth2 https://stackoverflow.com/questions/25409700/using-gitlab-token-to-clone-without-authentication
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
