# Bitbucket Examples

> Note: all command line arguments can be permanently set in your `$HOME/.config/ghorg/conf.yaml` for more information see the [configuration](https://github.com/gabrie30/ghorg#configuration) section of the README.md.

To view all additional flags see the [sample-conf.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml) or use `ghorg clone --help`

1. The `--preserve-scm-hostname` flag will always create a top level folder in your GHORG_ABSOLUTE_PATH_TO_CLONE_TO with the hostname of the instance you are cloning from. For bitbucket cloud it will be `bitbucket.com/` otherwise it will be what is set to the hostname of the `GHORG_SCM_BASE_URL`.

## Bitbucket Cloud

### API Token Authentication (Recommended)

Bitbucket has deprecated App Passwords in favor of API Tokens. This is the recommended authentication method.

**Creating the API Token:**
1. Go to your [Atlassian account settings](https://id.atlassian.com/manage/api-tokens)
2. Create a new API token
3. **Important**: Grant **all read scopes** (Account: Read, Workspace membership: Read, Projects: Read, Repositories: Read) to ensure ghorg can list and clone repositories

**Using the API Token:**

1. Clone the microsoft workspace using an API token

    ```
    ghorg clone microsoft --scm=bitbucket --bitbucket-api-email=<your-atlassian-email> --token=<api-token>
    ```

1. Using environment variables (recommended for scripts)

    ```
    export GHORG_BITBUCKET_API_TOKEN=<api-token>
    export GHORG_BITBUCKET_API_EMAIL=<your-atlassian-email>
    ghorg clone microsoft --scm=bitbucket
    ```

> Note: When using API tokens, ghorg automatically uses `x-bitbucket-api-token-auth` as the Git username for clone operations. The email is only used for API calls to list repositories.

### App Password Authentication (Legacy)

> Note: Bitbucket has deprecated App Passwords. Consider using API Tokens instead.

1. Clone the microsoft workspace using an app-password

    ```
    ghorg clone microsoft --scm=bitbucket --bitbucket-username=<your-username> --token=<app-password>
    ```

### OAuth Token Authentication

1. Clone the microsoft workspace using oauth token

    ```
    ghorg clone microsoft --scm=bitbucket --token=<oauth-token>
    ```

## Bitbucket Server (Self-hosted)

1. Clone a project using username and password

    ```
    ghorg clone PROJECT_KEY --scm=bitbucket --base-url=https://bitbucket.company.com --bitbucket-username=<your-username> --token=<your-password>
    ```

1. Clone a project with insecure HTTP connection

    ```
    GHORG_INSECURE_BITBUCKET_CLIENT=true ghorg clone PROJECT_KEY --scm=bitbucket --base-url=http://bitbucket.company.com --bitbucket-username=<your-username> --token=<your-password>
    ```

1. Clone all repositories the user has access to

    ```
    ghorg clone <username> --clone-type=user --scm=bitbucket --base-url=https://bitbucket.company.com --bitbucket-username=<your-username> --token=<your-password>
    ```
