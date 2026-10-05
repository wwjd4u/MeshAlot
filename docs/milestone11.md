# Milestone 11 — Provider Resource Controls

Status: **PASSED — 2026-10-05**

## Goal

Let users decide exactly how much of their machine can be sold while keeping the agent, not merely the website, responsible for enforcing those limits.

## Governing controls

The agent-side provider-control contract covers:

- maximum GPU use
- maximum VRAM offered
- maximum RAM offered
- maximum CPU use
- allowed hours
- manual pause
- normal, away, and maximum-earnings modes
- behavior when the local user returns

Implementation:

- `agent/provider_controls.go`
- `agent/provider_controls_test.go`

## Agent behavior

New-work admission enforces:

- provider-control validation
- configured GPU / VRAM / RAM / CPU ceilings
- allowed-hours windows, including overnight windows
- manual pause
- local-user return policy
- provider mode values

Already-running work is protected from abrupt termination. Provider-control changes by themselves do not terminate active work. Termination requires a safety condition or explicit authorization.

## Governing M11 test sequence

The required behavior has been implemented and exercised in focused tests:

1. Share 50 percent of a node and reject requests that exceed the offered GPU, VRAM, RAM, or CPU amount.
2. Increase from normal mode to maximum-earnings mode without bypassing owner limits.
3. Return local activity and stop inappropriate new-work admission while allowing existing work to continue.
4. Pause sharing manually and block new work without killing active work.
5. Re-enable sharing and resume compatible new-work admission while preserving the configured ceilings.

An end-to-end governing-sequence test repeats those transitions in order and confirms the final state still rejects over-limit work.

## Validation completed

- Gates 1–8 provider-control contract and policy tests: passed earlier on MS-02.
- Gate 9 50-percent sharing test: passed on MS-02.
- Gates 10–14 transition behavior: passed in an isolated Go harness using the same provider-control semantics.
- Final validation ran on the trusted GitHub self-hosted runner `meshalot-ms02`, physically hosted by the MS-02.
- Final validation used **Go 1.24.13** from the repository's declared Go 1.24 toolchain.
- The targeted M11 provider-control test suite passed.
- The full repository Go test suite passed.
- Branch-scope validation confirmed the M11 branch changed only:
  - `agent/provider_controls.go`
  - `agent/provider_controls_test.go`
  - `docs/current-status.md`
  - `docs/milestone11.md`
- The runner worktree was clean after validation.
- GitHub Actions final-validation run: `37311263984`.
- Pull request #2 merged the verified M11 branch into `main`.
- M11 implementation merge SHA: `2165afc7927b79028ebf9c37219a360c2d7c604a`.
- No production deployment, database migration, node re-enrollment, benchmark rerun, identity regeneration, or rollback-material change was performed.

## Final validation result

**PASS**

The required Go 1.24 MS-02 validation, branch-scope review, full repository test suite, and merge to `main` are complete.

## Production boundary

Milestone 11 does not require a production cutover to validate agent-side provider controls.

Production, node identities, M8/M9 benchmark evidence, M10 scoring history, and rollback material remain unchanged.

## Pass criterion

The owner remains in control and the marketplace cannot silently exceed configured resource limits.

## Result

**PROCEED TO MILESTONE 12**

Provider controls are verified under Go 1.24 as enforced by the agent rather than only represented in UI or configuration.
