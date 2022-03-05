# GitHub Cloud

clone a **users** repos, assumes ghorg conf.yaml is setup with a token

```
$ ghorg clone <github_username> --clone-type=user
```

clone an **org**, using a token on the commandline

```
$ ghorg clone <github_org> --token=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2
```

clone all repos from a **github org** that are **prefixed** with "frontend" **into a folder** called "design_only", assumes ghorg conf.yaml is setup with a token

```
$ ghorg clone <github_org> --match-regex=^frontend --output-dir=design_only
```

# GitHub Enterprise

clone a **users** repos

```
$ ghorg clone <github_username> --clone-type=user --base-url=https://internal.github.com --token=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2
```

clone an **org**

```
$ ghorg clone <github_org> --base-url=https://your.internal.github.com --token=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2
```

clone all repos from a **github org** that are **prefixed** with "frontend" **into a folder** called "design_only"

```
$ ghorg clone <github_org> --match-regex=^frontend --output-dir=design_only --base-url=https://your.internal.github.com
```
