CREATE TABLE IF NOT EXISTS authentication_settings (
  singleton BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (singleton),
  mode TEXT NOT NULL DEFAULT 'local',
  allow_registration BOOLEAN NOT NULL DEFAULT FALSE,
  trusted_username_headers TEXT NOT NULL DEFAULT '',
  trusted_email_headers TEXT NOT NULL DEFAULT '',
  trusted_display_name_headers TEXT NOT NULL DEFAULT '',
  trusted_group_headers TEXT NOT NULL DEFAULT '',
  trusted_admin_group TEXT NOT NULL DEFAULT '',
  oidc_issuer TEXT NOT NULL DEFAULT '',
  oidc_client_id TEXT NOT NULL DEFAULT '',
  oidc_group_claim TEXT NOT NULL DEFAULT '',
  oidc_group_sync BOOLEAN NOT NULL DEFAULT FALSE,
  oidc_groups_authoritative BOOLEAN NOT NULL DEFAULT FALSE,
  oidc_admin_group TEXT NOT NULL DEFAULT '',
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO authentication_settings (singleton)
VALUES (TRUE)
ON CONFLICT (singleton) DO NOTHING;
