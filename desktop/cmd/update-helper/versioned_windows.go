//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"inx/internal/installlayout"
	"inx/internal/repair"
)

// activateVersionedWindowsFromStaging publishes the versioned-v1 layout from a
// staged NSIS payload:
//
//	InstallRoot/
//	  inx-launcher.exe
//	  Inx.exe              (launcher alias when present or portable)
//	  inx-cli.exe          (CLI entry; full binary for now)
//	  current.json
//	  versions/<version>/
//	    inx-desktop.exe
//	    inx-cli.exe
//	    inx-update-helper.exe
//
// Any failure before the current.json pointer swap leaves the previous active
// version unchanged. The helper never counts crashes or selects prior versions.
func activateVersionedWindowsFromStaging(claimed *repair.UpdateTransaction, stagingDir string) error {
	if claimed == nil {
		return fmt.Errorf("versioned activate: transaction is nil")
	}
	installRoot := filepath.Clean(strings.TrimSpace(filepath.Dir(claimed.TargetPath)))
	// When the claimed primary is already under versions/<ver>/, climb to root.
	if root, err := installlayout.ResolveInstallRoot(claimed.TargetPath); err == nil && root != "" {
		installRoot = root
	}
	version := strings.TrimSpace(claimed.ToVersion)
	if err := installlayout.ValidateVersionName(version); err != nil {
		// Accept bare product versions from NSIS (1.20.0 → v1.20.0).
		if !strings.HasPrefix(version, "v") {
			version = "v" + version
		}
		if err := installlayout.ValidateVersionName(version); err != nil {
			return fmt.Errorf("versioned activate: %w", err)
		}
	}
	stagingDir = filepath.Clean(strings.TrimSpace(stagingDir))

	desktopSrc := filepath.Join(stagingDir, "inx-desktop.exe")
	cliSrc := filepath.Join(stagingDir, "inx-cli.exe")
	helperSrc := filepath.Join(stagingDir, "inx-update-helper.exe")
	launcherSrc := filepath.Join(stagingDir, "inx-launcher.exe")
	for _, path := range []string{desktopSrc, cliSrc, helperSrc, launcherSrc} {
		info, err := os.Lstat(path)
		if err != nil {
			return fmt.Errorf("versioned activate: staged %s: %w", filepath.Base(path), err)
		}
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("versioned activate: staged %s is not a regular file", filepath.Base(path))
		}
	}

	requestID := repair.UpdateTransactionID(claimed)
	if requestID == "" {
		requestID = "helper-" + version
	}
	if err := installlayout.ActivateVersion(installlayout.ActivationRequest{
		InstallRoot: installRoot,
		Version:     version,
		RequestID:   requestID,
		Members: []installlayout.Member{
			{Name: "inx-desktop.exe", Path: desktopSrc, Mode: 0o700},
			{Name: "inx-cli.exe", Path: cliSrc, Mode: 0o700},
			{Name: "inx-update-helper.exe", Path: helperSrc, Mode: 0o700},
		},
		RootMembers: []installlayout.Member{
			{Name: "inx-launcher.exe", Path: launcherSrc, Mode: 0o700},
			{Name: "Inx.exe", Path: launcherSrc, Mode: 0o700},
			{Name: "inx-cli.exe", Path: cliSrc, Mode: 0o700},
		},
		RequiredRootNames: []string{"inx-launcher.exe", "Inx.exe", "inx-cli.exe"},
	}); err != nil {
		return err
	}

	// Remove flat release-unit leftovers so the install root is the thin layout.
	// Do not remove the launcher/CLI/alias we just wrote.
	for _, name := range []string{
		"inx-desktop.exe",
		"inx-guard.exe",
		"inx-update-helper.exe", // helper lives only under versions/
	} {
		_ = os.Remove(filepath.Join(installRoot, name))
	}
	// Best-effort retention GC of older version trees.
	_ = installlayout.RetainPreviousVersions(installRoot, 0)
	_ = installlayout.CleanupStaleStaging(installRoot, 0)
	return nil
}

// preferVersionedWindowsActivation reports whether the staged payload is
// complete enough for versioned-v1 activation.
func preferVersionedWindowsActivation(stagingDir string) bool {
	for _, name := range []string{
		"inx-desktop.exe",
		"inx-cli.exe",
		"inx-update-helper.exe",
		"inx-launcher.exe",
	} {
		info, err := os.Lstat(filepath.Join(stagingDir, name))
		if err != nil || !info.Mode().IsRegular() {
			return false
		}
	}
	return true
}
