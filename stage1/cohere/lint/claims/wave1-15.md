# Lint wave 1 slot 15

Owner: codex/lint-wave1-15.

The ordered handoff is in HELPERS.md, referenced by helpers/REPORT.md:

43. nexus/import-require-node-namespace
44. structure/network-no-invalidate-cache-literal-key
45. structure/network-no-string-literal-query

All fetched origin branch stage1 trees were checked for implementation filenames;
none of these three rules was found. No rule is skipped.

Claim pushed before implementation. New Adamic files use .a.

Foundation merges were not both clean: helpers conflicted with registration in
six shared driver files. The directory registration versions were retained;
standalone helpers and inventory were merged. docs/parallel-work.md is absent
from main and both foundation branches; docs/lint-registration.md supplies the
available rule-directory contract.

## Continuation claim, October 7

After pushing all previous work and fetching all 297 origin refs, claim the next
three available rules under the user's main-plus-claims selection criterion:

1. structure/tailwind-no-physical-direction: the sole unclaimed candidate in the
   original 46-rule helper handoff referenced by helpers/REPORT.md.
2. @next/next/no-assign-module-variable: first available entry in inventory.json's
   syntax ready for AST/API adaptation list on origin/codex/lint-inventory.
3. @typescript-eslint/default-param-last: next available entry in that list.

Checked every Markdown/JSON claim file under stage1/cohere/lint/claims on all
origin refs (28 files), and production .ts/.a/rule.json source on origin/main.
No selected name is claimed or registered on main. Earlier ports on other origin
branches do not exclude a rule under this continuation's explicit criterion.

The inventory's syntax-ready wave is used after the original helper handoff;
syntax waiting on helpers and type-aware/binding entries are separate waves.
The newly added comment helper report is a separate handoff, not the original
46-rule list requested here. This update is pushed before writing code.

New implementations and their messages are .a files. The inherited default
registration/harness blockers remain. Validation will use a documented scratch
compatibility overlay; no shared-file changes are authorized or committed.

## Third batch

Fetched all 320 origin refs after pushing the previous work on 2026-10-07.
All 46 helper-ready rules are represented in origin claim Markdown. The next
three rules in the inventory syntax-ready wave, absent from main production
registrations and from claim Markdown on all fetched origin branches, are:

- @typescript-eslint/no-unnecessary-type-constraint
- @typescript-eslint/prefer-as-const
- @typescript-eslint/prefer-enum-initializers

Claim audit JSON files enumerate rejected candidates and are not interpreted as
claims. Rule-name mentions in actual claim Markdown, including reserved/skipped
dispositions, were conservatively excluded. This claim is pushed before code.

### Third-batch outcome

All three remain claimed and blocked on the shared repair contract. The exact
upstream diagnostics and current serializer refusals are reproduced by the
owned Go probes; no incomplete rule is registered. See wave1-15-third-report.md
for commands, shape evidence, Go-only probe mutants and untested backend scope.

## Fourth batch

Fetched all 341 origin refs on 2026-10-07 after pushing all nine owned rule
implementations and their comparison evidence. Checked all 52 distinct origin
claim Markdown blobs and main production registrations against the original
46 helper-ready rules followed by the inventory syntax-ready wave. Claim:

- no-multi-str
- no-nonoctal-decimal-escape
- no-octal

These are the first three remaining eligible entries. This update is pushed
before implementation. Only owned rule directories will be changed.
