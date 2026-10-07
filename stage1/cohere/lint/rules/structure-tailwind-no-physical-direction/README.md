# structure/tailwind-no-physical-direction

Candidate .a port of the unchanged Go tailwind.NoPhysicalDirection rule. It
visits every string and static template piece, reproduces the Go class-shape
regular expression and Unicode White_Space splitting, maps all twenty physical
families in upstream order, and preserves negative-value and rtl/ltr exceptions.
Messages use the resolved Go policy catalog's exact text. It proposes no fixes.
The file gate preserves Go's .ts/.tsx-only behavior, including silence on .a.

Supported non-JSX source is compared against original cohere tests, TypeScript's
compiler and all stage1 .ts/.a files. The single pinned JSX attribute test is
not passed: Go emits useLogicalClass, while the shared parser explicitly refuses
it on source Node and sanitized native. This candidate is not a complete JSX
port. The gap probe independently builds the unchanged Go rule with TSX parsing.

The direction_exemption_removed mutant compiles, exits cleanly without sanitizer
reports, and adds an unwanted rtl-prefixed finding. Go output comparison catches
it on source Node, emitted JavaScript and sanitized native.

## Reproduce validation

Source /workspace/adamic-tools/env.sh (or your setup script's actual path), then:

```
python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/validate.py \
  --scratch /tmp/lint-wave1-15-validation \
  --typescript /path/to/pinned-TypeScript-v6.0.3 \
  > /tmp/lint-wave1-15-validation.log 2>&1
```

The driver copies three shared Go files into scratch and applies the reviewable
compatibility.patch there. A Go overlay supplies those copies and the owned
validation_test.go.txt. No shared repository file is modified. The proposal
adds .a discovery/import/copy/mutant/corpus support, emitted JavaScript checks,
repairs the inherited profile test's old portFiles use, and preserves original
captured filename extensions. It excludes exactly one pinned JSX source and
requires that gap count to equal one; every other captured case is compared.

Default registration is still blocked by the .ts-only foundation. The proposal
is not installed or certified as a foundation migration. Generated registries
are ignored. The existing shared .ts sources are retained; no new Adamic .ts
implementation was written. The validation overlay extends the one published
by origin/codex/lint-wave1-12, with this unit's owned tests and explicit JSX gap.

The continuation report and per-rule rates are under
../../claims/wave1-15-next-report.md. Evidence is in evidence/.
