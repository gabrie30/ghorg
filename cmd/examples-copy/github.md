# GitHub Examples

> Note: all command line arguments can be permanently set in your `$HOME/.config/ghorg/conf.yaml` for more information see the [configuration](https://github.com/gabrie30/ghorg#configuration) section of the README.md.

To view all additional flags see the [sample-conf.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml) or use `ghorg clone --help`. You can also read this page in your terminal with `ghorg examples github`.

## Quick Start

Clone every repo in an org with a [Personal Access Token](https://github.com/gabrie30/ghorg#github-setup)

```
ghorg clone kubernetes --token=XXXXXX
```

Which will produce the following

```sh
$HOME/ghorg
└── kubernetes
    ├── apimachinery
    ├── kubectl
    ├── kubelet
    └── ...
```

## Things to know

1. Your token needs all `repo` scopes; if your org uses SAML SSO you must also [authorize the token for SSO](https://docs.github.com/en/github/authenticating-to-github/authenticating-with-saml-single-sign-on/authorizing-a-personal-access-token-for-use-with-saml-single-sign-on).

1. The `--token` flag also accepts a path to a file containing the token e.g. `--token=~/.config/ghorg/github-token.txt`, so you never have to put the token itself on the commandline.

1. By default repos are cloned to `$HOME/ghorg/<org>`. Change the base path with `--path` (must be absolute) and the folder name with `--output-dir`.

1. Running the same clone a second time will `git pull` and `git clean` every repo, overwriting local changes. If you work inside the clone directory use `--no-clean` or `--protect-local` (see [Don't miss these features](#dont-miss-these-features)).

1. The `--preserve-scm-hostname` flag will always create a top level folder in your GHORG_ABSOLUTE_PATH_TO_CLONE_TO with the hostname of the instance you are cloning from. For GitHub cloud it will be `github.com/` otherwise it will be the hostname of the `GHORG_SCM_BASE_URL` set for your GitHub Enterprise instance.

## GitHub Cloud

1. Clone an **org**, using a token on the commandline

    ```
    ghorg clone <github_org> --token=XXXXXX
    ```

    Will produce the following

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── github_org
        ├── repo1
        └── repo2
    ```

1. Clone all repos from a **github org** that are **prefixed** with "frontend" **into a folder** called "design_only"

    ```
    ghorg clone <github_org> --match-regex=^frontend --output-dir=design_only --token=XXXXXX
    ```

    Will produce the following

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── design_only
        ├── frontend-admin
        ├── frontend-dashboard
        └── frontend-website
    ```

1. Clone a **users** repos, assumes ghorg conf.yaml is setup with a token

    ```
    ghorg clone <github_username> --clone-type=user
    ```

    By default only repos the user **owns** are cloned. Use `--github-user-option` to widen the net to `all` or `member` (repos they contribute to)

    ```
    ghorg clone <github_username> --clone-type=user --github-user-option=all
    ```

1. Clone a **users** repos **and all their gists**; gists are placed in a `ghorg-gists` folder inside the clone

    ```
    ghorg clone <github_username> --clone-type=user --github-user-gists --token=XXXXXX
    ```

    Will produce the following

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── github_username
        ├── repo1
        ├── repo2
        └── ghorg-gists
            ├── gist1
            └── gist2
    ```

1. Clone only repos written in **go or ruby** (GitHub only feature)

    ```
    ghorg clone <github_org> --github-filter-language=go,ruby --token=XXXXXX
    ```

1. Clone only repos tagged with the **kubernetes or docker topic**

    ```
    ghorg clone <github_org> --topics=kubernetes,docker --token=XXXXXX
    ```

1. Clone an **org**, preserving the scm hostname

    ```
    ghorg clone <github_org> --preserve-scm-hostname --token=XXXXXX
    ```

    Will produce the following

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── github.com
        └── github_org
            ├── repo1
            └── repo2
    ```

## GitHub Enterprise

For self hosted GitHub instances set `--base-url` to the URL of your instance.

1. Clone an **org**

    ```
    ghorg clone <github_org> --base-url=https://<your-hosted-github>.com --token=XXXXXX
    ```

1. Clone all repos from a **github org** that are **prefixed** with "frontend" **into a folder** called "design_only"

    ```
    ghorg clone <github_org> --match-regex=^frontend --output-dir=design_only --base-url=https://<your-hosted-github>.com --token=XXXXXX
    ```

1. Clone a **users** repos

    ```
    ghorg clone <github_username> --clone-type=user --base-url=https://<your-hosted-github>.com --token=XXXXXX
    ```

1. Clone an **org**, preserving the scm hostname. The org folder may seem redundant here but it matters once you set an output directory.

    ```
    ghorg clone <github_org> --base-url=https://<your-hosted-github>.com --preserve-scm-hostname --token=XXXXXX
    ```

    Will produce the following

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── your-hosted-github.com
        └── github_org
            ├── repo1
            └── repo2
    ```

1. Clone an **org**, preserving the scm hostname **with an output directory**

    ```
    ghorg clone <github_org> --base-url=https://<your-hosted-github>.com --output-dir=my-repos --preserve-scm-hostname --token=XXXXXX
    ```

    Will produce the following

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── your-hosted-github.com
        └── my-repos
            ├── myrepo1
            └── myrepo2
    ```

## GitHub App Authentication

Instead of a personal token, ghorg can authenticate as a [GitHub App](https://github.com/gabrie30/ghorg#github-app-authentication-advanced), which is useful for org wide automation and backups

```
ghorg clone <github_org> --github-app-pem-path=/path/to/app.pem --github-app-id=123456 --github-app-installation-id=987654
```

If you already generated an app token outside of ghorg, pass it with `--token` plus `--github-token-from-github-app`

```
ghorg clone <github_org> --token=<app_token> --github-token-from-github-app
```

## Don't Miss These Features

Cross provider flags that are easy to overlook — `--dry-run`, `--protect-local`, `--prune`, `--backup`, `--clone-depth=1`, `--stats-enabled`, ghorgignore/ghorgonly files, and more — are documented in one place in [examples/features.md](https://github.com/gabrie30/ghorg/blob/master/examples/features.md), or run `ghorg examples features`.
