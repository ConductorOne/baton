package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

var Bundle = dsl.Bundle{}

func noFatal(m dsl.Matcher) { //nolint:unused // magically exported
	m.Match(`$caller.Fatal($*_)`, `$caller.fatal($*_)`).Where(!m["caller"].Type.Is(`*testing.T`) && !m["caller"].Type.Is(`*testing.B`) && !m["caller"].Type.Is(`*testing.F`)).Report(`propagate errors rather than calling $$`) // ignore *testing.T.Fatal and equiv, but match log.Fatal and equiv

	m.Match(`$caller.Fatalf($*_)`, `$caller.fatalf($*_)`).Where(!m["caller"].Type.Is(`*testing.T`) && !m["caller"].Type.Is(`*testing.B`) && !m["caller"].Type.Is(`*testing.F`)).Report(`propagate errors rather than calling $$`) // ignore *testing.T.Fatalf and equiv, but match log.Fatalf and equiv

	m.Match(`$_.Fatalln($*_)`, `$_.fatalln($*_)`, `$_.Fatalw($*_)`, `$_.fatalw($*_)`).Report(`propagate errors rather than calling $$`) // catches most, if not all, x.[fF]atal(ln|w)? patterns

	m.Match(`$_.FatalFn($*_)`).Report(`propagate errors rather than calling $$`) // logrus-specific pattern

	m.Match(`$_.Log($level, $*_)`, `$_.Logf($level, $*_)`, `$_.Logln($level, $*_)`, `$_.Logw($level, $*_)`, `$_.log($level, $*_)`, `$_.logf($level, $*_)`, `$_.logln($level, $*_)`, `$_.logw($level, $*_)`).Where(m["level"].Text.Matches(`.*([Ll]evel)?[fF]atal([Ll]evel)?.*`)).Report(`propagate errors rather than logging at $level`) // general "log at fatal"

	m.Match(`$_.LogFn($level, $*_)`).Where(m["level"].Text.Matches(`.*([Ll]evel)?[fF]atal([Ll]evel)?.*`)).Report(`propagate errors rather than logging at $level`) // logrus-specific pattern

	m.Match(`$_.WithLevel($level, $*_)`).Where(m["level"].Text.Matches(`.*([Ll]evel)?[fF]atal([Ll]evel)?.*`)).Report(`propagate errors rather than logging at $level`) // zerolog-style level setting
}

func noPanic(m dsl.Matcher) { //nolint:unused // magically exported
	m.Match(`$_.Panic($*_)`, `$_.Panicf($*_)`, `$_.Panicln($*_)`, `$_.Panicw($*_)`, `$_.panic($*_)`, `$_.panicf($*_)`, `$_.panicln($*_), $_.paniclw($*_)`).Report(`propagate errors rather than calling $$`) // catches most, if not all, x.[pP]anic(f|ln|w)? patterns

	m.Match(`$_.PanicFn($*_)`).Report(`propagate errors rather than calling $$`) // logrus-specific pattern

	m.Match(`$_.Log($level, $*_)`, `$_.Logf($level, $*_)`, `$_.Logln($level, $*_)`, `$_.Logw($level, $*_)`, `$_.log($level, $*_)`, `$_.logf($level, $*_)`, `$_.logln($level, $*_)`, `$_.logw($level, $*_)`).Where(m["level"].Text.Matches(`.*([Ll]evel)?[Pp]anic([Ll]evel)?.*`)).Report(`propagate errors rather than logging at $level`) // general "log at fatal"

	m.Match(`$_.LogFn($level, $*_)`).Where(m["level"].Text.Matches(`.*([Ll]evel)?[Pp]anic([Ll]evel)?.*`)).Report(`propagate errors rather than logging at $level`) // logrus-specific pattern

	m.Match(`$_.WithLevel($level, $*_)`).Where(m["level"].Text.Matches(`.*([Ll]evel)?[Pp]anic([Ll]evel)?.*`)).Report(`propagate errors rather than logging at $level`) // zerolog-style level setting
}
