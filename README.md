
![Baton Logo](./docs/images/baton-logo.png)

# Baton: A Toolkit for Auditing Infrastructure Access

The Baton toolkit gives developers the ability to extract, normalize, and interact with workforce identity data such as user accounts, permissions, roles, groups, resources, and more, so they can audit infrastructure access, start to automate user access reviews, and enforce the principle of least privilege. 

# Trying it out: Find all GitHub Repo Admins

Baton can installed via Homebrew (for other operating systems, see (docs/setup-and-install.md)[./docs/setup-and-install.md]):
```
brew install conductorone/baton/baton conductorone/baton/baton-github
```

Once installed, you can extract all Repository admins from GitHub with the following:

```
baton-github 
baton -o json | jq ....
```

Baton is structured as a _toolkit_ of related command line tools.  For each data source there is a "connector", eg, `baton-github` for interacting with GitHub's API.  This tool exports data in a format that the `baton` tool can understand, transform, and do other operations on.

# What can you do with Baton?

- Find all AWS IAM Users with a specific IAM Role
- Audit Github Repo Admin
- Find users in apps that aren't in your IdP
- Detect differences or changes in permissions in GitHub or AWS

## What Connectors exist in Baton today?

We're releasing 5 initial connectors with the open source launch of Baton.  The ConductorOne team has dozens of more connectors written in our precursor proprietary project from before Baton, and is aggressively porting them to the Baton ecosystem.

Additionally, making a new Connector is really easy -- we wrap up many complexities in the SDK, letting a Connector developer just focus on translating to the Baton data model.

| Connector                | Status     |
|--------------------------|------------|
| [baton-aws](https://github.com/ConductorOne/baton-aws) |   GA   |
| [baton-github](https://github.com/ConductorOne/baton-github) |   GA   |
| [baton-mysql](https://github.com/ConductorOne/baton-github) |   GA   |
| [baton-okta](https://github.com/ConductorOne/baton-okta) |   GA   |
| [baton-postgres](https://github.com/ConductorOne/baton-github) |   GA   |

# Contributing

We started Baton because we were tired of taking screenshots and manually building spreadsheets.  We welcome contributions, and ideas, no matter how small -- our goal is to make identity and permissions sprawl less painful for everyone.

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
 - data model

# Ecosystem and Repositories 

The Baton project lives inside multiple git repositories.  We have several core repos, which contain the core of Baton, and for each specific Connector to a SaaS or IaaS we have a "connector repo":

- [baton](https://github.com/ConductorOne/baton): Baton Command Line tool, which can be used to explore data extracted by a connector.
- [baton-sdk](https://github.com/ConductorOne/baton-sdk): Primary SDK library, which contains many core behavoirs, data strcutures, and utilities. 

Every individial connector also lives in their own repository (see table above)

## `baton` Command Line Usage

```
baton is a utility for working with the output of a baton-based connector

Usage:
  baton [command]

Available Commands:
  access         List effective access for a user
  completion     Generate the autocompletion script for the specified shell
  diff           Perform a diff between sync runs
  entitlements   List entitlements
  export         Export data from the C1Z for upload
  grants         List grants
  help           Help about any command
  principals     List principals
  resource-types List resource types for the latest (or current) sync
  resources      List resources for the latest sync
  stats          Simple stats about the c1z

Flags:
  -f, --file string            The path to the c1z file to work with. (default "sync.c1z")
  -h, --help                   help for baton
  -o, --output-format string   The format to output results in: (console, json) (default "console")
  -v, --version                version for baton

Use "baton [command] --help" for more information about a command.
```
