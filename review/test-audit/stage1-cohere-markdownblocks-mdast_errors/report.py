import pathlib,json,re,statistics,difflib,collections
root=pathlib.Path('/workspace/adamic');p=root/'review/test-audit/stage1-cohere-markdownblocks-mdast_errors';groups=json.loads((p/'groups.json').read_text());menu=json.loads((p/'menu.json').read_text());commands=json.loads((p/'matrix-commands.json').read_text())+json.loads((p/'additional-probe-commands.json').read_text());originals=json.loads((p/'originals.json').read_text());byname={n:g['test']for g in groups for n in g['members']}
def events(mid):
 out=[]
 for line in (p/(mid+'.log')).read_text().splitlines():
  try:out.append(json.loads(line))
  except:pass
 return out
matrix={};raw_matrix={};rebuilds={}
for m in menu:
 e=events(m['id']);raw_matrix[m['id']]={n:next((x['Action']for x in reversed(e)if x.get('Test')==n and x['Action']in ['fail','pass','skip']),'unknown')for n in m['members']};matrix[m['id']]=sorted(set(byname.get(n,n)for n,r in raw_matrix[m['id']].items()if r=='fail'));rebuilds[m['id']]=[x['Output'].strip()for x in e if 'build 'in x.get('Output','')]
(p/'matrix.json').write_text(json.dumps(matrix,indent=2));(p/'matrix-top-level.json').write_text(json.dumps(raw_matrix,indent=2));(p/'rebuilds.json').write_text(json.dumps(rebuilds,indent=2))
files={}
for f in (root/'stage1/cohere/markdownblocks').glob('*_test.go'):
 for i,line in enumerate(f.read_text().splitlines(),1):
  q=re.match(r'func (Test\w+)\(',line)
  if q:files[q.group(1)]=(str(f.relative_to(root)),i)
(p/'scope-locations.json').write_text(json.dumps({n:files[n]for n in byname},indent=2))
seconds={};samples={}
for g in groups:
 values=[]
 for i in range(1,4):
  text=''.join(x.get('Output','')for x in events(g['test'].replace(' family','Family')+'-'+str(i)));values.append(float(re.search(r'\bok\s+\S+\s+(\d+\.\d+)s',text).group(1)))
 seconds[g['test']]=statistics.median(values);samples[g['test']]=values
oracles={
'TestMdastMalformedEvents_Setup':('Build/readiness success of suite-owned malformed-event products; no runtime comparison in this row.','self'),
'TestMdastIdentifierWitnesses':('Actual Go cohere mdast and pinned Node fork, recorded Go/fork witness snapshots, and native/source/backend comparison with Go. Known casing disagreement is retained.', ['external-run','self']),
'TestNativeMdastConstruction':('Actual Go cohere mdast construction and pinned Node fork; full canonical tree bytes compared with native/source/backend, plus sanitizer/leak checks.','external-run'),
'TestOptionalStringInitializationWitness':('Original source on Node decides missing; handwritten missing label is also checked, then native/backend output compared with Node.',['external-run','self']),
'TestMarkdownAstPath':('Actual Go AstPath and pinned Node fork observations; full path bytes, plus handwritten key-gap values.',['external-run','self']),
'TestMarkdownParserPrefixes':('Actual Go micromark preprocessing/prefix primitives; native/source/backend exact bytes compared with Go.','external-run'),
'TestWholeDocumentOraclePreflightMutants':('Pinned Node fork output; row expects prebuilt Go mutants to disagree. W2 disables the witnessed difference check.','external-run'),
'TestMarkdownQuoteLayoutNative':('Suite-owned native-product readiness; returned products are ignored by the top-level row.','self'),
'TestMarkdownQuoteLayout_Setup':('Suite-owned quote readiness; returned products are ignored by the top-level row.','self'),
'TestSampleRetainsFixedMarkdownInputs':('Handwritten selected labels, counts and numeric-answer bytes.','self'),
'TestGeneratedLayoutSelection':('Handwritten generated/control selection and numeric index expectations.','self'),
'TestWholeDocumentOraclePreflight family':('Go cohere versus pinned Node fork, self-recorded auto gaps, union coverage and a planted off-output mismatch; W1 weakens the guarded comparison.',['external-run','self']),
'TestMarkdownQuoteLayout family':('Go cohere layout, pinned Node Markdown/doc printers and full output bytes; union coverage and planted mismatch use self expectations. W3 tests the built-in disagreement guard.',['external-run','self'])}
setup={'TestMdastMalformedEvents_Setup':'G1','TestMarkdownQuoteLayoutNative':'G2','TestMarkdownQuoteLayout_Setup':'G2','TestSampleRetainsFixedMarkdownInputs':'G3','TestGeneratedLayoutSelection':'G3'}
witness={'TestWholeDocumentOraclePreflight family':'W1','TestWholeDocumentOraclePreflightMutants':'W2'}
probe_entries={'TestMdastMalformedEvents_Setup':['P10'],'TestMarkdownQuoteLayoutNative':['P11'],'TestMarkdownQuoteLayout_Setup':['P12'],'TestSampleRetainsFixedMarkdownInputs':['P6','P7','P8'],'TestGeneratedLayoutSelection':['P6','P7','P9'],'TestMdastIdentifierWitnesses':['P1'],'TestNativeMdastConstruction':['P1'],'TestMarkdownAstPath':['P2'],'TestMarkdownParserPrefixes':['P3'],'TestMarkdownQuoteLayout family':['P4'],'TestOptionalStringInitializationWitness':['P5']}
def failure(mid,name):
 e=events(mid); failed=[n for n,r in raw_matrix[mid].items()if r=='fail'and byname.get(n,n)==name]
 for x in e:
  out=x.get('Output','').strip()
  if x.get('Test','').split('/')[0]in failed and re.search(r'_test\.go:\d+:',out)and not any(v in out for v in ['build ','corpus-files:','setup):','planted failure caught','build cache']):return out
 if mid=='G1':return next(x['Output'].strip()for x in e if 'Go errors: exit status'in x.get('Output',''))
 return next((x['Output'].strip()for x in e if x.get('Output','').startswith('panic:')),'No assertion line; inspect raw log.')
