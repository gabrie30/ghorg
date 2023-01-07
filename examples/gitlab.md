# GitLab Examples

> Note: all command line arguments can be permanently set in your `$HOME/.config/ghorg/conf.yaml` for more information see the [configuration](https://github.com/gabrie30/ghorg#configuration) section of the README.md.

To view all additional flags see the [sample-conf.yaml](https://github.com/gabrie30/ghorg/blob/master/sample-conf.yaml) or use `ghorg clone --help`

#### Things to know

1. There are differences in how ghorg works with GitLab on hosted instances vs GitLab cloud. Please make sure to follow the correct section below.

1. With GitLab, if ghorg detects repo naming collisions with repos being cloned from different groups/subgroups, ghorg will automatically append the group/subgroup path to the repo name. You will be notified in the output if this occurs.

1. For all versions of GitLab you can clone groups or subgroups individually although the behavior is slightly different on hosted vs cloud GitLab

## Hosted GitLab Instances

#### Cloning All Groups

> Note: "all-groups" only works on hosted GitLab instances running 13.0.1 or greater

> Note: You must set `--base-url` which is the url to your instance. If your instance requires an insecure connection you can use the `--insecure-gitlab-client` flag

1. Clone **all groups**, **preserving the directory structure** of subgroups

    ```
    ghorg clone all-groups --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX --preserve-dir
    ```

    This would produce a directory structure like

    ```
    /configured/ghorg-dir
    └── your.instance.gitlab
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

    ```
    ghorg clone all-groups --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX
    ```

    This would produce a directory structure like

    ```
    /configured/ghorg-dir
    └── your.instance.gitlab
        ├── project1
        ├── project2
        ├── project3
        └── project4
    ```

#### Cloning Specific Groups

1. Clone **a specific group**, **preserving the directory structure** of subgroups

    ```
    ghorg clone group3 --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX --preserve-dir
    ```

    This would produce a directory structure like

    ```
    /configured/ghorg-dir
    └── group3
        └── subgroup1
            ├── project3
            └── project4
    ```

1. Clone **a specific group**, **WITHOUT preserving the directory structure** of subgroups

    ```
    ghorg clone group3 --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX
    ```

    This would produce a directory structure like

    ```
    /configured/ghorg-dir
    └── group3
        ├── project3
        └── project4
    ```

1. Clone **a specific subgroup**

    ```
    ghorg clone group3/subgroup1 --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX
    ```

    This would produce a directory structure like

    ```
    /configured/ghorg-dir
    └── group3
        └── subgroup1
            ├── project3
            └── project4
    ```

#### Cloning a Specific Users Repos

1. Clone a **user** on a **hosted gitlab** instance using a **token** for auth

    ```
    ghorg clone <gitlab_username> --clone-type=user --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=bGVhdmUgYSBjb21tZW50IG9uIGlzc3VlIDY2
    ```

    This would produce a directory structure like

    ```
    /configured/ghorg-dir
    └── gitlab_username
        ├── project3
        └── project4
    ```

#### Cloning All Users Repos

> Note: "all-users" only works on hosted GitLab instances running 13.0.1 or greater

> Note: You must set `--base-url` which is the url to your instance. If your instance requires an insecure connection you can use the `--insecure-gitlab-client` flag

1. Clone **all users**, **preserving the directory structure** of users

    ```
    ghorg clone all-users --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX --preserve-dir
    ```

    This would produce a directory structure like

    ```
    /configured/ghorg-dir
    └── your.instance.gitlab_users
        ├── user1
        │   └── project1
        ├── user2
        │   └── project2
        └── user3
            ├── project3
            └── project4
    ```
1. Clone **all users**, **WITHOUT preserving the directory structure** of users

    ```
    ghorg clone all-users --base-url=https://<your.instance.gitlab.com> --scm=gitlab --token=XXXXXX
    ```

    This would produce a directory structure like

    ```
    /configured/ghorg-dir
    └── your.instance.gitlab_users
        ├── project1
        ├── project2
        ├── project3
        └── project4
    ```

## Cloud GitLab Orgs

Examples below use the `gitlab-examples` GitLab cloud organization https://gitlab.com/gitlab-examples

1. clone **all groups**, **preserving the directory structure** of subgroups

    ```
    ghorg clone gitlab-examples --scm=gitlab --token=XXXXXX --preserve-dir
    ```

    This would produce a directory structure like

    ```
    /configured/ghorg-dir
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

    ```
    ghorg clone gitlab-examples/wayne-enterprises --scm=gitlab --token=XXXXXX --preserve-dir
    ```

    This would produce a directory structure like

    ```
    /configured/ghorg-dir
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

1. clone only a **subgroup**, **WITHOUT preserving the directory structure** of subgroups

    ```
    ghorg clone gitlab-examples/wayne-enterprises --scm=gitlab --token=XXXXXX
    ```

    This would produce a directory structure like

    ```
    /configured/ghorg-dir
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
