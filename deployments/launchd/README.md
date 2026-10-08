# MeshAlot M13 macOS launch-at-login packaging (NOT installed)

The render script stages an XML launchd **LaunchAgent plist**. It does not
invoke launchctl, create an enrollment code or key, modify the Mac, or write
to the actual ~/Library/LaunchAgents directory.

A LaunchAgent starts at the existing user's login after reboot, not before
macOS permits that user to log in. The full M13 live reboot test must check
what happens in the selected Mac login/session configuration.

Stage only after separately verifying:
- The agent binary has been built and is owned by the existing node user.
- The existing Ed25519 identity is retained unchanged and has mode 0600.
- The HTTPS server URL, provider mode, and manual-pause setting were
  explicitly selected and reflect the owner's current provider controls.

Example syntax for a FUTURE approved staging operation (not executed here):

  python3 render_agent_launchd.py --binary /absolute/path/to/meshalot-agent \
    --identity /absolute/path/to/existing-agent-identity.json \
    --server https://api.meshalot.com --mode away --manual-pause true \
    --output /absolute/path/to/staging/com.meshalot.agent.plist

A separately approved, documented installation process would be required to
copy the staged plist to the user's LaunchAgents folder and load it.
Node002's identity, enrollment and historical benchmark records MUST NOT
be reset or overwritten. The M13 server must be deployed and its migration
tested before trying to connect a live Mac to the production API.

The Linux runner can unit-test this renderer but cannot validate actual
macOS launchd lifecycle, user-login behaviour, GPU telemetry or sleep/wake.
Those remain explicit live acceptance tests.
