# V3 array repair and runtime API receipt

The initial comparison base is rehearsal 2b6c032a. The push raced V2 and these repairs were rebased onto 4fb447d5; V2's temporary duplicate viewArrayElementType is removed in favor of the arrays-owned definition, kept restricted to ordinary arrays until own-field contracts exist; both slices' counts are preserved, and assertion 09 is recorded Compiles after V3 installs its standalone array field adapter. Mixed union array membership remains pending V2's switch. It installs lane 2's array adapter, selected-element payload readers and array helper APIs. The 21 original TestCheckedViewArrays cases and their parent now pass; none is marked pending on V2. Two additional cases cover an undefined array receiver and sparse slots. Array element contracts are checked at the selected read. Unread elements stay lazy, and receiver/index expressions are evaluated once.

The public runtime APIs needed by V2 are present: adamic_array.element_kind and adamic_array_holes_at. Numeric Array lengths preserve missing slots, reject invalid lengths with RangeError and release sparse storage. The flow graph includes the constructor's throwing edge. ArrayViewRead, viewArrayElementType, graphArray, viewArrayReadOwner and heldIn are defined for V4's array dependencies. This does not turn on V2's union-arm reads. Rebase dispatch sends ordinary arrays to V3's hook before V2's placeholder. Untagged union membership containing an array still returns false from supportsUntaggedRead, because both V2 runtime matchers still reject array contracts. V2 must update that boundary when installing its array matcher; TestArrayMembershipAwaitingUnionMatcher pins it.

## Source changes and ruled resolutions

The arrays/parser net is 4dbf3b6a..7cd02813; the B net is 70522aa1..7676bf57. Each net was applied once, then its array-only shared-file changes were resolved once. The payload reader and type-helper work includes 9fc5c01 and 3b1263b6. Original lane commits are listed below and in the two original views-v3 commit messages. The API changes are separate fixes sourced from af0a0ae7 and 41603d77. No cohere source was copied. No protected compiler file was edited.

The representation amendment preserves named representations and typed callback layouts. Merge judgments 19 and 21 preserve target evidence, source freshness and cycle refusal; judgment 23 preserves boxed union array storage while checking each selected element. Callables and their contracts remain V4's. Dictionaries and intersections remain V5's. Runtime graph admission 2724dabd and fresh_refused/graph_regions are excluded.

Of the 494 originalB witnesses, 473 require tsc's checker profile and are renamed byte-for-byte to .ts. The 21 actual Adamic fixtures remain .a. Every manifest, preparer and test reference is updated. The witness-extensions receipt intentionally retains historical from paths to attest byte identity; active paths use .ts. All 473 Node goldens run independently of Adamic's checker profile. These witnesses are not claimed as 473 native admissions.

## Comparison with the same-machine base

The full lower/IR/flow/JavaScript/oracle sweep was run on both 2b6c032a and the repair. The base has exactly the original 21 array subtest failures and their parent. The repair turns these green. Its first sweep found the numeric Array constructor's missing throw edge and a counts-table mismatch while new rows were being refreshed; the edge was fixed and the complete flow suite was rerun green, with a separate omission mutant. Counts were refreshed and then checked without update on the final fixture set; both commands passed. The final lower, IR and JavaScript package sweeps passed. The final aCheck examined 26 files and reported zero failures. The old counts rows are preserved; only V3's 26 fixture rows are added. The mixed union contract descriptor's readonly array arm moves from skipped to passing; union read dispatch itself is unchanged.

Stage 1's named Gap/Gaps/Probes results are identical: 112 pass and 6 skip on both trees, with no moved gap and no GAPS.md change. Stage 3 fixtures pass after npm ci in stage3/api; they also pass with the platform guard disabled in a local Go overlay. That overlay is not committed. The gate's aCheck is run over every lane2 .a plus the three hole fixtures. Build and vet run before push. Tests write to log files. On the rebased 4fb447d5 comparison, TestNodeFSFileScratchOptionsBorrow is already failing on the base at internal/lower/library_node_fs_file_test.go:93 (mkdtempSync options with evaluated expressions). Compiler supplied the current complete train-4 list: exactly the 21 TestCheckedViewArrays cases. All 21 pass here. The older 12-failure list was from unfixed V1 and is stale. Gate logs are origin branches.

## Checks and omission mutants

Runtime controls use Node, sanitized C, release C and generated JavaScript. Numeric Array fixtures also check leaks. TestCheckedViewArrayReadMutants removes checks for array-second, array-boolean, array-string-literal, array-undefined, array-iteration-bad, array-map-bad, non-array and array-missing; each pinned diagnostic catches the release C and JavaScript mutant.

TestArrayElementKindRuntimeMutant removes element-kind certification and is caught by the wrong-type array-boolean fixture. TestArrayHolesAbsentSlotMutant invents present zeros and is caught by Node's 4:true:true observation. TestArrayHolesRangeErrorMutant drops the invalid-length throw and is caught by Node's RangeError observations. A flow overlay drops the ArrayHoles throw edge; TestEveryPathNodeTakesIsInTheGraph fails on library_array_holes_range.a.

