Unicode true

####
## Inx per-user NSIS installer.
##
## This file is COMMITTED and customized (Wails leaves an existing project.nsi
## untouched and only regenerates wails_tools.nsh). The customizations vs.
## Wails' default template:
##
##   1. REQUEST_EXECUTION_LEVEL "user" + InstallDir under $LOCALAPPDATA - install
##      without administrator rights. This lets the auto-updater re-run a freshly
##      downloaded installer in a visible progress-only mode with no UAC prompt.
##   2. Uninstall registry under HKCU (not HKLM). Wails' wails.writeUninstaller /
##      wails.deleteUninstaller macros hard-code HKLM, which a non-admin install
##      cannot write - so we inline HKCU versions below instead.
##   3. InstallDir is remembered across updates via InstallDirRegKey +
##      InstallLocation (HKCU\...\Uninstall\InstallLocation). When upgrading from
##      a build that did not write InstallLocation yet, .onInit falls back to the
##      old DisplayIcon path before using the default. Without this, every release
##      forces the user back to %LOCALAPPDATA%\Programs\Inx even if they had
##      moved the install to a different drive (e.g. D:\Tools\Inx); the
##      auto-updater would overwrite the wrong dir, leaving the old install
##      orphaned.
##
## Everything else mirrors Wails' generated default. Defines below override the
## ProjectInfo values that wails_tools.nsh would otherwise populate.
####

## Install per-user (no admin). Must be defined BEFORE including wails_tools.nsh,
## which only sets the "admin" default when REQUEST_EXECUTION_LEVEL is undefined.
!define REQUEST_EXECUTION_LEVEL "user"

####
## Include the wails tools (auto-generated; provides INFO_* defines and the
## wails.* macros used below).
####
!include "wails_tools.nsh"
!include "FileFunc.nsh"
!include "LogicLib.nsh"

# The version information for this two must consist of 4 parts
VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

# Enable HiDPI support. https://nsis.sourceforge.io/Reference/ManifestDPIAware
ManifestDPIAware true

!include "MUI.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
# !define MUI_WELCOMEFINISHPAGE_BITMAP "resources\leftimage.bmp" #Include this to add a bitmap on the left side of the Welcome Page. Must be a size of 164x314
!define MUI_FINISHPAGE_NOAUTOCLOSE # Wait on the INSTFILES page so the user can take a look into the details of the installation steps
!define MUI_ABORTWARNING # This will warn the user if they exit from the installer.

!define MUI_PAGE_CUSTOMFUNCTION_PRE inx.skipSetupPageForUpdate
!insertmacro MUI_PAGE_WELCOME # Welcome to the installer page.
# !insertmacro MUI_PAGE_LICENSE "resources\eula.txt" # Adds a EULA page to the installer
!define MUI_PAGE_CUSTOMFUNCTION_PRE inx.skipSetupPageForUpdate
!insertmacro MUI_PAGE_DIRECTORY # In which folder install page.
!define MUI_PAGE_CUSTOMFUNCTION_SHOW inx.showUpdateProgress
!insertmacro MUI_PAGE_INSTFILES # Installing page.
!define MUI_PAGE_CUSTOMFUNCTION_PRE inx.skipFinishPageForUpdate
!insertmacro MUI_PAGE_FINISH # Finished installation page.

!insertmacro MUI_UNPAGE_INSTFILES # Uinstalling page

!insertmacro MUI_LANGUAGE "English"
!insertmacro MUI_LANGUAGE "SimpChinese"
!insertmacro MUI_LANGUAGE "TradChinese"

LangString inxUpdateTitle ${LANG_ENGLISH} "Updating Inx"
LangString inxUpdateTitle ${LANG_SIMPCHINESE} "正在更新 Inx"
LangString inxUpdateTitle ${LANG_TRADCHINESE} "正在更新 Inx"
LangString inxUpdateSubtitle ${LANG_ENGLISH} "Installing the verified update. Inx will restart automatically."
LangString inxUpdateSubtitle ${LANG_SIMPCHINESE} "正在安装已验证的更新，完成后 Inx 将自动重启。"
LangString inxUpdateSubtitle ${LANG_TRADCHINESE} "正在安裝已驗證的更新，完成後 Inx 將自動重新啟動。"

