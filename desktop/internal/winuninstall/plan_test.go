package winuninstall

import "testing"

func TestPlanDoesNotPromoteLegacyOnlyRegistrationWithoutManagedWailsInstall(t *testing.T) {
	legacy := &Registration{
		DisplayName:     "Inx",
		DisplayVersion:  "0.53.0",
		InstallLocation: `"D:\Inx"`,
		UninstallString: `"D:\Inx\uninstall.exe"`,
	}

	got, err := Plan(nil, legacy, `D:\Inx`, "v1.21.0", true)
	if err != nil {
		t.Fatal(err)
	}
	if got.Managed || got.DeleteLegacy {
		t.Fatalf("plan = %+v, want the full installer to migrate a legacy-only registration", got)
	}
}

func TestPlanRefreshesManagedWailsRegistrationAndDeletesMatchingLegacyAlias(t *testing.T) {
	current := &Registration{
		DisplayName:     "Inx",
		DisplayVersion:  "1.18.0",
		InstallLocation: `D:\Inx`,
		UninstallString: `"D:\Inx\uninstall.exe"`,
	}
	legacy := &Registration{
		DisplayName:     "Inx",
		DisplayVersion:  "0.53.0",
		InstallLocation: `D:\Inx`,
		UninstallString: `"D:\Inx\uninstall.exe"`,
	}

	got, err := Plan(current, legacy, `d:\inx\`, "1.21.0", true)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Managed || !got.DeleteLegacy || got.Desired.DisplayVersion != "1.21.0" {
		t.Fatalf("plan = %+v, want current registration refresh", got)
	}
	if got.Desired.InstallLocation != `d:\inx` ||
		got.Desired.UninstallString != `"d:\inx\uninstall.exe"` ||
		got.Desired.DisplayIcon != `d:\inx\inx-launcher.exe` {
		t.Fatalf("desired registration = %+v", got.Desired)
	}
}

func TestPlanDoesNotRegisterPortableOrUnrelatedInstall(t *testing.T) {
	tests := []struct {
		name      string
		current   *Registration
		legacy    *Registration
		uninstall bool
	}{
		{name: "portable", uninstall: false},
		{
			name: "unrelated legacy install",
			legacy: &Registration{
				DisplayName:     "Inx",
				InstallLocation: `C:\Other\Inx`,
				UninstallString: `"C:\Other\Inx\uninstall.exe"`,
			},
			uninstall: true,
		},
		{
			name: "foreign display name",
			legacy: &Registration{
				DisplayName:     "Another App",
				InstallLocation: `D:\Inx`,
				UninstallString: `"D:\Inx\uninstall.exe"`,
			},
			uninstall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Plan(tt.current, tt.legacy, `D:\Inx`, "v1.21.0", tt.uninstall)
			if err != nil {
				t.Fatal(err)
			}
			if got.Managed || got.DeleteLegacy {
				t.Fatalf("plan = %+v, want no registry mutation", got)
			}
		})
	}
}

func TestPlanRejectsInvalidInputs(t *testing.T) {
	managed := &Registration{
		DisplayName:     "Inx",
		InstallLocation: `D:\Inx`,
		UninstallString: `"D:\Inx\uninstall.exe"`,
	}
	for _, tc := range []struct {
		root    string
		version string
	}{
		{root: "", version: "v1.21.0"},
		{root: `D:\Inx`, version: ""},
		{root: `D:\Inx`, version: "dev"},
	} {
		if _, err := Plan(managed, nil, tc.root, tc.version, true); err == nil {
			t.Fatalf("Plan(%q, %q) succeeded, want error", tc.root, tc.version)
		}
	}
}
