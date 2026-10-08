Scanner core.ts:107:20 remains blocked: untyped numeric Array construction needs a type-contract ruling and hole-aware runtime arrays.
Topic branch codex/notyet-new-expression-topic; final evidence commit SHA accompanies the single push.
Exact scanner witness and typed control are held to Node; focused tests passed 0.465s including Uint16Array NotYet.
Forced any-element representation and accepted sparse construction mutants both fail exact NotYet assertions before emission.
No constructor lowering, IR, backend, runtime or other worker function changed; this does not clear the scanner site.

The user-selected witness is read directly from codex/stage3-scanner-native-3 a3a8c599, stage3/drivers/scanner/evidence/land-area-next/witnesses/08-array.a. Its bytes are retained as internal/oracle/testdata/new_expression_scanner_array_notyet.a. It creates new Array(length) in make(length: number), then prints make(2). Source Node prints 2. The scanner branch is fetched for evidence only, not merged. This topic remains main-based with only our own non-merge work above its main pin.

The current production lowering recognizes the Array constructor but first asks for an element representation. The exact untyped witness stops at an array of any. Our typed control changes only new Array(length) to new Array<number>(length); Node still prints 2, while Adamic stops at an Array constructor creating holes. The dynamic number is JavaScript's length form, not a single element. Replacing it by [length], zero-filled storage or dense undefined entries changes semantics: callbacks and property enumeration must skip holes. The existing dense-constructor lesson cannot safely lower this shape.

Dependencies: the type-contract/language owner must rule on the implicit any[] construction; the runtime/array owner (area/runtime) must supply a hole/presence representation with consistent iteration and methods, plus compiler integration through main. internal/lower/object.go newArrayFilled already explicitly keeps bare holes unsupported and only admits a whole fill; runtime/array.c likewise documents that 0.1 cannot hold holes. These observations establish the current implementation boundary; they do not claim a new language refusal. Neither checked views nor typed-array Uint16 integration supplies sparse ordinary-array semantics. No unlanded branch is merged and no existing NotYet is converted to Refused.

The dedicated scanner _test.go registers both fixtures as not lowering, checks Node output 2, and checks each exact NotYet before emission. Native/JavaScript/sanitizer execution is intentionally absent because production must reject these shapes. The two scratch-overlay mutants change real lowering decisions, never diagnostic strings: (1) force ir.Number instead of deriving the untyped element representation, caught because the exact witness reaches the different hole stop; (2) remove the single-numeric-argument guard, caught because typed construction incorrectly lowers with nil error. No clang failure is counted. Production files are unchanged by the mutants.

Commands, all redirected to logs:

- go test ./internal/oracle -run 'TestNewExpressionScannerArrayDependencies|TestNewExpressionUint16RemainsNotYet' -count=1 -timeout 10m: passed 0.465s.
- python3 stage3/notyet-new-expression/run-scanner-array-mutants.py: both killed by the fixture-specific NotYet assertions; logs and summary in evidence/scanner-mutant-*.log.txt and scanner-mutants.log.txt.
- go test ./internal/oracle -count=1 -timeout 30m: ok  	github.com/system-inc/adamic/internal/oracle	147.437s; output in evidence/scanner-oracle.log.txt.
- go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts: passed 34.403s; output in evidence/scanner-counts.log.txt. Neither unsupported fixture adds a count row.
- git diff --check: passed. No whole-repository gate. Uint16Array remains typed array element type Uint16Array NotYet; no runtime helper or typed-array lowering introduced.
