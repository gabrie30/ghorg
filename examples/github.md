# GitHub Examples

> Note: all command line arguments can be permanently set in your `$HOME/.config/ghorg/conf.yaml` for more information see the [configuration](https://github.com/gabrie30/ghorg#configuration) section of the README.md.

To view all additional flags see the [sample-conf.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml) or use `ghorg clone --help`

## Things to know

1. The `--preserve-scm-hostname` flag will always create a top level folder in your GHORG_ABSOLUTE_PATH_TO_CLONE_TO with the hostname of the instance you are cloning from. For github cloud it will be `github.com/` otherwise it will be what is set to the hostname of the `GHORG_SCM_BASE_URL` if set for enterpirse github orgs.

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

1. Clone a **users** repos, assumes ghorg conf.yaml is setup with a token

    ```
    ghorg clone <github_username> --clone-type=user --token=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2
    ```

1. Clone an **org**, preserving the cloud scm hostname
    ```
    ghorg clone <github_org> --preserve-cloud-scm-hostname
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

1. Clone an **org**, preserving the cloud scm hostname. The github_org may seem redudant in this example but when its needed for the case when you start to set an output directory.
    ```
    ghorg clone <github_org> --base-url=https://<your-hosted-github>.com --preserve-cloud-scm-hostname
    ```
    Will produce the following

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── your-hosted-github.com
        └── github_org
            ├── repo1
            └── repo2
    ```

1. Clone an **org**, preserving the cloud scm hostname **with an output directory**
    ```
    ghorg clone <github_org> --base-url=https://<your-hosted-github>.com --output-dir=my-repos --preserve-cloud-scm-hostname
    ```
    Will produce the following

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── your-hosted-github.com
        └── my-repos
            ├── myrepo1
            └── myrepo2
    ```
