## 1.10.0

### 🚀 Features

- feat: implement Runner Controller API ([!2634](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2634)) by [Duo Developer](https://gitlab.com/duo-developer)

### 🔄 Other Changes

- chore(deps): update module github.com/godbus/dbus/v5 to v5.2.1 ([!2635](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2635)) by [GitLab Dependency Bot](https://gitlab.com/gitlab-dependency-update-bot)



# [1.10.0](https://gitlab.com/gitlab-org/api/client-go/compare/v1.9.1...v1.10.0) (2025-12-19)


### Features

* implement Runner Controller API ([66f19f4](https://gitlab.com/gitlab-org/api/client-go/commit/66f19f4073ce87566c7751e0987f857eeb008849))

## 1.9.1

### 🐛 Bug Fixes

- fix: use parameters in config.NewClient and Jobs.DownloadArtifactsFile ([!2633](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2633)) by [Oleksandr Redko](https://gitlab.com/alexandear)

### 🔄 Other Changes

- test: fix TestCreateMergeRequestContextCommits failing locally ([!2631](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2631)) by [Oleksandr Redko](https://gitlab.com/alexandear)
- Code Refactor Using Request Handlers - 8 ([!2523](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2523)) by [Yashesvinee V](https://gitlab.com/yashes7516)
- Code Refactor Using Request Handlers - 6 ([!2521](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2521)) by [Yashesvinee V](https://gitlab.com/yashes7516)



## [1.9.1](https://gitlab.com/gitlab-org/api/client-go/compare/v1.9.0...v1.9.1) (2025-12-17)


### Bug Fixes

* use parameters in config.NewClient and Jobs.DownloadArtifactsFile ([28b7cd7](https://gitlab.com/gitlab-org/api/client-go/commit/28b7cd72f06777a2d3ec7772870c26565140341a))

## 1.9.0

### 🚀 Features

- feat(api): add support for matrix project integration ([!2630](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2630)) by [aishahsofea](https://gitlab.com/aishahsofea)



# [1.9.0](https://gitlab.com/gitlab-org/api/client-go/compare/v1.8.2...v1.9.0) (2025-12-16)


### Features

* **api:** add support for matrix project integration ([0a5b11b](https://gitlab.com/gitlab-org/api/client-go/commit/0a5b11b9e2e405fb0a22009d60ce38091cc96625))

## 1.8.2

### 🐛 Bug Fixes

- fix: correct omitempty tag in VariableFilter.EnvironmentScope field ([!2629](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2629)) by [Oleksandr Redko](https://gitlab.com/alexandear)

### 🔄 Other Changes

- feat(protectedTags): add support for `deploy_key_id` to `protected_tags` ([!2624](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2624)) by [Zubeen](https://gitlab.com/syedzubeen)
- chore(deps): update docker docker tag to v29.1.3 ([!2623](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2623)) by [GitLab Dependency Bot](https://gitlab.com/gitlab-dependency-update-bot)
- chore(deps): update module buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go to v1.36.11-20251209175733-2a1774d88802.1 ([!2622](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2622)) by [GitLab Dependency Bot](https://gitlab.com/gitlab-dependency-update-bot)
- chore(deps): update module google.golang.org/protobuf to v1.36.11 ([!2621](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2621)) by [GitLab Dependency Bot](https://gitlab.com/gitlab-dependency-update-bot)



## [1.8.2](https://gitlab.com/gitlab-org/api/client-go/compare/v1.8.1...v1.8.2) (2025-12-15)


### Bug Fixes

* correct omitempty tag in VariableFilter.EnvironmentScope field ([c117da1](https://gitlab.com/gitlab-org/api/client-go/commit/c117da1b123251ba86271d1ce3bf9750617e344f))


### Features

* **protectedTags:** add support for `deploy_key_id` to `protected_tags` ([c0fc3db](https://gitlab.com/gitlab-org/api/client-go/commit/c0fc3db793b51bfabb0ac8bb42442e6916b9df3f))

## 1.8.1

### 🐛 Bug Fixes

- fix(epics): handle datetime format in ISOTime UnmarshalJSON ([!2612](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2612)) by [Zubeen](https://gitlab.com/syedzubeen)

### 🔄 Other Changes

- chore(deps): update module buf.build/go/protovalidate to v1.1.0 ([!2619](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2619)) by [GitLab Dependency Bot](https://gitlab.com/gitlab-dependency-update-bot)
- chore: add deprecation notice for PersonalAccessTokens.RevokePersonalAccessToken ([!2615](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2615)) by [aishahsofea](https://gitlab.com/aishahsofea)
- chore(deps): update golangci/golangci-lint docker tag to v2.7.2 ([!2613](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2613)) by [GitLab Dependency Bot](https://gitlab.com/gitlab-dependency-update-bot)
- chore(deps): do not use the experimental package ([!2614](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2614)) by [Mikhail Mazurskiy](https://gitlab.com/ash2k)
- test: Replace SkipIfRunningCE with SkipIfNotLicensed ([!2616](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2616)) by [Patrick Rice](https://gitlab.com/PatrickRice)



## [1.8.1](https://gitlab.com/gitlab-org/api/client-go/compare/v1.8.0...v1.8.1) (2025-12-10)


### Bug Fixes

* **epics:** handle datetime format in ISOTime UnmarshalJSON ([257e0ac](https://gitlab.com/gitlab-org/api/client-go/commit/257e0acd29daf887456d924c0063b52ebc2e808f))

## 1.8.0

### 🚀 Features

- feat(hooks): add support for all hook event types ([!2606](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2606)) by [Heidi Berry](https://gitlab.com/heidi.berry)



# [1.8.0](https://gitlab.com/gitlab-org/api/client-go/compare/v1.7.0...v1.8.0) (2025-12-08)


### Features

* **hooks:** add support for all hook event types ([c3c9ca2](https://gitlab.com/gitlab-org/api/client-go/commit/c3c9ca275969adffca37908d63e5c70f634d7bbe))

## 1.7.0

### 🚀 Features

- feat(users): Add support for a user to see only one file diff per page ([!2597](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2597)) by [Zubeen](https://gitlab.com/syedzubeen)



# [1.7.0](https://gitlab.com/gitlab-org/api/client-go/compare/v1.6.0...v1.7.0) (2025-12-06)


### Features

* **users:** Add support for a user to see only one file diff per page ([e2a9e09](https://gitlab.com/gitlab-org/api/client-go/commit/e2a9e09e79e7949e0b19dcfc97e3b7b533541856))

## 1.6.0

### 🚀 Features

- feat: add admin compliance policy settings API ([!2610](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2610)) by [Hannes Lange](https://gitlab.com/hlange4)

### 🔄 Other Changes

- doc: fix typo ([!2603](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2603)) by [Guilhem Bonnefille](https://gitlab.com/gbonnefille)
- chore(deps): update golangci/golangci-lint docker tag to v2.7.1 ([!2611](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2611)) by [GitLab Dependency Bot](https://gitlab.com/gitlab-dependency-update-bot)
- chore(deps): update docker docker tag to v29.1.2 ([!2609](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2609)) by [GitLab Dependency Bot](https://gitlab.com/gitlab-dependency-update-bot)
- chore(deps): update golangci/golangci-lint docker tag to v2.7.0 ([!2608](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2608)) by [GitLab Dependency Bot](https://gitlab.com/gitlab-dependency-update-bot)



# [1.6.0](https://gitlab.com/gitlab-org/api/client-go/compare/v1.5.0...v1.6.0) (2025-12-05)


### Features

* add admin compliance policy settings API ([5c17773](https://gitlab.com/gitlab-org/api/client-go/commit/5c17773ca94ddece28978c7396bddcc6c65fb6a7))

## 1.5.0

### 🚀 Features

- feat(Project Mirrors): Add missing Mirror attributes when reading or updating Project Mirrors ([!2600](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2600)) by [Patrick Rice](https://gitlab.com/PatrickRice)



# [1.5.0](https://gitlab.com/gitlab-org/api/client-go/compare/v1.4.1...v1.5.0) (2025-12-03)


### Features

* **Project Mirrors:** Add missing Mirror attributes when reading or updating Project Mirrors ([a49b32d](https://gitlab.com/gitlab-org/api/client-go/commit/a49b32df59aeae97247d21a83be3fab97da1bbfe))

## 1.4.1

### 🐛 Bug Fixes

- Encode package managers as CSV in query for dependencies list ([!2604](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2604)) by [Timo Furrer](https://gitlab.com/timofurrer)



## [1.4.1](https://gitlab.com/gitlab-org/api/client-go/compare/v1.4.0...v1.4.1) (2025-12-02)

## 1.4.0

### 🚀 Features

- feat(integrations): Add attestations integrations ([!2582](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2582)) by [Sam Roque-Worcel](https://gitlab.com/sroque-worcel)

### 🔄 Other Changes

- chore(deps): update docker docker tag to v29.1.1 ([!2602](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2602)) by [GitLab Dependency Bot](https://gitlab.com/gitlab-dependency-update-bot)



# [1.4.0](https://gitlab.com/gitlab-org/api/client-go/compare/v1.3.1...v1.4.0) (2025-12-02)


### Features

* **integrations:** Add attestations integrations ([4f50db4](https://gitlab.com/gitlab-org/api/client-go/commit/4f50db4acfb19212bfdfc12eb808dbc7ed8d7ad2))

## 1.3.1

### 🐛 Bug Fixes

- fix(merge_requests): Reinstate missing request option ([!2601](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2601)) by [Heidi Berry](https://gitlab.com/heidi.berry)



## [1.3.1](https://gitlab.com/gitlab-org/api/client-go/compare/v1.3.0...v1.3.1) (2025-12-01)


### Bug Fixes

* **merge_requests:** Reinstate missing request option ([f5f912d](https://gitlab.com/gitlab-org/api/client-go/commit/f5f912ddc2dfb1af88de8710bde783f3f7ccd7c2))

## 1.3.0

### 🚀 Features

- feat(credentials): Add support for revoking group PATs, listing/deleting group SSH keys ([!2594](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2594)) by [Heidi Berry](https://gitlab.com/heidi.berry)

### 🔄 Other Changes

- refactor: moved comments to interface ([!2595](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2595)) by [Zubeen](https://gitlab.com/syedzubeen)
- refactor(users): moved comments to interface ([!2596](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2596)) by [Zubeen](https://gitlab.com/syedzubeen)
- refactor: moved comments to interface ([!2599](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2599)) by [Zubeen](https://gitlab.com/syedzubeen)
- Simplify more request functions, introducing NoEscape ([!2592](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2592)) by [Timo Furrer](https://gitlab.com/timofurrer)



# [1.3.0](https://gitlab.com/gitlab-org/api/client-go/compare/v1.2.0...v1.3.0) (2025-11-30)


### Features

* **credentials:** Add support for revoking group PATs, listing/deleting group SSH keys ([3439f4f](https://gitlab.com/gitlab-org/api/client-go/commit/3439f4f0345b97dea0abf926ecaac9d3a7eb6769))

## 1.2.0

### 🚀 Features

- feat(credentials): Add support for listing all SaaS enterprise user personal access tokens ([!2593](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2593)) by [Heidi Berry](https://gitlab.com/heidi.berry)

### 🔄 Other Changes

- Code Refactor Using Request Handlers - 10 ([!2525](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2525)) by [Yashesvinee V](https://gitlab.com/yashes7516)



# [1.2.0](https://gitlab.com/gitlab-org/api/client-go/compare/v1.1.0...v1.2.0) (2025-11-27)


### Features

* **credentials:** Add support for listing all SaaS enterprise user personal access tokens ([3697779](https://gitlab.com/gitlab-org/api/client-go/commit/369777938e435b043e37460ff1feffedd84b7dd1))

## 1.1.0

### 🚀 Features

- feat(service_account): allow providing email when update a Service Account ([!2589](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2589)) by [kilianpaquier](https://gitlab.com/u.kilianpaquier)

### 🔄 Other Changes

- Bump dependencies ([!2591](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2591)) by [Timo Furrer](https://gitlab.com/timofurrer)
- chore(deps): update docker docker tag to v29 ([!2586](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2586)) by [GitLab Dependency Bot](https://gitlab.com/gitlab-dependency-update-bot)



# [1.1.0](https://gitlab.com/gitlab-org/api/client-go/compare/v1.0.1...v1.1.0) (2025-11-26)


### Features

* **service_account:** allow providing email when update a Service Account ([324d080](https://gitlab.com/gitlab-org/api/client-go/commit/324d0806a5cd8cb6ae7f68381d09cf5e2a31a0cc))

## 1.0.1

### 🐛 Bug Fixes

- fix: fix ReviewerID() and let it accept int64 ([!2587](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2587)) by [Ilya Savitsky](https://gitlab.com/ipsavitsky234)



## [1.0.1](https://gitlab.com/gitlab-org/api/client-go/compare/v1.0.0...v1.0.1) (2025-11-25)


### Bug Fixes

* fix ReviewerID() and let it accept int64 ([6a6d439](https://gitlab.com/gitlab-org/api/client-go/commit/6a6d43952b70191358e7b726eff4f7f24a0f7ff6))

## 1.0.0

### 💥 Breaking Changes

- Release client-go 1.0 ([!2575](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2575)) by [Patrick Rice](https://gitlab.com/PatrickRice)



# [1.0.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.161.1...v1.0.0) (2025-11-24)


* Merge branch 'release-client-1.0' into 'main' ([f06b8c2](https://gitlab.com/gitlab-org/api/client-go/commit/f06b8c2cb4446e2e76a13bbc707c64e22a64d477))


### Bug Fixes

* **issues:** use AssigneeIDValue for ListProjectIssuesOptions.AssigneeID ([1dcb219](https://gitlab.com/gitlab-org/api/client-go/commit/1dcb219c343bc5b5622ff49933199c003a231bd4))


### Features

* **ListOptions:** Update ListOptions to use composition instead of aliasing ([60beef3](https://gitlab.com/gitlab-org/api/client-go/commit/60beef36d0f93a7dc66749f55d98defbc1b3fe28))


### BREAKING CHANGES

* Release 1.0
* **ListOptions:** ListOptions implementation changed from aliasing to composition
Changelog: Improvements

## 0.161.1

### 🐛 Bug Fixes

- fix(users): Fix a bug where error parsing causes user blocking to not function properly ([!2584](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2584)) by [Patrick Rice](https://gitlab.com/PatrickRice)



## [0.161.1](https://gitlab.com/gitlab-org/api/client-go/compare/v0.161.0...v0.161.1) (2025-11-24)


### Bug Fixes

* **users:** Fix a bug where error parsing causes user blocking to not function properly ([2ad5506](https://gitlab.com/gitlab-org/api/client-go/commit/2ad55065d624d27d1f539a3c41489989b9a0d036))

## 0.161.0

### 🚀 Features

- fix: return detailed API errors for BlockUser instead of generic LDAP message ([!2581](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2581)) by [Zubeen](https://gitlab.com/syedzubeen)



# [0.161.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.160.2...v0.161.0) (2025-11-24)


### Bug Fixes

* return detailed API errors for BlockUser instead of generic LDAP message ([2ba9fa6](https://gitlab.com/gitlab-org/api/client-go/commit/2ba9fa6995de6cadf0dae1bf600979b73ee471ce))

## 0.160.2

### 🐛 Bug Fixes

- Fix double escaping in paths ([!2583](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2583)) by [Timo Furrer](https://gitlab.com/timofurrer)



## [0.160.2](https://gitlab.com/gitlab-org/api/client-go/compare/v0.160.1...v0.160.2) (2025-11-24)

## 0.160.1

### 🐛 Bug Fixes

- fix: update input field from "key" to "name" in pipeline schedules to prevent an API error ([!2580](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2580)) by [Zubeen](https://gitlab.com/syedzubeen)

### 🔄 Other Changes

- Code Refactor Using Request Handlers - 9 ([!2524](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2524)) by [Yashesvinee V](https://gitlab.com/yashes7516)
- Code Refactor Using Request Handlers - 7 ([!2522](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2522)) by [Yashesvinee V](https://gitlab.com/yashes7516)
- Code Refactor Using Request Handlers - 5 ([!2518](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2518)) by [Yashesvinee V](https://gitlab.com/yashes7516)
- Code Refactor Using Request Handlers - 2 ([!2515](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2515)) by [Yashesvinee V](https://gitlab.com/yashes7516)
- Code Refactor Using Request Handlers - 4 ([!2517](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2517)) by [Yashesvinee V](https://gitlab.com/yashes7516)
- Code Refactor Using Request Handlers - 3 ([!2516](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2516)) by [Yashesvinee V](https://gitlab.com/yashes7516)
- chore(deps): update module github.com/godbus/dbus/v5 to v5.2.0 ([!2576](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2576)) by [GitLab Dependency Bot](https://gitlab.com/gitlab-dependency-update-bot)
- chore(deps): update golangci/golangci-lint docker tag to v2.6.2 ([!2577](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2577)) by [GitLab Dependency Bot](https://gitlab.com/gitlab-dependency-update-bot)
- Code Refactor Using Request Handlers - 1 ([!2514](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2514)) by [Yashesvinee V](https://gitlab.com/yashes7516)



## [0.160.1](https://gitlab.com/gitlab-org/api/client-go/compare/v0.160.0...v0.160.1) (2025-11-19)


### Bug Fixes

* update input field from "key" to "name" in pipeline schedules to prevent an API error ([062133f](https://gitlab.com/gitlab-org/api/client-go/commit/062133f0c24b32ca6ae64a9f7b80fd3fa7e58256))

## 0.160.0

### 🚀 Features

- feat (project_members): Add show_seat_info option to ProjectMembers ([!2572](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2572)) by [Zubeen](https://gitlab.com/syedzubeen)

### 🔄 Other Changes

- refactor: fix modernize lint issues ([!2574](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2574)) by [Oleksandr Redko](https://gitlab.com/alexandear)
- chore(deps): update module cel.dev/expr to v0.25.1 ([!2573](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2573)) by [GitLab Dependency Bot](https://gitlab.com/gitlab-dependency-update-bot)
- docs(no-release): format examples, update pkg doc url ([!2543](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2543)) by [Oleksandr Redko](https://gitlab.com/alexandear)



# [0.160.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.159.0...v0.160.0) (2025-11-12)

## 0.159.0

### 🚀 Features

- feat(integrations): add group integration API endpoints for Jira ([!2563](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2563)) by [Harsh Rai](https://gitlab.com/harshrai654)

### 🔄 Other Changes

- chore(deps): update golangci/golangci-lint docker tag to v2.6.1 ([!2564](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2564)) by [GitLab Dependency Bot](https://gitlab.com/gitlab-dependency-update-bot)



# [0.159.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.158.0...v0.159.0) (2025-11-04)


### Features

* **integrations:** add group integration API endpoints for Jira ([09e18ee](https://gitlab.com/gitlab-org/api/client-go/commit/09e18ee598bb7805ac8221f6a05426b1785f9011))

## 0.158.0

### 🚀 Features

- Add support to send variables for GraphQL queries ([!2562](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2562)) by [rafasf](https://gitlab.com/rafasf)

### 🔄 Other Changes

- chore(deps): update module cel.dev/expr to v0.25.0 ([!2560](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2560)) by [GitLab Dependency Bot](https://gitlab.com/gitlab-dependency-update-bot)
- chore(no-release): standardize GitLab name capitalization ([!2551](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2551)) by [Zubeen](https://gitlab.com/syedzubeen)
- chore(deps): update golangci/golangci-lint docker tag to v2.6.0 ([!2558](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2558)) by [GitLab Dependency Bot](https://gitlab.com/gitlab-dependency-update-bot)
- refactor: moved comments to interface 2 ([!2557](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2557)) by [Zubeen](https://gitlab.com/syedzubeen)
- refactor: moved comments to interface ([!2556](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2556)) by [Zubeen](https://gitlab.com/syedzubeen)
- refactor(test): avoid panic in tests with goroutines ([!2553](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2553)) by [Oleksandr Redko](https://gitlab.com/alexandear)



# [0.158.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.157.1...v0.158.0) (2025-11-03)

## 0.157.1

### 🐛 Bug Fixes

- fix(protected_packages): fix invalid types ([!2554](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2554)) by [Ruwen Schwedewsky](https://gitlab.com/RuwenSchwedewskySinch)

### 🔄 Other Changes

- chore: Update review instructions for mentioning GitLab ([!2552](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2552)) by [Zubeen](https://gitlab.com/syedzubeen)
- Implement do function to reduce boilerplate ([!2550](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2550)) by [Timo Furrer](https://gitlab.com/timofurrer)
- refactor(test): migrate to testify assertions 4 ([!2548](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2548)) by [Zubeen](https://gitlab.com/syedzubeen)
- refactor(test): migrate to testify assertions 2 ([!2546](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2546)) by [Zubeen](https://gitlab.com/syedzubeen)
- refactor(test): migrate to testify assertions ([!2545](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2545)) by [Zubeen](https://gitlab.com/syedzubeen)
- refactor(test): migrate to testify assertions 5 ([!2549](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2549)) by [Zubeen](https://gitlab.com/syedzubeen)
- test: add unit tests for cluster agents and deployments ([!2499](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2499)) by [Zubeen](https://gitlab.com/syedzubeen)
- refactor(test): migrate to testify assertions 3 ([!2547](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2547)) by [Zubeen](https://gitlab.com/syedzubeen)
- Fix: Helper Functions for Code Refactoring ([!2544](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2544)) by [Yashesvinee V](https://gitlab.com/yashes7516)
- test: adds UT for formatPackageURL ([!2527](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2527)) by [Zubeen](https://gitlab.com/syedzubeen)
- test: adds UT for getEpicLinks ([!2526](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2526)) by [Zubeen](https://gitlab.com/syedzubeen)
- test: add test for ApproveOrRejectProjectDeployment ([!2498](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2498)) by [Zubeen](https://gitlab.com/syedzubeen)
- test: adds UTs for packages ([!2529](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2529)) by [Zubeen](https://gitlab.com/syedzubeen)



## [0.157.1](https://gitlab.com/gitlab-org/api/client-go/compare/v0.157.0...v0.157.1) (2025-10-28)


### Bug Fixes

* **no-release:** Helper Functions for Code Refactoring ([6feffea](https://gitlab.com/gitlab-org/api/client-go/commit/6feffea6696a8e333fd0811eee8501e58ba743e3))
* **protected_packages:** fix invalid types ([c09943b](https://gitlab.com/gitlab-org/api/client-go/commit/c09943b0dde510dca32a2544a9c0f75f85943d96))

## 0.157.0

### 🚀 Features

- Add merge requests commit api ([!2539](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2539)) by [Ilya Savitsky](https://gitlab.com/ipsavitsky234)

### 🔄 Other Changes

- test: adds missing UTs for notifications ([!2528](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2528)) by [Zubeen](https://gitlab.com/syedzubeen)
- chore: Update review instructions ([!2537](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2537)) by [Patrick Rice](https://gitlab.com/PatrickRice)
- chore(no-release): Fix godoc comments; enable godoclint ([!2535](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2535)) by [Oleksandr Redko](https://gitlab.com/alexandear)



# [0.157.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.156.0...v0.157.0) (2025-10-13)

## 0.156.0

### 🚀 Features

- feat(api): add support for test report summary api ([!2487](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2487)) by [Daniela Filipe Bento](https://gitlab.com/danifbento)



# [0.156.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.155.0...v0.156.0) (2025-10-10)


### Features

* **api:** add support for test report summary api ([8a0c6dd](https://gitlab.com/gitlab-org/api/client-go/commit/8a0c6dde10a4c9c034274a439eaa060dc6e40995))

## 0.155.0

### 🚀 Features

- feat(group_relations_export): Added Group Relations API integration ([!2508](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2508)) by [Jose Gabriel Companioni Benitez](https://gitlab.com/elC0mpa)

### 🔄 Other Changes

- chore: use local protoc plugin with buf ([!2536](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2536)) by [Timo Furrer](https://gitlab.com/timofurrer)
- chore(no-release): Change generated file comment ([!2532](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2532)) by [Oleksandr Redko](https://gitlab.com/alexandear)
- docs(no-release): Fix the comment for EnvVarGitLabContext ([!2533](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2533)) by [Oleksandr Redko](https://gitlab.com/alexandear)
- feat(client_options): Added unit tests ([!2510](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2510)) by [Jose Gabriel Companioni Benitez](https://gitlab.com/elC0mpa)



# [0.155.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.154.0...v0.155.0) (2025-10-09)


### Features

* **client_options:** Added unit tests ([c148031](https://gitlab.com/gitlab-org/api/client-go/commit/c14803189aa47a0cc9e64e9b455b93e6d4c4e4b9))
* **group_relations_export:** Added Group Relations API integration ([956e039](https://gitlab.com/gitlab-org/api/client-go/commit/956e03950d6bc03c56fa1ea4c5d6e06bfd0b264f))

## 0.154.0

### 🚀 Features

- feat(protected_packages): Add api integration ([!2520](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2520)) by [Jose Gabriel Companioni Benitez](https://gitlab.com/elC0mpa)



# [0.154.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.153.0...v0.154.0) (2025-10-08)


### Features

* **protected_packages:** Add api integration ([2de15c7](https://gitlab.com/gitlab-org/api/client-go/commit/2de15c7875e232b0b0b1e5e5bb8e184cd11d0774))

## 0.153.0

### 🚀 Features

- feat(project_Statistics): Added api integration ([!2512](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2512)) by [Jose Gabriel Companioni Benitez](https://gitlab.com/elC0mpa)

### 🔄 Other Changes

- refactor: moved comments to interface ([!2509](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2509)) by [ajey muthiah](https://gitlab.com/ajeymuthiah)
- chore(no-release): Helper Functions for Code Refactoring ([!2503](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2503)) by [Yashesvinee V](https://gitlab.com/yashes7516)
- Add t.Parallel() to all tests and enable linters ([!2513](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2513)) by [Timo Furrer](https://gitlab.com/timofurrer)
- ci: Remove the `commitlint` job. ([!2511](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2511)) by [Florian Forster](https://gitlab.com/fforster)
- refactor: moved comments to interface ([!2507](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2507)) by [ajey muthiah](https://gitlab.com/ajeymuthiah)



# [0.153.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.152.0...v0.153.0) (2025-10-08)


### Features

* **project_Statistics:** Added api integration ([75b5a03](https://gitlab.com/gitlab-org/api/client-go/commit/75b5a03010a39d5353c975a558fda0b6f00cb697))

## 0.152.0

### 🚀 Features

- feat(api): add api support for listing users who starred a project ([!2486](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2486)) by [ajey muthiah](https://gitlab.com/ajeymuthiah)

### 🔄 Other Changes

- chore(no-release): Update Duo Review Instructions ([!2502](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2502)) by [Patrick Rice](https://gitlab.com/PatrickRice)
- feat(model_registry_api): Added api integration ([!2501](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2501)) by [Jose Gabriel Companioni Benitez](https://gitlab.com/elC0mpa)
- feat(no-release): Add AGENTS.md file ([!2479](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2479)) by [Patrick Rice](https://gitlab.com/PatrickRice)
- chore(no-release): Disable dependency scanning on personal forks ([!2500](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2500)) by [Patrick Rice](https://gitlab.com/PatrickRice)



# [0.152.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.151.0...v0.152.0) (2025-10-06)


### Features

* **api:** add api support for listing users who starred a project ([0cdb4ce](https://gitlab.com/gitlab-org/api/client-go/commit/0cdb4ce5399b43e47bf120a90b16d00c022e194c))
* **model_registry_api:** Added api integration ([065dd63](https://gitlab.com/gitlab-org/api/client-go/commit/065dd639bc8bd0f44cab4d92dbe3ea7f134b913f))
* **no-release:** Add AGENTS.md file ([b9febab](https://gitlab.com/gitlab-org/api/client-go/commit/b9febab3181c3f87edd1fd99b5e596f76bc8b7cc))

## 0.151.0

### 🚀 Features

- feat(api): add api support for delete enterprise user ([!2492](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2492)) by [ajey muthiah](https://gitlab.com/ajeymuthiah)

### 🔄 Other Changes

- docs(no-release): Make it easier to find the docs on issues ([!2497](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2497)) by [Heidi Berry](https://gitlab.com/heidi.berry)



# [0.151.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.150.0...v0.151.0) (2025-10-04)


### Features

* **api:** add api support for delete enterprise user ([36ca8ab](https://gitlab.com/gitlab-org/api/client-go/commit/36ca8ab7672c352a073d59dacae3d763d4089abb))

## 0.150.0

### 🚀 Features

- feat: add Project Aliases API support ([!2493](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2493)) by [Yashesvinee V](https://gitlab.com/yashes7516)

### 🔄 Other Changes

- chore(deps): update module buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go to v1.36.10-20250912141014-52f32327d4b0.1 ([!2495](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2495)) by [GitLab Dependency Bot](https://gitlab.com/gitlab-dependency-update-bot)
- chore(deps): update module github.com/danieljoos/wincred to v1.2.3 ([!2494](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2494)) by [GitLab Dependency Bot](https://gitlab.com/gitlab-dependency-update-bot)



# [0.150.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.149.0...v0.150.0) (2025-10-03)


### Features

* add Project Aliases API support ([4ece88e](https://gitlab.com/gitlab-org/api/client-go/commit/4ece88e6a8cfa0f53e68184b2905d4c2fb6e857a))

## 0.149.0

### 🚀 Features

- feat(no-release): Add dependency scanning ([!2480](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2480)) by [Patrick Rice](https://gitlab.com/PatrickRice)

### 🔄 Other Changes

- ci(semantic-release): migrate to `@gitlab/semantic-release-merge-request-analyzer` ([!2490](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2490)) by [Florian Forster](https://gitlab.com/fforster)
- ci: add the `autolabels` job ([!2489](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2489)) by [Florian Forster](https://gitlab.com/fforster)
- chore(deps): update module google.golang.org/protobuf to v1.36.10 ([!2488](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2488)) by [GitLab Dependency Bot](https://gitlab.com/gitlab-dependency-update-bot)
- refactor(no-release): added tests for delete project hook method ([!2482](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2482)) by [Jose Gabriel Companioni Benitez](https://gitlab.com/elC0mpa)
- docs(no-release): Add guide for adding new APIs and issue templates ([!2478](https://gitlab.com/gitlab-org/api/client-go/-/merge_requests/2478)) by [Heidi Berry](https://gitlab.com/heidi.berry)



# [0.149.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.148.1...v0.149.0) (2025-10-02)


### Features

* **no-release:** Add dependency scanning ([8b0ee10](https://gitlab.com/gitlab-org/api/client-go/commit/8b0ee10acb8adceb5d34be2165b7d587b1e42e49))

## [0.148.1](https://gitlab.com/gitlab-org/api/client-go/compare/v0.148.0...v0.148.1) (2025-09-26)


### Bug Fixes

* label unmarshaling for `BasicMergeRequest` list operations ([e80c453](https://gitlab.com/gitlab-org/api/client-go/commit/e80c453aa6a5a265ec8748ae3f3f761a70f4470e))

# [0.148.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.147.1...v0.148.0) (2025-09-23)


### Features

* **ResourceGroup:** add `newest_ready_first` to resource group `process_mode` ([fc8f743](https://gitlab.com/gitlab-org/api/client-go/commit/fc8f7431da4ca8594723105473687e8f1378df2b))

## [0.147.1](https://gitlab.com/gitlab-org/api/client-go/compare/v0.147.0...v0.147.1) (2025-09-22)


### Bug Fixes

* **client:** use default retry policy from retryablehttp ([2a72511](https://gitlab.com/gitlab-org/api/client-go/commit/2a725113118608712f668b159ca2dab11f4e588e))

# [0.147.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.146.0...v0.147.0) (2025-09-22)


### Features

* **Project:** add resource_group_default_process_mode ([7804faf](https://gitlab.com/gitlab-org/api/client-go/commit/7804fafa18cc15fec8a0886a081bf3311d72eb1f))

# [0.146.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.145.0...v0.146.0) (2025-09-18)


### Features

* **pipelines:** Add compile-time type-safe pipeline inputs support ([4b30e60](https://gitlab.com/gitlab-org/api/client-go/commit/4b30e60260e4f06e7684352693aac49abd748579)), closes [gitlab-org/api/client-go#2154](https://gitlab.com/gitlab-org/api/client-go/issues/2154)
* **PipelinesService:** Add support for pipeline inputs with type validation ([ab3056f](https://gitlab.com/gitlab-org/api/client-go/commit/ab3056f403ec0268e14b312de3f5b51b115ad97a)), closes [gitlab-org/api/client-go#2154](https://gitlab.com/gitlab-org/api/client-go/issues/2154)
* **PipelineTriggersService:** Add support for pipeline inputs to trigger API ([9ad770e](https://gitlab.com/gitlab-org/api/client-go/commit/9ad770e49e59b2a41c665dfc4781f3b56650e813)), closes [gitlab-org/api/client-go#2154](https://gitlab.com/gitlab-org/api/client-go/issues/2154)

# [0.145.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.144.1...v0.145.0) (2025-09-15)


### Features

* Add missing created_by field to ProjectMembers and GroupMembers ([5348e01](https://gitlab.com/gitlab-org/api/client-go/commit/5348e01913c358c53bdd3da46b069713273d6802))

## [0.144.1](https://gitlab.com/gitlab-org/api/client-go/compare/v0.144.0...v0.144.1) (2025-09-13)

# [0.144.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.143.3...v0.144.0) (2025-09-12)


### Features

* **client:** add http.RoundTripper Middleware Configuration Option to Client ([88f9d10](https://gitlab.com/gitlab-org/api/client-go/commit/88f9d1055acbd5e060ab13947b856ccc3a03da6f))

## [0.143.3](https://gitlab.com/gitlab-org/api/client-go/compare/v0.143.2...v0.143.3) (2025-09-10)

## [0.143.2](https://gitlab.com/gitlab-org/api/client-go/compare/v0.143.1...v0.143.2) (2025-09-09)

## [0.143.1](https://gitlab.com/gitlab-org/api/client-go/compare/v0.143.0...v0.143.1) (2025-09-08)

# [0.143.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.142.6...v0.143.0) (2025-09-08)


### Features

* **users:** Add support for PublicEmail to ListUsers ([74a3b6a](https://gitlab.com/gitlab-org/api/client-go/commit/74a3b6a7dd1340faa70ec1246b5b99394c56f90b))

## [0.142.6](https://gitlab.com/gitlab-org/api/client-go/compare/v0.142.5...v0.142.6) (2025-09-02)

## [0.142.5](https://gitlab.com/gitlab-org/api/client-go/compare/v0.142.4...v0.142.5) (2025-08-30)

## [0.142.4](https://gitlab.com/gitlab-org/api/client-go/compare/v0.142.3...v0.142.4) (2025-08-28)

## [0.142.3](https://gitlab.com/gitlab-org/api/client-go/compare/v0.142.2...v0.142.3) (2025-08-28)

## [0.142.2](https://gitlab.com/gitlab-org/api/client-go/compare/v0.142.1...v0.142.2) (2025-08-26)

## [0.142.1](https://gitlab.com/gitlab-org/api/client-go/compare/v0.142.0...v0.142.1) (2025-08-25)

# [0.142.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.141.2...v0.142.0) (2025-08-21)


### Features

* **tokens:** add expiration filters and sorting options to ListPersonalAccessTokens ([0a9f797](https://gitlab.com/gitlab-org/api/client-go/commit/0a9f79790ac87c7f7b8e32e9cdea27fbc613728b))

## [0.141.2](https://gitlab.com/gitlab-org/api/client-go/compare/v0.141.1...v0.141.2) (2025-08-20)

## [0.141.1](https://gitlab.com/gitlab-org/api/client-go/compare/v0.141.0...v0.141.1) (2025-08-18)

# [0.141.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.140.0...v0.141.0) (2025-08-18)


### Features

* **config:** support custom headers for instances ([76b0e82](https://gitlab.com/gitlab-org/api/client-go/commit/76b0e82ab57b21b7da915117fb37ac2bf56506e8))

# [0.140.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.139.2...v0.140.0) (2025-08-18)


### Features

* **client:** add support for cookie jars ([4b525e3](https://gitlab.com/gitlab-org/api/client-go/commit/4b525e3f14741176ea8cbf4e7ae988b87455f4d0))

## [0.139.2](https://gitlab.com/gitlab-org/api/client-go/compare/v0.139.1...v0.139.2) (2025-08-14)

## [0.139.1](https://gitlab.com/gitlab-org/api/client-go/compare/v0.139.0...v0.139.1) (2025-08-14)

# [0.139.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.138.0...v0.139.0) (2025-08-13)


### Features

* **terraform:** improve Terraform States service ([e08128b](https://gitlab.com/gitlab-org/api/client-go/commit/e08128bf87011455db06dc946e77b2a16ee36948))

# [0.138.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.137.0...v0.138.0) (2025-08-12)


### Bug Fixes

* deprecate ListProjectInvidedGroupOptions due to a typo ([322496a](https://gitlab.com/gitlab-org/api/client-go/commit/322496a8a4c3fd7393b4b2c2b427c42fff243861))
* Update config package name to v1beta1 ([f958e6b](https://gitlab.com/gitlab-org/api/client-go/commit/f958e6bd2935fddf4867d9992908e87288e89c20))


### Features

* add support for field "Created at" for Tags ([f363d57](https://gitlab.com/gitlab-org/api/client-go/commit/f363d57853f2e05c848e88946269c936f0b6bf76))
* **app settings:** Add support for CanCreateOrganization ([1db661d](https://gitlab.com/gitlab-org/api/client-go/commit/1db661de26e0d3a78134c6bd1d31fb24d9a60677))
* **hooks:** Add support for project webhook url variables ([efabed5](https://gitlab.com/gitlab-org/api/client-go/commit/efabed57d83eefe565aa2dbbb943d94212ec6167))
* update datadog integration with new fields and API endpoints ([660ef31](https://gitlab.com/gitlab-org/api/client-go/commit/660ef31daf884bde545cfaa88432ac5ec7e3bfe7))
* update external status checks to return the status check object ([2d78e8c](https://gitlab.com/gitlab-org/api/client-go/commit/2d78e8cc43971c4395c980672de7263c10401900))

# [0.137.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.136.0...v0.137.0) (2025-07-21)


### Features

* **integrations:** add group harbor integration ([220e4cb](https://gitlab.com/gitlab-org/api/client-go/commit/220e4cb524d9303d36384043f29f96f43e4d9387))

# [0.136.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.135.0...v0.136.0) (2025-07-21)


### Features

* **client:** add NewRequestToURL function for calls to absolute URLs ([524b571](https://gitlab.com/gitlab-org/api/client-go/commit/524b571339b7704e0e346a5a64f367265b96125f))
* **projects:** Add support for RestoreProject ([b33e888](https://gitlab.com/gitlab-org/api/client-go/commit/b33e8882ad6611b1ac19222d0abdbfc477846ea1))

# [0.135.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.134.0...v0.135.0) (2025-07-21)


### Features

* **config:** implement extensions API ([257f745](https://gitlab.com/gitlab-org/api/client-go/commit/257f74599727b6a006ba65b1c3efd7ff5fc7b86c))
* **config:** initial push of the ability to use a config file for auth ([575c0cc](https://gitlab.com/gitlab-org/api/client-go/commit/575c0cc6a1de48582ea9b3b19cef021dc3f1397a))
* **integrations:** add group integration for microsoft teams ([da0b1dd](https://gitlab.com/gitlab-org/api/client-go/commit/da0b1dd5b86fd6a41d7c043621611d0687fc628f))
* **merge-requests:** add auto_merge, deprecate old field, for merging a request ([9119eb0](https://gitlab.com/gitlab-org/api/client-go/commit/9119eb0e6662f136e589cdee74aaa410136ca664))

# [0.134.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.133.1...v0.134.0) (2025-07-07)


### Features

* **oauth:** implement OAuth2 helper package ([a44e8eb](https://gitlab.com/gitlab-org/api/client-go/commit/a44e8eb7743ff8d948f396b9849a82a7d7d6d6c4))

## [0.133.1](https://gitlab.com/gitlab-org/api/client-go/compare/v0.133.0...v0.133.1) (2025-07-07)


### Bug Fixes

* deprecate ProjectReposityStorage due to a typo ([38a9652](https://gitlab.com/gitlab-org/api/client-go/commit/38a965279a4c570fd4db4f08503a63c4e7177439))

# [0.133.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.132.0...v0.133.0) (2025-07-03)


### Features

* **testing:** allow to specify client options when creating test client ([9377147](https://gitlab.com/gitlab-org/api/client-go/commit/93771470166ce7c9097328b5e49f75a381c1720b))

# [0.132.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.131.0...v0.132.0) (2025-07-02)


### Bug Fixes

* **no-release:** fix body-max-line-length ([f5d6d05](https://gitlab.com/gitlab-org/api/client-go/commit/f5d6d05d5781cd4fc31fa647ed94d486a1f6fa72))


### Features

* add missing ref_protected property from PushWebhookEventType ([15d0224](https://gitlab.com/gitlab-org/api/client-go/commit/15d0224575e7a5415783466afffe6c6b7aaf5dec))
* add WithUserAgent client option ([3e8b80c](https://gitlab.com/gitlab-org/api/client-go/commit/3e8b80cd40b3d4ad54cb050ebd1b6e11b848869a))
* export various auth sources ([281e408](https://gitlab.com/gitlab-org/api/client-go/commit/281e4083beed2b88b035dddcb562982d4c412143))
* **serviceaccounts:** bring group service accounts in line with API ([a08974f](https://gitlab.com/gitlab-org/api/client-go/commit/a08974f284c043d4039495ed4b8f24ebeb256cdc))
* **serviceaccounts:** bring group service accounts in line with API ([fb582a4](https://gitlab.com/gitlab-org/api/client-go/commit/fb582a4bb523443984851bc1d4b0fb699cfa2a9f))

# [0.131.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.130.1...v0.131.0) (2025-07-01)


### Features

* add ScanAndCollect for pagination ([cbac9ae](https://gitlab.com/gitlab-org/api/client-go/commit/cbac9aed9bb3c7f8d175585a6d38baa3f2a7fbe1))
* add support for optional query params to get commit statuses ([e1b29ad](https://gitlab.com/gitlab-org/api/client-go/commit/e1b29adfd37db39aae4e1547f336b71d67efcdb8))

## [0.130.1](https://gitlab.com/gitlab-org/api/client-go/compare/v0.130.0...v0.130.1) (2025-06-11)


### Bug Fixes

* add missing nil check on create group with avatar ([3298a05](https://gitlab.com/gitlab-org/api/client-go/commit/3298a058f36962a86dea31587956863cd1ed7624))

# [0.130.0](https://gitlab.com/gitlab-org/api/client-go/compare/v0.129.0...v0.130.0) (2025-06-11)


### Bug Fixes

* **workflow:** the `release.config.mjs` file mustn't be hidden ([5d423a5](https://gitlab.com/gitlab-org/api/client-go/commit/5d423a55d5b577ebff50dc1a0905c6511b5a4d6f))


### Features

* add "emoji_events" support to group hooks ([c6b770f](https://gitlab.com/gitlab-org/api/client-go/commit/c6b770f350b11e1c9a7c4702ab25b865624b0d47))
* Add `active` to ListProjects ([7818155](https://gitlab.com/gitlab-org/api/client-go/commit/78181558db20647c22e7fed23e749ecafedad27b))
* add generated_file field for MergeRequestDiff ([4b95dac](https://gitlab.com/gitlab-org/api/client-go/commit/4b95dac3ef2b5aabe3040f592ba6378d081d7642))
* add support for `administrator` to Group `project_creation_level` enums ([664bbd7](https://gitlab.com/gitlab-org/api/client-go/commit/664bbd7e3c955c8068b895b1cf1540054ebc13c1))
* add the `WithTokenSource` client option ([6ccfcf8](https://gitlab.com/gitlab-org/api/client-go/commit/6ccfcf857a0a4a850168ecf9317e2e0b8a532173))
* add url field to MergeCommentEvent.merge_request ([bd639d8](https://gitlab.com/gitlab-org/api/client-go/commit/bd639d811c8e7965f426c2deccee84a12d32920f))
* implement a specialized `TokenSource` interface ([83c2e06](https://gitlab.com/gitlab-org/api/client-go/commit/83c2e06cbe76b5268e55589e8bc580582e65bb22))
* **projects:** add ci_push_repository_for_job_token_allowed parameter ([3d539f6](https://gitlab.com/gitlab-org/api/client-go/commit/3d539f66fd63ce4fec6fa7e4e546c9d2acd018f0))
* **terraform-states:** add Terraform States API ([082b81c](https://gitlab.com/gitlab-org/api/client-go/commit/082b81cd456d4b8020f6542daeb3f47c80ba38d0))
