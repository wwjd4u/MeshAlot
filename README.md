# MeshAlot

The distributed AI compute network. One account and one agent will be able to provide idle compute and consume remote compute.

This repository is the Milestone 0 vertical scaffold. It deliberately excludes real payments, arbitrary remote commands, public AI-runtime ports, multi-cloud networking, and peer-to-peer transport.

## Layout

- `agent/` — dual-role host agent and control-plane client
- `server/` — control API
- `web/` — React, TypeScript, and Vite dashboard
- `protocol/` — shared wire types
- `database/` — PostgreSQL core schema
- `deployments/` — local and future Google Cloud assets
- `docs/` — architecture, standards, rules, and milestone evidence

## Run the first vertical slice

```bash
go test ./...
go run ./server/cmd/meshalot-server
```

In another terminal:

```bash
go run ./agent/cmd/meshalot-agent -server http://127.0.0.1:8080 -node meshalot-development-node-001 -token dev-enrollment-token
curl http://127.0.0.1:8080/v1/nodes
```

The development token is only a placeholder. Milestone 6 replaces it with short-lived, single-use enrollment and locally generated persistent cryptographic identity.

See `docs/milestones.md` for the current gate and next step.