## Preserve the first-pass generated uninstaller so the release workflow can
## Authenticode-sign it together with the other installed payload files.
## The second pass provides ARG_INX_SIGNED_UNINSTALLER and embeds that
## signed binary instead of generating another unsigned uninstaller.
!ifndef ARG_INX_SIGNED_UNINSTALLER
!uninstfinalize 'cmd.exe /C copy /Y "%1" "inx-uninstall.exe" >NUL'
!endif
#!finalize 'signtool --file "%1"'

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe" # Name of the installer's file.
!define INX_DEFAULT_INSTALLDIR "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
!define INX_UPDATE_HELPER "inx-update-helper.exe"
!define INX_GUARD "inx-guard.exe"
!define INX_LAUNCHER "inx-launcher.exe"
!define INX_CLI "inx-cli.exe"
!define INX_PORTABLE_ENTRY "Inx.exe"
!define INX_LAYOUT_INSTALLER "inx-layout-installer.exe"
!define INX_PAYLOAD_MANIFEST "inx-payload.json"
!define INX_PAYLOAD_SIGNATURE "inx-payload.json.minisig"
!define INX_UNLOCK_RETRIES 60
Var InxUpdateMode
Var InxStageMode
InstallDirRegKey HKCU "${UNINST_KEY}" "InstallLocation" # Reuse the previous install path on update; .onInit falls back to the default on first install.
InstallDir "${INX_DEFAULT_INSTALLDIR}" # Per-user install location (no admin rights required).
ShowInstDetails show # This will always show the installation details.

####
## Per-user uninstaller registry (HKCU). Replaces wails.writeUninstaller /
## wails.deleteUninstaller, which write HKLM and would fail without admin rights.
####
!macro inx.writeUninstaller
    !ifdef ARG_INX_SIGNED_UNINSTALLER
    File "/oname=uninstall.exe" "${ARG_INX_SIGNED_UNINSTALLER}"
    !else
    WriteUninstaller "$INSTDIR\uninstall.exe"
    !endif

    WriteRegStr HKCU "${UNINST_KEY}" "Publisher" "${INFO_COMPANYNAME}"
    WriteRegStr HKCU "${UNINST_KEY}" "DisplayName" "${INFO_PRODUCTNAME}"
    WriteRegStr HKCU "${UNINST_KEY}" "DisplayVersion" "${INFO_PRODUCTVERSION}"
    WriteRegStr HKCU "${UNINST_KEY}" "DisplayIcon" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    WriteRegStr HKCU "${UNINST_KEY}" "UninstallString" "$\"$INSTDIR\uninstall.exe$\""
    WriteRegStr HKCU "${UNINST_KEY}" "QuietUninstallString" "$\"$INSTDIR\uninstall.exe$\" /S"
    # Persist the resolved install path so a subsequent update picks it up
    # via InstallDirRegKey above. Without this, every release would force the
    # user back to %LOCALAPPDATA%\Programs\Inx even if they had moved
    # the install to a different drive (e.g. D:\Tools\Inx). The auto-
    # updater trusts this persisted path, so it has to be present before the
    # visible progress-only re-install.
    WriteRegStr HKCU "${UNINST_KEY}" "InstallLocation" "$INSTDIR"

    ${GetSize} "$INSTDIR" "/S=0K" $0 $1 $2
    IntFmt $0 "0x%08X" $0
    WriteRegDWORD HKCU "${UNINST_KEY}" "EstimatedSize" "$0"
!macroend

!macro inx.deleteUninstaller
    Delete "$INSTDIR\uninstall.exe"
    DeleteRegKey HKCU "${UNINST_KEY}"
!macroend

Function .onInit
   !insertmacro wails.checkArchitecture

   ; The helper passes /INXUPDATE=1 and a final /D=<current directory>.
   ; This mode remains visible but skips every page that could change the
   ; destination, then closes automatically after the file copy so the helper
   ; can relaunch Inx. A normal manual installer keeps the full wizard.
   StrCpy $InxUpdateMode "0"
   StrCpy $InxStageMode "0"
   ${GetParameters} $R0
   ClearErrors
   ${GetOptions} $R0 "/INXUPDATE=" $R1
   IfErrors inx_update_mode_done
   StrCmp $R1 "1" 0 inx_update_mode_done
   StrCpy $InxUpdateMode "1"

inx_update_mode_done:
   ClearErrors
   ${GetOptions} $R0 "/INXSTAGE=" $R2
   IfErrors inx_stage_mode_done
   StrCmp $R2 "1" 0 inx_stage_mode_done
   StrCpy $InxStageMode "1"

inx_stage_mode_done:

   ; InstallDirRegKey leaves $INSTDIR empty when the InstallLocation value is
   ; missing. Older installers still wrote DisplayIcon, so use its parent folder
   ; as a compatibility bridge before falling back to the per-user default.
   StrCmp $INSTDIR "" 0 done
   ClearErrors
   ReadRegStr $0 HKCU "${UNINST_KEY}" "DisplayIcon"
   IfErrors fallback
   StrCmp $0 "" fallback
   ${GetParent} "$0" $INSTDIR
   StrCmp $INSTDIR "" fallback done

