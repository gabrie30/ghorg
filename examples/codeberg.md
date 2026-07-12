# Codeberg Examples

> Note: all command line arguments can be permanently set in your `$HOME/.config/ghorg/conf.yaml` for more information see the [configuration](https://github.com/gabrie30/ghorg#configuration) section of the README.md.

To view all additional flags see the [sample-conf.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml) or use `ghorg clone --help`

## Things to know

1. [Codeberg](https://codeberg.org) runs [Forgejo](https://forgejo.org), which is API-compatible with Gitea. The `codeberg` scm reuses ghorg's Gitea backend and simply defaults the base URL to `https://codeberg.org`, so no `--base-url` is required. Self-hosted Forgejo instances are also supported by setting `--base-url`.
1. Create a token at https://codeberg.org/user/settings/applications with at least the `read:organization` and `read:repository` scopes.
1. The `--preserve-scm-hostname` flag will always create a top level folder in your GHORG_ABSOLUTE_PATH_TO_CLONE_TO with the hostname of the `GHORG_SCM_BASE_URL` you are cloning from (e.g. `codeberg.org`).

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

1. Clone from a **self-hosted Forgejo instance**

    ```
    ghorg clone <org> --scm=codeberg --base-url=https://forgejo.yourinstance.com --token=XXXXXXX
    ```

1. Clone from a **self-hosted Forgejo instance using HTTP** (not recommended for production)

    ```
    ghorg clone <org> --scm=codeberg --base-url=http://forgejo.yourinstance.com --token=XXXXXXX --insecure-codeberg-client
    ```
