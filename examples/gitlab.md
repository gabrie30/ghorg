# GitLab Examples

> Note: all command line arguments can be set in your $HOME/.config/ghorg/conf.yaml for more information see the [configuration](https://github.com/gabrie30/ghorg#configuration) section of the readme

## Hosted GitLab Instances


### Cloning All Groups

**Note: "all-groups" only works on hosted GitLab instances running 13.0.1 or greater**

1. Clone all groups **preserving the directory structure** of subgroups

    ```
    ghorg clone all-groups --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX --preserve-dir
    ```

1. Clone all groups on an **insecure** instance **preserving the directory structure** of subgroups

    ```
    ghorg clone all-groups --base-url=http://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX --preserve-dir --insecure-gitlab-client
### Cloning Specific Groups

1. Clone a single group, **preserving the directory structure** of any subgroups within that group

    ```
    ghorg clone <gitlab_group> --base-url=https://<your.instance.gitlab.com> --scm=gitlab --preserve-dir
    ```

1. Clone only a **subgroup**

    ```
    ghorg clone <gitlab_group>/<gitlab_sub_group> --base-url=https://<your.instance.gitlab.com> --scm=gitlab
    ```

1. clone all repos that are **prefixed** with "frontend" **into a folder** called "design_only"

    ```
    ghorg clone <gitlab_group> --base-url=https://<your.instance.gitlab.com> --scm=gitlab --match-regex=^frontend --output-dir=design_only
    ```
### Cloning User Repos

1. Clone a **user** on a **hosted gitlab** instance using a **token** for auth

    ```
    ghorg clone <gitlab_username> --clone-type=user --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2
    ```

## Cloud GitLab Orgs

eg. https://gitlab.com/gitlab-examples

1. clone **all groups**, **preserving the directory structure** of subgroups

    ```
    ghorg clone gitlab-examples --scm=gitlab --token=XXXXXX --preserve-dir
    ```

1. clone only a **subgroup**

    ```
    ghorg clone gitlab-examples/wayne-enterprises --scm=gitlab --token=XXXXXX
    ```
