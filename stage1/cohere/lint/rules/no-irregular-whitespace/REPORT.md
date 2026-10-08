# no-irregular-whitespace

Ports the unchanged Go core rule through its typed options adapter. All 247 captured upstream cases and four witnesses match Go byte-for-byte on source Node, emitted JavaScript and sanitized native. Exact character set, run grouping, leading-BOM exemption and all five skip options are preserved; Go provides no fixes or suggestions. The compiling run-splitting mutant is caught on Node and emitted JavaScript; the package canary covers sanitized native.

Both held and clean branches passed the full lint package against wave2-02 at `95968dd9`, with every input enabled: 147 pass, 0 fail, 1 skip (top-level 41/0/1). The skip is the pending shared `TestCheckerBridgeRefusalPending` API test. `gofmt -l` is empty. This landing adds only this registered directory; no shared code, historical evidence or wave suites.
