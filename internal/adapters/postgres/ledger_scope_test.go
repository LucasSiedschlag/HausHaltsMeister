package postgres

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestQueriesScopedByLedgerID(t *testing.T) {
	if _, err := exec.LookPath("rg"); err != nil {
		t.Skip("rg not available; skipping ledger_id scan")
	}

	root, err := repoRoot()
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}

	args := []string{
		"-n",
		`WHERE\s+id\s*=\s*\$[0-9]+`,
		"internal/adapters/postgres",
		"--glob", "*.sql",
		"--glob", "*.go",
		"--glob", "!**/sqlc/**",
		"--glob", "!**/auth/**",
		"--glob", "!**/ledger/**",
		"--glob", "!**/budget/**",
		"--glob", "!**/ledger_scope_test.go",
		"--glob", "!**/access.go",
	}

	cmd := exec.Command("rg", args...)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("unscoped queries found (WHERE id = without ledger_id):\n%s", strings.TrimSpace(string(output)))
	}

	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		return
	}

	t.Fatalf("rg scan failed: %v\n%s", err, strings.TrimSpace(string(output)))
}

func repoRoot() (string, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", errRootPath
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "..")), nil
}

var errRootPath = runtimeError("unable to resolve repo root")

type runtimeError string

func (e runtimeError) Error() string {
	return string(e)
}
