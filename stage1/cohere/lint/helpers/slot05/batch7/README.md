# Slot 05 seventh batch

Three construction helpers, one per .a file. The interfaces are projections of the Go state at this seam, not full rule or parser implementations.

- addRepositoryFunctionalRoots(descriptors, roots): consider only true functional registrations. Replace collisions with fresh declining descriptors. Preserve the map and all untouched descriptor aliases. Root and PerDeclaration are exposed along with zero reading fields and nil-presence metadata; existing descriptor payloads are opaque to this helper.
- addThemeNamespaces(table, entryKeys, keysInNamespace): enumerate every proper segment prefix of a custom-property key, deduplicate candidates, retain namespaces with nonempty membership, deduplicate member keys, and replace both table containers. Sort by descending UTF-8 byte length, then ascending UTF-8 byte order. Entry keys come from Theme.Entries, including its prefix behavior. The callback is the existing pure Theme.KeysInNamespaces method; ordering of callback side effects is outside the contract because Go ranges a map.
- newUtilityEvaluator(theme, definitions, normalize): create a fresh lookup map, normalize every definition in input order, then register it under its current name. Last duplicate wins. Preserve theme and definition identities. The separately owned normalizeUtilityDefinition receives an opaque body handle; this helper never parses or normalizes CSS itself.

Stage 0 cannot lower nullable object unions at this seam. Nil metadata and nil themes use explicit presence bits. The initial refusal is retained in evidence/nullable-gap.log. The constructor accepts an absent-theme adapter with present=false and retains that adapter unchanged. Definitions are nonnil: Go itself faults on a nil definition after its nil-tolerant normalizer returns. Concurrent mutation, malicious callbacks and Go runtime crash wording are not covered. No compiler or shared harness was changed.

The overlay calls the actual private Go helpers pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db. Each of the six consumers supplies all nonempty Go string literals in its inventory-listed test files. There are 639 distinct literals, including descriptions and options, plus empty, dash, Unicode and NUL controls. Each helper receives 1,292 cases. These are derived helper inputs, not a replay of full rule diagnostics or external fixtures. Theme.Add constructs namespace states with ambiguous nested prefixes, ignored font subnamespaces, trailing separators, subvariables and prefixed entries. Evaluator normalization results come from real Go node mutation, and every overwritten definition is checked as well as the final winner. Old table containers, map/definition identities and post-call mutations are observed.

Baselines compare actual Go, source Node, emitted JavaScript and sanitized native. Each mutant must compile, exit zero without stderr and differ semantically from Go. Crashes and compile failures never count as kills. See REPORT.md, CONSUMERS.md and readiness.json for evidence and exact prerequisite removals.

After sourcing /workspace/adamic-tools/env.sh:

```sh
ADAMIC_SLOT05_BATCH7_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch7/evidence" ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch7 -count=1 -v -timeout=20m > /tmp/lint05-batch7-final2.log 2>&1
```

Tests regenerate their Go corpus in temporary directories and write no repository artifacts unless the evidence environment variable is set. Evidence contains consumer counts, deterministic corpus hashes and logs. Committed log copies trim trailing whitespace; raw run logs remain under /tmp. Production parser/linter integration, full Tailwind CSS compilation, arbitrary invalid UTF-8 and the full repository gate remain outside this comparison.
