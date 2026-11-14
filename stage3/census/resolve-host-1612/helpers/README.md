# Replay helpers

These are the exact measurement helpers retained from /tmp/resolve-host-1612.
They use that scratch root and sibling work/build trees. The census uses the
repository's existing stage3/census/latent/make_overlay.py and no-output guards;
no compiler code or checker option was changed for measurement.

The adapter, fixture, presence/parameter contract and restoring mutant live in
the implementation commit under stage3/adapt/99-wasm-crypto-import.
commands.json records the completed invocations; raw/ retains complete streams
and logs. lane/ is the fresh whole unfiltered run, not a targeted test selection.
