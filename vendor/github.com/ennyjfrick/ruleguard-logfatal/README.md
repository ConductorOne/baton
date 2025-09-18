# `ruleguard-logfatal`: A `ruleguard` bundle for better log hygiene

`ruleguard-logfatal` is a [`ruleguard`](https://github.com/quasilyte/go-ruleguard) bundle that enforces the absence of logging at `Fatal` or `Panic` levels in a codebase.

It supports most well-known logging libraries with a `Fatal` and/or `Panic` level, including [`zap`](https://github.com/uber-go/zap), [`logrus`](https://github.com/sirupsen/logrus), [`zerolog`](https://github.com/rs/zerolog), and the standard library's [`log`](https://pkg.go.dev/log) package. If you're not using one of these libraries, it should still work just fine if your library has an  API similar to any of the above; you can look at the custom logger interface in [`_example/examples.go`](./_example/example.go) as an example.

## Motivation

Quite simply: _it is better to propagate errors so they can be handled rather than crash an application_.

Errors should be propagated and handled accordingly, rather than cause an application to crash; using `Fatal` makes it too easy to be haphazard about failure modes. It is _especially_ important to be considerate of error handling in library; no one wants their application to crash because a dependency decided to call `os.Exit(1)`! Logging at `Panic` is slightly better due to `recover`, but suffers from the same problem: if you don't know it exists in your dependency, you won't know to use `recover`.

## Usage

### Installation

Install `ruleguard-logfatal` like you would any other `ruleguard` bundle:
1. If you're not using `golangci-lint`, then [get `ruleguard` first](https://github.com/quasilyte/go-ruleguard?tab=readme-ov-file#quick-start)
2. `go get -u github.com/ennyjfrick/ruleguard-logfatal@latest`
3. Create a `rules.go` with the following content somewhere in your project directory:
```go
//go:build ruleguard
// +build ruleguard

package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"

	logfatalrules "github.com/ennyjfrick/ruleguard-logfatal"
)

func init() {
	dsl.ImportRules("logfatal", logfatalrules.Bundle)
}
```
4. If you're using `ruleguard` as a standalone tool, just point it at your new `rules.go` file:
```shell
$ ruleguard -rules /path/to/rules.go ./...
```
5. Otherwise, if you're using `ruleguard` through `golangci-lint`, add the following to your `.golangci.yml`:
```yaml
linters:
  enable:
    - gocritic
linters-settings:
  gocritic:
    enabled-checks:
      - ruleguard
    settings:
      ruleguard:
        rules: "rules.go"
```

### Customization

As a `ruleguard` bundle, `ruleguard-logfatal` uses the parent `ruleguard` configuration.

#### Disabling or Enabling Rule Groups

`ruleguard-logfatal` has the following two rule groups:
- `noFatal`: checks for logging at the `Fatal` level
- `noPanic`: checks for logging at the `Panic` level

To enable or disable a group, simply pass the prefix you set in the `dsl.ImportRules` call plus the group name to the appropriate `ruleguard` flags. For example, to disable `noPanic`:
```shell
# using ruleguard as a standalone tool
$ ruleguard -rules rules.go -disable logfatal/noPanic . # to disable the check for logging at `Panic`
# alternatively: ruleguard -rules rules.go -enable logfatal/noFatal .
```
```yaml
# using ruleguard through golangci-lint
linters:
  enable:
    - gocritic
linters-settings:
  gocritic:
    enabled-checks:
      - ruleguard
    settings:
      ruleguard:
        rules: "rules.go"
        disable: "logfatal/noPanic"
        # alternatively
        # enable: "logfatal/noError"
```

## Contributing

Feel free to open an [issue](https://github.com/ennyjfrick/ruleguard-logfatal/issues/new) or create a [pull request](https://github.com/ennyjfrick/ruleguard-logfatal/pulls) for any features/bugs/etc. 

If you're creating a pull request, please include test cases when applicable.

## TODO

- [ ] add tests for example