results=[]
for g in groups:
 name=g['test'];caught=[m['id']for m in menu if m['id'].startswith('M')and name in matrix[m['id']]];unique=[mid for mid in caught if len(matrix[mid])==1];sub=[];subs=None;why=None
 if name in setup:verdict='setup-check';proof=setup[name]
 elif name in witness:verdict='witness';proof=witness[name]
 elif unique:verdict='slow-worthy'if seconds[name]>60 else'sacred';proof=caught[-1]
 elif caught:
  candidates=[other['test']for other in groups if other['test']!=name and all(other['test']in matrix[mid]for mid in caught)];verdict='subsumed'if candidates else'overlapping';sub=[min(candidates,key=lambda n:seconds[n])]if candidates else sorted(set(n for mid in caught for n in matrix[mid]if n!=name));subs=seconds[sub[0]]if candidates else None;proof=caught[-1]
 else:verdict='cannot-judge';proof=None;why='No compiler initialization mutant was included in the fixed four-native-mutant menu. Mutating the optional-string source would change the Node oracle fixture; Lower was only empty-probed. This row remains unaudited for production worthiness and truth.'
 relevant=[m for m in menu if any(n in g['members']for n in m['members'])];prod=[m for m in relevant if m['id'].startswith('M')];ran=sorted(set(byname.get(n,n)for m in (prod if prod else relevant)for n in m['members']));probes=probe_entries.get(name,[]);pk=[mid for mid in probes if name in matrix[mid]];vac=None if not probes else all(name not in matrix[mid]for mid in probes);oracle,kind=oracles[name];file,line=files[g['members'][0]];fail=failure(proof,name)if proof else None;command=next((x['command']for x in commands if x['id']==proof),'')if proof else '';evidence=command+' > '+str(p/(proof+'.log'))+' 2>&1; '+fail if proof else 'Three clean timings and P5.log; no production mutation claim.'
 item=dict(test=name,package='stage1/cohere/markdownblocks',file=file+':'+str(line),seconds=seconds[name],oracle=oracle,oracle_kind=kind,kills=caught,unique_kills=unique,last_proven_fail=(proof+' '+fail)if proof else None,verdict=verdict,subsumed_by=sub,mutants_in_matrix=len(prod),probe_kills=pk,subsumer_seconds=subs,vacuous=vac,bounded=True,matrix_rows=ran,evidence=evidence,members=g['members'],timing_samples=samples[name],entry_probe_results={mid:raw_matrix[mid]for mid in probes})
 if name in setup:item['construction_kills']=[setup[name]]
 if name in witness:item['witness_kills']=[witness[name]]
 if name=='TestMarkdownQuoteLayout family':item['witness_kills']=['W3'];item['probe_not_called_by']=['TestMarkdownQuoteLayoutUnion']
 if name=='TestWholeDocumentOraclePreflight family':item['weakened_check_passes']=[n for n,r in raw_matrix['W1'].items()if r=='pass']
 if sub:item['subsumption_mutants']=len(caught)
 if why:item['cannot_judge_reason']=why
 results.append(item)
