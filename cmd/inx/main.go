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

// version is injected at build time via -ldflags "-X main.version=...".
var version = "dev"

var runCLI = cli.Run

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
