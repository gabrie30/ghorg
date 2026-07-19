# Codeberg Examples

> Note: all command line arguments can be permanently set in your `$HOME/.config/ghorg/conf.yaml` for more information see the [configuration](https://github.com/gabrie30/ghorg#configuration) section of the README.md.

To view all additional flags see the [sample-conf.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml) or use `ghorg clone --help`. You can also read this page in your terminal with `ghorg examples codeberg`.

## Quick Start

Clone every repo in a Codeberg org, using an [Access Token](https://codeberg.org/user/settings/applications)

```
ghorg clone <codeberg_org> --scm=codeberg --token=XXXXXXX
```

Which will produce the following

```sh
$HOME/ghorg
└── codeberg_org
    ├── repo1
    ├── repo2
    └── ...
```

## Things to know

1. [Codeberg](https://codeberg.org) runs [Forgejo](https://forgejo.org), which is API-compatible with Gitea. The `codeberg` scm reuses ghorg's Gitea backend and simply defaults the base URL to `https://codeberg.org`, so no `--base-url` is required. Self-hosted Forgejo instances are also supported by setting `--base-url`.

1. Create a token at https://codeberg.org/user/settings/applications with at least the `read:organization` and `read:repository` scopes.

1. The `--token` flag also accepts a path to a file containing the token e.g. `--token=~/.config/ghorg/codeberg-token.txt`.

1. Running the same clone a second time will `git pull` and `git clean` every repo, overwriting local changes. If you work inside the clone directory use `--no-clean` or `--protect-local` (see [Don't miss these features](#dont-miss-these-features)).

1. The `--preserve-scm-hostname` flag will always create a top level folder in your GHORG_ABSOLUTE_PATH_TO_CLONE_TO with the hostname of the instance you are cloning from (e.g. `codeberg.org`).

1. Codeberg is a donation funded nonprofit; consider lowering `--concurrency` (default 25) or adding `--clone-delay-seconds=1` when cloning large orgs from codeberg.org.

## Examples

1. Clone an **org**

    ```
    ghorg clone <codeberg_org> --scm=codeberg --token=XXXXXXX
    ```

1. Clone a **users** repos

    ```
    ghorg clone <codeberg_username> --scm=codeberg --clone-type=user --token=XXXXXXX
    ```

1. Clone all repos from a **codeberg org** that are **prefixed** with "frontend" **into a folder** called "design_only"

    ```
    ghorg clone <codeberg_org> --scm=codeberg --match-regex=^frontend --output-dir=design_only --token=XXXXXXX
    ```

    Will produce the following

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── design_only
        ├── frontend-admin
        ├── frontend-dashboard
        └── frontend-website
    ```

1. Clone an **org**, preserving the scm hostname, useful when you clone from several SCM providers

    ```
    ghorg clone <codeberg_org> --scm=codeberg --token=XXXXXXX --preserve-scm-hostname
    ```

    Will produce the following

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── codeberg.org
        └── codeberg_org
            ├── repo1
            └── repo2
    ```

1. Clone from a **self-hosted Forgejo instance**

    ```
    ghorg clone <org> --scm=codeberg --base-url=https://forgejo.yourinstance.com --token=XXXXXXX
    ```

1. Clone from a **self-hosted Forgejo instance using HTTP** (not recommended for production)

    ```
    ghorg clone <org> --scm=codeberg --base-url=http://forgejo.yourinstance.com --token=XXXXXXX --insecure-codeberg-client
    ```

## Don't Miss These Features

Cross provider flags that are easy to overlook — `--dry-run`, `--protect-local`, `--prune`, `--backup`, `--clone-depth=1`, `--stats-enabled`, ghorgignore/ghorgonly files, and more — are documented in one place in [examples/features.md](https://github.com/gabrie30/ghorg/blob/master/examples/features.md), or run `ghorg examples features`.
