package control

import (
	"os"
	"testing"

	"inx/internal/testenv"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	cleanupUserState, err := testenv.IsolateUserState()
	if err != nil {
		panic(err)
	}
	if os.Getenv("INX_CREDENTIALS_STORE") == "" {
		_ = os.Setenv("INX_CREDENTIALS_STORE", "file")
	}
	goleak.VerifyTestMain(m, goleak.Cleanup(func(exitCode int) {
		cleanupUserState()
		os.Exit(exitCode)
	}))
}
