package db

import (
	"fmt"
	"testing"
)

func TestSchemaMigrationsAreContiguousAndEmbedded(t *testing.T) {
	if len(schemaMigrations) != schemaVersionLatest {
		t.Fatalf("latest version is %d but registry has %d migrations", schemaVersionLatest, len(schemaMigrations))
	}
	for index, migration := range schemaMigrations {
		expected := index + 1
		if migration.version != expected {
			t.Fatalf("migration %d has version %d", expected, migration.version)
		}
		if _, err := migrationFiles.ReadFile(migration.file); err != nil {
			t.Fatalf("migration %d file %q is unavailable: %v", migration.version, migration.file, err)
		}
		expectedPrefix := fmt.Sprintf("migrations/%03d_", expected)
		if len(migration.file) < len(expectedPrefix) || migration.file[:len(expectedPrefix)] != expectedPrefix {
			t.Fatalf("migration %d file %q does not use its version prefix", migration.version, migration.file)
		}
	}
}
