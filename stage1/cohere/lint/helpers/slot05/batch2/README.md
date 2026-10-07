# Slot 05 second batch

Two retained helpers; the third claim collided and was withdrawn, one per .a file:

- calleeValues requires a call expression, a present callee and a present arguments list. It asks the supplied readsCallee(index) predicate whether the callee matches the current reader name/path settings. It collects every argument in order into fresh literal/template lists. Supply collectClassValues(argument, Callee, {leading:false,trailing:false}) through the callback.
- variableValues requires a variable initializer and an Identifier declaration name. It asks the supplied matches(name) predicate, corresponding to the current reader variables.matches operation, then delegates to classValuesUnder(initializer, Variable). That dependency must start with false outer edges.

SurfaceNode records in types.a carry exact named fields from Go's parser. Index -1 is nil; other indexes must address immutable records in one file's arena. The argument-list presence flag distinguishes nil from a present empty list. Identifier text is decoded, kinds omit Go's Kind prefix, and wrappers are not silently stripped.

ClassValues lists contain non-owning indexes into the caller's immutable payload arena. A literal record retains node identity, decoded text, content range, origin and both edges; a template record retains node identity, origin and both edges. The helper never interprets, rewrites or loses any payload field. Empty Go nil slices are represented by empty arrays. No ownership cycles are introduced. Variable delegation preserves the dependency's returned arrays; the callee reader copies entries into newly accumulated arrays.

Go's private callee and variable methods assert their AST type before their apparent nil guards. Nil and wrong-kind inputs therefore panic in Go. These ports explicitly panic too; exact Go runtime crash prose is not an integration contract. The shared dispatcher supplies only the intended kind.

The isolated Go oracle uses an overlay to expose the actual private Go methods without changing cohere. It reads every nonempty string literal in every inventory-listed test file for each of the 11 consumers, refuses missing coverage, adds controls, parses each string as TSX, and repeats with default, empty and custom settings. The custom settings include an invalid regex, which actual Go compilation skips. All payload records are produced by Go, interned by their complete contents, and carried unchanged by arena index.

Downstream collection callbacks receive real Go-generated answers. Name and callee predicates receive the actual Go reader answers, preserving compiled selector semantics. These predicate seams are explicit dependencies; this package supplies no matcher or regex compiler.

The comparison covers every call for the callee reader and every declaration for the variable reader. Separate probes establish Go refusal on nil/wrong kinds and require Node and sanitized native to exit 70 with an explicit refusal. Successful runs and semantic mutants must exit 0 without stderr. Mutants must compile and differ from Go output; a compiler or sanitizer failure is not credited.

Run from the repository root after sourcing /workspace/adamic-tools/env.sh:

    ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch2 -count=1 -v -timeout=15m > /tmp/lint05-batch2-final.log 2>&1

Set ADAMIC_SLOT05_BATCH2_EVIDENCE to an existing directory for per-consumer counts and corpus hashes. Go pin drift from 7945d102a6c18dd36adf9114a758ce646e8b2359 fails the test. The runner and all new implementation/evidence files stay inside this slot's directory. Shared registration, harness and compiler files are unchanged.

Limits: no full lint diagnostics/fixes, external fixture replay, reconstruction of dynamically assembled Go fixture strings, production AST/payload adapter, regex compiler, or downstream collection implementation. Description and option strings are also parsed; literal counts are not full fixture counts. Go's namespace-attribute assertion is not turned into a supported result: the generator explicitly refuses any such consumer input. [CONSUMERS.md](CONSUMERS.md) and [readiness.json](readiness.json) distinguish dependency removals from fully ready rules.

The dispatcher reservation was withdrawn because slot 03 ce7f300 at 00:48:49 UTC preceded slot 05 dd9e8c3 at 00:48:53 UTC. Its passing experiment is not credited as delivered work. Ahra then instructed workers to finish existing claims and claim nothing more; this unit stops with these two uniquely owned continuation helpers. No replacement helper was claimed.

## Current-pin migration

The landing migration replaces the original exact-name map and RE2 callback array contracts with current Go reader predicates. These are API changes for callers: supply readsCallee for arbitrary callee nodes, and variables.matches for identifier text. Argument collection and payload handling remain unchanged. Default, empty and custom compiled-selector settings are replayed through actual Go, with direct, property, string-element and configured member-path call controls. The old REPORT.md and evidence describe the original pin; fresh current-pin results live in ../landing_7945d102/. No shared reader or matcher implementation is ported in this package.