fallback:
   StrCpy $INSTDIR "${INX_DEFAULT_INSTALLDIR}"
done:
FunctionEnd

Function inx.skipSetupPageForUpdate
   StrCmp $InxUpdateMode "1" 0 inx_show_setup_page
   Abort

inx_show_setup_page:
FunctionEnd

Function inx.showUpdateProgress
   StrCmp $InxUpdateMode "1" 0 inx_update_progress_done
   !insertmacro MUI_HEADER_TEXT "$(inxUpdateTitle)" "$(inxUpdateSubtitle)"
   SetDetailsView hide
   SetAutoClose true
   BringToFront

inx_update_progress_done:
FunctionEnd

Function inx.skipFinishPageForUpdate
   StrCmp $InxUpdateMode "1" 0 inx_show_finish_page
   Abort

inx_show_finish_page:
FunctionEnd

Function inx.waitForExecutableUnlock
   StrCpy $0 0

retry:
   IfFileExists "$INSTDIR\${PRODUCT_EXECUTABLE}" 0 check_versioned_target
   ClearErrors
   FileOpen $1 "$INSTDIR\${PRODUCT_EXECUTABLE}" a
   IfErrors locked
   FileClose $1

check_versioned_target:
   ; A same-version recovery install replaces this directory transactionally.
   ; Detect the running active binary before asking the Go activator to rename it.
   IfFileExists "$INSTDIR\versions\v${INFO_PRODUCTVERSION}\${PRODUCT_EXECUTABLE}" 0 check_guard
   ClearErrors
   FileOpen $1 "$INSTDIR\versions\v${INFO_PRODUCTVERSION}\${PRODUCT_EXECUTABLE}" a
   IfErrors locked
   FileClose $1

check_guard:
   IfFileExists "$INSTDIR\${INX_GUARD}" 0 check_launcher
   ClearErrors
   FileOpen $1 "$INSTDIR\${INX_GUARD}" a
   IfErrors locked
   FileClose $1

check_launcher:
	IfFileExists "$INSTDIR\${INX_LAUNCHER}" 0 check_cli
	ClearErrors
	FileOpen $1 "$INSTDIR\${INX_LAUNCHER}" a
	IfErrors locked
	FileClose $1

check_cli:
	IfFileExists "$INSTDIR\${INX_CLI}" 0 check_portable_entry
	ClearErrors
	FileOpen $1 "$INSTDIR\${INX_CLI}" a
	IfErrors locked
	FileClose $1

check_portable_entry:
   IfFileExists "$INSTDIR\${INX_PORTABLE_ENTRY}" 0 done
   ClearErrors
   FileOpen $1 "$INSTDIR\${INX_PORTABLE_ENTRY}" a
   IfErrors locked
   FileClose $1
   Goto done

locked:
   IntOp $0 $0 + 1
   IntCmp $0 ${INX_UNLOCK_RETRIES} failed 0 0
   Sleep 1000
   Goto retry

failed:
   IfSilent silent interactive

interactive:
   MessageBox MB_RETRYCANCEL|MB_ICONEXCLAMATION "Inx is still running. Close Inx, then click Retry to continue the installation." IDRETRY retry IDCANCEL abort
   Goto retry

silent:
   SetErrorLevel 1618

abort:
   Abort "Inx is still running. Close Inx and run the installer again."

done:
FunctionEnd

