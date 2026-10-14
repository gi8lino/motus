package bootstrap

import (
	"testing"

	appkitauth "github.com/containeroo/appkit/auth"
	"github.com/stretchr/testify/require"
)

func TestAuthenticationModeOverride(t *testing.T) {
	require.Empty(t, authenticationMode(nil))
	mode := appkitauth.ModeNone
	require.Equal(t, "none", string(authenticationMode(&mode)))
	mode = appkitauth.ModeProxy
	require.Equal(t, "proxy", string(authenticationMode(&mode)))
}
