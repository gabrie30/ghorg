> Note: all flags can be permanently set in your $HOME/.config/ghorg/conf.yaml

clone a **user** on a **hosted gitlab** instance using a **token** for auth

```
$ ghorg clone <gitlab_username> --clone-type=user --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXXXXXXXXX
```

clone a **group** on a **hosted gitlab** instance **preserving** the directory structure of subgroups

```
$ ghorg clone <gitlab_group> --base-url=https://<your.instance.gitlab.com> --scm=gitlab --preserve-dir
```

clone only a **subgroup** on a **hosted gitlab**

```
$ ghorg clone <gitlab_group>/<gitlab_sub_group> --base-url=https://<your.instance.gitlab.com> --scm=gitlab
```

clone all repos that are **prefixed** with "frontend" **into a folder** called "design_only" from a **group** on a **hosted gitlab** instance

```
$ ghorg clone <gitlab_group> --base-url=https://<your.instance.gitlab.com> --scm=gitlab --match-prefix=frontend --output-dir=design_only
```
