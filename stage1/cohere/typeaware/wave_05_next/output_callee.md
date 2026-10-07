# output-callee

An exact `CallExpression` request returns raw resolved-signature metadata. Fields use the bridge's decimal UTF-16-length framing: version 1, `output-callee`, declaration-present; then file, declaration kind, byte start/end, external-module flag, generator flag, async flag, raw pinned-checker return-type flags, body-present; if present, body kind and byte start/end. Native code decides whether and how to follow it. The pinned Go checker assigns `Never` 262144, rather than TypeScript's enum ordering. Suffixes and non-call requests are refused.