(p/'rows.json').write_text(json.dumps(results,indent=2));table='| ID | Origin file:line | Change | Failing grouped rows |\n|---|---|---|---|\n'
for m in menu:
 change=m['old']+' -> '+m['new']if m['id'].startswith('M')else m['kind'];table+='| '+m['id']+' | '+m['file']+':'+str(m['line'])+' | `'+change.replace('|','\\|').replace('\n',' ')+'` | '+', '.join(matrix[m['id']])+' |\n'
(p/'mutants.md').write_text(table)
timingwall=sum(x['wall']for x in json.loads((p/'timing-commands.json').read_text()));matrixwall=sum(x['wall']for x in commands);vetwall=sum(x['wall']for x in json.loads((p/'validation.json').read_text())+json.loads((p/'additional-probe-validation.json').read_text()))
report='''u128: all 29 supplied names exist at ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2; grouped into 13 rows.
Clean whole package cooked at 90s; bounded combined run also cooked, without assertion failures; all selected rows passed alone.
Four port mutants caught; two bounded sacred rows, one slow-worthy, two mutually subsumed on one mutant.
Five setup-check rows, two witness rows, one cannot-judge compiler-initialization row; three readiness rows are vacuous.
Warm setup skipped; nproc 5; source restored; standalone diffs and all logs saved here.

CODE UNDER TEST AND ORACLES were named in commentary before mutation. Port entries compile(), observePath(), preprocess(), printQuote(); optional row calls Lower. Setup rows check suite construction. Preflight has no Adamic product, so only its witnessed check was weakened. Go cohere and Node fork source were never mutated by this audit. Existing tests build their own built-in mutants; this audit changed only its recorded production sites or permitted comparison/construction edits.

'''+json.dumps(results,indent=2)+'\n\n'+table+'''
Survivors:
None among M1-M4. Each changed behavior and was caught after native build. W/G edits are witness/construction evidence, not production survivors. Passing P10/P11/P12 are empty-answer findings, not survivors or sacred kills.

Brief ambiguities, costs and limits:
- The brief says 13 rows and lists 29 functions. Grouping the numbered preflight and quote shards with their unions produces exactly 13. The union members assert coverage and a planted mismatch rather than invoking the same runtime checker. This follows the brief's explicit numbered-shards-plus-union family rule; readiness functions and the full-corpus preflight mutant witness remain separate. Members are listed exhaustively.
- All 29 names remain in the supplied files at this starting commit; none moved or vanished. Current inventory, not the historical 8de93800f4 file mapping, determined scope.
- Whole package 90.041s cooked during TestMarkdownListLayout_Setup, outside this unit. Bounded all-29 run 90.037s also cooked with NativeMdastConstruction and AstPath unfinished. They then passed alone in 34.953s and approximately 62s. Thirty-nine isolated timing runs all passed. No red assertion baseline was audited.
- Native port/compiler run limit of four overrides the approximate three-mutants-per-row target. Four constants were chosen from separate reached code functions before seeing kills; no test-derived string failure was used as an oracle mutation. Only four production mutations support these verdicts.
- Optional-string initialization has no separate markdown port implementation. Its source is the original Node fixture; mutating it would mutate the oracle input. A compiler initialization mutation was not included under the four-mutant menu. This is explicit incomplete coverage, marked cannot-judge, not a claim that a meaningful compiler mutant is impossible. P5 proves it rejects an empty Lower answer but cannot establish worthiness or production truth.
- Empty probes are additional to the production menu. P1-P4 were validated through native compilation and execution in their target tests; P5 returns nil IR and the already isolated optional row fails with a recovered Go panic. No later row was silently counted. Go probe early returns use an always-true branch to retain referenced imports and make standalone diffs vet-clean.
- Readiness functions are the setup rows' direct construction entries. P10/P11/P12 return an empty product set at entry, and each top-level row ignores that return. Thus they pass their own probes and are vacuous even though G1/G2 demonstrate construction failures. Port probes passing a compile-only setup row are not used to judge that row.
- Sampling checks call multiple construction entries. P6-P9 probe each selected input/numeric/answer/selection entry directly. Each relevant probe fails; no vacuous inference comes from an unprobed helper.
- Preflight compares Go cohere against Node, with no Adamic port. Production mutants would be wrong here. W1 weakens the off comparison: the union and shard 000 fail their planted guard, while shards 001-003 still pass. W2 makes the separate built-in-mutant witness fail. These are witness verdicts, not production kills or uniqueness claims.
- Quote runtime checks include both real external comparisons and built-in source mutants. M4 proves production failure. W3 independently disables the shared byte check and makes union/runtime planted-mismatch checks fail. The family keeps its production verdict; the extra witness evidence is recorded separately.
- Mdast identifier and optional rows contain Witness in their names but run actual native behavior checks, not solely planted disagreement checks. Their labels alone do not determine witness classification. Identifier also checks self-recorded snapshots and a known Go/fork gap.
- M1 is shared by Identifier and NativeMdastConstruction. Reciprocal subsumption rests on exactly one mutant, not a deletion recommendation. Identifier's measured median is much cheaper. Path's bounded unique kill plus median over 60s supports slow-worthy. Package/repository uniqueness outside each named caller set is unknown.
- The static inventory follows transitive relative imports from the actual probe entries and lists functions/methods conservatively. It is a reachability superset, not dynamic proof that every method executed. Mutation sites themselves are confirmed reached by failures. Compiler internals beyond the empty Lower entry were not exhaustively traced. This limitation prevents claiming a complete function-level coverage audit.
- Source mutations were applied sequentially as standalone diffs, then restored. Each source change invalidates content-keyed native products. Go harness edits use overlays, keeping source files and native products untouched. Construction G1/G2 use their own cache directories. P5 uses /tmp/u128/cache/P5, so the compiler probe cannot read an old native product.
- G1 deliberately changes the setup's child build output flag and catches construction failure; it does not mutate Go cohere or count a child compile failure as a production kill. G2 drops the constructed sanitized-product value; G3 drops generated-input inclusion. Their parent Go code passes vet.
- No full package mutation matrix was rerun after its clean timeout. Per-mutation row lists are explicit in matrix-commands.json and matrix-top-level.json. Results outside each list are unknown, not passes. All verdicts are bounded accordingly.
- Timing medians use three isolated -count=1 binary ok lines, including each family as a complete selected run. They measure warm product-cache behavior where the test uses caches; direct native build rows still rebuild. Outer Go compilation is reported separately.
- Node dependencies in stage3/api were installed before baseline. Selected rows use pinned fork bundles or the repository Node runner, not additional node_modules directories. No selected row skipped. The full package's width opt-in row did not reach execution before timeout; its external npm SDK was not installed or enabled in this bounded unit. Other package skips remain unknown.
- Function/source/test reads and two cooked baseline attempts consumed budget. No setup installation was needed. Go overlay failure line numbers in empty probes can shift by one; raw logs retain scratch positions. Origin locations in the mutant table and row file fields are authoritative; production/W/G edits do not shift lines.
- Every diff applies to the starting commit. All Go edits pass vet via their exact overlays; port edits compile with the tests' native build flags. rebuilds.json retains product lowering/native-build times. Direct natively builds do not emit separate build timers, so command wall is an upper bound rather than an invented native-only time.

Build/run timing:
'''+f'Setup 0s; nproc 5. npm install not separately timed. Whole baseline shell 92.311s, binary 90.041s cooked; bounded shell 92.071s, binary 90.037s cooked. Isolated diagnostic baseline shells 37.111s and 64.118s. Thirty-nine timing commands total {timingwall:.3f}s shell wall. Matrix/probe commands total {matrixwall:.3f}s shell wall. Exact-overlay Go vet checks total {vetwall:.3f}s.\n'+'''M1-M4 command walls and explicit per-product rebuild measurements are in matrix-commands.json/rebuilds.json. M4 lowered 35.73s, release native 6.79s, sanitized native 24.93s. P4 has its own rebuilds, not a stale product. Initial whole-package compile/setup and native builds overlap, so these sums are work totals, not elapsed session time.
Audit started about 12:45 UTC and evidence completed within the 30-minute stage1 budget. Final restoration and smoke logs retained.

Not covered: package rows outside the supplied set; complete dynamic function reachability; compiler initialization production mutation; all port branches; census/full external corpora; SDK-gated width oracle; repository-wide uniqueness. No PR or main push.
'''
(p/'REPORT.md').write_text(report);print([(x['test'],x['seconds'],x['verdict'],x['vacuous'])for x in results]);print('matrix wall',matrixwall,'vet wall',vetwall)
