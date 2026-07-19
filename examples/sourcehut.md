# Sourcehut Examples

> Note: all command line arguments can be permanently set in your `$HOME/.config/ghorg/conf.yaml` for more information see the [configuration](https://github.com/gabrie30/ghorg#configuration) section of the README.md.

To view all additional flags see the [sample-conf.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml) or use `ghorg clone --help`. You can also read this page in your terminal with `ghorg examples sourcehut`.

## Quick Start

Clone all of a user's repos using a [Personal Access Token](https://meta.sr.ht/oauth2) with REPOSITORIES and OBJECTS permissions

```
ghorg clone <sourcehut_username> --scm=sourcehut --token=XXXXXXX
```

Which will produce the following

```sh
$HOME/ghorg
└── sourcehut_username
    ├── repo1
    ├── repo2
    └── ...
```

## Things to know

1. Sourcehut uses a GraphQL API which requires a Personal Access Token with REPOSITORIES and OBJECTS permissions.

1. For **private repos you must use SSH** (`--protocol=ssh`) as sourcehut does not accept PATs in HTTPS URLs for authentication.

1. **Username format**: You can provide usernames with or without the `~` prefix (e.g., both `gabrie30` and `~gabrie30` work). Local folder paths will never include the `~` prefix to avoid shell expansion issues.

1. **Organizations**: Sourcehut's organization concept differs from other SCM providers. Currently, ghorg treats user and org cloning identically, fetching all repos owned by the specified username.

1. The `--token` flag also accepts a path to a file containing the token e.g. `--token=~/.config/ghorg/sourcehut-token.txt`.

1. The `--preserve-scm-hostname` flag will always create a top level folder in your GHORG_ABSOLUTE_PATH_TO_CLONE_TO with the hostname of the instance you are cloning from (`git.sr.ht` for sourcehut cloud).

1. Sourcehut is a small independent service; consider lowering `--concurrency` (default 25) or adding `--clone-delay-seconds=1` when cloning many repos.

## Examples

1. Clone a **user's** repos (default behavior for sourcehut)

    ```
    ghorg clone <sourcehut_username> --scm=sourcehut --token=XXXXXXX
    ```

    Or with the `~` prefix (both work):

    ```
    ghorg clone ~<sourcehut_username> --scm=sourcehut --token=XXXXXXX
    ```

1. Clone a **user's** repos using **SSH protocol** (required for private repos)

    ```
    ghorg clone <sourcehut_username> --scm=sourcehut --token=XXXXXXX --protocol=ssh
    ```

1. Clone all repos from a **sourcehut user** that are **prefixed** with "go-" **into a folder** called "golang_projects"

    ```
    ghorg clone <sourcehut_username> --scm=sourcehut --match-regex=^go- --output-dir=golang_projects --token=XXXXXXX
    ```

    Will produce the following

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── golang_projects
        ├── go-cli
        ├── go-parser
        └── go-server
    ```

1. Clone from a **self-hosted sourcehut instance**

    ```
    ghorg clone <username> --scm=sourcehut --base-url=https://git.yourinstance.com --token=XXXXXXX
    ```

1. Clone from a **self-hosted sourcehut instance using HTTP** (not recommended for production)

    ```
    ghorg clone <username> --scm=sourcehut --base-url=http://git.yourinstance.com --token=XXXXXXX --insecure-sourcehut-client
    ```

1. Clone and checkout a specific **branch** for all repos

    ```
    ghorg clone <sourcehut_username> --scm=sourcehut --token=XXXXXXX --branch=develop
    ```

## API Limitations

Sourcehut's GraphQL API does not expose everything ghorg can use on other providers, so a few flags have no effect:

- `--skip-archived` and `--skip-forks`: archived/fork status is not exposed by the API.
- `--topics` / `GHORG_TOPICS`: topic information isn't available in the GraphQL API response.
- The API doesn't support server-side filtering by owner, so ghorg fetches all accessible repos and filters client-side.

## Don't Miss These Features

Cross provider flags that are easy to overlook — `--dry-run`, `--protect-local`, `--prune`, `--backup`, `--clone-depth=1`, `--stats-enabled`, ghorgignore/ghorgonly files, and more — are documented in one place in [examples/features.md](https://github.com/gabrie30/ghorg/blob/master/examples/features.md), or run `ghorg examples features`.
