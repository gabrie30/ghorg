# Sourcehut Examples

> Note: all command line arguments can be permanently set in your `$HOME/.config/ghorg/conf.yaml` for more information see the [configuration](https://github.com/gabrie30/ghorg#configuration) section of the README.md.

To view all additional flags see the [sample-conf.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml) or use `ghorg clone --help`

## Things to know

1. Sourcehut uses GraphQL API which requires a Personal Access Token with REPOSITORIES and OBJECTS permissions.
2. For private repos, you'll need to use SSH protocol as sourcehut does not accept PATs in HTTPS URLs for authentication.
3. The `--preserve-scm-hostname` flag will always create a top level folder in your GHORG_ABSOLUTE_PATH_TO_CLONE_TO with the hostname of the `GHORG_SCM_BASE_URL` you are cloning from.
4. **Username format**: You can provide usernames with or without the `~` prefix (e.g., both `gabrie30` and `~gabrie30` work). Local folder paths will never include the `~` prefix to avoid shell expansion issues.

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

## Sourcehut-Specific Notes

- **Organizations**: Sourcehut's organization concept differs from other SCM providers. Currently, ghorg treats user and org cloning identically, fetching all repos owned by the specified username.
- **Private Repositories**: Must use SSH protocol (`--protocol=ssh`) as sourcehut doesn't support token authentication in HTTPS clone URLs.
- **API Limitations**:
  - Sourcehut's GraphQL API doesn't currently expose archived/fork status, so `--skip-archived` and `--skip-forks` flags won't have any effect.
  - The API doesn't support server-side filtering by owner, so ghorg fetches all accessible repos and filters client-side.
- **Topics/Tags**: Topic filtering (`GHORG_TOPICS`) is not currently supported as this information isn't available in the GraphQL API response.

