from pathlib import Path
import json
p=Path('review/compiler/rethink-check'); rows=json.loads((p/'results.json').read_text()); lookup={r['name']:r for r in rows}
def cell(x):
 if x is None:return 'not run: compile stopped'
 return 'exit '+str(x['exit'])+'; stdout '+json.dumps(x['stdout'])+'; stderr '+json.dumps(x['stderr'])
def safe(s): return s.replace('|','&#124;').replace('`','\\`')
notes={
 'bxpmash':'Main commit 13213182: git log -S found classType = l.concrete(classType), before inner layout selection. Historical static generic specialization requirements are outside this minimal probe.',
 'drtmvb3':'Candidate attribution: main commit 4556d540 added the Optional exclusion in the shared checked-view reader (git log -S). No claim that this search proves it is the first fixing commit.',
 'vk7ed2m':'Safe NotYet stop. Workaround: construct { value: box.value } directly rather than spreading it. The diagnostic itself does not supply a fix.',
 'q9hz1bf':'Safe NotYet stop. Workaround: read source.item explicitly inside the try, then construct a data record. No accessor spread exception edge was exercised; no checkThrown mutant is claimed.',
 'dv99xzy':'Safe NotYet stop. Workaround: declare function visit(depth: number): number inside run instead of a self-capturing arrow initializer. The diagnostic itself does not supply a fix.',
 'eep1m5z':'Primary minimal tuple-union length probe. The nine original review witnesses follow below; the brief truncation was resolved using the repository evidence.'}
workarounds={
 'classfeat_maybe_setter':'Use explicit getLevel/setLevel methods, or a non-optional numeric setter.',
 'iterators_derived_symbol':'Use a non-derived iterator class with its own limit field.',
 'classfeat_static_virtual':'Initialize Derived.label before Derived.early evaluates this.m().',
 'classfeat_method_view':'Use an ordinary run method, or a callable property interface with an arrow field.',
 '93122fa_t_t5':'Use an array representation for dynamic lengths or discriminate the tuple variants before their individual fixed-length reads.',
 '93122fa_t_v1':'Test undefined explicitly before reading the known tuple length.',
 '93122fa_t_v2':'Use an array representation for dynamic lengths or discriminate the tuple variants.',
 '93122fa_t_v5':'Use an array representation for dynamic length or test the possibly missing tuple explicitly.',
 '9984394_uncaught_name':'Strict stderr comparison differs intentionally: current uncaught language exceptions exit 1, and native omits engine stack rendering (runtime/exceptions.c). This does not reproduce the former colon formatting bug.'}
counts={status:sum(r['classification']==status for r in rows) for status in ['FIXED','REFUSED','STILL WRONG']}
lines=[
 'Built: evidence-only repros for six old compiler tasks and all nine review-lane witnesses; no compiler fixes or fixtures.',
 'Commits: base 50654a40; delivery is the commit containing this report on compiler/rethink-check.',
 'Commands and outputs: Node source, native release, JavaScript backend and ASan/UBSan/LSan runs; '+', '.join(str(v)+' '+k for k,v in counts.items())+' across '+str(len(rows))+' programs.',
 'Mutants: three executed JavaScript artifact mutants independently caught by stdout, stderr and exit comparisons; no compiler implementation mutants.',
 'Not covered: full gate, generic static specializations, accessor-spread throw edges while refused, or new fixtures/count rows.',
 '',
 'The explicitly requested base takes precedence over the generic current-main start rule. No later main is merged into this audit. Programs are committed only as .a.txt and copied to /tmp/rethink-check/*.a for execution. No cohere source was copied. This supplies prelanding evidence for the lowering chain and roadmap steps 16 (generics) and 22 (accessor spread); it lands no implementation.',
 '',
 'Classification uses exact stdout, stderr and exit agreement with source Node, plus a clean sanitized run. REFUSED includes a sound path-bearing NotYet implementation stop, not just a permanent language refusal. A workaround below is reviewer guidance, not a fix supplied by a diagnostic unless explicitly stated. STILL WRONG includes bad C, crashes and exact output differences. In particular, an intentional engine-stderr difference is labeled and not inferred to be a new compiler bug.',
 '',
 '| Task | Classification | Evidence |', '| --- | --- | --- |']
