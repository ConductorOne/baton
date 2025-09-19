//go:build ruleguard
// +build ruleguard

package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"

	logfatalrules "github.com/ennyjfrick/ruleguard-logfatal"
)

func init() {
	dsl.ImportRules("logfatal", logfatalrules.Bundle) // checks for uses of log.Fatal or log.Panic
}
