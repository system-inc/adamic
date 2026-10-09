import pathlib,json,gzip,subprocess,shutil
r=pathlib.Path('/tmp/defend-lower-taste'); out=pathlib.Path('review/test-defend/internal-lower-taste_representation');out.mkdir(parents=True,exist_ok=True)
m=json.loads((r/'matrix.json').read_text());assert len(m)==3
pairs=[('TestViewObjectWritesNeedSourceCertificate','TestDefaultTaggedInterfaceAdmission'),('TestMixedUnionContractGraph','TestMixedUnionContractRecursiveMember'),('TestUntaggedViewStructuralFallback','TestUntaggedViewMemberTags')]
rows=[]
for (test,subsumer),x in zip(pairs,m):
 assert x['rows_failed']==[test],x['rows_failed']
 ev=x['failures'][0]['Output'].strip()
 rows.append({'test':test,'package':'internal/lower','prior_verdict':'subsumed','subsumed_by':[subsumer],'defense':'defended','unique_mutant':x['mutant']+' '+x['file_line'],'attempts':[{k:x[k] for k in ['mutant','file_line','change','rows_failed']}],'evidence':x['command']+' > '+x['mutant']+'.log 2>&1; '+ev})
(out/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
for f in r.iterdir():
 if f.is_file():
  if f.suffix=='.log': (out/(f.name+'.gz')).write_bytes(gzip.compress(f.read_bytes()))
  else:shutil.copyfile(f,out/f.name)
base=subprocess.check_output(['git','rev-parse','HEAD']).decode().strip()
report=f'''All three requested rows are defended by package-unique production mutants.
Base: {base}. Scope: all 276 top-level tests, including 37 additions since the audit.
Three full-package matrices each failed exactly one requested row; each passed 273 other rows and skipped two rows.

Code under test: Adamic Lower's checked-view assignment admission (view), strictViewContract member metadata, and UntaggedViewMembers structural-member planning. Oracles: handwritten diagnostic/IR assertions; TestMixedUnionContractGraph additionally compares member names against the live Microsoft TypeScript-Go checker. No test, harness, or oracle was mutated.

D1 drops the whole assignment-refusal if statement at interface_cast.go:74, including the representation local, so no unused variable remains. The target alone covers the refusal body at :78 compared with its former subsumer. It checks object writes through direct and destructuring assignments, while the admission row's alias-write input writes a scalar string.
D2 changes the contract Name option to an empty string at view_contracts.go:59. This is a defense on shared lines: the graph row compares each member name with checker.TypeToString(member); the recursive row checks identities, object fields, and total graph size without checking member names. Coverage exclusive blocks are separately recorded; exclusivity was not required for this semantic difference.
D3 flips the rejection condition connective from OR to AND at view_unions_untagged.go:45. It admits an unsupported member kind when its Of remains Object. The structural row alone exercises the refusal body at :46 compared with the tags row. Its negative subcases cover Unknown, Array, Callable, and an invalid ID; the tags row supplies only valid object members.

The complete list of passed rows for each mutant is in matrix.json (rows_passed). The former subsumers passed all three full-package matrices. No panic, timeout, bounded narrowing, or unknown matrix result occurred. Baseline passed in 38.822 seconds. Every standalone diff compiled with go vet ./internal/lower/ while applied. Production source was restored after each run. Separate ADAMIC_BUILD_CACHE_DIR paths appear in matrix.json. No probes count as defense mutants.

Coverage: six exact-name runs with -coverpkg=./internal/lower, -coverprofile, -count=1, and -timeout 90s. Coverage differences and raw profiles are included. Run scripts preserve exact commands. Two skipped top-level rows: TestOriginalCycleLedger and TestOptionalWideningCensus, both unavailable generated census inputs. The existing Graph array-member subcase also skips pending compiler/views-v3. All newly added rows were included, listed in scope.json.

Brief issues and limits: /tmp has only 8.8 GB total capacity, so the requested 15 GB free threshold cannot be met. Earlier native-math scratch/cache was removed; /workspace had 19 GB free throughout. The audit's rows.json contains names, while results.json contains verdicts; all report pieces were read. The audit baseline is older than current main; 37 added rows were included. Matrix counts exclude two skipped rows; no claim covers those skipped rows or the skipped array subcase. No repo-wide uniqueness is claimed. All rows were defended, so there is no undefended name/assertion mismatch finding. The Graph oracle's broader limitation remains: scalar member kinds are not exhaustively checked, as the original audit notes.

Warm setup was skipped. nproc=5. npm ci in stage3/api and the clean baseline were run before mutations. Per-mutant elapsed wall seconds (vet plus package build/run): {[(x['mutant'],round(x['wall_seconds'],3)) for x in m]}. Test-binary durations are preserved in the JSON logs. No tests were deleted, rewritten, or weakened.
'''
(out/'report.md').write_text(report)
print(json.dumps(rows,indent=2))
