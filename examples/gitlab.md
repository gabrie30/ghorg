# GitLab Examples

> Note: all command line arguments can be permanently set in your `$HOME/.config/ghorg/conf.yaml` for more information see the [configuration](https://github.com/gabrie30/ghorg#configuration) section of the README.md.

To view all additional flags see the [sample-conf.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml) or use `ghorg clone --help`

## Things to know

1. There are differences in how ghorg works with GitLab on hosted instances vs GitLab cloud. Please make sure to follow the correct section below.

1. The `--preserve-dir` flag will mirror the nested directory structure of the groups/subgroups/projects locally to what is on GitLab. This prevents any name collisions with project names. If this flag is omitted all projects will be cloned into a single directory. If there are collisions with project names and `--preserve-dir` is not used the group/subgroup name will be prepended to those projects. An informational message will also be displayed during the clone to let you know if this happens.

1. For all versions of GitLab you can clone groups or subgroups.

1. The `--output-dir` flag overrides the default name given to the folder ghorg creates to clone repos into. The default will be the instance name when cloning `all-groups` or `all-users` or the `group` name when cloning a specific group. The exception is when you are cloning a subgroup and preserving the directory structure, then it will preserve the parent groups of the subgroup.

1. The `--preserve-scm-hostname` flag will always create a top level folder in your GHORG_ABSOLUTE_PATH_TO_CLONE_TO with the hostname of the instance you are cloning from. For gitlab cloud it will be `gitlab.com/` otherwise it will be what is set to the hostname of the `GHORG_SCM_BASE_URL`.

1. If the group name you are cloning has spaces, substitute the spaces with "-" e.g.

    ```sh
        # incorrect
        ghorg clone "my group" --scm=gitlab
        ghorg clone my group --scm=gitlab
    ```

    ```sh
        # correct
        ghorg clone my-group --scm=gitlab
    ```

## Hosted GitLab Instances

#### Cloning All Groups

> Note: "all-groups" only works on hosted GitLab instances running 13.0.1 or greater

> Note: You must set `--base-url` which is the url to your instance. If your instance requires an insecure connection you can use the `--insecure-gitlab-client` flag

1. Clone **all groups**, **preserving the directory structure** of subgroups

    ```sh
    ghorg clone all-groups --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX --preserve-dir
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── your.instance.gitlab.com
        ├── group1
        │   └── project1
        ├── group2
        │   └── project2
        └── group3
            └── subgroup1
                ├── project3
                └── project4
    ```

1. Clone **all groups**, **WITHOUT preserving the directory structure** of subgroups

    ```sh
    ghorg clone all-groups --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── your.instance.gitlab.com
        ├── project1
        ├── project2
        ├── project3
        └── project4
    ```

1. Clone **all groups**, **preserving the directory structure** of users, preserving scm hostname

    ```sh
    ghorg clone all-groups --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX --preserve-dir --preserve-scm-hostname
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── your.instance.gitlab.com
        └── all-groups
            ├── group1
            │   └── project1
            ├── group2
            │   └── project2
            └── group3
                ├── project3
                └── project4
    ```

#### Cloning Specific Groups

1. Clone **a specific group**, **preserving the directory structure** of subgroups

    ```sh
    ghorg clone group3 --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX --preserve-dir
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── group3
        └── subgroup1
            ├── project3
            └── project4
    ```

1. Clone **a specific group**, **WITHOUT preserving the directory structure** of subgroups

    ```sh
    ghorg clone group3 --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── group3
        ├── project3
        └── project4
    ```

1. Clone **a specific subgroup**, **WITHOUT preserving the directory structure** of subgroups

    ```sh
    ghorg clone group3/subgroup1 --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX
    ```

    This would produce a directory structure like, where `projectX` is a project in a subgroup nested inside `subgroup1`

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── group3
        └── subgroup1
            ├── projectX
            ├── project3
            └── project4
    ```

1. Clone **a specific subgroup**, **preserving the directory structure** of subgroups

    ```sh
    ghorg clone group3/subgroup1 --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX --preserve-dir
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── group3
        └── subgroup1
            ├── project3
            └── project4
                └── subgroup2
                    └── projectX

    ```

#### Cloning a Specific Users Repos

1. Clone a **user** on a **hosted gitlab** instance using a **token** for auth

    ```sh
    ghorg clone <gitlab_username> --clone-type=user --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── gitlab_username
        ├── project3
        └── project4
    ```

#### Cloning All Users Repos

> Note: "all-users" only works on hosted GitLab instances running 13.0.1 or greater

> Note: You must set `--base-url` which is the url to your instance. If your instance requires an insecure connection you can use the `--insecure-gitlab-client` flag

1. Clone **all users**, **preserving the directory structure** of users

    ```sh
    ghorg clone all-users --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX --preserve-dir
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── your.instance.gitlab.com
        ├── user1
        │   └── project1
        ├── user2
        │   └── project2
        └── user3
            ├── project3
            └── project4
    ```
1. Clone **all users**, **WITHOUT preserving the directory structure** of users

    ```sh
    ghorg clone all-users --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── your.instance.gitlab.com
        ├── user1-repo1
        └── user2-repo1
    ```

1. Clone **all users**, **preserving the directory structure** of users, preserving scm hostname

    ```sh
    ghorg clone all-users --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX --preserve-dir --preserve-scm-hostname
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── your.instance.gitlab.com
        └── all-users
            ├── user1
            │   └── project1
            ├── user2
            │   └── project2
            └── user3
                ├── project3
                └── project4
    ```

## Cloud GitLab Orgs

Examples below use the `gitlab-examples` GitLab cloud organization https://gitlab.com/gitlab-examples

1. clone **all groups**, **preserving the directory structure** of subgroups

    ```sh
    ghorg clone gitlab-examples --scm=gitlab --token=XXXXXX --preserve-dir
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── gitlab-examples
        ├── aws-sam
        ├── ci-debug-trace
        ├── clojure-web-application
        ├── cpp-example
        ├── cross-branch-pipelines
        ├── docker
        ├── docker-cloud
        ├── functions
        └── ...
    ```

1. clone only a **subgroup**, **preserving the directory structure** of subgroups

    ```sh
    ghorg clone gitlab-examples/wayne-enterprises --scm=gitlab --token=XXXXXX --preserve-dir
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
    └── gitlab-examples
        └── wayne-enterprises
            ├── wayne-aerospace
            │   └── mission-control
            ├── wayne-financial
            │   ├── corporate-website
            │   ├── customer-upload-tool
            │   ├── customer-web-portal
            │   ├── customer-web-portal-security-policy-project
            │   ├── datagenerator
            │   ├── mobile-app
            │   └── wayne-financial-security-policy-project
            └── wayne-industries
                ├── backend-controller
                └── microservice
    ```

1. clone only a **subgroup**, **WITHOUT preserving the directory structure** of subgroups

    ```sh
    ghorg clone gitlab-examples/wayne-enterprises --scm=gitlab --token=XXXXXX
    ```

    This would produce a directory structure like

    ```sh
    /GHORG_ABSOLUTE_PATH_TO_CLONE_TO
        └── wayne-enterprises
            ├── backend-controller
            ├── corporate-website
            ├── customer-upload-tool
            ├── customer-web-portal
            ├── customer-web-portal-security-policy-project
            ├── datagenerator
            ├── microservice
            ├── mission-control
            ├── mobile-app
            └── wayne-financial-security-policy-project
    ```
