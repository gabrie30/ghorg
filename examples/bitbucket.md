# Bitbucket Examples

> Note: all command line arguments can be set in your $HOME/.config/ghorg/conf.yaml

## Bitbucket Cloud

Clone the microsoft workspace using an app-password

```
ghorg clone microsoft --scm=bitbucket --bitbucket-username=<your-username> --token=<app-password>
```

Clone the microsoft workspace using oauth token

```
ghorg clone microsoft --scm=bitbucket --token=<oauth-token>
```

## Hosted Bitbucket

Clone a workspace on a hosted bitbucket instance using an app-password

```
ghorg clone <workspace> --scm=bitbucket --bitbucket-username=<your-username> --token=<app-password> --base-url=https://<api.myhostedbb.com>/v2
```
