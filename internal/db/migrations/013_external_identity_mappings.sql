ALTER TABLE auth_identities
  ADD COLUMN IF NOT EXISTS email TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS display_name TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE INDEX IF NOT EXISTS auth_identities_user_id_idx ON auth_identities(user_id);

INSERT INTO auth_identities (
  provider,
  subject,
  user_id,
  email,
  display_name,
  last_seen_at,
  created_at
)
SELECT
  'oidc:' || issuer,
  subject,
  user_id,
  email,
  display_name,
  last_seen_at,
  created_at
FROM oidc_identities
ON CONFLICT (provider, subject) DO NOTHING;

DROP TABLE oidc_identities;
