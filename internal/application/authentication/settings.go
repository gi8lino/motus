package authentication

import "strings"

// Settings is the persisted, non-secret browser authentication configuration.
type Settings struct {
	// Mode selects the effective browser authentication flow.
	Mode string
	// AllowRegistration permits unknown provider identities to create native users.
	AllowRegistration bool
	// TrustedUsernameHeaders identifies accepted trusted-proxy username headers.
	TrustedUsernameHeaders []string
	// TrustedEmailHeaders identifies accepted trusted-proxy email headers.
	TrustedEmailHeaders []string
	// TrustedDisplayNameHeaders identifies accepted trusted-proxy display-name headers.
	TrustedDisplayNameHeaders []string
	// TrustedGroupHeaders identifies accepted trusted-proxy group headers.
	TrustedGroupHeaders []string
	// TrustedAdminGroup names the trusted-proxy group that grants administrator access.
	TrustedAdminGroup string
	// OIDCIssuer identifies the configured OpenID Connect issuer.
	OIDCIssuer string
	// OIDCClientID identifies Motus at the configured OpenID Connect issuer.
	OIDCClientID string
	// OIDCGroupClaim identifies the token claim containing group memberships.
	OIDCGroupClaim string
	// OIDCGroupSync enables group synchronization during OIDC login.
	OIDCGroupSync bool
	// OIDCGroupsAuthoritative removes missing synchronized memberships.
	OIDCGroupsAuthoritative bool
	// OIDCAdminGroup names the OIDC group that grants administrator access.
	OIDCAdminGroup string
}

// HeaderList normalizes a comma-separated persisted trusted-header list.
func HeaderList(value string) []string {
	values := strings.Split(value, ",")
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}
