# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)

## [1.8.1] - unreleased
### Added
- Reclone command
- Gitea token check
### Changed
- Simplified token check
### Deprecated
### Removed
### Fixed
### Security

## [1.8.0] - 6/25/22
### Added
- Exit 1 when any issue messages are produced; thanks @i3v
- GHORG_EXIT_CODE_ON_CLONE_INFOS to allow for control of exit code when any info messages are produced
- GHORG_EXIT_CODE_ON_CLONE_ISSUES to allow for control of exit code when any issue messages are produced
- Remotes updated count to clone stats
### Changed
### Deprecated
### Removed
### Fixed
- Backup flag not working with prune; thanks @i3v
- Clone wiki flag with prune; thanks @i3v
- Cloning of Gitlab subgroups with prune; thanks @i3v
### Security

## [1.7.16] - 6/1/22
### Added
- GHORG_PRUNE setting which allows a user to have Ghorg automatically remove items from their local
  org clone which have been removed (or archived, if GHORG_SKIP_ARCHIVED is set) upstream; thanks @toothbrush
- GHORG_PRUNE_NO_CONFIRM which disables the interactive yes/no prompt for every item to be deleted
  when pruning; thanks @toothbrush
### Changed
### Deprecated
### Removed
### Fixed
### Security

## [1.7.15] - 5/29/22
### Added
- CodeQL security analysis action
- Security policy
- Ghorg version in run details output
### Changed
### Deprecated
### Removed
### Fixed
### Security

## [1.7.14] - 5/25/22
### Added
### Changed
### Deprecated
### Removed
### Fixed
### Security
- Upgrade all dependencies, including gopkg.in/yaml.v3 to 3.0.0

## [1.7.13] - 4/17/22
### Added
### Changed
### Deprecated
### Removed
### Fixed
- User cloning self on github now finds all public/private repos; thanks @iDoka
### Security

## [1.7.12] - 5/19/22
### Added
- Dockerfile
- Ability to set GHORG_IGNORE_PATH; thanks @jeffreylo
- Ability to set configuration file as `ghorg.yaml` at root of repo; thanks @jeffreylo
- GHORG_QUIET mode; thanks @jeffreylo
### Changed
### Deprecated
### Removed
### Fixed
### Security

## [1.7.11] - 3/5/22
### Added
- Automatic env setting for viper to allow for overriding env vars; thanks @hojerst
- Exclude gitlab groups by regex match; thanks @schelhorn
### Changed
### Deprecated
### Removed
### Fixed
### Security

## [1.7.10] - 3/2/22
### Added
### Changed
### Deprecated
### Removed
### Fixed
- Configuration env var; thanks @Rabattkarte
### Security

## [1.7.9] - 2/27/22
### Added
- Bitbucket base url support
### Changed
### Deprecated
### Removed
### Fixed
- Deprecated bitbucket api calls; thanks @Riduidel
### Security

## [1.7.8] - 2/14/22
### Added
### Changed
### Deprecated
### Removed
### Fixed
- GitLab `--preserve-dir` flag not being respected; thanks @attachmentgenie
### Security

## [1.7.7] - 2/12/22
### Added
- Filtering repos by topics for gitlab; thanks @dschafhauser
- Exclude filtering for prefix and regex; thanks @dschafhauser
### Changed
### Deprecated
### Removed
### Fixed
### Security
## [1.7.6] - 1/15/22
### Added
- goreleaser
- GHORG_BITBUCKET_OAUTH_TOKEN to support oauth tokens for bitbucket; thanks @
skupjoe
### Changed
### Deprecated
### Removed
### Fixed
- Gitlab token length requirements; thanks @dschafhauser
- Appending trailing slashes on urls and filepaths; thanks @dschafhauser
### Security

## [1.7.5] - 12/11/21
### Added
- GHORG_DRY_RUN to do dry runs on clones
- GHORG_FETCH_ALL to run fetch all on each repo
- output for long running repo fetches
- support for cloning github enterprise repos
- log repos cloned vs pulled at end of run
### Changed
- go-github versions v32 -> v41
### Deprecated
### Removed
### Fixed
- Setting new gitlab token check from config file; thanks @vegas1880
### Security
## [1.7.4] - 11/11/21
### Added
- GHORG_CLONE_WIKI to clone wiki pages of repos; thanks @ahmadalli
### Changed
### Deprecated
### Removed
### Fixed
- Setting new gitlab token check from config file; thanks @vegas1880
### Security

## [1.7.3] - 11/1/21
### Added
- GHORG_INSECURE_GITLAB_CLIENT to to skip verification of ssl certificates for hosted gitlab servers
### Changed
### Deprecated
### Removed
### Fixed
- Gitlab token length validation; thanks @vegas1880
### Security
## [1.7.2] - 10/14/21
### Added
- GHORG_NO_CLEAN only clones new repos and does not perform a git clean on existing repos; thanks @harmathy
### Changed
### Deprecated
### Removed
### Fixed
### Security

## [1.7.1] - 9/27/21
### Added
- all-groups for cloning all groups on a hosted gitlab instance
### Changed
- go version in go.mod to 1.17 and updated all dependencies
### Deprecated
### Removed
### Fixed
- Pagination with gitlab cloud; thanks @brenwhyte
### Security

## [1.7.0] - 9/2/21

> Big thanks to @cugu and @Code0x58

### Added
- integration tests on windows, ubuntu, and mac for github
- GHORG_MATCH_REGEX to filter cloning repos by regex; thanks @petomalina
### Changed
- initial clone will try to checkout a branch if specified; thanks @dword-design
- default clone directory to $HOME/ghorg
- users/orgs directory no longer appends "\_ghorg" or forces underscores
- make $HOME/.config/ghorg/conf.yaml optional
- color is off by default
- color flag configuration options are enabled/disabled
### Deprecated
### Removed
### Fixed
- file pathing to be windows compatible
### Security
## [1.6.0] - 4/9/21
### Added
### Changed
- how github users clone their own repos thanks @dword-design
### Deprecated
### Removed
### Fixed
### Security

## [1.5.2] - 4/5/21
### Added
- ghorg clone me to clone all of your own private repos from github
### Changed
### Deprecated
### Removed
### Fixed
### Security

## [1.5.1] - 3/4/21
### Added
### Changed
- error messages for ls command
- GHORG_BRANCH if not set, will first look for the repos default branch, if no default branch is found on repo will fall back to using master as default
### Deprecated
### Removed
### Fixed
### Security

## [1.5.0] - 10/31/20
### Added
- gitea support; thanks @6543
- skip forks flag; thanks @6543
- ls command
- scm package
### Changed
- go-gitlab version 0.33.0 -> 0.38.2
### Deprecated
### Removed
### Fixed
- dry'd scm user/org filter; thanks @6543
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
