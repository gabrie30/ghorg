# GitLab Examples

> Note: all command line arguments can be permanently set in your `$HOME/.config/ghorg/conf.yaml` for more information see the [configuration](https://github.com/gabrie30/ghorg#configuration) section of the README.md.

To view all additional flags see the [sample-conf.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml) or use `ghorg clone --help`. You can also read this page in your terminal with `ghorg examples gitlab`.

## Quick Start

Clone a top level group and mirror its subgroup structure locally, using a [Personal Access Token](https://github.com/gabrie30/ghorg#gitlab-setup) with the `read_api` scope

```
ghorg clone gitlab-examples --scm=gitlab --preserve-dir --token=XXXXXX
```

Which will produce the following

```sh
$HOME/ghorg
└── gitlab-examples
    ├── project1
    ├── project2
    └── subgroup1
        └── project3
```

## Things to know

GitLab works differently in ghorg than every other SCM provider. Read this section before cloning.

1. **Hosted GitLab vs GitLab cloud behave differently.** The special targets `all-groups` and `all-users` only work on self hosted GitLab instances (13.0.1 or greater). On gitlab.com you always clone a specific group or subgroup. Make sure to follow the correct section below.

1. GitLab organizes projects into **groups and subgroups**, and ghorg can clone at any level: a top level group, a single subgroup (`group/subgroup`), or every group on a hosted instance (`all-groups`).

1. The `--preserve-dir` flag will mirror the nested directory structure of the groups/subgroups/projects locally to what is on GitLab. This prevents any name collisions with project names. If this flag is omitted all projects are cloned into a single flat directory. If there are collisions with project names and `--preserve-dir` is not used the group/subgroup name will be prepended to those projects and an informational message will be displayed during the clone.

1. The `--output-dir` flag overrides the default name given to the folder ghorg creates to clone repos into. The default will be the instance hostname when cloning `all-groups` or `all-users`, or the `group` name when cloning a specific group. The exception is when you are cloning a subgroup and preserving the directory structure, then it will preserve the parent groups of the subgroup.

1. The `--preserve-scm-hostname` flag will always create a top level folder in your GHORG_ABSOLUTE_PATH_TO_CLONE_TO with the hostname of the instance you are cloning from. For GitLab cloud it will be `gitlab.com/` otherwise it will be the hostname of the `GHORG_SCM_BASE_URL`.

1. When cloning a group whose name contains spaces, use the group **path** (dashes) not its display name e.g.

    ```sh
        # incorrect
        ghorg clone "my group" --scm=gitlab
        ghorg clone my group --scm=gitlab
    ```

    ```sh
        # correct
        ghorg clone my-group --scm=gitlab
    ```

1. Your token can also be given as a path to a file containing the token e.g. `--token=~/.config/ghorg/gitlab-token.txt`.

## Hosted GitLab Instances

> Note: You must set `--base-url` which is the url to your instance. If your instance requires an insecure connection you can use the `--insecure-gitlab-client` flag

#### Cloning All Groups

> Note: "all-groups" only works on hosted GitLab instances running 13.0.1 or greater

1. Clone **all groups**, **preserving the directory structure** of subgroups

    ```sh
    ghorg clone all-groups --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX --preserve-dir
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── your.instance.gitlab.com
        ├── group1
        │   └── project1
        ├── group2
        │   └── project2
        └── group3
            └── subgroup1
                ├── project3
                └── project4
    ```

1. Clone **all groups**, **WITHOUT preserving the directory structure** of subgroups

    ```sh
    ghorg clone all-groups --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── your.instance.gitlab.com
        ├── project1
        ├── project2
        ├── project3
        └── project4
    ```

1. Clone **all groups**, **preserving the directory structure** of subgroups, preserving scm hostname

    ```sh
    ghorg clone all-groups --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX --preserve-dir --preserve-scm-hostname
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── your.instance.gitlab.com
        └── all-groups
            ├── group1
            │   └── project1
            ├── group2
            │   └── project2
            └── group3
                ├── project3
                └── project4
    ```

1. Clone **all groups except those matching a regex**, useful for skipping archived or sandbox groups

    ```sh
    ghorg clone all-groups --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX --preserve-dir --gitlab-group-exclude-match-regex=^sandbox
    ```

    The opposite flag `--gitlab-group-match-regex` clones **only** groups matching the regex. Both flags are GitLab only.

#### Cloning Specific Groups

1. Clone **a specific group**, **preserving the directory structure** of subgroups

    ```sh
    ghorg clone group3 --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX --preserve-dir
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── group3
        └── subgroup1
            ├── project3
            └── project4
    ```

1. Clone **a specific group**, **WITHOUT preserving the directory structure** of subgroups

    ```sh
    ghorg clone group3 --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── group3
        ├── project3
        └── project4
    ```

1. Clone **a specific subgroup**, **WITHOUT preserving the directory structure** of subgroups

    ```sh
    ghorg clone group3/subgroup1 --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX
    ```

    This would produce a directory structure like, where `projectX` is a project in a subgroup nested inside `subgroup1`. Note that projects from nested subgroups will appear both flattened and in their original structure.

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── group3
        └── subgroup1
            ├── project3
            ├── project4
            ├── projectX
            └── subgroup2
                └── projectX
    ```

1. Clone **a specific subgroup**, **preserving the directory structure** of subgroups

    ```sh
    ghorg clone group3/subgroup1 --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX --preserve-dir
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── group3
        └── subgroup1
            ├── project3
            └── project4
                └── subgroup2
                    └── projectX

    ```

#### Cloning a Specific Users Repos

1. Clone a **user** on a **hosted gitlab** instance using a **token** for auth

    ```sh
    ghorg clone <gitlab_username> --clone-type=user --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── gitlab_username
        ├── project3
        └── project4
    ```

#### Cloning All Users Repos

> Note: "all-users" only works on hosted GitLab instances running 13.0.1 or greater

> Note: When using "all-users", you must include the `--clone-type=user` flag

1. Clone **all users**, **preserving the directory structure** of users

    ```sh
    ghorg clone all-users --clone-type=user --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX --preserve-dir
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── your.instance.gitlab.com
        ├── user1
        │   └── project1
        ├── user2
        │   └── project2
        └── user3
            ├── project3
            └── project4
    ```

1. Clone **all users**, **WITHOUT preserving the directory structure** of users

    ```sh
    ghorg clone all-users --clone-type=user --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── your.instance.gitlab.com
        ├── user1-repo1
        └── user2-repo1
    ```

1. Clone **all users**, **preserving the directory structure** of users, preserving scm hostname

    ```sh
    ghorg clone all-users --clone-type=user --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX --preserve-dir --preserve-scm-hostname
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── your.instance.gitlab.com
        └── all-users
            ├── user1
            │   └── project1
            ├── user2
            │   └── project2
            └── user3
                ├── project3
                └── project4
    ```

## GitLab Cloud (gitlab.com)

> Note: `all-groups` and `all-users` are **not** available on gitlab.com; clone a specific group or subgroup instead.

Examples below use the `gitlab-examples` GitLab cloud group https://gitlab.com/gitlab-examples

1. Clone a **group**, **preserving the directory structure** of subgroups

    ```sh
    ghorg clone gitlab-examples --scm=gitlab --token=XXXXXX --preserve-dir
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── gitlab-examples
        ├── aws-sam
        ├── ci-debug-trace
        ├── clojure-web-application
        ├── cpp-example
        ├── cross-branch-pipelines
        ├── docker
        ├── docker-cloud
        ├── functions
        └── ...
    ```

1. Clone only a **subgroup**, **preserving the directory structure** of subgroups

    ```sh
    ghorg clone gitlab-examples/wayne-enterprises --scm=gitlab --token=XXXXXX --preserve-dir
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── gitlab-examples
        └── wayne-enterprises
            ├── wayne-aerospace
            │   └── mission-control
            ├── wayne-financial
            │   ├── corporate-website
            │   ├── customer-upload-tool
            │   ├── customer-web-portal
            │   ├── customer-web-portal-security-policy-project
            │   ├── datagenerator
            │   ├── mobile-app
            │   └── wayne-financial-security-policy-project
            └── wayne-industries
                ├── backend-controller
                └── microservice
    ```

1. Clone only a **subgroup**, **WITHOUT preserving the directory structure** of subgroups

    ```sh
    ghorg clone gitlab-examples/wayne-enterprises --scm=gitlab --token=XXXXXX
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
        └── wayne-enterprises
            ├── backend-controller
            ├── corporate-website
            ├── customer-upload-tool
            ├── customer-web-portal
            ├── customer-web-portal-security-policy-project
            ├── datagenerator
            ├── microservice
            ├── mission-control
            ├── mobile-app
            └── wayne-financial-security-policy-project
    ```

## GitLab Only Features

1. `--clone-snippets` additionally clones every snippet. Snippets belonging to a project are placed in a `<project>.snippets` folder next to the project, and instance level snippets go into `_ghorg_root_level_snippets`; each snippet folder is named `<title>-<id>` so they never collide

    ```sh
    ghorg clone gitlab-examples --scm=gitlab --clone-snippets --token=XXXXXX
    ```

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── gitlab-examples
        ├── docker
        ├── docker.snippets
        │   └── my-snippet-2891763
        └── _ghorg_root_level_snippets
            └── some-root-snippet-1747392
    ```

1. `--gitlab-include-shared-projects=false` skips projects that are only *shared with* a group rather than owned by it (shared projects are included by default)

1. `--gitlab-group-match-regex` and `--gitlab-group-exclude-match-regex` filter entire groups/subgroups, while `--match-regex`/`--exclude-match-regex` filter individual projects; they can be combined

## Don't Miss These Features

Cross provider flags that are easy to overlook — `--dry-run`, `--protect-local`, `--prune`, `--backup`, `--clone-depth=1`, `--stats-enabled`, ghorgignore/ghorgonly files, and more — are documented in one place in [examples/features.md](https://github.com/gabrie30/ghorg/blob/master/examples/features.md), or run `ghorg examples features`.
