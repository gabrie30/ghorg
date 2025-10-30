# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)

## [1.11.5] - unreleased
### Added
- Self hosted Bitbucket support
- `ghorgonly` file support to clone only repos matching substring patterns (include-only filter); thanks @fgouteroux
### Changed
- Empty interface to any
### Deprecated
### Removed
### Fixed
- File path in tokens; thanks @blackcodersec
### Security

## [1.11.4] - 8/29/25
### Added
- Total clone time to ghorg stats
- Gitea integration test suite
- Clone delay option with auto concurrency adjustment; thanks @batagy
### Changed
- GitLab clone examples and notes
### Deprecated
### Removed
### Fixed
- Gitlab prune and preseve dir on windows; thanks @derekhes
- Directory structure of repos and wikis cloned from users on Gitlab; thanks @MickaelFontes
- GHORG_GITLAB_GROUP_EXCLUDE_MATCH_REGEX to work on subgroups; thanks @deepzeafish
- GHORG_FETCH_ALL is incompatible with cloning via fine-grained token over https; thanks @brainplot
### Security
- Bump go-github from v62 to v72; thanks @alexandre-orca
- Bump golang.org/x/crypto in /scripts/local-gitea/test-runner (#565)
- Bump github.com/hashicorp/go-retryablehttp (#555)
- Bump golang.org/x/oauth2 in /scripts/local-gitlab/seeder (#553)
- Bump github.com/go-viper/mapstructure/v2 from 2.3.0 to 2.4.0 (#550)

## [1.11.3] - 7/4/25
### Added
### Changed
### Deprecated
### Removed
### Fixed
- Reclone cron zombie processes; thanks @steinbrueckri
### Security
- Bump github.com/ktrysmt/go-bitbucket from 0.9.85 to 0.9.86 (#540)

## [1.11.2] - 6/29/25
### Added
- Post exec script to reclone cmd; thanks @steinbrueckri
### Changed
### Deprecated
### Removed
### Fixed
- Gitlab prune not working with preserve dir; thanks @Fraguinha
### Security
- Bump github.com/go-viper/mapstructure/v2 from 2.2.1 to 2.3.0 (#539)
- Bump github.com/ktrysmt/go-bitbucket from 0.9.83 to 0.9.85 (#537)
- Bump github.com/bradleyfalzon/ghinstallation/v2 from 2.15.0 to 2.16.0 (#536)
- Bump golang.org/x/oauth2 from 0.29.0 to 0.30.0 (#535)

## [1.11.1] - 5/1/25
### Added
### Changed
### Deprecated
### Removed
### Fixed
- Gitea internal repos token injection on https clones; thanks @drmacdon
- BitBucket backup with prune removes repositories just cloned; thanks @dannuralogix
- Empty repo clones no longer error; thanks @efasel
### Security
- Bump golang.org/x/net from 0.36.0 to 0.38.0 (#510)
- Bump github.com/spf13/viper from 1.19.0 to 1.20.1 (#506)
- Bump github.com/ktrysmt/go-bitbucket from 0.9.81 to 0.9.82 (#507)
- Bump golang.org/x/oauth2 from 0.27.0 to 0.28.0 (#505)
- Bump github.com/golang-jwt/jwt/v4 from 4.5.1 to 4.5.2 (#498)
- Bump golang.org/x/net from 0.33.0 to 0.36.0 (#495)
- Bump github.com/spf13/cobra from 1.8.1 to 1.9.1 (#492)
- Bump golang.org/x/oauth2 from 0.25.0 to 0.27.0 (#491)
- Bump github.com/bradleyfalzon/ghinstallation/v2 from 2.13.0 to 2.14.0 (#490)
- Bump golang.org/x/oauth2 from 0.24.0 to 0.25.0 (#485)
- Bump github.com/bradleyfalzon/ghinstallation/v2 from 2.12.0 to 2.13.0 (#486)
- Bump code.gitea.io/sdk/gitea from 0.19.0 to 0.20.0 (#487)
- Bump github.com/briandowns/spinner from 1.23.1 to 1.23.2 (#488)
- Bump golang.org/x/net from 0.25.0 to 0.33.0 (#484)
- Bump github.com/xanzy/go-gitlab from 0.114.0 to 0.115.0 (#480)
- Bump golang.org/x/crypto from 0.23.0 to 0.31.0 (#478)
- Bump golang.org/x/oauth2 from 0.23.0 to 0.24.0 (#475)
- Bump github.com/bradleyfalzon/ghinstallation/v2 from 2.11.0 to 2.12.0 (#476)
- Bump github.com/xanzy/go-gitlab from 0.112.0 to 0.114.0 (#474)
- Bump github.com/golang-jwt/jwt/v4 from 4.5.0 to 4.5.1 (#471)
- Bump github.com/ktrysmt/go-bitbucket from 0.9.80 to 0.9.81 (#469)
- Bump github.com/xanzy/go-gitlab from 0.109.0 to 0.112.0 (#468)
- Bump github.com/fatih/color from 1.17.0 to 1.18.0 (#470)
- Bump rexml from 3.3.6 to 3.3.9 in /site (#467)
- Bump golang.org/x/oauth2 from 0.22.0 to 0.23.0 (#462)
- Bump github.com/xanzy/go-gitlab from 0.108.0 to 0.109.0 (#463)

## [1.11.0] - 9/28/2024
### Added
- Reclone name and description to reclone output
- GHORG_PRESERVE_SCM_HOSTNAME, to organize clones by a top level SCM hostname folder; thanks @rrrix
- GHORG_PRUNE_UNTOUCHED, to prune repos that users make no changes in; thanks @MaxG87
- GHORG_GITHUB_TOKEN_FROM_GITHUB_APP to handle github app tokens; thanks @PaarthShah
- Command reclone-server, to run ad hoc reclone commands via HTTP requests
- Command reclone-cron, to run periodic reclone commands on a timer
### Changed
- **Breaking Change:** Self hosted gitlab clones will now include the top level domain in the default output directory
### Deprecated
### Removed
### Fixed
### Security

## [1.10.0] - 9/20/2024
### Added
- GHORG_NO_DIR_SIZE flag to turn off directory size output which is now enabled by default
- GHORG_STATS_ENABLED flag to track clone data over time, set to false by default
- Added two new flags to the `ghorg ls` command: `--long` and `--total`, which provide additional information about the cloned directories.
- GHORG_GITHUB_FILTER_LANGUAGE to filter clones by repo language
### Changed
### Deprecated
### Removed
- GHORG_RECLONE_VERBOSE flag
### Fixed
### Security
- Bump code.gitea.io/sdk/gitea from 0.18.0 to 0.19.0 (#441)
- Bump github.com/xanzy/go-gitlab from 0.106.0 to 0.107.0 (#442)
- Bump rexml from 3.2.8 to 3.3.3 in /site (#443)
- Bump rexml from 3.3.3 to 3.3.6 in /site (#444)
- Bump golang.org/x/oauth2 from 0.21.0 to 0.22.0 (#447)
- Bump github.com/xanzy/go-gitlab from 0.107.0 to 0.108.0 (#446)

## [1.9.13] - 7/20/2024
### Added
- GHORG_CLONE_SNIPPETS as a way to clone all snippets, gitlab only
- Commit count to clone output
### Changed
### Deprecated
### Removed
### Fixed
- Refs to deprecated io/ioutil (#415); thanks @testwill
- Bitbucket clones without app password; thanks @SparklingAperioso
- Github user repo clone options to be configurable; thanks @anton-kotik, @alexcurtin
- Failing checkout when default branch changes on remote; thanks @zarthross
### Security
- Update go version to 1.22
- Updated go github from v41 to v62
- Bump rexml from 3.2.5 to 3.2.8 in /site (#414)
- Bump golang.org/x/oauth2 from 0.19.0 to 0.20.0 (#419)
- Bump github.com/bradleyfalzon/ghinstallation/v2 from 2.10.0 to 2.11.0 (#420)
- Bump github.com/fatih/color from 1.16.0 to 1.17.0 (#422)
- Bump github.com/xanzy/go-gitlab from 0.103.0 to 0.105.0 (#423)
- Bump github.com/ktrysmt/go-bitbucket from 0.9.79 to 0.9.80 (#421)
- Bump github.com/hashicorp/go-retryablehttp from 0.7.2 to 0.7.7 (#427)
- Bump golang.org/x/image from 0.10.0 to 0.18.0 (#428)
- Bump golang.org/x/oauth2 from 0.20.0 to 0.21.0 (#431)
- Bump github.com/spf13/viper from 1.18.2 to 1.19.0 (#432)
- Bump github.com/spf13/cobra from 1.8.0 to 1.8.1 (#433)

## [1.9.12] - 5/15/24
### Added
- Additional info to git commands when debug is enabled; thanks @fey
- GHORG_TARGET_REPOS_PATH now outputs target repos not found in the org
- ssh-client to docker image; thanks @fey
### Changed
### Deprecated
### Removed
### Fixed
- GHORG_TARGET_REPOS_PATH working with GHORG_CLONE_WIKI; thanks @han911976
### Security
- Bump github.com/xanzy/go-gitlab from 0.101.0 to 0.103.0 (#405)
- Bump code.gitea.io/sdk/gitea from 0.17.1 to 0.18.0 (#403)
- Bump golang.org/x/oauth2 from 0.18.0 to 0.19.0 (#402)
- Bump github.com/ktrysmt/go-bitbucket from 0.9.77 to 0.9.79 (#404)

## [1.9.11] - 4/27/24
### Added
- GHORG_TARGET_REPOS_PATH
### Changed
### Deprecated
### Removed
### Fixed
- GitLab group calls by id instead of by name; thanks @angelabad (#400)
### Security
- Bump golang.org/x/net from 0.22.0 to 0.23.0 (#401)
- Bump github.com/ktrysmt/go-bitbucket from 0.9.76 to 0.9.77 (#398)
- Bump github.com/xanzy/go-gitlab from 0.98.0 to 0.101.0 (#396) (9 days ago)
- Bump github.com/bradleyfalzon/ghinstallation/v2 from 2.9.0 to 2.10.0 (#397)
- Bump golang.org/x/oauth2 from 0.17.0 to 0.18.0 (#395)
- Bump google.golang.org/protobuf from 1.31.0 to 1.33.0 (#394)
- Bump golang.org/x/oauth2 from 0.16.0 to 0.17.0 (#392)
- Bump github.com/ktrysmt/go-bitbucket from 0.9.74 to 0.9.76 (#393)
- Bump github.com/xanzy/go-gitlab from 0.96.0 to 0.98.0 (#391)
- Bump github.com/xanzy/go-gitlab from 0.95.2 to 0.96.0 (#386)
- Bump github.com/ktrysmt/go-bitbucket from 0.9.73 to 0.9.74 (#387)
- Bump github.com/bradleyfalzon/ghinstallation/v2 from 2.8.0 to 2.9.0 (#388)

## [1.9.10] - 1/21/24
### Added
- Automated docker images (#354); thanks @afonsoc12
- Docker support for armv7 (#361); thanks @LuckyTurtleDev
### Changed
- Allow using file paths in --token (#358); thanks @dvonessen
### Deprecated
### Removed
### Fixed
- Release version name; thanks @gbloquel
### Security
- Various dependency updates

## [1.9.9] - 8/26/23
### Added
### Changed
### Deprecated
### Removed
### Fixed
- Examples command; thanks @vlcinsky
### Security

## [1.9.8] - 8/25/23
### Added
- Examples command; thanks @MaxG87
- GHORG_RECLONE_ENV_CONFIG_ONLY; thanks @vlcinsky
### Changed
### Deprecated
### Removed
### Fixed
### Security
- Update go version to 1.20
- Bump github.com/ktrysmt/go-bitbucket from 0.9.60 to 0.9.63 (#329)
- Bump github.com/bradleyfalzon/ghinstallation/v2 from 2.5.0 to 2.6.0 (#327)
- Bump github.com/xanzy/go-gitlab from 0.86.0 to 0.89.0 (#326)
- Bump golang.org/x/image from 0.0.0-20220302094943-723b81ca9867 to 0.5.0 (#333)

## [1.9.7] - 7/23/23
### Added
- Reclone list flag
### Changed
### Deprecated
### Removed
### Fixed
- Clone help text referencing old default config path; thanks @MaxG87
### Security
- Bump github.com/ktrysmt/go-bitbucket from 0.9.58 to 0.9.60 (#320)
- Bump github.com/bradleyfalzon/ghinstallation/v2 from 2.4.0 to 2.5.0 (#319)
- Bump golang.org/x/oauth2 from 0.8.0 to 0.9.0 (#318)
- Bump github.com/xanzy/go-gitlab from 0.83.0 to 0.86.0 (#317)

## [1.9.6] - 6/10/23
### Added
- GHORG_CLONE_DEPTH; thanks @elliot-wasem
### Changed
### Deprecated
### Removed
### Fixed
### Security
- Bump github.com/spf13/viper from 1.15.0 to 1.16.0 (#314)
- Bump github.com/ktrysmt/go-bitbucket from 0.9.56 to 0.9.58 (#312)
- Bump golang.org/x/oauth2 from 0.7.0 to 0.8.0 (#311)

## [1.9.5] - 3/29/23
### Added
- GHORG_NO_TOKEN to allow cloning no token present; thanks @6543
- GitHub App authentication; thanks @duizabojul
### Changed
### Deprecated
### Removed
### Fixed
- Gitea SCM x509: certificate signed by unknown authority (#307); thanks @oligot
### Security
- Bump github.com/ktrysmt/go-bitbucket from 0.9.55 to 0.9.56 (#305)
- Bump github.com/spf13/cobra from 1.6.1 to 1.7.0 (#303)
- Bump github.com/xanzy/go-gitlab from 0.81.0 to 0.83.0 (#302)


## [1.9.4] - 2/26/23
### Added
### Changed
### Deprecated
### Removed
- Token length checks
### Fixed
- Gitea tokens from not being found in config.yaml; thanks @Antfere
- HTTPS GitHub clones with new GH fine grain tokens; thanks @verybadsoldier
### Security
- Bump github.com/spf13/viper from 1.14.0 to 1.15.0
- Bump github.com/fatih/color from 1.13.0 to 1.14.1
- Bump github.com/xanzy/go-gitlab from 0.77.0 to 0.79.1
- Bump golang.org/x/net from 0.4.0 to 0.7.0
- Bump golang.org/x/oauth2 from 0.3.0 to 0.5.0
- Bump github.com/xanzy/go-gitlab from 0.79.1 to 0.80.2

## [1.9.3] - 1/7/23
### Added
- Better examples for GitLab
- Better tests for local gitlab enterprise
### Changed
### Deprecated
### Removed
### Fixed
- gitlab hash concurrency issues
- all-users command directory nesting
- ls command to work with output dirs
- gitlab groups and subgroup nesting; thanks @nudgegoonies
### Security
- Bump github.com/ktrysmt/go-bitbucket from 0.9.54 to 0.9.55
- Bump github.com/xanzy/go-gitlab from 0.76.0 to 0.77.0

## [1.9.2] - 12/31/22
### Added
### Changed
### Deprecated
### Removed
### Fixed
- GitLab nested group names; thanks @MaxG87
### Security

## [1.9.1] - 12/11/22
### Added
- Ability to clone all users repos on hosted GitLab instances; thanks @mlaily
### Changed
### Deprecated
### Removed
### Fixed
- Top level GitLab groups on hosted GitLab instances now fetch all groups; thanks @mlaily
### Security
- Bump github.com/xanzy/go-gitlab from 0.74.0 to 0.76.0
- Bump github.com/spf13/viper from 1.13.0 to 1.14.0

## [1.9.0] - 11/5/22
### Added
- GHORG_INSECURE_GITEA_CLIENT to be explicit when cloning from Gitea instances using http; thanks @zerrol
### Changed
### Deprecated
### Removed
- Logging errors from security command
### Fixed
- GHORG_RECLONE_PATH from getting unset; thanks @afonsoc12
### Security
- Bump github.com/xanzy/go-gitlab from 0.73.1 to 0.74.0
- Bump github.com/spf13/cobra from 1.5.0 to 1.6.1

## [1.8.8] - 10/11/22
### Added
- Filename length limit on gitlab repos with name collisions
### Changed
### Deprecated
### Removed
### Fixed
### Security

## [1.8.7] - 7/19/22
### Added
### Changed
### Deprecated
### Removed
### Fixed
- SCM providers prompting for credentials
### Security

## [1.8.6] - 7/18/22
### Added
### Changed
- Default behavior of gitlab repo naming collisions to append namespace'd path to repo name instead of auto adding preserve dir behavior; thanks @sding3
### Deprecated
### Removed
### Fixed
### Security

## [1.8.5] - 7/13/22
### Added
- GHORG_GIT_FILTER flag to clone command; thanks @ryanaross
### Changed
### Deprecated
### Removed
### Fixed
### Security

## [1.8.4] - 7/12/22
### Added
- GHORG_INCLUDE_SUBMODULES flag to clone command; thanks @Antfere
### Changed
### Deprecated
### Removed
### Fixed
### Security

## [1.8.3] - 7/8/22
### Added
### Changed
### Deprecated
### Removed
### Fixed
- repo name collisions with gitlab clones not using --preserve-dir
- gitlab cloud clones with --preserve-dir creating nested output dirs; thanks @llama0815
### Security
## [1.8.2] - 7/6/22
### Added
- GHORG_RECLONE_QUIET flag
### Changed
- reclone logging format
### Deprecated
### Removed
### Fixed
### Security
- Updated to go 1.18
- Updated dependencies

## [1.8.1] - 7/3/22
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
- GHORG_CLONE_DELAY_SECONDS; thanks @batagy
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