for task in ['bxpmash','drtmvb3','vk7ed2m','q9hz1bf','dv99xzy','eep1m5z']:
 r=lookup[task]; status=r['classification']
 if task=='eep1m5z':
  family=[x for x in rows if x['name'].startswith('eep1m5z-')]; status='STILL WRONG' if any(x['classification']=='STILL WRONG' for x in family) else ('REFUSED' if any(x['classification']=='REFUSED' for x in family) else 'FIXED'); detail='; '.join(str(sum(x['classification']==s for x in family))+' '+s for s in counts)
 else: detail=cell(r.get('native') or r.get('native_compile'))
 lines.append('| #'+task+' | '+status+' | '+safe(detail)+' |')
lines += ['', '| Program | Classification | Node | Native release | JavaScript | Sanitized native |', '| --- | --- | --- | --- | --- | --- |']
for r in rows:
 vals=[r['name'],r['classification'],cell(r['node']),cell(r.get('native') or r['native_compile']),cell(r.get('js') or r['js_compile']),cell(r.get('sanitized') or r['sanitized_compile'])]
 vals[0]='['+r['name']+']('+Path(r['source']).name+')'; lines.append('| '+' | '.join(safe(x) for x in vals)+' |')
lines += ['', 'Observations and attribution:', '']
for key,note in notes.items(): lines.append('- #'+key+': '+note)
for r in rows:
 if r['name'].startswith('eep1m5z-'):
  n=r['name'][len('eep1m5z-'):]; lines.append('- '+r['name']+': '+workarounds.get(n,'No fixing commit identified by the saved searches.'))
lines += ['', 'Complete per-command outputs, invocation arguments and elapsed seconds are in [results.json](results.json). Every compiler invocation and execution has separate .out.log and .err.log files. Generated JavaScript is retained in the js-compile.out.log files. The program links above contain the exact inputs. [run.py](run.py) recreates the suite with a 60-second limit per subprocess and progress saved after every program. [run-comparison-mutants.py](run-comparison-mutants.py) executes the three mutants, saving their complete artifacts as .mjs.txt.', '', 'Toolchain setup:', '', 'GOPROXY=https://proxy.golang.org|direct was set before setup; /workspace/adamic-tools/env.sh was sourced for builds and runs. nproc=5. The initial timeout 600 bash cloud/setup.sh and retry with GOFLAGS=-p=2 under timeout 300 did not complete dependency warming. The retry exited 124. The initial surrounding shell continued to nproc, so it did not preserve the setup exit code. Go, clang with its sanitizer overflow check, Node and the submodule finished installation. Initial CLI build limits 180 and 300 also expired (124). The final bounded build used timeout 600 go build -p 2 -o /tmp/rethink-adamic ./cmd/adamic; succeeded with exit 0; see build-final.log. The installed toolchain is Go 1.27.1, clang 20.1.8, Node 24.19.0.', '', 'Initial setup timing lines:', '', '```text']
lines += [x for x in (p/'setup-initial.log').read_text().splitlines() if x.startswith('setup:')]
lines += ['```', '', 'Retry setup timing lines:', '', '```text']
lines += [x for x in (p/'setup-retry.log').read_text().splitlines() if x.startswith('setup:')]
lines += ['```', '', 'Validation commands:', '', '```sh', 'source /workspace/adamic-tools/env.sh', 'timeout 600 python3 review/compiler/rethink-check/run.py > review/compiler/rethink-check/run.log 2>&1', 'timeout 60 python3 review/compiler/rethink-check/run-comparison-mutants.py > review/compiler/rethink-check/comparison-mutants.log 2>&1', 'git diff --check', "git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -", '```', '', 'No Go tests or .a fixtures were added or touched, so counts.md regeneration and TestCallTargetReaders do not apply. No whole package tests or full gate were run. Integration lane output is saved in lane-checks.log. An automatic approval-review attempt for the evidence runner timed out; its authorized retry succeeded.']
(p/'report.md').write_text('\n'.join(lines)+'\n')
