package config

import (
	"os"
	"testing"

	"inx/internal/testenv"
)

func TestMain(m *testing.M) {
	if os.Getenv("INX_CONFIG_LOCK_HELPER") == "1" {
		os.Exit(m.Run())
	}
	testenv.RunWithIsolatedUserState(m)
}