TestArrayViewMissingReceiverMutant drops the receiver check; nullable-array-missing catches it in both backends. TestArrayViewPhysicalWriteMutant drops the physical-layout guard; number-to-boolean and heap-pointer-to-number writes catch it in both backends. TestArrayViewHolesAbsenceMutant drops absent-slot handling; array-holes-view catches it against Node. TestViewArraySourceCertificateOmission removes the lowering guard; the required refusal catches it. Source-level push and indexed-set controls also require that refusal. A Go overlay replaces the pending array membership boundary with element-only acceptance; TestArrayMembershipAwaitingUnionMatcher fails and catches the omission.

Earlier standalone mutants remove array-kind, selected-element, field-presence, field-initialization, child-error propagation, tuple refusal and callback-once checks. TestViewArraysNode, TestViewArrayFieldPresence and TestViewArrayContract catch all seven.

## Runtime clearance outside view_* files

- 04420be9: adamic.h includes view_arrays.h and temporarily adds private array view metadata; array.c initializes and copies that metadata. The next API commit replaces the private metadata with element_kind.
- ef762440: adamic.h exposes uint8_t element_kind after references, keeping heap first; array.c initializes element_kind and preserves it in slices.
- 3de13ddb: adamic.h adds sparse storage and array_holes/is_range_error/holes_at/holes_set declarations. array.c initializes sparse to NULL. heap.c releases sparse storage instead of traversing dense slots for sparse arrays. The new array_holes.c implements numeric-length validation, RangeError construction, constant-space sparse allocation, absent-slot lookup and bounded slot writes.

All other runtime changes are in view_* files. These are implementation receipts for runtime clearance, not a claim of that clearance.

## Still refused and language questions

This is not a whole V3 completion. Tuples-as-arrays, array subclasses with own fields and the ranked native witness certifications remain incomplete. Taking the full lane tree would also import other slices. e3cb66a1's tuple source-storage work depends on Narrow.Tuple and tuple slot/union dispatch; f1f6355a and 0a6edef2 mix tuple alternatives with referenced Map payloads. cf1765d1 and 3dd953f7 mix tuple arity/rest positions with shared constructor and Map contracts. Their full commits cannot be imported without those dependencies. The 3b1263b6 own-field readers require an object real_type representation and ScalarWriteContracts absent on this base. Callable contracts remain excluded. No unsupported family is silently admitted.

A writable asserted array needs its original logical element certificate, even when its current contents happen to match. For example, writing 2 through a number[] view of a 1[] source succeeds under Node's erased assertion but breaks Adamic's source promise. This repair conservatively refuses such writes with the source-slot certificate diagnostic; readonly views can still observe writes through a correctly typed source alias. Proposed rule: check the source's declared element contract at writes, separately from the physical-layout guard.

For tuples, Node uses the source array's actual length rather than the asserted tuple arity. Proposed rule: preserve dynamic length, check only selected positions and retain the source storage/ownership certificate. Array-owned-field reads need the predecessor's object identity representation before their callbacks can safely recover the containing array. These are open dependencies, not new admissions.

## Logs and commands

Logs are under /tmp/views-v3-*. The same-machine package sweeps are views-v3-api-final-base-internal.jsonl and views-v3-api-final-internal.jsonl; the repaired full flow sweep is views-v3-final-flow-sweep.jsonl. Stage 1 comparisons are views-v3-api-final-base-gaps.jsonl and views-v3-api-final-gaps.jsonl. Stage 3 logs are views-v3-api-stage3.log and views-v3-api-final-stage3-linux.log.

Final focused commands: go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts; go test ./internal/oracle -run 'TestCountsAreRecorded|TestCheckedViewArrays|TestArray.*Mutant|TestOriginalArrayWitnessNode'; go test ./internal/oracle -run 'TestCheckedViewArrayReadMutants|TestArrayHoles'; go test ./internal/lower -run 'TestViewArrayWritesNeedSourceCertificate|TestViewArraySourceCertificateOmission'; go test ./internal/lower ./internal/javascript -run 'TestViewArrayContract|TestViewArraysNode|TestViewArrayFieldPresence'; the fast-gate aCheck harness; go build ./...; go vet ./internal/.... The unit explicitly requested the full five-package and stage 3 sweeps; the whole fast gate was not run.

Toolchain setup reported node 0.024s, Go 0.026s, Markdown 0.079s, clang 0.235s, submodules 159.015s, Go build 345.334s and deferred completion 345.532s. nproc reported 5. These are observed setup timings, not forecasts.

## Original lane commits


arrays-parser

