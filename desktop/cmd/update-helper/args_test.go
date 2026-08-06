package main

import (
	"strings"
	"testing"
)

func TestInstallerCommandLineUsesVisibleUpdateModeAndLeavesDFlagLast(t *testing.T) {
	got := installerCommandLine(`C:\Temp\Inx Installer.exe`, `D:\Tools\Inx App`)
	want := `"C:\Temp\Inx Installer.exe" /INXUPDATE=1 /INXSTAGE=1 /D=D:\Tools\Inx App`
	if got != want {
		t.Fatalf("installerCommandLine = %q, want %q", got, want)
	}
	if strings.Contains(got, " /S") {
		t.Fatalf("auto-update must expose progress instead of using silent mode, got %q", got)
	}
	if !strings.HasSuffix(got, `/D=D:\Tools\Inx App`) {
		t.Fatalf("/D= must be the final unquoted NSIS token, got %q", got)
	}
	if !strings.Contains(got, " /INXSTAGE=1") {
		t.Fatalf("auto-update must extract away from the live install, got %q", got)
	}
}
