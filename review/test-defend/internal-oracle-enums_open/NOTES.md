# Code, oracle, scope, and defense limits

Base: e77a4ae41f473c149aee910c51b73637686a804a. Audit base: 859dee825a4ef89f5c4ecc0f00e6620ac75c4994.

Code under test: Adamic's lowerer and native/JavaScript backends, including fresh-write refusal diagnostics, ordinary-array assignment bounds, arithmetic and conditional emission, and imported module functions. Node, fixtures, tests, agreement helpers, and expected answers were unchanged.

Oracles:
- TestStage3EnumSparseArrayBoundary uses a self-written exact empty stdout, panic stderr, and exit-70 pin for native and JavaScript. Node is run but only its successful exit is checked.
- TestFreshWriteProbesStayRefused uses self-written rejection class, cycle-capable rule label, and the location of each fixture's marked closing write. Its ordinary rejection path does not execute Node.
- TestImportCycleRuntimeCalls runs Node with a fresh runner importing the mutually recursive functions, pins Node's output, adds calls to the already-lowered IR, and compares JavaScript and native output. It also checks native heap accounting. The fixture's generic run only runs module-body logging, not the imported functions after initialization.

Coverage profiles use -coverpkg=./internal/lower,./internal/native,./internal/javascript. Each target was measured separately from its named subsumer. Exclusive covered blocks are leads; they do not prove unique behavior. Go coverage does not instrument runtime C. The sparse-array path differs by ordinary-array indexed assignment; fresh refusal differs by exact marked-write location; import calls differ by mutually recursive execution history after initialization.

The whole clean package timed out at 90.066 binary seconds with no preceding test failure. The narrowed clean matrix and selected generic fixtures pass. The inventory grew from 193 to 198 tests. All five added tests are replayed separately for every mutant. Uniqueness is bounded to the recorded rows and selected generic fixture subcases. Other package rows and other packages are unknown. No deletion recommendation follows from these limited attempts.

Mutants were planned before running the defense matrix. D1 changes the marked-write diagnostic prefix. D2 changes native ordinary-array panic wording. D3 changes JavaScript array-write panic wording. D4 changes the same wording consistently in ordinary-array C, typed-array C, and JavaScript. This is a single shared-diagnostic fault implemented in three constant sites; the diff is intentionally multi-file. D5 changes subtraction lowering to remainder to perturb recursive decrement. D6 swaps binary operands to perturb recursive decrement and comparisons. D7 flips the emitted conditional branch to perturb recursive termination. These use change-constant/option, swap-arguments, and flip-condition operations. No switch, oracle, test, or harness mutation was used.

Each standalone diff applies to the recorded main base. All seven variants passed go vet for lower/native/javascript. Native binaries in the matrix were rebuilt using each variant's separate ADAMIC_BUILD_CACHE_DIR. The matrix also supplies native product compilation evidence for C edits. ADAMIC_GATE_UNCACHED=1 prevents cached agreement observations.

Disk check: /tmp has an 8.8 GB total filesystem, so it cannot meet a 15 GB free requirement. It had 8.6 GB free initially, 8.6 GB after removing the earlier /tmp/defend-typed scratch directory, and 8.5 GB during this run. /workspace had 16 GB free. No repository or tools were deleted. No disk-related failure occurred.

Toolchain was warm, so no setup install ran. nproc=5, Go go1.27.1, Node v24.19.0. npm ci in stage3/api completed and reported 346 ms. No other node_modules directory is loaded by these target rows. Exact command wall times are in commands.json and new-commands.json; test-binary elapsed times are in the JSON logs. Building and execution are not separately instrumented in those totals.

Brief friction: the whole-package baseline cannot finish inside the budget, so the matrix is bounded. The original audit's subsumption rests on one coarse dropped-main mutant for two targets and one diagnostic mutant for fresh refusal. Native C reachability cannot be read from Go coverage. The fresh defense proves a diagnostic contract, not a new cycle-safety rejection. Witness failures in the raw matrix can be broken preconditions and are not proof that their own disagreement guard works. The review lane contains pending sidecars; its skipped subcases are recorded rather than treated as passes. The instruction to include added tests required an additional clean baseline and seven matrix replays. The disk threshold exceeds /tmp's total capacity.

Names and assertions: the sparse-array row checks a specific intentional panic boundary, and the import-cycle row actually invokes imported mutually recursive functions after initialization. Neither name promises a behavior absent from its assertions. Failure to find uniqueness in these attempts is not proof of redundancy. There are no cost rows or Node/native twins among the three assigned rows.
