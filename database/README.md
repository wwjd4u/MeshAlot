# Database

The migration preserves the planned core model. Wallet balances must be derived from append-only `wallet_transactions`. The Milestone 0 API intentionally uses memory until Milestone 4 adds persistent database access.

## Milestone 4 database verification (pending)

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

Database execution is pending verification. Keep Milestone 4 in progress until
the tests pass on PostgreSQL and the application persistence path is verified.
Before deployment, add migration version/checksum tracking and use a separate
non-owner, non-superuser application role without schema/trigger administration
or TRUNCATE privileges. Owner/superuser access can disable triggers: these guards
are not protection against a database administrator. The live API remains
in-memory; this change does not connect it to PostgreSQL or alter live services.
