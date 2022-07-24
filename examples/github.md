# GitHub Examples

> Note: all command line arguments can be set in your $HOME/.config/ghorg/conf.yaml

## GitHub Cloud

clone an **org**, using a token on the commandline

```
ghorg clone <github_org> --token=XXXXXX
```

clone all repos from a **github org** that are **prefixed** with "frontend" **into a folder** called "design_only", assumes ghorg conf.yaml is setup with a token

```
ghorg clone <github_org> --match-regex=^frontend --output-dir=design_only --token=XXXXXX
```

clone a **users** repos, assumes ghorg conf.yaml is setup with a token

```
ghorg clone <github_username> --clone-type=user --token=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2
```

## GitHub Enterprise

clone an **org**

```
ghorg clone <github_org> --base-url=https://<your-hosted-github>.com --token=XXXXXX
```

clone all repos from a **github org** that are **prefixed** with "frontend" **into a folder** called "design_only"

```
ghorg clone <github_org> --match-regex=^frontend --output-dir=design_only --base-url=https://<your-hosted-github>.com --token=XXXXXX
```

clone a **users** repos

```
ghorg clone <github_username> --clone-type=user --base-url=https://<your-hosted-github>.com --token=XXXXXX
```
