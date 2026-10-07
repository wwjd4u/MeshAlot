# V100 evidence preservation and remote runner proof

Verified on 2026-10-07 at 20:33 UTC (15:33 America/Chicago).

- Runner: `meshalot-ms02`; machine: `wwjd4u-MS-02-Ultra`.
- Workflow branch: `infra/ms02-runner-gateway`.
- Pushed commit, independently fetched and checked out by MS-02: `47b04dc972eef5938bda7cbe32e8783c6fb45881`.
- Successful run: https://github.com/wwjd4u/MeshAlot/actions/runs/37682886330
- Job: `113003180653`.
- Evidence directory: `/home/jason_guynes/meshalot-v100-evidence/post-v100-20261006-134015`.
- Created `SHA256SUMS.txt` exclusively; no existing manifest overwritten.
- All three original files verified with `sha256sum --check --strict`.
- File hashes were checked again after preservation; original evidence remained unchanged.
- Existing benchmark JSON was read; no benchmark was rerun.
- Confirmed prompt processing: 1098.282826 tok/s; generation: 131.700661 tok/s.
- This proves connector push, runner fetch/execution, and connector retrieval of results without user copy/paste. It does not prove Git write credentials from the runner itself.
- Main and M13 were not changed; M13 remains paused pending governing-plan and current-state review.

## Verified SHA-256 hashes and result

```text
9f4ae6bfc486b737fc10a219a0eec194154885c75bf8a8e4d924cbd891d78314  llama-bench-v100.json
2c13ae0e7da3392389f59b82803443615f638c69624ac3ef642c5637c1a8d946  nvidia-after.txt
2ebece9cdb5a5f64bcc170c3b77c619eb3fa6efe8fd2bcf110e5a83b5c083c1c  nvidia-before.txt
d46f07d99d24dac09265bcc9598bfc7a715de36074a21b2f72fb62867f68a5f0  SHA256SUMS.txt
PROMPT_TOK_S=1098.282826
GENERATION_TOK_S=131.700661
EVIDENCE_UNCHANGED=PASS
EVIDENCE_HASH_VERIFICATION=PASS
```
