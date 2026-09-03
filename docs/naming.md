# Naming standards

- Nodes: `meshalot-<environment>-node-<number>`
- Control servers: `meshalot-<environment>-control-<number>`
- Jobs and users: server-generated UUIDs
- Regions: provider region names such as `us-central1`
- Releases: semantic versions
- Environments: `development`, `poc`, `beta`, `production`

Never reuse credentials, databases, or enrollment tokens across environments.
