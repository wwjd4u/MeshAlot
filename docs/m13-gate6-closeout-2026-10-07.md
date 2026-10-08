# MeshAlot M13 — Gate 6 Closeout — 2026-10-07

**Gate 6 CODE PASS — Darwin telemetry parsing, macOS cross-compilation, measured WSS RTT.**
**Live execution on Node002/Mac remains PENDING and will be verified at the M13 two-node integration gate. Do not mark M13 complete or deploy production.**

## Source and behavior
- Gate 5 closeout parent: `2b421f827fd4e07f847a2ec2eb51f8825ee9a92d`.
- Initial code: `6e7a0bed3172a7cf7461ceb72e12e6f0f177d324`.
- Fixed parser escaping before validation: `d7dafef66250b20ea15f6da512ef1b1c2b3ef031`.
- Runner gofmt and fast-forward push: `85f39447eb86d18fd705d12a54ac6770ec96006e`.

Implementation:
- `agent/m13_darwin_parse.go` parses the final macOS `top` CPU idle sample and reconstructs approximate available RAM from `vm_stat` page size and reclaimable page counters.
- `agent/m13_sampler_darwin.go` bounds read-only `top`/`vm_stat` collection using a context timeout. Unsupported GPU/VRAM counters remain nil rather than falsely zero.
- `agent/m13_sampler_other.go` explicitly rejects unsupported OS sampling.
- `agent/m13_connection.go` overwrites the caller-supplied latency with measured WSS handshake timing on the first heartbeat, followed by measured heartbeat acknowledgement RTT on subsequent heartbeats. No ICMP/bandwidth testing or port forwarding.
- No new node keys, enrollment, arbitrary command execution or provider-policy writes.

## Runner evidence
- Validation run: https://github.com/wwjd4u/MeshAlot/actions/runs/37711808677
- Job: `113099192059`.
- Baseline verified: `d7dafef66250b20ea15f6da512ef1b1c2b3ef031`.
- Focused `TestM13*`, full Go suite, focused Go race tests, `go vet`, and cross-compile of the agent tests for Darwin x86_64 and arm64 succeeded.
- `M13_GATE6_DARWIN_CROSSCOMPILE=PASS`, `M13_GATE6_CODE_VALIDATION=PASS`.
- First Gate 6 run failed during gofmt due malformed source escape sequences; corrected in the next commit and verified by successful runner run. No code was pushed from the failed run.

## Preservation and next step
- No Node002 enrollment/regeneration or live Mac sampling was attempted. Node002 historic reports and V100 evidence remain unchanged.
- The live macOS `top` and `vm_stat` results must be checked against actual Node002 host before POC closeout.
- No production migration or M13 server deployment.
- **Next — M13 Gate 7:** wire explicit agent `connect` startup mode to the existing identity, real host sampler and bounded session lifecycle, ensuring it does not silently enroll or alter resource controls. Follow with offline detection, reconnect/backoff, and internet/reboot recovery gates.
