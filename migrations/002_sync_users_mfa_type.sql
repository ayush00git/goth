-- users.mfa_type is now the source of truth for "is MFA on".
-- One-off sync for rows written before that change.

-- users with an mfa_secrets row get its mfa_type
UPDATE users u
SET mfa_type = m.mfa_type, updated_at = NOW()
FROM mfa_secrets m
WHERE m.user_id = u.id
  AND u.mfa_type IS DISTINCT FROM m.mfa_type;

-- users without one have MFA off
UPDATE users u
SET mfa_type = 0, updated_at = NOW()
WHERE u.mfa_type IS DISTINCT FROM 0
  AND NOT EXISTS (SELECT 1 FROM mfa_secrets m WHERE m.user_id = u.id);
