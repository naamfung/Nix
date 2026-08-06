package main

import (
	"os"
	"strings"
	"testing"
)

func TestInstallerCommandLineUsesVisibleUpdateModeAndKeepsDFlagLast(t *testing.T) {
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
}

func TestWindowsUpdateHandoffArgsCarryParentInstallAndRelaunch(t *testing.T) {
	got := windowsUpdateHandoffArgs(
		4242,
		`C:\Users\Jane Doe\AppData\Local\Inx\updates\Inx-windows-amd64-installer.exe`,
		strings.Repeat("a", 64),
		`D:\Tools\Inx App`,
		`D:\Tools\Inx App\inx-desktop.exe`,
		"v1.6.0",
		"2026-07-29T00:00:00Z",
		"transaction-1",
	)
	want := []string{
		"--parent-pid", "4242",
		"--installer", `C:\Users\Jane Doe\AppData\Local\Inx\updates\Inx-windows-amd64-installer.exe`,
		"--installer-sha256", strings.Repeat("a", 64),
		"--to-version", "v1.6.0",
		"--created-at", "2026-07-29T00:00:00Z",
		"--transaction-id", "transaction-1",
		"--install-dir", `D:\Tools\Inx App`,
		"--relaunch", `D:\Tools\Inx App\inx-desktop.exe`,
	}
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("args = %#v, want %#v", got, want)
	}
}

func TestWindowsVersionedUpdateHandoffArgsDoNotRequireLegacyPendingIdentity(t *testing.T) {
	got := windowsVersionedUpdateHandoffArgs(
		4242,
		`C:\Temp\Inx-installer.exe`,
		strings.Repeat("b", 64),
		`D:\Tools\Inx`,
		`D:\Tools\Inx\inx-launcher.exe`,
		"v1.20.0",
	)
	want := []string{
		"--parent-pid", "4242",
		"--installer", `C:\Temp\Inx-installer.exe`,
		"--installer-sha256", strings.Repeat("b", 64),
		"--to-version", "v1.20.0",
		"--install-layout", "versioned-v1",
		"--install-dir", `D:\Tools\Inx`,
		"--relaunch", `D:\Tools\Inx\inx-launcher.exe`,
	}
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("args = %#v, want %#v", got, want)
	}
	for _, legacy := range []string{"--created-at", "--transaction-id"} {
		if strings.Contains(strings.Join(got, " "), legacy) {
			t.Fatalf("versioned handoff must not carry legacy field %s", legacy)
		}
	}
}

func TestWindowsInstallerScriptWaitsBeforeCopyingExecutable(t *testing.T) {
	data, err := os.ReadFile("build/windows/installer/project.nsi")
	if err != nil {
		t.Fatal(err)
	}
	script := string(data)
	for _, want := range []string{
		`!define INX_UPDATE_HELPER "inx-update-helper.exe"`,
		`!define INX_GUARD "inx-guard.exe"`,
		`!define INX_LAUNCHER "inx-launcher.exe"`,
		`!define INX_CLI "inx-cli.exe"`,
		`!define INX_PORTABLE_ENTRY "Inx.exe"`,
		`!define INX_LAYOUT_INSTALLER "inx-layout-installer.exe"`,
		`!define INX_PAYLOAD_MANIFEST "inx-payload.json"`,
		`!define INX_PAYLOAD_SIGNATURE "inx-payload.json.minisig"`,
		"Var InxUpdateMode",
		"Var InxStageMode",
		`${GetOptions} $R0 "/INXUPDATE=" $R1`,
		`${GetOptions} $R0 "/INXSTAGE=" $R2`,
		"Function inx.skipSetupPageForUpdate",
		"Function inx.showUpdateProgress",
		`!define MUI_PAGE_CUSTOMFUNCTION_PRE inx.skipFinishPageForUpdate`,
		"Function inx.skipFinishPageForUpdate",
		`StrCmp $InxUpdateMode "1" 0 inx_show_finish_page`,
		"SetAutoClose true",
		"BringToFront",
		`LangString inxUpdateTitle ${LANG_ENGLISH} "Updating Inx"`,
		`LangString inxUpdateTitle ${LANG_SIMPCHINESE} "正在更新 Inx"`,
		`LangString inxUpdateTitle ${LANG_TRADCHINESE} "正在更新 Inx"`,
		`LangString inxUpdateSubtitle ${LANG_ENGLISH} "Installing the verified update. Inx will restart automatically."`,
		`LangString inxUpdateSubtitle ${LANG_SIMPCHINESE} "正在安装已验证的更新，完成后 Inx 将自动重启。"`,
		`LangString inxUpdateSubtitle ${LANG_TRADCHINESE} "正在安裝已驗證的更新，完成後 Inx 將自動重新啟動。"`,
		"Function inx.waitForExecutableUnlock",
		`FileOpen $1 "$INSTDIR\${PRODUCT_EXECUTABLE}" a`,
		`FileOpen $1 "$INSTDIR\versions\v${INFO_PRODUCTVERSION}\${PRODUCT_EXECUTABLE}" a`,
		`FileOpen $1 "$INSTDIR\${INX_GUARD}" a`,
		`FileOpen $1 "$INSTDIR\${INX_LAUNCHER}" a`,
		`FileOpen $1 "$INSTDIR\${INX_CLI}" a`,
		`FileOpen $1 "$INSTDIR\${INX_PORTABLE_ENTRY}" a`,
		"SetErrorLevel 1618",
		"Call inx.waitForExecutableUnlock",
		`File "/oname=${INX_UPDATE_HELPER}" "${INX_UPDATE_HELPER}"`,
		`File "/oname=${INX_CLI}" "${INX_CLI}"`,
		`File "/oname=${INX_LAYOUT_INSTALLER}" "${INX_GUARD}"`,
		`--activate-staging "$R9" --no-relaunch`,
		`File "/oname=${INX_PAYLOAD_MANIFEST}" "${INX_PAYLOAD_MANIFEST}"`,
		`File "/oname=${INX_PAYLOAD_SIGNATURE}" "${INX_PAYLOAD_SIGNATURE}"`,
		`Delete "$INSTDIR\${INX_UPDATE_HELPER}"`,
		`Delete "$INSTDIR\${INX_CLI}"`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("project.nsi missing %q", want)
		}
	}
	finishPageHook := strings.Index(script, "!define MUI_PAGE_CUSTOMFUNCTION_PRE inx.skipFinishPageForUpdate")
	finishPage := strings.Index(script, "!insertmacro MUI_PAGE_FINISH")
	if finishPageHook < 0 || finishPage < 0 || finishPageHook > finishPage {
		t.Fatalf("update-only finish page hook must be attached to MUI_PAGE_FINISH (hook=%d page=%d)", finishPageHook, finishPage)
	}
	wait := strings.Index(script, "Call inx.waitForExecutableUnlock")
	copyFiles := strings.Index(script, "!insertmacro wails.files")
	if wait < 0 || copyFiles < 0 || wait > copyFiles {
		t.Fatalf("installer must wait for the running exe to unlock before wails.files (wait=%d copy=%d)", wait, copyFiles)
	}
	stageBranch := strings.Index(script, "StrCmp $InxStageMode \"1\" inx_stage_payload")
	if stageBranch < 0 || stageBranch > copyFiles {
		t.Fatalf("staging mode must bypass live executable unlock before payload extraction (branch=%d copy=%d)", stageBranch, copyFiles)
	}
	if !strings.Contains(script, "Goto inx_section_done") {
		t.Fatal("staging mode must skip registry, shortcuts, associations, and uninstaller")
	}
	if strings.Contains(script, `FileOpen $0 "$INSTDIR\current.json" w`) {
		t.Fatal("normal installer must delegate the current.json commit to the atomic Go activator")
	}
	metadataBranch := strings.Index(script, `inx_stage_payload:`)
	metadataFile := strings.Index(script, `File "/oname=${INX_PAYLOAD_MANIFEST}"`)
	normalInstall := strings.Index(script, `inx_normal_install:`)
	if metadataBranch < 0 || metadataFile < 0 || normalInstall < 0 || metadataBranch > metadataFile || metadataFile > normalInstall {
		t.Fatalf("payload manifest must be extracted only in staging mode (branch=%d file=%d)", metadataBranch, metadataFile)
	}
}

