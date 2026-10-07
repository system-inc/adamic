Built: collapse.isValidThemePrefix, collapse.namespaceForVariantRoot and collapse.convertUnderscoresToWhitespace in separate .a files.
Commits: claim c332c14 pushed before code; implementation 12018bd; prior nine helpers already tested and pushed through ad1df33.
Commands and outputs: setup 32s, nproc 5; isolated batch PASS 35.258s; full helper package PASS 235.975s; filtered uncached input oracle PASS 2.653s; touched-package go vet PASS.
Mutants: empty prefix, reject z, namespace substring, ignore skip and escaped underscore changed to space; all compile and run on three backends, then fail the actual Go comparison.
Not covered: full repository gate, integrated rule findings/fixes/suggestions, invalid UTF-8 byte strings, isolated UTF-16 surrogates and eight unavailable Tailwind live-engine/corpus gates.

# Selection and claim ownership

Finished and pushed the previous nine helpers before this batch. Fetched all six origin codex/lint-helpers* branches and read every claim. All named helpers with more than six remaining consumers were reserved. The generic seven-consumer strict option decoding and schema validation entry describes remaining per-rule configuration, rather than an unclaimed Go helper symbol. The chosen three leaf helpers tie the highest available named count, six each. Claim c332c14 was pushed at 01:41:08 UTC before implementation.

Subsequent all-branch refreshes exposed later overlapping claims: slot 03 2390671 at 01:41:11 and slot 04 a33e02d at 01:42:42. Both explicitly withdrew isValidThemePrefix and namespaceForVariantRoot to this slot. No earlier competing claim was found. Full snapshots are in evidence/claims.log and final-claims.log. No additional helper is claimed.

Changes stay inside slot-owned helper, fixture, test and report files. No shared registration generator, shared rule harness, protected compiler file or tracked cohere source is edited. Oracle and capture instrumentation use temporary Go overlays. New Adamic files are .a. The inherited options_json.ts stays unchanged and is renamed .a only in temporary mutant copies. No PR opened.

# Observations

The actual Go helpers are the external oracle, pinned to cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. The capture records 157 deduplicated runtime rule/file/source inputs across all six consumers, including dynamically assembled fixtures and inputs captured before external-engine skips. Coverage is checked against the frozen readiness ledger. The Go TypeScript parser supplies decoded literal/template text. Supplementary theme_test.go and framework_variants_test.go string literals are statically extracted controls, not a claim that every Go fixture’s reconstructed runtime input is captured.

The Go oracle also calls actual parseThemeOptions, framework registration, segment and ParseVariant on collected text. It records 202 successful parsed variants and compound roots. Deduplication leaves 1,342 texts. Every text is queried through all three actual private helpers and both conversion flags. All 16,384 ASCII byte pairs are queried as well, including NUL and every backslash/underscore order. Prefix and namespace probes sweep integer code points 0 through 0x10ffff (1,114,112 candidates); surrogate code points map to Go’s replacement rune. Prefix acceptance is exactly ASCII a-z (26), and only @ selects a container namespace (1). Each exceptional scalar observation is recorded, not just aggregate counts.

isValidThemePrefix rejects empty text and accepts only lowercase ASCII letters. namespaceForVariantRoot tests the raw beginning of the string, without validating the remaining root. convertUnderscoresToWhitespace always turns backslash-underscore into an underscore, including when skip is true. Only a bare underscore is controlled by skip. Runs of backslashes preserve all but the backslash immediately before the underscore; other escapes are untouched. The implementation copies chunks at ASCII boundaries, preserving valid Unicode pairs.

The common input domain is valid Unicode text. Malformed Go byte strings and isolated UTF-16 surrogates are not claimed equivalent. Unicode preservation is checked with BMP/astral conversion controls and consumer text; the exhaustive scalar sweep specifically covers prefix and namespace classification.

Expected Go output is removed from runtime input before Adamic reads it. Baselines compare byte-for-byte on Node source, emitted JavaScript and ASan/UBSan native, then against Go. All successful executions exit zero without stderr or sanitizer findings. Compilation or runtime failure is not credited as a semantic mutant. Isolated batch PASS 35.258s. Full regression result: PASS 235.975s with all twenty-one semantic mutants caught.

# Mutants and what catches them

All five are compiling semantic changes; the Node, emitted JavaScript and sanitized native answers agree with one another and differ from actual Go:

| Mutant | First observed difference |
|---|---|
| Empty prefix returns true | Line 1: prefix:true versus prefix:false. |
| Reject z by changing > 122 to >= 122 | Line 56,045: prefix:false versus prefix:true. |
| Namespace uses includes instead of startsWith | Line 28: --container versus --breakpoint. |
| Ignore conversion skip flag | Line 10: bare underscore becomes a space when Go preserves it. |
| Escaped underscore becomes a space | Line 8: escaped underscore is lost when Go preserves it. |

The full package also reruns all sixteen prior compiling mutants: permissive option JSON, permissive schema, missing policy placeholder replacement, ignored strict option fields; accepting non-Identifier JSX attribute names, losing computed Identifier/PrivateIdentifier property names, class-reader cache key collisions; React namespace substring, discarded partial class fields, NBSP excluded as a separator, aliased class slice; dropped standalone factory, dropped bare factory calls, ignored React predicate, math substring matching, wrong calc table entry. Prior reports describe the individual witnesses and their backends. Every prior mutant must still be caught to pass the full package.

# Regeneration and external limits

Capture workflow is adapted from slot 03’s earlier workflow and writes only temporary overlays. Repeated regeneration reproduces sources.jsonl.gz and coverage.json byte for byte; evidence/reproducibility.log records both SHA-256 hashes. Missing consumers or unexpected failures reject regeneration. Tailwind capture exits 1 for these eight known external live/corpus checks:

- TestClassOrderFixturesActuallyRan
- TestUnknownClassFixturesActuallyRan
- TestConflictFixturesActuallyRan
- TestConflictingClassesPlacementIsAccountedFor
- TestCanonicalFixturesActuallyRan
- TestCanonicalClassesPlacementIsAccountedFor
- TestUnknownClassesPlacesEveryCorpusClass
- TestClassOrderLiveMatchesTheEngineOverTheCorpus

Unavailable installations and empty live corpora prevent these gates from passing. They do not block comparison of the actual leaf helpers on captured inputs. No passing integrated Tailwind rule gate, live-engine parity, findings, fixes or suggestion serialization is claimed. See evidence/capture.log for exact failures.

Commands and outputs are retained in evidence/commands.log, setup.log, helpers.log, helpers-full.log, oracle-input.log and vet.log. The first oracle filter selected no tests; oracle.log preserves that result and is not counted as validation. The corrected uncached TestInputAgreesWithNode selection runs all six input probes with zero cache hits and passes in 2.653s. Setup reports Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 32s, done 32s; nproc is 5. The touched helper package is vetted. The full repository gate was not run.

# Readiness inference

Each helper removes one listed blocker from the same six rules: better-tailwindcss/enforce-canonical-classes, enforce-consistent-class-order, enforce-consistent-variant-order, enforce-shorthand-classes, no-conflicting-classes and no-unknown-classes. That is 18 dependency edges across six distinct rules, with zero final blockers removed by this batch. RULES.md names every consumer per helper; readiness.json lists remaining dependencies after all twelve helpers from this slot only. Other workers’ integration is not assumed. Helper parity observations do not imply these rules are fully ported.
