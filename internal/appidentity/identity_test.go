package appidentity

import (
	"strings"
	"testing"
)

func TestAppUserModelIDIsStableAndVersionIndependent(t *testing.T) {
	if AppUserModelID != "Inx" {
		t.Fatalf("AppUserModelID = %q, want stable current-generation identity %q", AppUserModelID, "Inx")
	}
	if strings.ContainsAny(AppUserModelID, " \t\r\n") || len(AppUserModelID) > 128 {
		t.Fatalf("invalid AppUserModelID %q", AppUserModelID)
	}
}

func TestAppUserModelIDDoesNotMergeLegacyTauriDesktop(t *testing.T) {
	if AppUserModelID == legacyTauriAppUserModelID {
		t.Fatalf("current AppUserModelID %q must remain distinct from Inx Desktop 0.53", AppUserModelID)
	}
}
