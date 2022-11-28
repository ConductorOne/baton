
![Baton Logo](./docs/images/baton-logo.png)

# Baton: A Toolkit for Auditing Infrastructure Access

The Baton toolkit gives developers the ability to extract, normalize, and interact with workforce identity data such as user accounts, permissions, roles, groups, resources, and more. Through the Baton CLI, developers can audit infrastructure access on-demand, run diffs, and extract access data. This can be used for automating user access reviews, exports into SIEMs, real-time visibility, and many other use cases.

Baton is structured as a toolkit of related command line tools. For each data source there is a "connector", eg, baton-github for interacting with GitHub's API. This tool exports data in a format that the baton tool can understand, transform, and use to perform operations on the application

# What can you do with Baton?

As a generic toolkit for auditing access, Baton can be used for several use cases. Illustratively:
- Find all AWS IAM Users with a specific IAM Role
- Audit Github Repo Admin
- Find users in apps that aren't in your IdP
- Detect differences or changes in permissions in GitHub or AWS
- Export GitHub permissions into a CSV for loading into a user access review
- Discovering all access for an user / account across all SaaS and IaaS systems
- These are just a few of the use cases that Baton can be leveraged for.


# Trying it out: Find all GitHub Repo Admins

Baton can installed via Homebrew (for other operating systems, see [docs/setup-and-install.md](./docs/setup-and-install.md]:

```
brew install conductorone/baton/baton conductorone/baton/baton-github
```

Once installed, you can audit GitHub access with the following:

```
baton-github 
baton -o json | jq ....
```

# What Connectors exist in Baton today?

We're releasing 5 initial connectors with the open source launch of Baton. The ConductorOne team has dozens of more connectors written in our precursor proprietary project from before Baton, and is aggressively porting them to the Baton ecosystem.

Additionally, making a new Connector is really easy -- we wrap up many complexities in the SDK, letting a Connector developer just focus on translating to the Baton data model.

| Connector                | Status     |
|--------------------------|------------|
| [baton-aws](https://github.com/ConductorOne/baton-aws) |   GA   |
| [baton-github](https://github.com/ConductorOne/baton-github) |   GA   |
| [baton-mysql](https://github.com/ConductorOne/baton-github) |   GA   |
| [baton-okta](https://github.com/ConductorOne/baton-okta) |   GA   |
| [baton-postgres](https://github.com/ConductorOne/baton-github) |   GA   |

# Contributing, Support and Issues

We started Baton because we were tired of taking screenshots and manually building spreadsheets.  We welcome contributions, and ideas, no matter how small -- our goal is to make identity and permissions sprawl less painful for everyone.  If you have questions, problems, or ideas: Please open a Github Issue!

See [CONTRIBUTING.md](./CONTRIBUTING.md) for more details.

# `baton` Command Line Usage

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