- 7127080756a1904cc8d65ed766b66bbfaa767771 Lower closed numeric and string enums with Node oracles
- fd2ef77e166c06ca8dca6757c65cc57ba3087258 Add process exit status and terminal and environment observations
- 7ccbcd17e97d0bd74af28149aee8f23539d44b8f Add realpath and match Go directory-link enumeration
- 28d1db9983f2f0447399f40c3413fe4b94cccdca Admit flag enum combinations with proven domains
- e816797f07a70952e214be079a1da54db80fbd7c Check enum and nominal class downcasts before narrowing
- dd4b67e7bdedd27f31d2e3900e7caf1fd5575374 Drain JavaScript process exit output and pin the Node exception
- c63a961be1f617f0ed9b3bffc0a0b2cfaa051821 Cache regex UTF-16 input views with the string lifetime
- 75b7492f9f03865650f79f9204f76a624aafe2db Propagate UTF-16 counts and the ASCII fact while building strings
- f1fbc74d5b3cb1ffcc27331de340a65f752e3ea0 Read cached UTF-16 units directly in supplementary strings
- 0d3d76ccbc1fa1738b7e5d2cf66d9b0f319c984b Copy string boundaries that cannot join surrogate halves
- ed6bad6abd6332be24b578c9246336a1d73fefe4 Inline proven data and optional field reads and preserve inherited statics
- eef83ac19215884ab83872f96510d17a2c03d44a Add Node directory and POSIX path host members
- 3c6818554dde639879bb3dfa4a0fa13f94e7ad59 Use guarded inline slots for field writes
- ad56ba382e1e3124807758766c6aa852bf6e9c49 Extend synchronous process and performance host support
- 3b607c39ac08e974a5d8781a3b28406b95591eee Store small Maps inline and preserve live iteration through promotion
- b15216dabf65ffaa7152f6e64709b7b062ea01a9 Lower named nested functions through one counted frame environment
- f5702191f4aff7028464df3ab2b334389b7ca496 Use shared Node declarations and refuse unsupported host members
- 7e2bc734ce77dfe432ce593025fdc066a5ef973c Lower pure string records through counted dictionary operations
- 5f77d3315f3a6472ee2a6cc28e23e2f7cb1eb8ff Lower switch fallthrough and implicit returns before relaxing checker options.
- 409738474db5f9d5611bc5f320683f43832a663b Preserve flag aliases and inline iteration proofs and enumerate enum names
- 7f6d7d2ea23e1582107b62ef55a41c290b70dc3a Make WASI stack checks, sort throws and stream aliases agree with Node
- e4b14d9f1eebfa55a100c2665dc88a6793b98a5f Lower named array rest parameters, census nonplain 22 to 1
- 9b8057003a3bb72b067da9dfcd9ab84380251bcb Distinguish Node member owners and qualified types
- e8c79a94fb16524777f45171c18919e3d95c4a70 Keep three boolean field values, census fields 37 to 0
- aacac5791cb5ec26f50270efc7b37dcd1223744d Implement census Buffer encodings and SHA-256 host calls
- cfc23f7377da4320b8365e5356891a1f54e5715f Use shared Node declarations and refuse unsupported host members
- ab8e11fa76b9f9369f3a4155b6427eecfce62df9 Build synchronous fs file host and System wrappers
- 2ca5e36a5c7757a572fb20f9490401150c5c370f Distinguish Node member owners and qualified types
- 34e4aaa50724889f5176cccca719a05dabfcb702 Handle fs file operations in the freshness proof
- cd10dbd8e1bfe92a7dec252aec233e6e99740397 Complete Buffer fs reads and writes and record real host blockers
- ef1435e5c724e2bd9c090a8309ef66ed73bc59ed Check overloads strictly and lower implementations, census bodyless 194 to 0
- f5a6917029567f4e7621d404052ad985e716ed4b Merge branch 'codex/loop-array-hold'
- bc9f5d7b8d39cd26d359bead2b363ff53ac6a697 Admit scanner var definite declarations with captured readiness checks
- 5b682673b9d29cc378d3d3e4763e2604296f7a12 Fit resolved overload results and re-green after main integration
- 287a5ec543b10b51d7944b3855815f0fa8d96445 Open numeric enums and check impossible narrowing paths
- 2724dabd9f22d2e659968c56452e6b777073f3df Carry unproven cyclic ownership through graph regions
- 45664902869d263a84b4e79318b4bf4667dbc257 Restore switch empty cases
- 9f10b92b059dba1f5c777f7d74dfd72591b98346 Restore narrowed union member checks
- 4045e6428ddd06cc343267c776a8ee7c110e8563 Distinguish Node member owners and qualified types
- c8abaeb039ae3ba4442c600983388c579818c937 Implement census Buffer encodings and SHA-256 host calls
- c027ae7150fbeeb75d83b1ccb501ac04e7be6afe Use shared Node declarations and refuse unsupported host members
- 145c3cbbcbcfc26c043b1408947c754e5b867cce Build synchronous fs file host and System wrappers
- 999a58c8825f4d66181d7ab4bae02e2896bfd099 Handle fs file operations in the freshness proof
- a02613eff851e07f567ab9934adfc8a2dc98eeca Deinitialize slots on literal assertion assignments
- 26485b1142a221f6018a5746208508a7c04b6b73 Complete Buffer fs reads and writes and record real host blockers
- d06ed474e90a4fc5270f242dea8b1ec8d960fb54 Support pinned Node performance objects and tsc hook discovery
- 0dbd25cf6851647518368a47c4b4027ffbdd3cb0 Record scratch landing oracles and refusal proof
- f69bf6082b6db19ec0037ef8d2a51a9a5d46a8d8 Lower optional and default parameters on function values
- 157f53e0a6928635c1c43dcdb4efe368c09ca188 Add POSIX tmpdir and own code-bearing process errors
- e812d5ae10790d8694504207f74d8f95bc21a332 Lower declared functions in object shorthand
- 320b762dfebed000ca07288e51eb6d78cd29b5f7 Lower boolean or undefined conditions and operators
- cdddfee346e70a5cba3e982f808b9ebe426a6dfb Lower unary findings 85 to 0 and overload signatures 191 to 0
- 2f5277e74b548bb81018dbbd603a303a8f4b18bb Accept initialized cycle reads and check undecided reads at runtime
- 779ff9d337ed3e1f3b27d71dd3152dcb1ca8614b Classify original cycle ledger and pin initialized and early reads
- 6d61c52d9ff5f0ae8bcbabf5bd1c1fa7bb1ff892 Classify graph allocations through value flow before publication
- 8280fd0873a3d5c425f8a1a342b44af20cf883d9 Guard the as-const check against qualified type names
- 6439f4cf6a7f2d1d9de86bfc8f941d6f2d9087b3 Preserve validated performance projection through main relation checks
- d2d3c77e308cc26d4af196dedeba92d7d38c4cda Add phantom brands: value refusals 66 to 45, __String refusals 35 to 0
- 191d2d64ba1bb068feb4ef938f6926fda6f057d4 Store function values in erased never-rest marker slots
- 33a90f4b0763d797d33ed41eb9db79ef12ff6609 Prove nullable overload results or check them at the resolved call
- ae694d1900c36f19c37a1d00d44deda9addabafc Trap evaluated never values before they can reach a write
- 41603d77d89c99b9634766d8ce01b0c13d727843 Lower numeric Array lengths and preserve absent indexed slots
- f7e1008ad3ad70213ff9c5fc21686e230721685b Preserve holes in callbacks searches joins and length changes
- 41f14055ac334d0f667b06a88c3018090723ec8d Reconcile graph allocations with checked view readiness metadata
- 929448776a7c32a893bb39b6b26e7b3f9ab4d792 Preserve JavaScript readiness declarations after graph merge
- 3985ee2ca1637e46ea6989459712f5a236902351 Keep nested forward declarations distinct from readiness slots
- 8346b28c11f5cad7b1586efab5e891bb6b56ddaf Add standalone array and callable view hooks with semantic mutants
- 2a16ce35bd801338299fd1c039edb80591ea2c68 Record current main validation and callable kind mutant evidence
- 8576456768b498eeb1e5be94f8801903c2ba658e Refuse nullable Array element representations explicitly
- df7bc5da3155fe7811f758d67aed0c7e97cdf0d8 Keep object field reads checked through nested aliases
- 609ed3955d9223beaf59b1e0de97bb894d50fbf4 Check tagged object union membership before returning a viewed child
- 59db5d63a4caebcfeca67322712691d2f1ae85c2 Expose shared allocation flow and erase proven scalar record checks
- f8ed279d14169a05df03808f495e63b5f24d3768 Check widened optional reads through shared transitive views
- bde00304d8f8631aee4814cfb0f3f6c752938e67 Allow predicate values in never-rest markers, probe refusal 1 to 0
- 198ff780b594869165aa6d98b6d6f9ac1a410515 Lower detached own-property calls on records and objects
- 9fc5c01ab850d92b24b66d2a6782f2a820bdb929 Integrate arrays and callable members through checked views
- 0b141c26788b2bc0114b040f60e35f9cdf74bb9f Report integrated view evidence and sameMap site result
- 3d48edbb701282a175f59c7e6956a1d384412c38 Check predicate overload directions consumed by narrowed reads
- e3d5836cc1da8dddab96a1503f24d152c1e8f3be Add void array phantom brands, adapted probe refusals 1 to 0
- 46a282951bae034fa7a4f6daf9d919e14351a364 Check widened optional writes against real slot contracts
- af0a0ae70f58c5e63ef7cab7d6eb7ff89bbb4518 Measure view blockers and expose array integration prerequisites
- 460beecade437ac6277ae01489458545cd2edb6c Report predicate call-site directions in driver and oracle counts
- 31c1305fdab62fb31bbbb5cb80d48ad57c909ed3 Defer checked views and phantom casts to their proven lowerers
- 4ed5e301d5a120139e2984b1449647771099f198 Verify merged parser array and phantom dispatch against Node
- 6d0d71a427860c57fad440d49880b61dd7aab5ce Verify checked optional writes on current main and measure stage 3
- 04b3f073992d6712bfb850677def5d47673fbaea Fix three census panics and preserve named proof refusals
- c898009bfb8520b03c5a366ac23c308e1e56c882 Check nullable array returns and readonly empty conditionals
- d90994daa68bdba300f22acdad849a22ad5192a2 Accept phantom overload results; fixture refusals 1 to 0, census brand diagnostics 45 to 45
- 2de6d1ba1262a4eb9cb9e9bdd6eac5b36fb8ef5b Measure checked view demand at actual compiler field reads
- 42625435fa4bc120dae04d34d707bfc38eb6baf0 Use shared storage contracts for native array reads and mutation
- bf544d5c9fc9181fc966b18d9006943059fc5bac Record native array oracle and mutant evidence
- 21f4c0ea466763aa0767885f1148a9e29eec4370 Check primitive phantom brand fields through their approved scalar base
- de8b2ec763fb781bd35e42d21e30238d0b18af07 Admit unread view members and refuse demanded callable reads using shared flow
- f7100da75d253a0ff33b00dc58e4d399b640f0ab Retain demanded view refusals through wider helpers and array consumers
- 12482f1ff9cb45c7984471ee3ce99c0ea554d440 Check array searches lazily and require source contracts for reference writes
- 25ed1d3f7a1f27e9a471522c856647dad475bb8b Record integration conflict boundary without resolving shared hunks
- 5ca21a57b56cee7540f10a8e64f920a675982c32 Keep optional array views lazy through descriptor and backend dispatch
- 2b767776278b36f8644257b443fbaabe164d2796 Check complete branded string fields lazily at required reads
- dc5e5f26f0153896cae730d84c34a42054b02b0a Check structural object and primitive view reads with lazy descendant obligations
- 1baa2640053571d369cde31fa4bf750d92d8806c Wire named untagged object union hooks at demanded reads
- 47c6c23a4a5f9223c1b5b93f5dc719e5c19774c6 Check structural intersection contracts in production reads
- c9ad59e4dd76497571be6065bd6d449d757324b2 Wire checked dictionary source reads and preserve transitive guards
- 3deae80936daee800f06db56eb17525deb3815bc Check nullable view fields with distinct null and undefined identities
- f4e5993e3ed2817a9060ed53ca0b957ea5588c64 Check lazy array union members and preserve boxed field producers
- 36de78996acaa3843c90990391f021344783bfb2 Preserve callable proofs for nullable views and name collection read gaps
- 456c981b1c108abae5f2c8d6ae665cfd6b92e2fe Use dictionary storage for finite partial record views
- 60d50afce8aa6e57dca9aabcc6596d9f4b4c653b Refuse unimplemented intersection combinations at demanded reads
- 37fa06d393b73779bef134229a50370c168b1943 Check fixed scalar callable views at each read
- 0f722679dd7ddd16933c86d88880076cfc7a8418 Prove generic non-null assertions through direct failure helpers
- 3b1263b650b374183da1cd879b29a73dfb10135e Support NodeArray views and certified scalar own-field writes
- c809598e4443d2a20f4a8b2e9b260ce6069b4b5e Select nullable union members before checked payload reads
- 37161988d3eaaf5f18b6120c9c37fef35b871b2e Report ordinary predicate directions from admitted proofs and checked reads
- 1191e3f74b1e60a62aa1eb947b0bc45136522ffc Pin mixed union callback refusals after finding an unsafe result view
- 9c8c720e32ee3cb798951bd7c3ef4a646d9af128 Check finite tagged intersection arms and verify original witnesses
- 9ba1f722976d6d150bb97f1f0fc3e5bc80daff03 Hold the next five ranked array contracts to both backends
- f7a784c4e094e0c97dae9dc8309a900e4cd3d666 Check Map view reads against constructor key and value certificates
- 0af43ea7213a80b80193d885e25eb60c2ebb0cf3 Add primitive view probes and correct open key candidate counts
- 2b3a5ceea209e578cc4ad79ad03ef6fa8790f4c0 Check recursive object intersection payloads through nullable dispatch
- b6e53fb46bc623c7017ba8d86b9e995ff4a1d912 Check stored erased marker calls and release discarded results
- db7720971c2533ab375c4b8fe5a611c720c583be Certify Map array entries with readonly variance and recursive schemas
- 99089aa11253bd879e9b5ecad8a65c870eee8fc7 Preserve NodeArray own-field and optional entry evidence in Map views
- 8a660bcd50fa547e64dc148c13c3af334d1e834a Check required dictionary fields through optional receivers
- 4fb1c603e4494a0fad49858f56a835469eaa46f8 Check broad and overlapping object kinds for leading candidate reads
- 45b334d08dbe6fa6355d054f1f82b95ace40d847 Check erased dictionary sources and generic lookup descriptors
- e540f4d7de11090fd82053f1671ed2ab5df2063f Check original contracts for scalar record array writes
- 69c398621c76738e9482b087c94375508997b0a3 Hold ranked parser array contracts to both backends
- 44f1108b78c88d40d1b8af7c5ed29d255feca7af Store mixed and null-containing Map entries as owned boxed unions
- 9e4e13fb1bb9c755c1743eb0937906eeea988de8 Certify original tracker intersection field with complete declarations
- deec3c933543c7b5e7ba0d9db0964265d03c494c Check finite dictionary keys and optional named fields
- e3be8faf54cfdb308b52838a56e028367f4548ae Check recursive field-only and homogeneous array union projections
- 85e958e2deed551a539a665a35fd492a02a7f4ea Prove callable Map entry producers before certifying storage
- f201b579b587663f707e10612edf91c3a20ee395 Check readonly array union elements lazily
- edc2e0c080b3e5ec7fb934c948d302713404f7d1 Certify fixed tuple Map entries and check helper position reads
- f426929a0cf100ab74d6dde9f78b62a852008055 Check recursive flow projections and fixed callable unions at reads
- cb6adb25e01ef77c308c4bb9fc4a59051620a02b Adapt callable scalar and boxed union boundaries with directional checks
- 73c778ae6b3cbfaf08b7f4d20c054d19366da1b7 Hold the fourth ranked array contracts and fix graph adoption  size
- 8de074ccbce2c7470fe46980ff23c30d82d0c373 Reuse callable producer contracts in nested union projections
- 5c609c0109e7dcad3d9169f866d17f20cfbb3fd3 Carry the fourth ranked array contracts' 33 fixtures
- cd32db4234fb0a65a6ad199474b802c9e7491054 Preserve recursive array contracts and callable mismatch diagnostics
- c94b2a7be9eea1e963237b70480d78aa029ca600 Hold the fifth ranked array contracts to Node in both backends
- 65d8a138927fe240c4ede0854e9bd00084fd0f16 Convert readonly Map reads without changing producer storage
- b23327017e4a4a32447c9a07fd37245c7a531ea4 Hold the sixth ranked array contracts and name undefined elements alike
- 5a7f9db78001ddf73d1f749ad85f5360b9eb3e82 Hold the seventh ranked array contracts to Node in both backends
- dd198e0d1337ed5740fa2c61bc61137f70b097a5 Hold the eighth ranked array contracts to Node in both backends
- 88647abed401797689f7760a69531eff1b056b17 Hold the ninth ranked array contracts to Node in both backends
- 257568b0eb90669bd0585f04e81f8b1e11039760 Keep the combined lane 2 array oracle log at the ninth group's tip
- 60d9bf9f250f049ba91b887975140794f51d0b52 Hold the tenth ranked array contracts to Node in both backends
- bd05075f5b8cbd592e8ece1a098d582eb43cf510 Revert rejected owner merge and retain failure evidence
- adc4dd1c817aec017fa58161e03e254deda09717 Convert readonly numeric Map keys and nested array reads
- f753567dd8837e90914e7c229df88d8536a3bf8b Check original EmitSignature tuple reads using the shared certificates
- dc925ee2b88a6e190e724552be39dce49419c6be Hold the eleventh ranked array contracts to Node in both backends
- 6b3e88a07f42c4388c5c2f690474290c7d89343f Hold the twelfth ranked array contracts to Node in both backends
- 924eccb0b393f3d93a787a9e988029c8fd0abf33 Hold the thirteenth ranked array contracts to Node in both backends
- 20f64b3fad7681fcad613392cfa060682d6b3d77 Hold the fourteenth ranked array contracts and pin the text name fallback
- 1ff63ff942647d7eafe3ba2d4bebb71da7ba0e8e Keep the combined lane 2 array oracle log at the fourteenth group's tip
- 3e1a7f6fb0bd002fb87c7f97a1842b0f0d90a6d9 Certify six aggregate callable members and transitive payload reads
- 0ea22091e93375f5f03e3b34b76edf7b96891898 Verify nominal data-class Map producers with existing class identities
- d5738b8e863f47ff798ea88fed8ae53901b5d84d Admit broad object casts to uniquely tagged union targets
- ebb1f8a61aa5bc9dc9b1a37c063873ffafbc7889 Certify four mixed aggregate callable members and optional numeric arguments
- ca475993e3e6f0c89021da191a8aab027fc9daf7 Keep checked optional boolean stores coherent with alias reads
- d55340c06a13ad2076a4288a3ccf3a16fb462793 Hold group fifteen arrays against complete original declarations
- 19b40cd487e497a007bf6467a6daa038bc96492c Keep group fifteen validation and baseline failure evidence
- d9914ec2d92a2bd031b852d9ae1a93a6bd447228 Check nested primitive boxed array reads with temporary owners
- 6ed3ced09f64164e724763939de2d0f3ff958884 Scope lazy fallback obligations to actual viewed allocations
- 9ce75693f0b60dc76d3b5eb8600cc146e70e9997 Hold group sixteen arrays against complete original declarations
- d732e93978394206fbbe55fc0057b6fecc6d0cf6 Hold union defect fixes with witnesses and name tuple obligations
- 04a0c107b50ef2a47b4839255603f7af37a8cfab Decode packed boolean array reads and retain undefined presence
- 1ebb3c58746a1d515cc560221421a8db70f1b6d6 Hold group seventeen JSDoc and generic array contracts
- 94b200ef805d5fab4d8990133d22b42a8ec22e5b Check primitive union destructuring reads and retain boxing conversion
- a3b0e3570fdf83fa3071b7f34ff9780d5334d5f7 Hold group eighteen original diagnostic and string array contracts
- a6997ef44211f61b8a95b860a957ffa52c791e2f Convert readonly primitive Map key domains without changing source storage
- 4a5bc26befb0a00451a7a18ee0458ca5d636b9c7 Certify original Bindable intersection reads with a bounded walk
- f11e945d8af54fbcdd41e37df8f23e47c9c81005 Select complete primitive unions at property reads
- 2f6ed528661d936308646cbb6f4a2787335b5166 Check nested immutable nominal Map entries and escaped class reads
- f1f6355adf5cfb4a48e766b9d575c0c7386dfe5f Certify original Root indices with explicit tuple union dispatch
- b26025fefdef283bb41667bce8477abaeb8a6df9 Check deferred intersection members at every read path
- 3c143841114922f672fad316938fa7ac383b9bd5 Certify original emit signature tuple positions
- 799433e037670cf854bdf9fa6a225c1dbf6f5f2a Preserve optional tuple source arity and prepare shared contract review
- 503cd3f950d76e31d3f476617fb030b2fce4fadd Certify synthetic heritage intersection field reads
- 64cc98b8af19c74667b988ac9b359f9350218514 Transfer flagged tuple forEach snapshots through callback cleanup
- f6c07dea689e0ff5267d603bc3e615e0f8695f66 Check nominal array producers reads and alias writes
- 8b37f903a7621a964bab2eb7e9b2a15813d4563d Add shared tuple arity plan validators before source admission
- cf1765d1172bafe7ca65b58b01aecaa9ffa005e0 Certify optional tuple arity through the shared view constructor
- ae7470da61fc58f6d72901b364bdfb8799518444 Certify mutable nominal field writes and recheck escaped reads
- 3dd953f759c912177f197ff1499e5515d34bac53 Certify trailing rest tuples with complete positional and Map contracts
- 21fd3587d6a32fef3f195a31be84605c02f58953 Certify original watcher event tuple alternatives through one read path
- dca0b478e236c9ce0855e001191f55a14cb28941 Check primitive union array indexes without changing producer storage
- c41eaf85f7417bf3d33532fce30457d70be1440e Select primitive and nullish dictionary arms without erasing reference obligations
- bfd61815136f461e9540787eed51f4f22f403d8e Check boxed and packed JSON array extraction at each read
- 1dbeeff770c9f25fef0eb4b4c425e1025d3a8e7e Check nullable class array elements and preserve source write contracts
- bbbb2f4694e303534d168ddc7cb1bbb0069356d5 Certify original function and class union array reads
- e659d860ad63837e6c0dee781e2ebfaab2a79059 Record union array certification evidence
- 48d166cf7ba1954de46ca2a4b30d72f6872153a9 Check dictionary value and entry snapshots lazily
- 1c410de9fce6df90644e2f87e3774d594cf0ba7f Move checked view non-null assertion fixtures to TypeScript
- 6513b728f0f5085d0f983e12df6e7d3ef8c4057f Record array recheck after integration merge
- d03c7e1576f8d62a287469f7055f7cf585a8eb1b Certify original argument modifier and type union arrays
- 5bb11e735bef9241892f864e7e65c832bfd162cc Certify original heritage parameter JSX and statement arrays
- 5a51f54ae7171ec3f69e91a483d1af50efcd5b02 Certify original signature parameter array union
- 44d64b64b5532918654461dac1a2748051ee35ec Admit flat scalar and object arrays with checked element selection
- b8c9c1554b500bbfc60c000b23038ebf7f95261c Certify original mixed array storage and consumer joints
- e3cb66a17d7590e92df1afd5d425a25535b7010a Preserve source array storage through checked tuple reads and writes
- f4563b82f3e2b350281c1688955b66bd0f3f3a52 Adapt shared view and length hooks for array backed tuples
- e24b37f39705350ddf5a3fe1486c94a6fbd968f8 Certify original array to tuple reads with alias and stopping witnesses
- 1fbb892e422d6487a8f6649c262818ece8d4d872 Admit shared array fields through complete untagged object unions
- cddc09c1e6538503e82189747270573ccd21efb6 Connect shared array union admission to lazy view checks
- 70522aa1d73a80e295df608dee44b38414301c97 Certify original build info file names through checked union views
- 4712e29081aaf9f0a19caa3fcd92293fec89b861 Check callable array reads and original storage signatures
- 908cafeb42ad01dfcfab2250dbb0454de1a791a7 Certify original watcher callback arrays and source writes
- d0ddb5a7e1ad46f0ed3da85903266e562f1c41d9 Record finished array joints and the next intersection boundary
- 8303bd50e0a0103103afc237d9dfcb309a582b93 Certify original mapper and flow array reads
- 5dcbad439f5246ae43ceafc644e29858d5ed5052 Certify original accessor and signature array reads
- f947b7b10e307565fd21614fb8b37f5fe8da1a97 Certify original modifier and JSX type argument arrays
- 1b416b41503a6a56e06f384c825a02178ff82e75 Certify original function and JSX child arrays
- d75544c8c513e1b7453ee3aedfcbb31dfd384e1f Certify original configuration string arrays
- 7cd0281388208bd439b5a20d4f4294a55a4ac91e Certify original private directory file arrays

