ALTER TABLE IF EXISTS accounts
    DROP CONSTRAINT IF EXISTS "accounts_owner";

ALTER TABLE IF EXISTS transfers
    DROP CONSTRAINT IF EXISTS "transfers_from_account_id";

ALTER TABLE IF EXISTS transfers
    DROP CONSTRAINT IF EXISTS "transfers_to_account_id";

ALTER TABLE IF EXISTS entries
    DROP CONSTRAINT IF EXISTS "entries_account_id_fkey";

ALTER TABLE IF EXISTS transfers
    DROP CONSTRAINT IF EXISTS "transfers_compound";

ALTER TABLE IF EXISTS accounts
    DROP CONSTRAINT IF EXISTS "transfers_from_account_id_fkey";

ALTER TABLE IF EXISTS accounts
    DROP CONSTRAINT IF EXISTS "transfers_to_account_id_fkey";

ALTER TABLE IF EXISTS transfers
    DROP CONSTRAINT IF EXISTS "transfers_from_account_id_fkey";

DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS transfers;
DROP TABLE IF EXISTS entries;
