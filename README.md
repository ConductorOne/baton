
![Baton Logo](./docs/images/baton-logo.png)

# Baton: A toolkit for adding identity governance to any application

The Baton toolkit gives developers the ability to extract, normalize, and interact with workforce identity data such as user accounts, permissions, roles, groups, resources, and more. Through the Baton CLI, developers can audit infrastructure access on-demand, run diffs, and extract access data. This can be used for automating user access reviews, exports into SIEMs, real-time visibility, and many other use cases.

Baton is structured as a toolkit of related command line tools. For each data source there is a "connector", such as `baton-github` for interacting with GitHub's API. This tool exports data in a format that the `baton` tool can understand, transform, and use to perform operations on the application

## :tada: :tada: Launching Baton as an Open Source Project!
- [Announcing Baton, an Open Source Toolkit for Auditing Infrastructure User Access](https://www.conductorone.com/blog/announcing-baton-open-source-for-auditing-infrastructure-access/)
- [Technical Deep Dive: Using Baton to Audit Infrastructure Access](https://www.conductorone.com/blog/technical-deep-dive-using-baton-to-audit-infrastructure-access/)
- [Baton and the Journey to Identity Security and Unified Access Control](https://www.conductorone.com/blog/baton-journey-to-identity-security-and-unified-access/)


# What can you do with Baton?

As a generic toolkit for auditing access, Baton can be used for many use cases, such as:

 - [Export GitHub access updates to a CSV file using Baton](https://www.conductorone.com/docs/baton/github_integration/)
 - [Use Baton to get Splunk alerts when a new Github admin is added](https://www.conductorone.com/docs/baton/siem_integration/)
 - [Set up a daily check for GitHub user rights updates using Baton](https://www.conductorone.com/docs/baton/github_action_schedule/)
 - [Diff access rights from two SaaS systems with Baton](https://www.conductorone.com/docs/baton/saas_integration/)
- Finding all AWS IAM Users with a specific IAM Role
- Auditing Github Repo Admins
- Finding users in apps that aren't in your IdP
- Detecting differences or changes in permissions in GitHub or AWS
- Discovering all access for an user or account across all SaaS and IaaS systems
- Calculating the effective access of a user based on group membership

These are just a few of the use cases that Baton can be leveraged for.

# Trying it out: Find all GitHub repo admins

Baton can installed via Homebrew:

```
brew install conductorone/baton/baton conductorone/baton/baton-github
```

Once installed, you can audit GitHub access with the following:

```
# Run the baton github connector
baton-github 
# Output the resources discovered
baton resources
# Output the same data to JSON and parse it with jq
baton resources -o json | jq '.resources[].resource.displayName'
```

We have also recorded a short video exploring some of the data Baton can extract from Github:
[![Alt Video demo of using Baton with Github](./docs/images/baton-github-video.jpg)](http://www.youtube.com/watch?v=mgoPNvIc1U8 "VIDEO: Using Baton - How to export account and access data from GitHub")

# What connectors exist in Baton today?

We released five initial connectors with the open source launch of Baton. The ConductorOne team has dozens more connectors written in our precursor proprietary project from before Baton, and is aggressively porting them to the Baton ecosystem.

Additionally, making a new connector is really easy -- we wrap up many complexities in the SDK, letting a connector developer focus on translating to the Baton data model.

 - [1Password](https://github.com/ConductorOne/baton-1password)
 - [Amazon Web Services](https://github.com/ConductorOne/baton-aws)
 - [Asana](https://github.com/ConductorOne/baton-asana)
 - [BambooHR](https://github.com/ConductorOne/baton-bamboohr)
 - [Bitbucket](https://github.com/ConductorOne/baton-bitbucket)
 - [Bitbucket Data Center](https://github.com/ConductorOne/baton-bitbucket-datacenter)
 - [Box](https://github.com/ConductorOne/baton-box)
 - [Broadcom SAC](https://github.com/ConductorOne/baton-broadcom-sac)
 - [Calendly](https://github.com/ConductorOne/baton-calendly)
 - [Celigo](https://github.com/ConductorOne/baton-celigo)
 - [CloudAMQP](https://github.com/ConductorOne/baton-cloudamqp)
 - [Cloudflare](https://github.com/ConductorOne/baton-cloudflare)
 - [Cloudflare Zero Trust](https://github.com/ConductorOne/baton-cloudflare-zero-trust)
 - [Confluence](https://github.com/ConductorOne/baton-confluence)
 - [Confluence Data Center](https://github.com/ConductorOne/baton-confluence-datacenter)
 - [CrowdStrike](https://github.com/ConductorOne/baton-crowdstrike)
 - [Databricks](https://github.com/ConductorOne/baton-databricks)
 - [Datadog](https://github.com/ConductorOne/baton-datadog)
 - [Discord](https://github.com/ConductorOne/baton-discord)
 - [Dockerhub](https://github.com/ConductorOne/baton-dockerhub)
 - [Duo](https://github.com/ConductorOne/baton-duo)
 - [Elastic](https://github.com/ConductorOne/baton-elastic)
 - [Expensify](https://github.com/ConductorOne/baton-expensify)
 - [Fastly](https://github.com/ConductorOne/baton-fastly)
 - [Formal](https://github.com/ConductorOne/baton-formal)
 - [Fullstory](https://github.com/ConductorOne/baton-fullstory)
 - [Galileo Financial Technologies](https://github.com/ConductorOne/baton-galileo-ft)
 - [GitHub](https://github.com/ConductorOne/baton-github)
 - [Google BigQuery](https://github.com/ConductorOne/baton-google-bigquery)
 - [Google Identity Platform](https://github.com/ConductorOne/baton-google-identity-platform)
 - [Google Workspace](https://github.com/ConductorOne/baton-google-workspace)
 - [Hubspot](https://github.com/ConductorOne/baton-hubspot)
 - [Jamf](https://github.com/ConductorOne/baton-jamf)
 - [JD Edwards](https://github.com/ConductorOne/baton-jd-edwards)
 - [Jenkins](https://github.com/ConductorOne/baton-jenkins)
 - [Jira](https://github.com/ConductorOne/baton-jira)
 - [Jira Data Center](https://github.com/ConductorOne/baton-jira-datacenter)
 - [JumpCloud](https://github.com/ConductorOne/baton-jumpcloud)
 - [LDAP](https://github.com/ConductorOne/baton-ldap)
 - [Linear](https://github.com/ConductorOne/baton-linear)
 - [Litmos](https://github.com/ConductorOne/baton-litmos)
 - [Microsoft 365](https://github.com/ConductorOne/baton-ms365)
 - [Microsoft SQL Server](https://github.com/ConductorOne/baton-sql-server)
 - [Miro](https://github.com/ConductorOne/baton-miro)
 - [MongoDB](https://github.com/ConductorOne/baton-mongodb)
 - [MySQL](https://github.com/ConductorOne/baton-mysql)
 - [New Relic](https://github.com/ConductorOne/baton-newrelic)
 - [Notion](https://github.com/ConductorOne/baton-notion)
 - [Okta](https://github.com/ConductorOne/baton-okta)
 - [OneLogin](https://github.com/ConductorOne/baton-onelogin)
 - [Opsgenie](https://github.com/ConductorOne/baton-opsgenie)
 - [PagerDuty](https://github.com/ConductorOne/baton-pagerduty)
 - [Panorama](https://github.com/ConductorOne/baton-panorama)
 - [Panther](https://github.com/ConductorOne/baton-panther)
 - [PostgreSQL](https://github.com/ConductorOne/baton-postgresql)
 - [Privx](https://github.com/ConductorOne/baton-privx)
 - [Retool](https://github.com/ConductorOne/baton-retool)
 - [Segment](https://github.com/ConductorOne/baton-segment)
 - [SentinelOne](https://github.com/ConductorOne/baton-sentinel-one)
 - [ServiceNow](https://github.com/ConductorOne/baton-servicenow)
 - [Slack](https://github.com/ConductorOne/baton-slack)
 - [Snipe-IT](https://github.com/ConductorOne/baton-snipe-it)
 - [Snowflake](https://github.com/ConductorOne/baton-snowflake)
 - [Snyk](https://github.com/ConductorOne/baton-snyk)
 - [Splunk](https://github.com/ConductorOne/baton-splunk)
 - [Tableau](https://github.com/ConductorOne/baton-tableau)
 - [Temporal Cloud](https://github.com/ConductorOne/baton-temporalcloud)
 - [Torq](https://github.com/ConductorOne/baton-torq)
 - [Verkada](https://github.com/ConductorOne/baton-verkada)
 - [Very Good Security](https://github.com/ConductorOne/baton-vgs)
 - [XSOAR](https://github.com/ConductorOne/baton-xsoar)
 - [Zendesk](https://github.com/ConductorOne/baton-zendesk)
 - [Zoom](https://github.com/ConductorOne/baton-zoom)
 <!--[Xero](https://github.com/ConductorOne/baton-xero)-->

# Learn more about Baton

The [Baton documentation site contains more documentation and example use cases](https://www.conductorone.com/docs/baton/intro/).

# Contributing, support and issues

We started Baton because we were tired of taking screenshots and manually building spreadsheets.  We welcome contributions, and ideas, no matter how small -- our goal is to make identity and permissions sprawl less painful for everyone.  If you have questions, problems, or ideas: Please open a Github Issue!

See [CONTRIBUTING.md](./CONTRIBUTING.md) for more details.

# `baton` command line usage

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
