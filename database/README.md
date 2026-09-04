# Database

The migration preserves the planned core model. Wallet balances must be derived from append-only `wallet_transactions`. The Milestone 0 API intentionally uses memory until Milestone 4 adds persistent database access.

## Milestone 4 database verification

Core schema and restricted-role tests passed on the user's PostgreSQL 16 test
database. Migration-runner PostgreSQL verification remains pending.

`000001_core.up.sql` is unchanged. `000002_audit_and_ledger.up.sql` adds
append-only protections for wallet transactions, usage, and job events (including
TRUNCATE), immutable node ownership and job parties, automatic status events,
terminal-job preservation, and ledger sign/party/idempotency checks.

Only simulated `earning` (positive) and `usage` (negative) entries are supported.
They require a completed job with usage. Failed jobs retain audit/usage records
but cannot be charged. One entry of each kind per job prevents duplicate payments
under a different key. `wallet_balances` reconstructs balances from the ledger;
`job_usage_parties` traces usage to both parties. Negative test balances are
deliberately allowed: funding, spending limits, pricing, fees, refunds, and atomic
settlement are later accounting work, not implemented by this migration.

### Isolated PostgreSQL 16 test

Run from the repository using an existing empty database named `meshalot_test`:

```sh
sudo -u postgres psql -X --dbname=meshalot_test --set=ON_ERROR_STOP=1 --file=database/test-bootstrap.sql
```

The harness refuses any other database name or an existing public table. It
applies the two migrations in order, then tests all 13 core tables, successful
and failed jobs, reverse-role spending, balances, and rejection cases. Test
fixtures are rolled back, but migrated schema remains. No database is dropped.
If migration 2 fails, migration 1 remains committed; inspect the error instead
of rerunning bootstrap or deleting anything. After both migrations succeed,
repeat only the tests with:

```sh
sudo -u postgres psql -X --dbname=meshalot_test --set=ON_ERROR_STOP=1 --file=database/tests/core.sql
```

Never source these SQL files in a shell. Run `psql` as a separate process;
ON_ERROR_STOP makes failures nonzero without closing your terminal.

### Remaining release gates

Keep Milestone 4 in progress until migration tracking passes PostgreSQL tests
and the application persistence path is verified. Before deployment use a separate
non-owner, non-superuser application role without schema/trigger administration
or TRUNCATE privileges. Owner/superuser access can disable triggers: these guards
are not protection against a database administrator. The live API remains
in-memory; this change does not connect it to PostgreSQL or alter live services.

## Tracked migrations

`migrate.py` requires Python 3 and `psql`; it uses no additional Python packages.
It runs trusted repository SQL as the migration administrator, not the application
role. Run as a separate process, never source it into your interactive shell.

- `apply`: validate all recorded filenames and SHA-256 hashes, require contiguous
  history, then execute only pending migrations and record them atomically.
- `status`: show verified/pending versions or report an untracked database. No
  persistent schema changes; an altered/missing applied file is an error.
- `adopt-tested`: explicit one-time registration of the existing empty
  `meshalot_test` schema. Restricted to the two pinned, previously tested files.
  Locks the 13 tables, reruns core tests inside a savepoint, rolls back fixtures,
  and records `adopted_test` rather than claiming it executed migrations.

Adoption verifies the known tests, not a full schema equivalence proof. Use it
only for the already reviewed test schema. It refuses populated tables; do not
delete data to make adoption succeed. Test sequence values can advance despite
fixture rollback, which is normal PostgreSQL behavior.

```sh
sudo -u postgres python3 database/migrate.py adopt-tested --database meshalot_test
sudo -u postgres python3 database/tests/check_tracking.py
sudo -u postgres python3 database/migrate.py status --database meshalot_test
```

The integration checks verify repeated application, checksum mismatch rejection,
and rollback of an intentionally failing pending migration. They do not edit
migration files or drop databases. A nonzero result requires inspection.

For a future EMPTY application database use `apply` instead of test adoption or
bootstrap. Existing untracked schemas are rejected. New migrations must retain
the exact outer `BEGIN;` / `COMMIT;` convention; the runner replaces those outer
boundaries with one transaction around the pending batch and history entries.
Nontransactional DDL, transaction control inside the migration, and psql commands
are unsupported. This is a trusted-code runner, not a general SQL parser.

Tracking lives in `meshalot_meta.schema_migrations`, with no PUBLIC privileges.
It detects file drift, not administrator tampering or arbitrary schema changes.
A transaction-scoped advisory lock serializes runner invocations; administrators
must not run manual schema changes concurrently. A 15-second lock timeout fails
closed. Connection errors near COMMIT can have an uncertain outcome: inspect
`status` before retrying. Never rewrite recorded hashes to hide a mismatch.

Local construction tests (do not require PostgreSQL):

```sh
python3 -m unittest discover -s database/tests -p 'test_migrate.py' -v
```
