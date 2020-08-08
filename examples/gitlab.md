clone a **user** on a **hosted gitlab** instance

```
$ ghorg clone <gitlab_username> --clone-type=user --base-url=https://<your.instance.gitlab.com> --scm=gitlab
```

clone a **group** on a **hosted gitlab** instance

```
$ ghorg clone <gitlab_group> --base-url=https://<your.instance.gitlab.com> --scm=gitlab
```

clone all repos that are **prefixed** with "frontend" **into a folder** called "design_only" from a **group** on a **hosted gitlab** instance

```
$ ghorg clone <gitlab_group> --base-url=https://<your.instance.gitlab.com> --scm=gitlab --match-prefix=frontend --output-dir=design_only
```