arrays-b

- 57fa2c98ed282dbe86afb3263ab68237670b55e4 Certify one-read array views from original declarations
- 4be16240660318473e294465cbd170fefd58aa92 Certify remaining one-read array fields from original declarations
- 4bf6da09a3c8a35c99ae76b26bb892879de0ee90 Certify mapper and two-read array fields from original declarations
- 7676bf57386e256ac4dd3b3a6e11a1827ff21889 Certify the final two-read array fields and record refused reads

## Merge with the standalone V2 source

Merge compiler/views-v2 26e74db38d1ead7530f9645f469247090857138f into b065fa5764b27478bb9b8acb1edef254731dd9a2 without rebasing either history. Preserve V2's resolutions against main, its optional-read guards, fs-option-boxing fixture, moved stage3 records and the three source-absent a-check refusal headers. Preserve V3's array adapter before the aggregate placeholder, the owning array element-type helper, the Compiles record for assertions/09_interface_kind.a and all 26 V3 count rows. Untagged array membership remains refused until the native and JavaScript matchers implement it; accepting its descriptor alone would be unsound.

Conflict resolutions apply to views-v2-dispatch-review.md, view_lazy.go, view_unions_callable_members.go, view_unions_untagged.go, view_unions_mixed_test.go, the four checked_views_v2 test files, counts.md, assertions/status.json and the three source-absent fixtures. V2's new tests and census pins are kept. The duplicate array element helper is removed from the callable file because its identical implementation is owned by V3. No protected emitter or orchestration file needs manual conflict resolution.

