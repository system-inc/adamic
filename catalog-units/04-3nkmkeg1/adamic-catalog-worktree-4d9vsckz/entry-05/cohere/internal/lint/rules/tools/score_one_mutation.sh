#!/usr/bin/env bash
# Score one mutation, refusing to score a mutant that changed nothing.
#
# A mutation sweep proves a fixture set can see. The failure it cannot see is its own: a replacement
# that matched nothing produces a mutant identical to the original, the suite passes, and that reads
# exactly like a surviving mutant that the fixtures are blind to. No error, no empty output, no
# friction. The probe ran perfectly against the wrong object.
#
# That happened twice in this project on one day, in two nodes that had not spoken. A rule file was
# indented one level deeper than its sibling so a replacement silently matched nothing, and a
# research pass counted the wrong diagnostic marker across 844 of 867 upstream snapshots and read the
# silence as a clean corpus.
#
# So this refuses to report a score until it has proven the input was real. The byte comparison is
# the half that a known-dirty control sails straight past, because a control proves the harness can
# fail while saying nothing about whether this particular edit landed.
#
#   internal/lint/rules/tools/score_one_mutation.sh <file> <package> <python-expression-rewriting-stdin>
#
# The mutant must also compile. A build failure is neither a catch nor a survival, and scoring it as
# a catch is how a sweep congratulates itself for a syntax error.
set -uo pipefail

if [ "$#" -lt 3 ] || [ "$#" -gt 4 ]; then
    echo "usage: internal/lint/rules/tools/score_one_mutation.sh <file> <package> <python-rewrite> [test-name-pattern]" >&2
    exit 2
fi

file="$1"
package="$2"
rewrite="$3"
# A fourth argument scopes the run to one rule's tests with `-run`. That exists because this tree
# has several agents in one package and a sibling's half-written file makes the whole package red
# for reasons unrelated to any mutant. Scoping keeps the baseline green and the measurement honest.
#
# The cost is real and has to be stated: a scoped run cannot see a guard living in another package.
# A mutation renaming a rule reads as a survivor under scoping, because the parity guard that would
# catch it is in internal/registry. So a scoped sweep measures the rule's own fixtures and nothing
# else, and anything relying on a cross-package guard needs an unscoped run to score.
pattern="${4:-}"

if [ ! -f "$file" ]; then
    echo "sweep: no such file: $file" >&2
    exit 2
fi

# The baseline has to be green before a mutation means anything. In a shared package another
# agent's in-flight file can make the suite fail for reasons unrelated to the mutant, and every
# verdict after that reports the package's state rather than the mutation's effect. This is the
# tool's own documented failure mode arriving one level up: a probe running perfectly against
# the wrong object. It scored a comment-only edit as caught by 32 lines before this existed.
if ! go vet "$package" >/dev/null 2>&1; then
    echo "REFUSED: $package does not compile before any mutation, so nothing can be scored." >&2
    exit 2
fi
baseline_command=(go test -count=1 "$package")
if [ -n "$pattern" ]; then
    baseline_command=(go test -count=1 "$package" -run "$pattern")
fi

if [ "$("${baseline_command[@]}" 2>&1 | grep -c 'FAIL')" -gt 0 ]; then
    echo "REFUSED: $package is already failing before any mutation, so every mutant would score" >&2
    echo "as caught. Pass a fourth argument to scope the run, e.g. -run TestYourRule, or wait." >&2
    exit 2
fi

backup="$(mktemp)"
cp "$file" "$backup"
restore() { cp "$backup" "$file"; rm -f "$backup"; }
trap restore EXIT

python3 -c "
import sys
source = open('$file').read()
mutated = $rewrite
open('$file', 'w').write(mutated)
" || { echo "sweep: rewrite raised" >&2; exit 2; }

if cmp -s "$backup" "$file"; then
    echo "REFUSED: the rewrite changed no bytes, so this scores nothing."
    echo "A no-op mutation is indistinguishable from one the fixtures cannot see."
    exit 2
fi

# `go vet` on the package under test, not `go build ./...` on the whole tree. The tree-wide form
# meant any sibling's broken file reported as "your mutant does not compile", and two porters each
# lost four verdicts to that before reading this script. The message named your mutation and was
# wrong about whose fault it was.
#
# Scoped here, so a refusal at this point really is your rewrite. A package that will not compile
# without your mutation was already refused by the baseline check above, in different words.
if ! go vet "$package" >/dev/null 2>&1; then
    echo "REFUSED: the mutant does not compile, which is neither a catch nor a survival."
    echo "This checked only $package, so the breakage is your rewrite rather than a sibling's file."
    exit 2
fi

failures="$("${baseline_command[@]}" 2>&1 | grep -c 'FAIL')"
if [ "$failures" -gt 0 ]; then
    echo "CAUGHT: $failures failing line(s) in $package"
else
    echo "SURVIVED: the mutant compiles, changed bytes, and no fixture noticed."
fi
