# MS-02 DEG1 / Tesla V100 Upgrade Baseline

Status: **PRE-UPGRADE BASELINE CAPTURED — 2026-10-05**

This lab record preserves the comparison point for the MS-02 before connecting
the DEG1/Tesla V100.

## Hardware state at baseline

- Host: `wwjd4u-MS-02-Ultra`
- CPU: Intel Core Ultra 5 235HX
- System RAM: approximately 122 GiB usable
- Kernel: `7.0.0-38-generic`
- Current graphics device: Intel Arrow Lake-S integrated graphics
- NVIDIA PCI device present: **No**
- `nvidia-smi` present: **No**
- Ollama version observed: `0.32.5`
- Ollama active models before preflight: **0**

The preflight was executed remotely through the trusted GitHub self-hosted runner
`meshalot-ms02`.

## Authoritative MeshAlot M9 baseline

The immutable Milestone 9 Node001 baseline remains the authoritative MeshAlot
before-upgrade comparison.

- Compute Score: **43**
- Mean prompt processing: `44.840360 tok/s`
- Mean generation: `9.690870 tok/s`
- Mean TTFT: `785.457734 ms`
- Prompt CV: `0.005573`
- Generation CV: `0.001912`
- TTFT CV: `0.017696`
- Successful runs: `5/5`
- Thermal throttling observed: **No**
- Benchmark SHA256:
  `a0e212eaf23c53d457e2fb8a10200bf53065938180cdcfddc26b2419e5bb76ba`
- Score SHA256:
  `43d86076567ac70421f81a2b1d81a7de1f9e282582ec57ab40b7500f2b29ef9e`

Standard M9 workload:

- Schema: `m9-v1`
- Workload: `m9-standard-v1`
- Model: `Qwen_Qwen3-30B-A3B-Instruct-2507-Q4_K_M.gguf`
- Quantization: `Q4_K_M`
- Context: `4096`
- Prompt tokens: `512`
- Generated tokens: `128`
- Repetitions: `5`
- llama.cpp commit: `4df29be4f4c3673f428170fda944a5b19f743bb8`
- llama.cpp build: `10454`

## Fresh practical pre-V100 baseline

A separate five-run practical baseline was captured through the current Ollama
runtime using `qwen3:30b-a3b-instruct-2507-q4_K_M`.

- Repetitions: `5`
- Prompt-eval tokens per measured run: `1608`
- Mean prompt throughput: `7939.070038 tok/s`
- Median prompt throughput: `7900.516383 tok/s`
- Prompt CV: `0.019613`
- Mean generation throughput: `5.112801 tok/s`
- Median generation throughput: `5.105840 tok/s`
- Generation CV: `0.003060`
- Mean observed TTFT: `336.812 ms`
- Median observed TTFT: `337.771 ms`
- TTFT CV: `0.022065`

This practical test used a warmed, resident model and therefore its prompt
throughput/TTFT should be compared only against the same practical test after the
V100 upgrade. It is not a replacement for the authoritative M9 Compute Score.

Local evidence:

`/home/jason_guynes/meshalot-v100-evidence/ollama-pre-v100-20261005-081806/ollama-pre-v100.json`

Evidence SHA256:

`c8e3165f169fe0bfdb4d6352c6d88ac55793a245ac0ab3866ee252925ac820f4`

## SYCL runtime note

A same-day attempt to reproduce the old M9 SYCL baseline was intentionally
stopped after the benchmark-only llama-server failed before health.

The existing `hermes-llama.service` also failed on restart with Intel SYCL JIT
errors including:

`bf conversion instruction not supported`

and:

`backend compiler failed build`

The current oneAPI/SYCL device discovery still sees `SYCL0: Intel(R) Graphics`,
but the old SYCL llama.cpp runtime no longer successfully compiles the required
kernel at model load. No failed same-day run is being treated as baseline data.

The service was stopped to prevent a repeated crash/restart loop. Ollama remained
healthy on `127.0.0.1:11434`.

## Post-upgrade comparison plan

After the DEG1/Tesla V100 is connected and validated:

1. Record PCI/NVIDIA driver/GPU/VRAM state.
2. Re-run the same practical Ollama test.
3. Build/run the preserved M9 llama.cpp commit for CUDA on the V100.
4. Use the exact M9 model and workload.
5. Calculate the authoritative MeshAlot Compute Score with the existing scoring
   function.
6. Compare:
   - Compute Score `43 -> post-V100`
   - prompt throughput
   - generation throughput
   - TTFT
   - variance/stability
   - throttling
   - practical Ollama generation throughput
7. Preserve raw evidence and SHA256 hashes for both sides of the upgrade.
