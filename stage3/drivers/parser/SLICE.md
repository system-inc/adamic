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
