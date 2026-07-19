# Gitea Examples

> Note: all command line arguments can be permanently set in your `$HOME/.config/ghorg/conf.yaml` for more information see the [configuration](https://github.com/gabrie30/ghorg#configuration) section of the README.md.

To view all additional flags see the [sample-conf.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml) or use `ghorg clone --help`. You can also read this page in your terminal with `ghorg examples gitea`.

## Quick Start

Clone every repo in an org from your Gitea instance, using an [Access Token](https://github.com/gabrie30/ghorg#gitea-setup) (Settings -> Applications -> Generate Token)

```
ghorg clone <gitea_org> --scm=gitea --base-url=https://<your-internal-gitea>.com --token=XXXXXXX
```

Which will produce the following

```sh
$HOME/ghorg
└── gitea_org
    ├── repo1
    ├── repo2
    └── ...
```

## Things to know

1. Gitea is self hosted only, so `--base-url` pointing at your instance is always required. If your instance is served over plain HTTP add `--insecure-gitea-client`.

1. The `--token` flag also accepts a path to a file containing the token e.g. `--token=~/.config/ghorg/gitea-token.txt`.

1. Running the same clone a second time will `git pull` and `git clean` every repo, overwriting local changes. If you work inside the clone directory use `--no-clean` or `--protect-local` (see [Don't miss these features](#dont-miss-these-features)).

1. The `--preserve-scm-hostname` flag will always create a top level folder in your GHORG_ABSOLUTE_PATH_TO_CLONE_TO with the hostname of the `GHORG_SCM_BASE_URL` you are cloning from.

## Examples

1. Clone an **org**

    ```
    ghorg clone <gitea_org> --scm=gitea --base-url=https://<your-internal-gitea>.com --token=XXXXXXX
    ```

1. Clone a **users** repos

    ```
    ghorg clone <gitea_username> --scm=gitea --clone-type=user --base-url=https://<your-internal-gitea>.com --token=XXXXXXX
    ```

1. Clone all repos from a **gitea org** that are **prefixed** with "frontend" **into a folder** called "design_only"

    ```
    ghorg clone <gitea_org> --scm=gitea --match-regex=^frontend --output-dir=design_only --base-url=https://<your-internal-gitea>.com --token=XXXXXXX
    ```

    Will produce the following

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── design_only
        ├── frontend-admin
        ├── frontend-dashboard
        └── frontend-website
    ```

1. Clone an **org** from an instance served over **HTTP**

    ```
    ghorg clone <gitea_org> --scm=gitea --base-url=http://<your-internal-gitea>.com --token=XXXXXXX --insecure-gitea-client
    ```

1. Clone an **org**, preserving the scm hostname, useful when you clone from several SCM providers or instances

    ```
    ghorg clone <gitea_org> --scm=gitea --base-url=https://<your-internal-gitea>.com --token=XXXXXXX --preserve-scm-hostname
    ```

    Will produce the following

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── your-internal-gitea.com
        └── gitea_org
            ├── repo1
            └── repo2
    ```

## Don't Miss These Features

Cross provider flags that are easy to overlook — `--dry-run`, `--protect-local`, `--prune`, `--backup`, `--clone-depth=1`, `--stats-enabled`, ghorgignore/ghorgonly files, and more — are documented in one place in [examples/features.md](https://github.com/gabrie30/ghorg/blob/master/examples/features.md), or run `ghorg examples features`.
