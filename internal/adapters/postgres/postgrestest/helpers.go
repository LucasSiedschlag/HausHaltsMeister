package postgrestest

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func ApplyMigrations(ctx context.Context, connStr string) error {
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	migrationPath, err := migrationFilePath("001_init.sql")
	if err != nil {
		return err
	}

	sqlBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		return err
	}

	statements := strings.Split(string(sqlBytes), ";")
	for _, stmt := range statements {
		trimmed := strings.TrimSpace(stmt)
		if trimmed == "" {
			continue
		}
		if _, err := conn.Exec(ctx, trimmed); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}
	return nil
}

func WaitForPostgres(ctx context.Context, connStr string) error {
	deadline := time.Now().Add(30 * time.Second)
	for {
		conn, err := pgx.Connect(ctx, connStr)
		if err == nil {
			conn.Close(ctx)
			return nil
		}
		if time.Now().After(deadline) {
			return err
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func SkipIfDockerUnavailable(t *testing.T) {
	t.Helper()

	if os.Getenv("SKIP_DOCKER_TESTS") != "" {
		t.Skip("SKIP_DOCKER_TESTS set")
	}

	host := os.Getenv("DOCKER_HOST")
	if strings.HasPrefix(host, "tcp://") {
		target := strings.TrimPrefix(host, "tcp://")
		conn, err := net.DialTimeout("tcp", target, 2*time.Second)
		if err != nil {
			t.Skip("docker host not reachable")
		}
		conn.Close()
		return
	}

	if runtime.GOOS == "windows" {
		return
	}

	conn, err := net.DialTimeout("unix", "/var/run/docker.sock", 2*time.Second)
	if err != nil {
		t.Skip("docker socket not available")
	}
	conn.Close()
}

func migrationFilePath(name string) (string, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("unable to resolve helper path")
	}

	base := filepath.Dir(currentFile)
	root := filepath.Clean(filepath.Join(base, "..", "..", "..", ".."))
	return filepath.Join(root, "migrations", name), nil
}