Run build, vet, stage3, V3 Node/native/JavaScript fixtures and omission mutants, lower array-boundary tests and regenerated counts on the merged tree. The fast gate's a-check implementation checks exactly the 90 existing .a paths added or changed against fetched origin/main: 87 checked programs and three deliberately refused programs with V2's exact headers, zero failures. Command output is recorded in /tmp/views-v3-v2-{build,vet,audit,stage3,owned,lower,counts}.log. The top merge commit carries Train-slice: views-v3. The c2 compatibility source receives this merge afterward and keeps its runtime audit and dynamic-spread initialization tests.

The regenerated count comparison against V2 has 26 added V3 rows, 0 changed existing V2 row values, and 0 changed V3 row values. views-v3-v2-counts-comparison.json names every added and moved row. New rows are the existing V3 array fixtures and numeric hole constructor witnesses; no additional source admissions are introduced by this merge.

## Hole constructor and RangeError identity analyses

ArrayHoles evaluates its scalar length, including nested writes, and returns fresh storage with no present reference elements. It is the length-form constructor, not a hole reader. Creating it stores no user reference; subsequent writes still require the ordinary cycle proof. A back edge through an object stored in this array remains unproven. ArrayRangeErrorIs evaluates and borrows its operand to test error identity, returns a primitive boolean, and neither stores nor escapes that operand. Nested writes in either operand remain recorded and judged.

Fresh classifies both nodes explicitly. Flow assigns the constructor a fresh result and the identity test a primitive result, preserving operand mutations. Existing generic SSA/read traversal and the ArrayHoles throw edge already cover the nodes. Direct tests hold those facts, including a cycle-closing operand write and constructor failure nested inside identity inspection.

Node rejects the old trace snapshot of new Array(4294967295) with RangeError: Invalid string length because joining absent slots expands the holes. The observer now records length and present own keys. Its Node control distinguishes an absent slot from a present undefined value, records the highest valid index, and observes shrinking length.

Omission mutants remove each fresh node classification, each fresh operand traversal, the flow constructor classification, each flow operand traversal, the constructor throw edge, snapshot length and present undefined slots. All ten must fail their focused regression tests. Private overlays and their logs are under /tmp/views-v3-fresh-mutants; the checkout keeps all checks. Full package checks use go test -timeout 30m ./internal/fresh ./internal/flow, with outputs in /tmp/views-v3-fresh-flow-final.log and /tmp/views-v3-c2-fresh-flow-final.log. No fixture or census row is added or moved.
