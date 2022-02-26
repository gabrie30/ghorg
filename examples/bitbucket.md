# Bitbucket Cloud

Clone the microsoft workspace using an app-password

```
ghorg clone microsoft --scm=bitbucket --bitbucket-username=<your-username> --token=<app-password>
```

Clone the microsoft workspace using oauth token

```
ghorg clone microsoft --scm=bitbucket --token=<oauth-token>
```

# Hosted Bitbucket

Clone the foobar workspace on a hosted bitbucket instance using an app-password

```
ghorg clone foobar --scm=bitbucket --bitbucket-username=<your-username> --token=<app-password> --base-url=https://api.myhostedbb.com/v2
```
