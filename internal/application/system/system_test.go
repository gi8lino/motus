package system

import "testing"

import "github.com/stretchr/testify/require"

// TestSetupStateCompletesWithoutDatabasePolling keeps first-run state process-local.
func TestSetupStateCompletesWithoutDatabasePolling(t *testing.T) {
	t.Parallel()

	system := New(NewSetupState(true))
	require.True(t, system.SetupRequired())

	system.CompleteSetup()
	require.False(t, system.SetupRequired())
}
