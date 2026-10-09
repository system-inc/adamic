# GraphQL printer

Adamic port of cohere's GraphQL printer, composed with the existing parser in
`../parser.ts`. `printer.ts` maps all 45 visitor node kinds into the small document
algebra in `doc.ts`; comments, descriptions, block strings and line wrapping are
included. Unicode display width comes from the JSON slice's captured Go tables.
The file wrapper matches cohere's BOM and line-ending normalization.

Build and format one file to stdout from the repository root:

```sh
source /workspace/adamic-tools/env.sh
go run ./cmd/adamic build stage1/cohere/graphql/printer/main.ts -o /tmp/adamic-graphql-printer > /tmp/graphql-printer-build.log 2>&1
/tmp/adamic-graphql-printer example.graphql > /tmp/formatted.graphql
```

Install the external oracle in a scratch directory, not the repository:

```sh
mkdir -p /tmp/graphql-printer-prettier
npm install --prefix /tmp/graphql-printer-prettier --save-exact prettier@3.9.6 graphql@17.0.2 > /tmp/graphql-printer-npm.log 2>&1
export ADAMIC_GRAPHQL_PRETTIER=/tmp/graphql-printer-prettier
export ADAMIC_GRAPHQL_LIBRARY=/tmp/graphql-printer-prettier
mkdir -p /tmp/graphql-printer-corpus
ADAMIC_GRAPHQL_PRINTER_KEEP=/tmp/graphql-printer-corpus ADAMIC_GRAPHQL_PRINTER_BENCH=1 go test -v -count=1 ./stage1/cohere/graphql/printer > /tmp/graphql-printer-tests.log 2>&1
```

The test asserts Prettier 3.9.6 and also runs cohere's embedded fork at that version.
Without `ADAMIC_GRAPHQL_PRETTIER`, only the external comparison and whitespace
proof skip; native, Node, Go, compiler gap and mutant checks still run. Benchmarks
are opt-in. Go oracle programs use overlays, leaving the submodule untouched.
`--cases <path> [defaults|narrow|tight|tabs]` is the test's escaped batch transport.

At base main `5d4c8012a0877094134e6c6bac367ff68f9313e8` and cohere
`715ba94f3608a6500086b1076ce5cb7e51b836db`, recursive discovery found **zero
.graphql files**. It traverses the repository and initialized submodules, including
cohere's nested TypeScript submodule at
`8d550c837c90bd1805b047b7eeccc2baac2d5e7a`, excluding only .git directories.

The 3,564 generated texts comprise 3,000 seeded parser documents, 497 Unicode
width cases, 40 argument-list lengths, 15 upstream formatting cases and 12 deep
valid/malformed documents. They visit every one of the 45 upstream AST kinds.
Each option set has 2,513 formatted results and 1,051 syntax refusals. Four option
sets exercise defaults (120 columns, four spaces), narrow (80 columns, two
spaces), no bracket spacing and tabs. Native, Node source and the JavaScript
backend agree with Go on every byte, including refusal messages. Native runs
under ASan, UBSan and a separate LeakSanitizer pass.

Both Prettier oracles agree on every formatted result and reject 1,046 other
texts. **Five whitespace texts differ upstream**: Go rejects them; Prettier returns
empty text. These are explicit assertions, not an unrestricted exception list.
See [GAPS.md](GAPS.md) for exact inputs and runnable proofs,
[PERFORMANCE.md](PERFORMANCE.md) for throughput, and [VALIDATION.md](VALIDATION.md) for commands and [results](results/) for logs.
Prettier syntax error code frames are not compared byte for byte; shared refusals
are checked as refusals. The Go error message is always compared byte for byte.

The scratch dependency lock is retained in results/oracle-package-lock.json.
Embedded asset SHA-256 values:

- standalone.js: `0c1acd3ad53d96bb66a4c530e4eb6240693f37dd0ecb81f51920edffa8c05567`
- plugins/graphql.js: `b0faef34893f8033b9adeaad048f4cec9456e95baa05f97b51f8315a5a504fec`
