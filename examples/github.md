clone a **users** repos

```
$ ghorg clone <github_username> --clone-type=user
```

clone an **org**

```
$ ghorg clone <github_org>
```

clone all repos that are **prefixed** with "frontend" **into a folder** called "design_only" from a **group** on a **hosted gitlab** instance

```
$ ghorg clone <github_org> --match-prefix=frontend --output-dir=design_only
```
