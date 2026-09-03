# POC architecture

The centralized POC uses one dual-role agent. A node initiates outbound TLS/WebSocket control traffic; the server never receives arbitrary shell capability. The first permitted runtime will be llama.cpp through localhost only.

Current slice: `agent -> enroll/heartbeat -> control API -> node store -> dashboard`.

Planned job path: `consumer -> API -> scheduler -> provider agent -> localhost llama.cpp -> agent -> API -> consumer`.

Peer networking, payments, AWS, and Azure remain out of scope until their gates pass.
