# Milestone 12 — llama.cpp Runtime Adapter

Status: **PASSED — 2026-10-05**

## Goal

Integrate llama.cpp as MeshAlot's first local AI runtime while keeping the runtime private to the provider machine and exposing only explicitly permitted AI operations.

## Governing runtime boundary

The agent-side runtime contract provides:

- llama.cpp as the first supported runtime
- loopback-only runtime URLs
- explicit health reporting
- model discovery
- model capability reporting
- non-streaming text completion
- an explicit operation allow-list
- no arbitrary shell, command, process, or generic execution operation

Implementation:

- `agent/runtime_adapter.go`
- `agent/runtime_adapter_test.go`
- `agent/runtime_llamacpp.go`
- `agent/runtime_llamacpp_test.go`

## llama.cpp HTTP integration

The adapter uses llama-server's documented local HTTP interfaces:

- `GET /health`
- `GET /v1/models`
- `POST /v1/completions`

Runtime configuration accepts only HTTP loopback endpoints with an explicit TCP port. Public, LAN, wildcard, HTTPS, credential-bearing, redirected, query-bearing, fragment-bearing, and subpath runtime URLs are rejected by the M12 boundary.

The adapter does not follow HTTP redirects.

## Runtime timeout policy

Live validation exposed a real timeout bug: the preserved 30B CPU-only Qwen3 completion required several minutes, while the original adapter imposed a 30-second HTTP timeout and the CLI imposed a 60-second completion timeout.

The fix separates fast control operations from potentially slow inference:

- health/model control timeout: 10 seconds
- completion timeout: 15 minutes
- a caller-supplied context deadline takes precedence

Regression coverage verifies that completion receives a longer default while caller deadlines are preserved.

## Agent test entrypoint

The MeshAlot agent CLI includes a narrow `runtime` diagnostic command with explicit actions:

- `meshalot-agent runtime health`
- `meshalot-agent runtime models`
- `meshalot-agent runtime complete`

The default runtime endpoint is:

`http://127.0.0.1:8080`

A different endpoint may be supplied only if it still passes the loopback-only runtime validator.

Implementation:

- `agent/cmd/meshalot-agent/runtime.go`
- `agent/cmd/meshalot-agent/runtime_test.go`
- `agent/cmd/meshalot-agent/main.go`

## Source validation

The source implementation is covered by focused tests for:

1. accepted loopback forms: IPv4, IPv6, and localhost
2. rejection of wildcard, LAN, public, credential-bearing, HTTPS, subpath, query, fragment, missing-port, and invalid-port runtime URLs
3. explicit completion-only marketplace operation policy
4. completion request validation
5. healthy and loading/unhealthy runtime states
6. model discovery and context-window reporting
7. capability reporting
8. OpenAI-compatible completion request/response normalization
9. runtime error propagation
10. refusal to follow redirects
11. MeshAlot agent runtime health output
12. MeshAlot agent model/capability output
13. MeshAlot agent completion path
14. rejection of remote runtime URLs through the agent command path
15. slow-completion timeout policy and caller-deadline preservation

Trusted MS-02 Go 1.24 validation after the timeout fix:

- GitHub Actions run: `37386270583`
- Runner: `meshalot-ms02`
- Host: `wwjd4u-MS-02-Ultra`
- Go: `go1.24.13 linux/amd64`
- Exact M12 branch scope check: **PASS**
- Targeted M12 tests: **PASS**
- Full repository `go test ./...`: **PASS**
- Final runner worktree clean: **PASS**

## Live llama.cpp validation

The governing live runtime test ran on the MS-02 through the trusted self-hosted gateway.

Evidence:

- GitHub Actions run: `37386318160`
- Tested implementation commit: `308bfe336eb59413072d11ca0cdba545dbad2998`
- llama.cpp binary: `~/llama.cpp/build/bin/llama-server`
- llama.cpp version: build `10454`, commit `4df29be4f`
- model: `Qwen_Qwen3-30B-A3B-Instruct-2507-Q4_K_M.gguf`
- model SHA256: `382b4f5a164d200f93790ee0e339fae12852896d23485cfb203ce868fea33a95`
- isolated runtime address: `127.0.0.1:18192`
- Go: `go1.24.13 linux/amd64`

Live checks:

1. llama.cpp started successfully with the preserved Qwen3 model: **PASS**
2. listener bound only to loopback: **PASS**
3. model discovery through `/v1/models`: **PASS**
4. direct llama.cpp completion: **PASS**
5. MeshAlot runtime health against the live server: **PASS**
6. MeshAlot model/capability reporting: **PASS**
7. completion through the MeshAlot agent: **PASS**
8. stopping llama.cpp caused MeshAlot health to become non-healthy: **PASS**
9. restarting llama.cpp restored healthy state: **PASS**
10. cleanup removed the isolated test runtime: **PASS**

Observed live-model details:

- reported model context capability: `262144` tokens
- direct completion prompt tokens: `11`
- direct completion generated tokens: `24`

## Host-state preservation

The live test used an isolated temporary llama.cpp process and did not re-enable the known-broken legacy SYCL service.

Final host state after cleanup:

- `hermes-llama.service`: disabled
- `hermes-llama.service`: inactive
- Ollama: active
- V100/OCuLink upgrade track: unchanged and still waiting on the replacement bracket

No production deployment, database migration, node re-enrollment, identity regeneration, provider-control change, or V100 configuration change was performed.

## Branch scope

M12 changes remain limited to:

- `agent/cmd/meshalot-agent/main.go`
- `agent/cmd/meshalot-agent/runtime.go`
- `agent/cmd/meshalot-agent/runtime_test.go`
- `agent/runtime_adapter.go`
- `agent/runtime_adapter_test.go`
- `agent/runtime_llamacpp.go`
- `agent/runtime_llamacpp_test.go`
- `docs/milestone12.md`

Pull request #3 contains the M12 implementation.

## Pass criterion

MeshAlot can safely discover and use a real local llama.cpp runtime through localhost-only interfaces, report health and model capabilities, execute an allowed completion through the agent, detect runtime loss and recovery, and expose no arbitrary command execution path.

## Result

**PASS — PROCEED TO MILESTONE 13**

Milestone 12 is complete. The llama.cpp runtime adapter is validated in source tests and against the real local Qwen3 runtime on the MS-02.
