# Milestone 12 — llama.cpp Runtime Adapter

Status: **IN PROGRESS — SOURCE VALIDATION PASSED; LIVE RUNTIME VALIDATION PENDING**

## Goal

Integrate llama.cpp as MeshAlot's first local AI runtime while keeping the runtime private to the provider machine and exposing only explicitly permitted AI operations.

## Governing runtime boundary

The agent-side runtime contract now provides:

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

## Agent test entrypoint

The MeshAlot agent CLI now includes a narrow `runtime` diagnostic command with explicit actions:

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

## Source validation completed

The source implementation has been exercised in focused tests covering:

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

## Trusted MS-02 Go 1.24 validation

Final source validation ran through the trusted GitHub self-hosted runner on the MS-02.

Evidence:

- GitHub Actions run: `37384908306`
- Runner: `meshalot-ms02`
- Host: `wwjd4u-MS-02-Ultra`
- User: `jason_guynes`
- Tested M12 commit: `daa9c2b0b927ac45a8e9eb4c476004afe176171f`
- Go: `go1.24.13 linux/amd64`
- Exact M12 branch scope check: **PASS**
- Targeted M12 tests: **PASS**
- Full repository `go test ./...`: **PASS**
- Final runner worktree clean: **PASS**

The final M12 branch scope at that validation contained only:

- `agent/cmd/meshalot-agent/main.go`
- `agent/cmd/meshalot-agent/runtime.go`
- `agent/cmd/meshalot-agent/runtime_test.go`
- `agent/runtime_adapter.go`
- `agent/runtime_adapter_test.go`
- `agent/runtime_llamacpp.go`
- `agent/runtime_llamacpp_test.go`
- `docs/milestone12.md`

Pull request #3 is open on branch `m12-llama-runtime-adapter`.

The ordinary PR workflow did not start for connector-created updates, so the trusted MS-02 gateway was used to obtain the required Go 1.24 validation evidence.

## Live validation still required

M12 must not be closed until the live runtime gate proves all of the following on an actual llama-server instance:

1. a completion succeeds directly through llama.cpp
2. the same class of completion succeeds through the MeshAlot agent runtime command
3. stopping llama.cpp changes MeshAlot runtime health away from healthy
4. restarting llama.cpp restores healthy state
5. the llama.cpp listener is bound only to loopback and is not publicly or LAN exposed

## Production boundary

No production deployment, database migration, node re-enrollment, identity regeneration, provider-control change, V100 configuration change, or llama.cpp service change has been performed as part of the source implementation.

The V100 upgrade remains a separate track.

## Current result

**SOURCE VALIDATION PASS — LIVE RUNTIME GATE PENDING**

Do not proceed to Milestone 13 until the M12 live runtime validation gate passes and M12 is formally closed.
