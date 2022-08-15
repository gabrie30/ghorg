# Bitbucket Examples

> Note: all command line arguments can be permanently set in your `$HOME/.config/ghorg/conf.yaml` for more information see the [configuration](https://github.com/gabrie30/ghorg#configuration) section of the README.md.

To view all additional flags see the [sample-conf.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml) or use `ghorg clone --help`

## Bitbucket Cloud

1. Clone the microsoft workspace using an app-password

    ```
    ghorg clone microsoft --scm=bitbucket --bitbucket-username=<your-username> --token=<app-password>
    ```

1. Clone the microsoft workspace using oauth token

    ```
    ghorg clone microsoft --scm=bitbucket --token=<oauth-token>
    ```

## Hosted Bitbucket

1. Clone a workspace on a hosted bitbucket instance using an app-password

    ```
    ghorg clone <workspace> --scm=bitbucket --bitbucket-username=<your-username> --token=<app-password> --base-url=https://<api.myhostedbb.com>/v2
    ```
