-- +migrate Up
-- Оставляем самый ранний аккаунт, у остальных сбрасываем email перед unique-индексом
UPDATE drivers d
SET email = NULL, updated_at = NOW()
WHERE d.id IN (
  SELECT id FROM (
    SELECT id,
           ROW_NUMBER() OVER (
             PARTITION BY lower(email)
             ORDER BY
               CASE WHEN hub_id = 'DH-DEMOTST2' THEN 0 ELSE 1 END,
               CASE WHEN COALESCE(password_hash, '') <> '' THEN 0 ELSE 1 END,
               created_at ASC NULLS LAST,
               id ASC
           ) AS rn
    FROM drivers
    WHERE email IS NOT NULL AND btrim(email) <> ''
  ) t
  WHERE rn > 1
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_drivers_email_unique
  ON drivers (lower(email))
  WHERE email IS NOT NULL AND email <> '';

-- +migrate Down
DROP INDEX IF EXISTS idx_drivers_email_unique;
