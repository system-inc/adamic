# cloud/gates: the checklist of what an Adamic head owes before it submits

`cloud/submit-gates.py` reads a head's diff against main's tip and prints every gate and row the head owes, with the
evidence it found for each, then a verdict naming every one it couldn't prove (#v3q9ayd). Its input is Compiler's map,
`diff-to-gates.json`. Each gate is a fixed recipe here, `<gate>.sh`, the whole command of a Loom job at the head's sha,
matched by its sha256 against the job's `commandSha256`.

## It is advisory

Adamic ruled on Nov 23 at 14:33Z that the script is the checklist every accept reads, never the authority. It never
authorizes a submit. A gate it shows as met is one the accepter still checks, and the accepter, Compiler or Adamic, is
the authority, whose accept stays a comment on the head's task. Every lane head carries its printout in that accept
comment. A script measured inside the head's own tree can't be the authority, because the head can always reach an
input the judge doesn't list.

Its exit code reports and gates nothing:

- **0:** it found evidence for every owed gate and row.
- **1:** it didn't. Exit 1 is a checklist with items open, never a refusal, and the verdict names each open item.
- **2:** the head, the map or the rows can't be read.

A gate it can't read exactly, such as `stock-js-identical` or the pass ledger row, prints as owed by hand, with what
the accepter checks.

## What it reads, and where

It shows evidence only where it checks it by exact equality: a recipe's command sha256, the repository, the head's
own sha, an exit code, or a fixed line the recipe itself prints. The map, the recipes, `steps.sh` and
`units-compare.py` are read at main's tip, never at the head, and the script stops unless its own bytes are main's
copy of itself.

A job still runs in the head's tree, so a head that changes judging code owes `gates-changed` by hand. Judging code is
this directory, the script, `cloud/sandboxed-job.sh`, and the census tooling `unit-census-entry.sh` and
`unit-census-directory.sh` run from the tree:

- `stage3/census/`
- `stage3/apply.sh`, `stage3/apply.py` and `stage3/caches.py`
- `stage3/meter/entry_overlay.py`
- `stage3/api`'s package files

`JudgePaths` names each one with why. That list is the checklist's reach, not a proof. A head can reach other inputs,
which is why the accepter reads the change itself.

## The map

`diff-to-gates.json` is Compiler's, copied here at version 7 (Compiler on #fj2ca6r, Nov 24, and on #hj3ajhz, Nov 26). Version 7
replaces `unit-census-pair` with `unit-census-entry` and `unit-census-directory`, one census pass a job. Version 6 added
`stage1-gaps-whole` and `stage3-fixtures-whole`, aligns `green` with main's recipe, and records `exposure`.
Adaptations owe directory-local ledgers, an owed-by-hand plant proof and table review, and an `-after` comparison
in every output test, while `mutants-red` is triggered by added or changed `*.mutants.md` files.

## One head

A head is one change's diff against main's tip. A merge of main into it is still one head. Two heads merged together
are two, and the report says so. Several `--head` arguments are each reported on their own.

## The bootstrap

The landing that first brings this directory and the script to main is read by code main doesn't hold yet. It runs
with `--gates <directory>` naming the caller's copy, and its report notes that its judging code is the caller's. Every
head after it is read by main's copy.
