# Core assign ownership-helper proof

The generic helper is implemented in the existing owner, adaptation 40, with a
shared final entrypoint in 99-correctness-core-assign. assign and all callers are
unchanged. No Adamic compiler, checker, refusal rule or backend source is modified.

| Check | Result |
| --- | --- |
| Before and accepted full apply, --build-product | Both exit 0 |
| Fresh accepted source versus measured candidate | All 746 files byte-identical |
| Source delta versus baseline | Only hasProperty's pinned header in core.ts |
| assign and every caller | Byte-identical |
| Ordinary compiler builds before and after | Both exit 0 |
| Adamic types, all 79 sorted roots before and after | Both exit 0; zero diagnostics |
| Core JavaScript, comments kept and removed | Byte-identical |
| All nine emitted compiler JavaScript files | Byte-identical |
| typescript.d.ts and tsserverlibrary.d.ts | Both byte-identical |
| typescript.internal.d.ts | Only the hasProperty generic signature changes |
| Whole-owner contract | 12 no-write refusals; foreign edit and LF/CRLF preserved |
| Owner replay | Zero writes; source bytes and mtime preserved |
| Paired adapter replay | 83 pairs; only amended owner output differs |
| Inherited replay failures | Same 14 nonzero exits before and after |
| Replay source trees | Zero changed source bytes on both sides |
| New shared final phase replay | Zero files, edits and writes |
| Original-helper mutant types | All 79 roots, exit 0 |
| Original-helper mutant source | All 746 files equal baseline |
| Original-helper mutant ordinary binder census | NotYet core.ts:1364:29 |
| Accepted per-file census | All 73 wall roots reach Refused core.ts:1484:12 |
| Other six root outcomes | Unchanged per file |
| Temporary observer | Removed; scratch build deleted; no live observer process |

The accepted helper's body is unchanged. Its T extends object parameter keeps
the actual caller type at the existing own-property library operation. The
object-only helper previously required the opaque parameter's native Object
storage, which an abstract assign T did not prove. No dictionary cast or new
string index signature is introduced.

Both full applies use the pinned TypeScript 6.0.3 checkout. The baseline adapter
set was frozen at the clean getownvalues cherry-pick before implementation.
The accepted apply keeps every early adapter input stable: 40 retains its
original early object output, and its shared owner planner introduces the
final generic spelling at the requested late entrypoint. Whole-owner pins
cover the upstream, object and generic states and reject partial/foreign owners.

The ordinary per-file census uses the same unmodified inspect function as
stage3/census/tool: Load, checking, then Lower. Its only Go overlay enumerates
bounded root groups. The fresh accepted tree equals every measured candidate
source byte, not a partial-root fallback. All authored Go evidence is .go.txt.

The paired replay runs each frozen baseline adapter on the completed baseline
copy and each accepted adapter on the completed accepted copy. All source
hashes are checked after every adapter. Only source/code root paths are
normalized in output comparison; elapsed timings remain raw and are not
compared. The amended owner's output is the sole common-adapter difference.
Its 14 inherited nonzero outcomes are retained, not waived or claimed passes.

The mutant restores exactly hasProperty(map: object, key: string), with its
original body. Every source file then equals baseline and the real binder
census returns the original opaque-object-parameter NotYet at core.ts:1364:29.
The new cast refusal at 1484:12 is an existing next wall, not a native compiler
acceptance claim or a new cast.

MEASURE.md reports all four spelling results and complete caller types.
The whole oracle passes its original guards: 106366 passes, the same one
expected failure and zero skips. It records 30782 calls and 61568 argument
appearances in the accepted traversal. A retained hook-timeout attempt adds
one call and two appearances, for 30783 calls and 61570 appearances across
all actual attempts. The only accessor is inherited Object.prototype.__proto__;
there are no inherited enumerable properties, own accessors or symbol keys.

comparison.json, AFTER-PHASES.json, REPLAY-COMPARISON.json, CENSUS.json and the
archives retain complete hashes, diagnostics, commands, raw logs and mutants.
Each archive has an internal SHA256SUMS and an adjacent SHA256. Transformation
file hashes stayed frozen through the accepted apply; later proof metadata
changes product keys without changing the transformation. No Loom product is
published and no force-push or main push is used.
