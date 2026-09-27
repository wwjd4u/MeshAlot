# Milestone 9 — Compute Benchmark and Compute Score

Status: **PASSED — 2026-09-27**

## Goal

Measure useful AI inference performance rather than ranking machines only by
hardware labels.

## Standard workload

- Schema: `m9-v1`
- Workload: `m9-standard-v1`
- Model: `Qwen_Qwen3-30B-A3B-Instruct-2507-Q4_K_M.gguf`
- Quantization: `Q4_K_M`
- Context size: `4096`
- Prompt tokens: `512`
- Generated tokens: `128`
- Repetitions: `5`
- llama.cpp commit: `4df29be4f4c3673f428170fda944a5b19f743bb8`
- llama.cpp build: `10454`

## Node001 — MS-02 Ultra

Baseline Compute Score: **43**

Baseline measurements:

- Mean prompt processing: `44.840360 tok/s`
- Mean generation: `9.690870 tok/s`
- Mean TTFT: `785.457734 ms`
- Prompt CV: `0.005573`
- Generation CV: `0.001912`
- TTFT CV: `0.017696`
- Five of five runs successful
- No thermal throttling observed

Baseline benchmark SHA256:

`a0e212eaf23c53d457e2fb8a10200bf53065938180cdcfddc26b2419e5bb76ba`

Baseline score SHA256:

`43d86076567ac70421f81a2b1d81a7de1f9e282582ec57ab40b7500f2b29ef9e`

Deliberate active-load validation reduced the score from `43` to `37`.

Measured active-load degradation relative to baseline:

- Prompt throughput: `-54.640150%`
- Generation throughput: `-57.192615%`
- TTFT: `+154.401626%`

This proved that meaningful competing AI load is visible in the benchmark.

Node001 closure manifest SHA256:

`25ebda494e01ce90a17f89b58d2290b5d1f4b0ad0cbe96a7a3bd3e1698e81d2b`

## Node002 — Intel MacBook Pro

Baseline Compute Score: **47**

Backend:

`CPU + Apple Accelerate`

Baseline measurements:

- Mean prompt processing: `69.422600 tok/s`
- Mean generation: `13.759880 tok/s`
- Mean TTFT: `238.792706 ms`
- Prompt CV: `0.022634`
- Generation CV: `0.013264`
- TTFT CV: `0.047030`
- Peak system RAM: `30.370 GiB`
- Five of five runs successful
- No throttling observed

Baseline benchmark SHA256:

`1b8a22fcff6791266a991499c31f9c9dbfc32c36a40e857ad4461c588d62af0d`

Baseline score SHA256:

`3afee50379cbb60aa4bdb1716ccb608cc91b623b4c8f3b4a0491927fa9777c7e`

Node002 closure manifest SHA256:

`3aedb72c1ff57db43b781475d1f447507f78b9f06360f785022df23b3100b377`

## Two-node measured ranking

1. Node002 — Compute Score `47`
2. Node001 — Compute Score `43`

Node002 relative to Node001:

- Median prompt throughput: `+55.75%`
- Median generation throughput: `+43.29%`
- Median TTFT: `69.97% lower`
- Compute Score: `+4 points`

Comparison artifact SHA256:

`6e6eb336ed771eeae602bf7141a9f4d59c38a3c68eb0a639bea1beac7b9a0014`

Both nodes completed five successful baseline repetitions and neither showed
thermal throttling.

### Cross-platform comparison caveat

Node001 Linux kept the benchmark-owned llama-server resident during throughput
collection.

Node002 Darwin runs throughput and TTFT sequentially because two approximately
30 GiB model processes would place the 64 GiB Mac near physical-memory capacity.

The ranking therefore records the observed M9 benchmark results while explicitly
not claiming identical process-memory contention between the two platforms.

## Signed persistence validation

The complete M9 signed submission path was proven against the isolated
`meshalot_m6_test` database.

Verified:

- Ed25519 node authentication
- HTTP `201` acceptance
- Server-side authoritative Compute Score calculation
- Historical `compute_benchmarks` persistence
- Runtime role cannot directly INSERT into `compute_benchmarks`
- Runtime role can execute only the controlled benchmark insert function
- Production remained untouched throughout the isolated gate

Successful isolated report:

`ace19901-8571-4a58-8c5f-0c380a425caf`

Stored score:

`44`

Condition:

`signed-path-validation`

## Production release

Production source commit:

`bd0acd826651af02e8d9ace3f6107235a727a676`

Production release:

`bd0acd826651a-20260927-094555`

Production database:

`meshalot`

Production migration:

`000008_compute_benchmark_m9.up.sql`

Production verification:

- Migration version `8`
- M9 compute benchmark insert function installed
- Restricted runtime grant installed
- Direct runtime INSERT permission denied
- M9 backend live
- Local API HTTP `200`
- Public API HTTP `200`

## Historical benchmark note

The real Node001 and Node002 baseline measurements remain preserved as immutable
benchmark evidence.

The M9 API intentionally limits the difference between benchmark `collected_at`
and signature time to five minutes. Therefore historical measurements were not
re-labeled with fresh collection timestamps merely to insert them into production.
No benchmark was rerun solely for database registration.

## Milestone 9 pass criteria

**PASS**

The platform can rank Node001 and Node002 by measured AI performance and explain
the primary reasons for the ranking.

Benchmark variance is understood, repeated runs are preserved, abnormal loaded
performance was demonstrated, and historical evidence is retained.

## Result

**PROCEED TO MILESTONE 10 — NODE RATING AND ELIGIBILITY**
