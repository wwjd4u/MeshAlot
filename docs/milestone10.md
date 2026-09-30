# Milestone 10 — Node Rating and Eligibility

Status: **PASSED — 2026-09-29**

## Goal

Decide which computers are allowed to receive which kinds of jobs.

## Rating model

Scheduling preserves five independent dimensions:

- Compute
- Network
- Reliability
- Availability
- Trust

These dimensions are not collapsed into one authoritative scheduling score.

Each dimension also carries explicit evidence state so an unknown measurement is
not treated as a valid zero or silently accepted.

A node with incomplete evidence is provider-facing `unrated`.

## Provider-facing tiers

Initial tiers are:

- Basic
- Standard
- Gold
- Platinum
- Datacenter

Tier is descriptive only.

Tier does not replace the five scheduling dimensions and cannot make an
otherwise-ineligible node eligible for a workload.

## Workload classes

Initial M10 workload classes:

- `text-inference`
- `large-data`

Eligibility is evaluated independently for each workload class.

## Governing test matrix

The required M10 cases passed:

1. High-compute / slower-network node.
2. Lower-compute / faster-network node.
3. Eligibility differs by workload type.
4. A node can remain eligible for text inference while being ineligible for a
   large-data workload.
5. Strong Compute and Network scores cannot hide inadequate Reliability.
6. Unknown required evidence fails closed.
7. Unknown workload classes fail closed.

## Tier validation

Tier behavior passed validation for:

- Basic
- Standard
- Gold
- Platinum
- Datacenter
- Unrated when required evidence is incomplete

A weak measured dimension limits the provider-facing tier rather than being
hidden by stronger dimensions.

## Persistence design

M10 migration:

`database/migrations/000009_node_rating_m10.up.sql`

The persistence schema stores the five dimensions separately and contains no
`overall_score`.

Evidence-known flags are persisted independently for:

- Compute
- Network
- Reliability
- Availability
- Trust

Incomplete evidence requires tier `unrated`.

The migration includes a controlled `SECURITY DEFINER` function:

`public.insert_node_rating(...)`

The runtime role has no direct INSERT, UPDATE, or DELETE permission on
`node_ratings`.

PUBLIC execution of the controlled insert function is revoked.

M10 runtime privilege assertions are defined in:

`database/tests/milestone10_runtime_privileges.sql`

## Validation evidence

Pure rating and eligibility implementation:

- `server/node_rating.go`
- `server/node_rating_test.go`

Checkpoint commits:

- `938af187390346abedb0c6282116fad798b29d93`
  `Add M10 node rating and eligibility core`
- `53d03bb8014e1c6dfab1db4eee3a570a2fce7e03`
  `Make M10 ratings evidence aware`
- `33f7f607c967b21b346c52179de258163606e658`
  `Add M10 node rating persistence schema`
- `358614882424b660873d1a205af717b1a90e1d5c`
  `Secure M10 node rating persistence`

Repository integration tests passed after the M10 eligibility and tier logic was
implemented.

Migration construction tests passed.

The migration/function/grant/test PostgreSQL signatures were verified consistent.

## Production boundary

Milestone 10 did not require a production database cutover.

Migration 9 has not been applied to production as part of this milestone.

No M9 benchmark was rerun.

No node was re-enrolled.

No node identity, historical benchmark evidence, rollback material, or existing
production release was altered.

## Pass criteria

**PASS**

Eligibility is workload-aware and does not incorrectly reject useful
lower-bandwidth hosts for light workloads.

A node's strengths and limitations are reflected in scheduling eligibility.

Compute, Network, Reliability, Availability, and Trust remain separate
scheduling inputs.

## Result

**PROCEED TO MILESTONE 11 — PROVIDER RESOURCE CONTROLS**
