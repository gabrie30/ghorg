clone a **users** repos

```
$ ghorg clone <github_username> --clone-type=user
```

clone an **org**

```
$ ghorg clone <github_org>
```

clone all repos from a **github org** that are **prefixed** with "frontend" **into a folder** called "design_only"

```
$ ghorg clone <github_org> --match-regex=^frontend --output-dir=design_only
```
