package cmd

import (
	"runtime/debug"
	"strings"
)

// These are set at build time via -ldflags, e.g.
//
//	go build -ldflags "-X github.com/chrishrb/go-grip/cmd.version=v1.2.3"
//
// When they are left empty the values are derived from the embedded build
// info, which covers `go install` and builds from a checkout.
var (
	version = ""
	commit  = ""
	date    = ""
)

// buildVersion assembles the string reported by `go-grip --version`.
func buildVersion() string {
	v, c, d := version, commit, date
	modified := false

	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				if c == "" {
					c = s.Value
				}
			case "vcs.time":
				if d == "" {
					d = s.Value
				}
			case "vcs.modified":
				modified = s.Value == "true"
			}
		}
	}

	// A version stamped by the go tool is either a release tag or a
	// pseudo-version that already contains the commit and date, so it is
	// reported as-is.
	if v == "" && ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	if v == "" {
		v = "dev"
	}

	var details []string
	if c != "" {
		if len(c) > 12 {
			c = c[:12]
		}
		if modified {
			c += "-dirty"
		}
		details = append(details, "commit "+c)
	}
	if d != "" {
		details = append(details, "built "+d)
	}

	if len(details) == 0 {
		return v
	}
	return v + " (" + strings.Join(details, ", ") + ")"
}