func TestDesktopBuildScriptCompilesAndPackagesWindowsUpdateHelper(t *testing.T) {
	data, err := os.ReadFile("../scripts/desktop-build.sh")
	if err != nil {
		t.Fatal(err)
	}
	script := string(data)
	for _, want := range []string{
		`UPDATE_HELPER="inx-update-helper.exe"`,
		`go build -trimpath -o "$windows_resource_tool" ./cmd/windows-resource`,
		`GOOS=windows GOARCH="$arch" go build`,
		`./cmd/update-helper`,
		`build/windows/installer/$UPDATE_HELPER`,
		`stamp_windows_executable "build/windows/installer/$UPDATE_HELPER"`,
		`cp "build/windows/installer/$UPDATE_HELPER" "$payload_dir/$UPDATE_HELPER"`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("desktop-build.sh missing %q", want)
		}
	}

	packageData, err := os.ReadFile("../scripts/package-windows-desktop.sh")
	if err != nil {
		t.Fatal(err)
	}
	packager := string(packageData)
	for _, want := range []string{
		`cp "$PAYLOAD/$UPDATE_HELPER" "$INSTALLER_DIR/$UPDATE_HELPER"`,
		`cp "$PAYLOAD/$UPDATE_HELPER" "$portable_staging/versions/$version_label/$UPDATE_HELPER"`,
		`"$ROOT/scripts/verify-windows-portable.sh" "$portable_staging"`,
	} {
		if !strings.Contains(packager, want) {
			t.Fatalf("package-windows-desktop.sh missing %q", want)
		}
	}
}

func TestWindowsUpdateRequiresObservedHelperHandoff(t *testing.T) {
	data, err := os.ReadFile("updater_windows.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	if !strings.Contains(source, "return startWindowsUpdateHelper(") {
		t.Fatal("Windows update handoff does not require the observed helper path")
	}
	if strings.Contains(source, "return installerCommand(installerPath, installDir).Start()") {
		t.Fatal("Windows update silently falls back to an unobserved installer")
	}
	if !strings.Contains(source, "cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}") {
		t.Fatal("Windows handoff helper should stay hidden while NSIS shows update progress")
	}
	helperData, err := os.ReadFile("cmd/update-helper/main_windows.go")
	if err != nil {
		t.Fatal(err)
	}
	helperSource := string(helperData)
	if strings.Contains(helperSource, "installerCommandLine(installer, installDir), HideWindow: true") {
		t.Fatal("update helper still hides the NSIS progress window")
	}
}
