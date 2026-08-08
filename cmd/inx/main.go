// Command inx is a config- and plugin-driven coding agent CLI.
package main

import (
	"os"
	"runtime/debug"

	"inx/internal/cli"
	"inx/internal/config"
	"inx/internal/crashreport"

	// Blank imports wire compile-time built-ins into their registries.
	_ "inx/internal/provider/anthropic"
	_ "inx/internal/provider/openai"
	_ "inx/internal/provider/responses"
	_ "inx/internal/tool/builtin"
)

// Build identity injected via -ldflags (see Makefile). version remains the
// single-line contract for `inx --version`; gitCommit/buildTimeUTC feed
// `inx version --verbose` / `--json` without embedding config paths.
var (
	version      = "dev"
	gitCommit    = ""
	buildTimeUTC = ""
)

// runCLI is the CLI entry; tests may stub it. Production routes through
// RunWithBuildInfo so ldflags metadata is available to version --verbose/--json.
var runCLI = func(args []string, buildVersion string) int {
	return cli.RunWithBuildInfo(args, cli.BuildInfo{
		Version:      buildVersion,
		GitCommit:    gitCommit,
		BuildTimeUTC: buildTimeUTC,
	})
}

func main() {
	os.Exit(runWithCrashCapture(os.Args[1:], version))
}

func runWithCrashCapture(args []string, buildVersion string) (exitCode int) {
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = crashreport.CapturePanic(config.InxHomeDir(), buildVersion, recovered, debug.Stack())
			panic(recovered)
		}
	}()
	return runCLI(args, buildVersion)
}
