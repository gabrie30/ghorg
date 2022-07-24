# GitLab Examples

> Note: all command line arguments can be set in your $HOME/.config/ghorg/conf.yaml

## Hosted GitLab Instances


**Note: "all-groups" only works on hosted GitLab instances running 13.0.1 or greater**

clone all groups on a **hosted gitlab** instance **preserving** the directory structure of subgroups

```
ghorg clone all-groups --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX --preserve-dir
```

clone a **user** on a **hosted gitlab** instance using a **token** for auth

```
ghorg clone <gitlab_username> --clone-type=user --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2
```

clone a **group** on a **hosted gitlab** instance **preserving** the directory structure of subgroups

```
ghorg clone <gitlab_group> --base-url=https://<your.instance.gitlab.com> --scm=gitlab --preserve-dir
```

clone only a **subgroup** on a **hosted gitlab**

```
ghorg clone <gitlab_group>/<gitlab_sub_group> --base-url=https://<your.instance.gitlab.com> --scm=gitlab
```

clone all repos that are **prefixed** with "frontend" **into a folder** called "design_only" from a **group** on a **hosted gitlab** instance

```
ghorg clone <gitlab_group> --base-url=https://<your.instance.gitlab.com> --scm=gitlab --match-regex=^frontend --output-dir=design_only
```

## Cloud GitLab Orgs

eg. https://gitlab.com/gitlab-examples

clone all groups **preserving** the directory structure of subgroups

```
ghorg clone gitlab-examples --scm=gitlab --token=XXXXXX --preserve-dir
```

clone only a **subgroup**

```
ghorg clone gitlab-examples/wayne-enterprises --scm=gitlab --token=XXXXXX
```
