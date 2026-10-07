# Parser proof: stopped before the closure experiment

This is an incomplete integration report, not the requested exact, ordered
parser lowering blocker list. No parser-specific lowering blocker was measured.
No temporary adaptation was made. No native parser proof is claimed.

## Pinned inputs

Started from origin/main ef3d907ecdc4c771b016f7d9c52372def057a340 on
codex/stage3-parser-proof. Fetched the required branches explicitly because the
checkout's default fetch brought down only main.

| Input | Commit |
| --- | --- |
| taste-not-soundness | aa896b5d5ccc82210184fd01b8fe4d0ce0730a50 |
| flag-enums | f7d62772fa8e52fcfae754e047ceb66dba88b782 |
| namespaces-tsc | ce8a2acf14e420a9c82345236845a377cd4c7a50 |
| nested-functions | b15216dabf65ffaa7152f6e64709b7b062ea01a9 |
| stage3-base | 8728405135d329efc12c837a7a6c293234abbe1c |
| tsc-census | 429c1177f0130f785c19cf590d1860513b2ddbfc |
| stage3-type-imports | a3ef0dc93d5b2a6cf58f74669c763dc83a1aad0e |
| stage3-optional-declarations | e3535e702e285a5dcb4d06364c4f889b8178ea0c |

## Observed integration sequence

The scratch worktree is /tmp/adamic-parser-scratch, branch
scratch/parser-proof. That branch has never been pushed.

1. An octopus merge of taste, flags, namespaces and nested functions failed
   with exit 2. Git reported "Should not be doing an octopus."
2. A sequential merge of taste fast-forwarded successfully.
3. The sequential flag-enums merge conflicted in four files:
   internal/lower/class_inheritance.go, internal/lower/lower.go,
   internal/lower/refusals.go and internal/oracle/counts.md.
4. I stopped without resolving these conflicts. Namespaces and nested
   functions were not subsequently merged. The scratch tree is still in its
   conflicted merge state; it is not a buildable four-feature compiler.

The shell command which printed the sequential merge logs returned zero
because its final command was git diff, but the flag-enums merge itself
failed. Its log explicitly says "Automatic merge failed". Do not count that
shell exit as a successful merge.

The conflict diff is preserved in evidence/merge-conflicts.diff. The conflicts
join nominal checks with enum checks, accessor discovery with enum
initialization, definite-assignment refusal with enum refusal, and independent
count rows. These are integration issues, not evidence about parser.ts.
They may be resolvable; I did not establish that they are impossible.

## Second witness

The stage1/typescript/parser implementation is a separate indexed-node port
following typescript-go's parser. Its --whole traversal has kind names,
positions, ends, cooked identifier/literal text and additional selected
metadata. main.ts maps its internal UTF-16 positions to UTF-8 byte offsets.
nodes.ts prints only the optional-chain bit as its node-flags column, plus
separate literal flags. Declaration semantics are another column. The node
table has no full context/NodeFlags field. expect and semicolon panic on
unsupported or malformed grammar; the driver does not emit parse diagnostics.

Consequently its current output cannot be reformatted into the requested
createSourceFile dump with full flags and parse diagnostics. Producing that
answer would require adding parser state and diagnostic behavior, beyond a
printer change. A kind/position/text projection could be compared after
normalizing positions and kind spellings, but that would be a weaker witness.
I did not run that projection or a Node corpus comparison. There is no measured
list of corpus differences in this report.

The historical WHOLE_REPORT.md records Go/Node/native agreement for its own
protocol over 77 compiler files. That is prior evidence, not a run tonight,
and does not establish this unit's full-flags dump agreement.

## Setup and limits

bash cloud/setup.sh completed successfully. Source the printed tool environment
at /workspace/adamic-tools/env.sh. Go 1.27.1, clang 20.1.8, Node v24.19.0.
Timing lines: Go 0s, clang 0s, Node 0s, submodules 0s, build-cache warm 98s,
total 98s. nproc: 5; cgroup cpu.max: 400000 100000. The complete setup log is
in evidence/setup.log. Its warm run builds test binaries without running tests.

Not done: main.a, run.sh, adapted-tree creation, scanner/parser closure
subtraction, sequential source blocker removal, temporary adaptations, stage-3
baseline oracle, native comparison and the planted Node node.end mutant.
No dump comparison exists yet, so no claim that it catches that mutant is made.
No compiler or scanner-worker file was edited on the deliverable branch.