Section
    !insertmacro wails.setShellContext

    ; /INXSTAGE=1: flat six-member payload for 1.18–1.19.1 helpers (and
    ; the new helper's staging extract). Do not write shortcuts/uninstaller.
    ; Normal install: versioned-v1 layout under versions/v${INFO_PRODUCTVERSION}/
    ; with a permanent thin launcher at InstallRoot. Guard is only present in
    ; STAGE payloads (as the one-shot legacy migrator) and is not persisted on
    ; a normal install.
    StrCmp $InxStageMode "1" inx_stage_payload
    !insertmacro wails.webview2runtime
    Call inx.waitForExecutableUnlock
    Goto inx_normal_install

inx_stage_payload:
    SetOutPath $INSTDIR
    !if /FileExists "${INX_PAYLOAD_MANIFEST}"
    File "/oname=${INX_PAYLOAD_MANIFEST}" "${INX_PAYLOAD_MANIFEST}"
    !endif
    !if /FileExists "${INX_PAYLOAD_SIGNATURE}"
    File "/oname=${INX_PAYLOAD_SIGNATURE}" "${INX_PAYLOAD_SIGNATURE}"
    !endif
    !insertmacro wails.files
    !if /FileExists "${INX_UPDATE_HELPER}"
    File "/oname=${INX_UPDATE_HELPER}" "${INX_UPDATE_HELPER}"
    !endif
    !if /FileExists "${INX_GUARD}"
    File "/oname=${INX_GUARD}" "${INX_GUARD}"
    !endif
    !if /FileExists "${INX_LAUNCHER}"
    File "/oname=${INX_LAUNCHER}" "${INX_LAUNCHER}"
    !endif
    !if /FileExists "${INX_CLI}"
    File "/oname=${INX_CLI}" "${INX_CLI}"
    !endif
    Goto inx_section_done

inx_normal_install:
    ; Extract into an install-local temporary directory, then let the signed Go
    ; activator validate the complete release unit, transactionally publish the
    ; version/root entries, and strictly atomically replace current.json last.
    ; The normal/recovery installer therefore shares the same commit protocol as
    ; automatic updates instead of writing live files or current.json in place.
    System::Call 'kernel32::GetCurrentProcessId() i .R8'
    CreateDirectory "$INSTDIR\versions"
    StrCpy $R9 "$INSTDIR\versions\.installer-v${INFO_PRODUCTVERSION}-$R8"
    RMDir /r "$R9"
    CreateDirectory "$R9"
    SetOutPath "$R9"
    !insertmacro wails.files
    !if /FileExists "${INX_UPDATE_HELPER}"
    File "/oname=${INX_UPDATE_HELPER}" "${INX_UPDATE_HELPER}"
    !else
    !warning "${INX_UPDATE_HELPER} was not found; Windows auto-update will fail safely until the helper is installed."
    !endif
    !if /FileExists "${INX_CLI}"
    File "/oname=${INX_CLI}" "${INX_CLI}"
    !else
    !warning "${INX_CLI} was not found; remote upload installation will be unavailable."
    !endif
    !if /FileExists "${INX_LAUNCHER}"
    File "/oname=${INX_LAUNCHER}" "${INX_LAUNCHER}"
    !endif

    SetOutPath "$PLUGINSDIR"
    !if /FileExists "${INX_GUARD}"
    File "/oname=${INX_LAYOUT_INSTALLER}" "${INX_GUARD}"
    !else
    !error "${INX_GUARD} was not found; normal installs require the signed layout activator."
    !endif
    ExecWait '"$PLUGINSDIR\${INX_LAYOUT_INSTALLER}" --install-root "$INSTDIR" --version "v${INFO_PRODUCTVERSION}" --activate-staging "$R9" --no-relaunch' $0
    StrCmp $0 "0" inx_layout_activated
    DetailPrint "Inx layout activation failed with exit code $0; the previous version remains active."
    RMDir /r "$R9"
    SetErrorLevel 1
    Abort "Inx could not activate the verified release. The previous version was left unchanged."

inx_layout_activated:
    RMDir /r "$R9"
    SetOutPath "$INSTDIR"

    ; Remove flat leftovers from prior 1.18–1.19 installs when overwriting.
    Delete "$INSTDIR\${PRODUCT_EXECUTABLE}"
    Delete "$INSTDIR\${INX_GUARD}"
    Delete "$INSTDIR\${INX_UPDATE_HELPER}"

    !if /FileExists "${INX_LAUNCHER}"
    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${INX_LAUNCHER}" "" "$INSTDIR\versions\v${INFO_PRODUCTVERSION}\${PRODUCT_EXECUTABLE}" 0
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${INX_LAUNCHER}" "" "$INSTDIR\versions\v${INFO_PRODUCTVERSION}\${PRODUCT_EXECUTABLE}" 0
    !else
    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\versions\v${INFO_PRODUCTVERSION}\${PRODUCT_EXECUTABLE}"
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\versions\v${INFO_PRODUCTVERSION}\${PRODUCT_EXECUTABLE}"
    !endif

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols
    !insertmacro inx.writeUninstaller

inx_section_done:
SectionEnd

Section "uninstall"
    !insertmacro wails.setShellContext

    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}" # Remove the WebView2 DataPath

    ; Precision uninstall: flat leftovers, thin entry points, and version trees.
    Delete "$INSTDIR\${PRODUCT_EXECUTABLE}"
    Delete "$INSTDIR\${INX_UPDATE_HELPER}"
    Delete "$INSTDIR\${INX_GUARD}"
    Delete "$INSTDIR\${INX_LAUNCHER}"
    Delete "$INSTDIR\${INX_CLI}"
    Delete "$INSTDIR\${INX_PORTABLE_ENTRY}"
    Delete "$INSTDIR\current.json"
    RMDir /r "$INSTDIR\versions"

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    !insertmacro inx.deleteUninstaller

    ; Only remove the installation directory if it is empty to prevent data loss
    RMDir $INSTDIR
SectionEnd
