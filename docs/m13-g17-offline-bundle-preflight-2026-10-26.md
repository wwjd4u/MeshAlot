# M13 G17: offline bundle integrity prerequisite

Date: October 26, 2026 (America/Chicago).

**Local implementation/test PASS; G17/1 BLOCKED; G17/2 NOT STARTED;
M13 OPEN / NOT RELEASE-READY.**

## Scope and objective

The operator expanded the original read-only scope to permit implementation and
verification, and subsequently authorized delivery through deployment when
required gates pass. Work remains on the existing milestone path in the original
checkout, without a parallel preparation branch. No gate criterion was relaxed.
The rejected service-identity diagnostic was not recreated, retried or dispatched.

The new `tools/recovery/check_bundle.py` checks an explicitly synthetic nine-file
rollback bundle against independently reviewed manifest and source hashes. It
rejects unsupported material, mismatched PG16/schema/grants versions, tampering,
unsafe paths/types/links/modes, malformed JSON and oversized inputs. It never
executes artifacts, connects to services, extracts archives or certifies a gate.
The accompanying README specifies invocation, trust boundaries and the remaining
baseline-recovery versus M13-runtime acceptance contract.

## Verified source and environment

- Original checkout: `/workspace/MeshAlot`.
- Local branch: `m13-persistent-heartbeat`.
- Base SHA: `3b67f57afcda8bdd98989a2b3ae68d42c1d1b19c`.
- Accepted application candidate: `4f67ee54e5ebca084e38e306e723ae0acd08e77d`.
- Local comparison confirms only docs and two audit scripts differ between those
  commits; no app/schema/test change justified repeating prior accepted tests.
- The initial local baseline was M12 main
  `1016c9b058690f34058dfab9cde7927e421fa469`. A normal targeted Git fetch succeeded
  after granting the executor network access for that command. No proxy, keys,
  credential bindings or network policy were modified.
- Main-only fetch configuration prevented automatic upstream tracking; an
  explicit local checkout of the fetched M13 ref completed successfully. The
  index and tracked worktree matched that exact ref before this change.
- Python 3.12.14; no dependencies, containers or new external service resources.

## Verification contract

Required outcome: valid synthetic bundles are ready only for offline review;
bad integrity/version/path inputs fail closed with sanitized output and nonzero
exit status. No result claims runtime or gate acceptance.

Resources: Python standard library, temporary owner-controlled synthetic files,
64-KiB manifest cap, 64-MiB per-artifact cap and streaming hashing. CLI tests have
a five-second subprocess timeout. These bounds do not certify future rehearsal
CPU/RAM/disk or timing requirements.

Command:

```sh
python3 -B -m unittest discover -s tools/recovery -p 'test_*.py' -v
```

The initial 12 focused tests passed. Review then identified that hardcoded gate
states could become stale; output was changed to `gate_acceptance=NOT_ASSESSED`
and the final focused suite was rerun for that specific behavior change. No
historical application, migration, ACL, Caddy, V100 or node test was repeated.

Final result: **12 tests PASS in 0.099 seconds**. New-file whitespace checks
PASS. At that checkpoint, existing tracked files and index matched HEAD; all
four additions were local and uncommitted. No application rebuild or historical CI run was needed for this
independent standard-library tool.

## Publication and CI integration

The next delivery step adds the same focused Python command to the existing
Go CI workflow, preserving its triggers, read-only token permission and timeout.
Future PR/main CI includes the new tool's tests; M13 pushes do not trigger that
workflow. There is no deployment or diagnostic-dispatch step. Workflow structure
and the staged diff are reviewed before committing to the existing M13 branch.
The actual resulting commit and remote publication must be verified separately;
this record does not claim that an intended push already succeeded.

Coverage includes CLI valid/invalid runs, source and independently pinned
manifest integrity, changed bytes, schema/purpose/grants pairing, PG major,
production-material rejection, file composition, duplicate keys/paths, malformed
and deeply nested JSON, size limits, symlinks including intermediate directories,
hardlinks, unsafe write modes, and preservation of input bytes/modes/mtime.

## Remaining work and preservation

This utility proves neither artifact origin nor actual runtime recovery.
Synthetic classification and manifest provenance need independent review. It
does not validate actual systemd/Caddy semantics, database owners/ACLs or live
isolation. G17/1 still requires real PG16 globals/ownership/authentication
recovery evidence and actual Caddy/TLS dependencies. G17/2's full archived
API/systemd/Caddy fault/restore rehearsal has not started.

No archive or vault key was accessed, no production connection/SSH renewal or
workflow execution occurred during local validation. Publishing this preparation
change is separate from merge, release approval and deployment.
Existing encrypted backups, Toshiba copies, evidence and Node001/Node002
identities were preserved. The support email watch remains independent of
technical execution. A support or Daybreak update alone does not close a gate.

References: `m13-gate17-fixed-completion-sequence-2026-10-13.md`,
`m13-gate17-step3e2-isolated-rollback-staging-plan-2026-10-12.md`,
`m13-g17-g1-postgres16-real-globals-recovery-runbook-2026-10-13.md`,
`m13-g17-g1-tls-certificate-recovery-strategy-2026-10-13.md`, and
`m13-g17-g1-approved-service-diagnostic-tool-safe-stop-2026-10-13.md`.
