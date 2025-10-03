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
