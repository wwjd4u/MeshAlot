# Milestone 11 — Provider Resource Controls

Status: **IN VALIDATION — final Go 1.24 verification pending**

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

## Validation completed so far

- Gates 1–8 provider-control contract and policy tests: passed earlier on MS-02.
- Gate 9 50-percent sharing test: passed on MS-02.
- Gates 10–14 transition behavior: passed in an isolated Go harness using the same provider-control semantics.
- Feature-branch scope review: branch remained based on the Milestone 10 closeout commit and only the provider-control implementation/tests were changed before documentation was added.
- No production deployment, database migration, node re-enrollment, benchmark rerun, identity regeneration, or rollback-material change was performed.

## Final validation still required

Before Milestone 11 can be marked PASSED/CLOSED:

1. Check out the `m11-provider-controls` branch on MS-02 without overwriting any local untracked work.
2. Run the targeted M11 provider-control tests under the repository's declared **Go 1.24** toolchain.
3. Run the full repository Go test suite.
4. Confirm the worktree/diff contains only the intended M11 changes.
5. Merge the verified branch into `main` and record the final closeout SHA.

Until those checks pass, this document intentionally does **not** claim Milestone 11 is closed.

## Production boundary

Milestone 11 does not require a production cutover to validate agent-side provider controls.

Production, node identities, M8/M9 benchmark evidence, M10 scoring history, and rollback material remain unchanged.

## Pass criterion

The owner remains in control and the marketplace cannot silently exceed configured resource limits.

## Do-not-proceed criterion

Do not proceed to Milestone 12 until provider controls are verified under Go 1.24 as enforced by the agent rather than only represented in UI or configuration.
