
![Baton Logo](./docs/images/baton-logo.png)

# Baton: A Toolkit for Auditing Infrastructure Access

The Baton toolkit gives developers the ability to extract, normalize, and interact with workforce identity data such as user accounts, permissions, roles, groups, resources, and more, so they can audit infrastructure access, start to automate user access reviews, and enforce the principle of least privilege. 

# Trying it out: Find all GitHub Repo Admins

Baton can installed via Homebrew (for other operating systems, see (SETUP_AND_INSTALL.md)[./docs/setup-and-install.md]):
```
brew install conductorone/baton/baton conductorone/baton/baton-github
```

Once installed, you can extract all Repository admins from GitHub with the following:

```
baton-github 
baton -o json | jq ....
```

Baton is structured as a _toolkit_ of several command line tools.  For each source there is a tool, eg, `baton-github` for interacting with GitHub's API.  This tool exports data in a format that the `baton` tool can understand.

# What can you do with Baton?

- Find all AWS IAM Users with a specific IAM Role
- Audit Github Repo Admin
- Find users in apps that aren't in your IdP
- Detect differences or changes in permissions in GitHub or AWS

# Contributing

We started Baton because we were tired of taking screenshots and manually building spreadsheets.   We welcome contributions, and ideas, no matter how small -- our goal is to make identity and permissions sprawl less painful for everyone.

See [CONTRIBUTING.md](./CONTRIBUTING.md) for more details.

# Why Baton?

## The Authorization and Identity challenge

TODO: expand
- The authorization and identity challenge
  - Objectives
    - User & Resource Discovery, Access Provisioning
  - There’s no unified way to manage identity and authz in applications
    - Python Scripting (~/scripts)
    - IdP based Provisioning (SCIM)
    - Webhooks
    - Terraform
- Bespoke authz and identity layer for each app
    OAuth
    API Keys
    Impersonation
    Rate Limits

    SCIM
    APIs

## What does Baton provide?

- Unified Interface for users, resources, and grants
- Fully Hostable
- Natively handles rate limiting, etc.

### Open Source Connectors built on an SDK Interface  

Baton, an open source project to extract, normalize, and interact with identity data such as user accounts, permissions, roles, groups, resources, and more for any security or governance initiative.

### Shared Tooling and Data Models
 - Diff'
 - `baton` CLI
 - `jq|` pipelines

# Ecosystem and Repositories 

The Baton project lives over multiple git repositories.  We have several core repos, which contain the core of Baton, and for each specific Connector to a SaaS or IaaS we have a "connector repo":
- [baton](https://github.com/ConductorOne/baton): Baton Command Line tool, which can be used to explore data extracted by a connector.
- [baton-sdk](https://github.com/ConductorOne/baton-sdk): Primary SDK library, which contains many core behavoirs, data strcutures, and utilities. 

Every individial connector also lives in their own repository:
- [baton-github](https://github.com/ConductorOne/baton-github) contains the implementation of a GitHub connector.


## Command Line Usage
```
baton is a utility for working with the output of a baton-based connector

Usage:
  baton [command]

Available Commands:
  completion     Generate the autocompletion script for the specified shell
  diff           Perform a diff between sync runs
  entitlements   List entitlements
  export         Export data from the C1Z for upload
  grants         List grants
  help           Help about any command
  resource-types List resource types for the latest (or current) sync
  resources      List resources for the latest (or current) sync
  stats          Simple stats about the c1z
  users          List user resources with more detail

Flags:
  -f, --file string   The path to the c1z file to work with. (default "sync.c1z")
  -h, --help          help for baton
  -v, --version       version for baton

Use "baton [command] --help" for more information about a command.
```
