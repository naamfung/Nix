package builtin

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTryOpenPathVariantsMissingFile checks the enrichment contract for the
// common failure a model hits: a wrong filename inside a directory that does
// exist. The error must name the caller's original path (not the last variant
// a Unix-style shell conversion tried) and say the directory exists, while
// staying classifiable as fs.ErrNotExist.
func TestTryOpenPathVariantsMissingFile(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "llama-prefix-cache.h")

	_, _, err := tryOpenPathVariants(tryWinPathVariants(missing))
	if err == nil {
		t.Fatal("expected an error for a missing file")
	}
	var pe *os.PathError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *os.PathError, got %T", err)
	}
	if pe.Path != missing {
		t.Errorf("PathError.Path = %q, want the original path %q", pe.Path, missing)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("error should stay classified as not-exist: %v", err)
	}
	msg := err.Error()
	if !strings.Contains(msg, "no such file") {
		t.Errorf("message should say the file is missing, got: %v", msg)
	}
	if !strings.Contains(msg, dir) {
		t.Errorf("message should name the existing directory %q, got: %v", dir, msg)
	}
	if !strings.HasPrefix(msg, "open "+missing+":") {
		// The error must be keyed to the caller's original path, never a
		// converted /drive/ variant.
		t.Errorf("message should report the original path %q, got: %v", missing, msg)
	}
}

// TestEnrichPathErrorMissingParent checks that a wrong directory is diagnosed
// by naming the nearest existing ancestor and the first missing segment.
func TestEnrichPathErrorMissingParent(t *testing.T) {
	dir := t.TempDir()
	deep := filepath.Join(dir, "missing-a", "missing-b", "file.txt")
	err := enrichPathError("open", deep, &os.PathError{Op: "open", Path: deep, Err: os.ErrNotExist})
	if err == nil {
		t.Fatal("expected an error")
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("error should stay classified as not-exist: %v", err)
	}
	msg := err.Error()
	if !strings.Contains(msg, "no such directory") {
		t.Errorf("message should say the directory is missing, got: %v", msg)
	}
	if !strings.Contains(msg, dir) {
		t.Errorf("message should name the nearest existing ancestor %q, got: %v", dir, msg)
	}
	if !strings.Contains(msg, filepath.Join(dir, "missing-a")) {
		t.Errorf("message should name the first missing segment, got: %v", msg)
	}
}

// TestEnrichPathErrorParentIsFile checks the "a path segment is a file" case,
// which on Windows surfaces as a not-exist syscall error.
func TestEnrichPathErrorParentIsFile(t *testing.T) {
	dir := t.TempDir()
	plain := filepath.Join(dir, "plain.txt")
	if err := os.WriteFile(plain, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(plain, "child.txt")
	err := enrichPathError("open", bad, &os.PathError{Op: "open", Path: bad, Err: os.ErrNotExist})
	if err == nil {
		t.Fatal("expected an error")
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("error should stay classified as not-exist: %v", err)
	}
	if !strings.Contains(err.Error(), "is a file, not a directory") {
		t.Errorf("message should flag the file-as-directory parent, got: %v", err)
	}
}

// TestEnrichPathErrorKeepsCause checks that non-not-exist failures pass through
// with their original cause (permission, ...) intact.
func TestEnrichPathErrorKeepsCause(t *testing.T) {
	dir := t.TempDir()
	err := enrichPathError("open", filepath.Join(dir, "x.txt"), os.ErrPermission)
	if err == nil {
		t.Fatal("expected an error")
	}
	var pe *os.PathError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *os.PathError, got %T", err)
	}
	if !errors.Is(err, os.ErrPermission) {
		t.Errorf("cause should be preserved, got: %v", err)
	}
}

// TestReadFileMissingFileMessage is the end-to-end contract the model sees: a
// wrong filename inside a real directory must produce an actionable message,
// not a bare /drive/ variant syscall error.
func TestReadFileMissingFileMessage(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "llama-prefix-cache.h")

	_, err := readFile{}.Execute(context.Background(), argsJSON(t, map[string]any{"path": missing}))
	if err == nil {
		t.Fatal("read_file on a missing file should error")
	}
	msg := err.Error()
	if !strings.Contains(msg, missing) {
		t.Errorf("message should name the original path, got: %v", msg)
	}
	if !strings.Contains(msg, "no such file") || !strings.Contains(msg, dir) {
		t.Errorf("message should say the directory exists, got: %v", msg)
	}
	if strings.Contains(msg, "The system cannot find") {
		t.Errorf("message should not be the raw syscall text, got: %v", msg)
	}
}

// TestReadFileEncodedMissingFile covers the edit_file/multi_edit/delete_range
// read path, which shares the same enrichment.
func TestReadFileEncodedMissingFile(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "nope.txt")

	_, _, err := readFileEncoded(missing)
	if err == nil {
		t.Fatal("expected an error for a missing file")
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("error should stay classified as not-exist: %v", err)
	}
	if !strings.Contains(err.Error(), "no such file") || !strings.Contains(err.Error(), dir) {
		t.Errorf("message should say the directory exists, got: %v", err)
	}
}
