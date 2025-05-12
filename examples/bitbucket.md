# Bitbucket Examples

> Note: all command line arguments can be permanently set in your `$HOME/.config/ghorg/conf.yaml` for more information see the [configuration](https://github.com/gabrie30/ghorg#configuration) section of the README.md.

To view all additional flags see the [sample-conf.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml) or use `ghorg clone --help`

1. The `--preserve-scm-hostname` flag will always create a top level folder in your GHORG_ABSOLUTE_PATH_TO_CLONE_TO with the hostname of the instance you are cloning from. For bitbucket cloud it will be `bitbucket.com/` otherwise it will be what is set to the hostname of the `GHORG_SCM_BASE_URL`.

## Bitbucket Cloud

1. Clone the microsoft workspace using an app-password

    ```
    ghorg clone microsoft --scm=bitbucket --bitbucket-username=<your-username> --token=<app-password>
    ```

1. Clone the microsoft workspace using oauth token

    ```
    ghorg clone microsoft --scm=bitbucket --token=<oauth-token>
    ```
