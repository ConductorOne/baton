# Welcome to Baton contributing guide 

Thank you for investing your time and interest in contributing to our project!

Read our [Code of Conduct](./CODE_OF_CONDUCT.md) to keep our community approachable and respectable.  Contributions are accepted under the [Apache 2.0 License](./LICENSE). 

## New contributor guide

To get an overview of the project, read the [README](./README.md).

### Where are things?

The Baton project lives over multiple git repositories.  To get started:
- [baton](https://github.com/ConductorOne/baton): Baton Command Line tool, which can be used to explore data extracted by a connector.
- [baton-sdk](https://github.com/ConductorOne/baton-sdk): Primary Go SDK library, which contains many core behavoirs, data strcutures, and utilities. 
- `baton-X`, where X is a specific connector.  For example [baton-github](https://github.com/ConductorOne/baton-github) contains the implementation of a GitHub connector.  It leverages the `baton-sdk` repository as a dependency.

## Getting Started

### Issues

#### Create a new issue

If you spot a problem with the Baton, [search if an issue already exists](https://github.com/ConductorOne/baton/issues). If a related issue doesn't exist, you can open a new issue using a relevant [issue form](https://github.com/ConductorOne/baton/issues/new). 

#### Solve an issue

Scan through our [existing issues](https://github.com/ConductorOne/baton/issues) to find one that interests you. As a general rule, we don’t assign issues to anyone. If you find an issue to work on, you are welcome to open a PR with a fix.

### Contribution Flow

This is a rough outline of what a contributor's workflow looks like:

- Fork the repository on GitHub
- Read the [README](./README.md) for build and test instructions
- Play with the project, submit bugs, submit patches!

Thanks for your contributions!

# Reporting Security Issues

The Baton project takes security seriously. If you discover a security issue, please bring it to our attention right away!

Please DO NOT file a public issue or pull request, instead send your report privately to the ConductorOne Security Team, reachable at [security@conductorone.com](mailto:security@conductorone.com).

Security reports are greatly appreciated and we will publicly thank you for them.
