# createSourceFile declaration slice

Step 1 uses integration source at 7ad8666 and the shared scanner slice tool
at faae0e9, including namespace-member slicing. Both were fetched explicitly.
The integration branch was merged into codex/stage3-parser-proof; no rebase.

The fully adapted tree is /tmp/parser-area-adapted, made with that integration
commit's apply.sh and every registered adaptation. No 60-69 edit was made.

createSourceFile reaches 79 files, 5,103 declarations and 190,844 copied-span
lines, comprising 3,958 value and 1,145 type declarations. Its 5,125 spans
include 22 namespace wrapper pieces. All pass the shared tool's byte audit.
The driver helper entries produce the exact same inventory and counts.

The compressed manifest records every original/output span, name, source line
and hash; slice-files.json records the full file list. Source files stay in
/tmp/parser-createSourceFile-slice and /tmp/parser-driver-slice outside Git.
This is the observed output of the requested shared tool, not an eight-file
parser slice or proof of runtime equivalence. Node and stage0 results follow
in separate commits.

```sh
SLICE_TYPESCRIPT=/home/agent/.cache/adamic-stage3/api/node_modules/typescript/lib/typescript.js bash /tmp/parser-slice-tool/stage3/slice/run.sh /tmp/parser-area-adapted /tmp/parser-createSourceFile-slice src/compiler/parser.ts:createSourceFile > /tmp/parser-slice-create.log 2>&1
node /tmp/parser-slice-tool/stage3/slice/verify.cjs /tmp/parser-createSourceFile-slice > /tmp/parser-slice-audit.log 2>&1
```

main.a now imports declaring modules directly. run.sh accepts --inputs TREE
so parser source can come from a slice while the corpus remains the original
81 files with the 2014ef06 full-tree dump hash.

## Step 2: full tree green, raw slice fails module initialization

The fully adapted parser ran on the fixed original 81-file corpus. Its output
is byte-identical to the retained original full-tree dump: 35,456,964 bytes,
SHA256 2014ef06f9db928b50d001787c44e490bbdbd8ecd9e912d3f676d58148ffefc5.
The actual Node Identifier-end increment is still caught, exactly one line.

The gathered parser slice emits zero bytes and exits 70. Its first failure is
semver.ts Version.zero's static initialization calling Debug.assert while
Debug is undefined. Diagnostic tracing (normal Node exit 1) retains the stack
in evidence/slice-node-trace.stderr. No trace mode changes input declarations.
A driver rooted in the preserved namespace facade still fails at that read.
A semver-first driver control instead fails in binder.createFlowNode reading
Debug.attachFlowNodeDebugInfo. Neither control is a delivered source repair.

A source-level cause of the large gather is reached Debug.formatSyntaxKind
at adapted debug.ts:444: (ts as Record<"SyntaxKind", Record<string, string | number>>).SyntaxKind.
The shared tool's qualified-reference test only sees an Identifier's immediate
PropertyAccessExpression/QualifiedName parent. Here ts has an AsExpression
parent, so add(symbol, false) gathers every export of its module namespace.
This is an inspected tool path, not proof that only that one path causes the
complete expansion. No second tool or tool modification was made.

The omission mutant removes exactly the exported createSourceFile span at
original parser.ts:1344 (there is also a same-name namespace member at :1978).
The unmodified byte audit passes; after omission it exits 1 at that span.
Node then exits 1 for the missing createSourceFile export, a distinct error
from the unmodified slice's load-time read. This proves omission detection by
the byte audit and loader, not by comparison against a passing slice dump.
The slice runtime equality requirement is unmet.

```sh
bash stage3/drivers/parser/run.sh /tmp/parser-area-adapted /tmp/parser-area-node-final --inputs /tmp/parser-adapted10 > /tmp/parser-area-node-final.log 2>&1
cmp /tmp/parser-area-node-final/node.dump /tmp/parser-node10-final/node.dump > /tmp/parser-area-full-comparison.log 2>&1
bash stage3/drivers/parser/run.sh /tmp/parser-driver-slice /tmp/parser-slice-node --inputs /tmp/parser-adapted10 > /tmp/parser-slice-node.log 2>&1
```

The driver's PrivateIdentifier text now goes through unescapeLeadingUnderscores
so the adapted branded __String type is accepted. The full-tree oracle proves
this leaves all emitted bytes unchanged.
