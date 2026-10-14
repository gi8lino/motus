package db

import (
	"context"
	"strings"
)

// AuthenticationSettings returns the persisted browser authentication configuration.
func (s *Store) AuthenticationSettings(ctx context.Context) (AuthenticationSettings, error) {
	var settings AuthenticationSettings
	var usernames, emails, names, groups string
	err := s.pool.QueryRow(ctx, `
		SELECT mode, allow_registration, trusted_username_headers, trusted_email_headers,
		trusted_display_name_headers, trusted_group_headers, trusted_admin_group,
		oidc_issuer, oidc_client_id, oidc_group_claim, oidc_group_sync,
		oidc_groups_authoritative, oidc_admin_group
		FROM authentication_settings WHERE singleton=TRUE
	`).Scan(&settings.Mode, &settings.AllowRegistration, &usernames, &emails, &names, &groups,
		&settings.TrustedAdminGroup, &settings.OIDCIssuer, &settings.OIDCClientID,
		&settings.OIDCGroupClaim, &settings.OIDCGroupSync, &settings.OIDCGroupsAuthoritative,
		&settings.OIDCAdminGroup)
	if err != nil {
		return AuthenticationSettings{}, err
	}
	settings.TrustedUsernameHeaders = headerList(usernames)
	settings.TrustedEmailHeaders = headerList(emails)
	settings.TrustedDisplayNameHeaders = headerList(names)
	settings.TrustedGroupHeaders = headerList(groups)
	return settings, nil
}

// SaveAuthenticationSettings persists non-secret browser authentication configuration.
func (s *Store) SaveAuthenticationSettings(ctx context.Context, settings AuthenticationSettings) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE authentication_settings SET mode=$1, allow_registration=$2,
		trusted_username_headers=$3, trusted_email_headers=$4, trusted_display_name_headers=$5,
		trusted_group_headers=$6, trusted_admin_group=$7, oidc_issuer=$8, oidc_client_id=$9,
		oidc_group_claim=$10, oidc_group_sync=$11, oidc_groups_authoritative=$12,
		oidc_admin_group=$13, updated_at=NOW() WHERE singleton=TRUE
	`, settings.Mode, settings.AllowRegistration, strings.Join(settings.TrustedUsernameHeaders, ","),
		strings.Join(settings.TrustedEmailHeaders, ","), strings.Join(settings.TrustedDisplayNameHeaders, ","),
		strings.Join(settings.TrustedGroupHeaders, ","), settings.TrustedAdminGroup, settings.OIDCIssuer,
		settings.OIDCClientID, settings.OIDCGroupClaim, settings.OIDCGroupSync,
		settings.OIDCGroupsAuthoritative, settings.OIDCAdminGroup)
	return err
}

// headerList normalizes a comma-separated persisted trusted-header list.
func headerList(value string) []string {
	values := strings.Split(value, ",")
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}
