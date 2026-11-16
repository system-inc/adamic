#!/usr/bin/env python3
"""Render hand-checked source trace data; no inferred allocation/reach algorithm."""
import hashlib
import json
from pathlib import Path
import sys
here = Path(__file__).resolve().parent
root = Path(sys.argv[1]).resolve()
provenance = json.loads((here / 'provenance.json').read_text())
for file, digest in provenance['compiler_sources_sha256'].items():
    assert hashlib.sha256((root / file).read_bytes()).hexdigest() == digest, file

def line(file, number):
    return (root / ('src/compiler/' + file)).read_text().splitlines()[number - 1].strip()

def at(file, number, why):
    return {'file': 'src/compiler/' + file, 'line': number, 'quote': line(file, number), 'why': why}

allocations = {
    'L1': at('program.ts', 474, 'createCompilerHostWorker allocates and returns compilerHost; System-backed methods do not change the receiver allocation.'),
    'L2': at('watch.ts', 755, 'createCompilerHostFromProgramHost allocates a new compilerHost; copied or bound methods do not make the input host its receiver.'),
    'L3': at('watch.ts', 844, 'createProgramHost returns this literal. Watch and solution-host factories extend this same object.'),
    'L4': at('tsbuildPublic.ts', 495, 'createSolutionBuilderState allocates state and returns it at line 551.'),
}
boundaries = {
    'P1': at('program.ts', 1517, 'Public positional createProgram: external _host; createCreateProgramOptions forwards it to options.host.'),
    'P2': at('program.ts', 1501, 'Public options-object createProgram: external createProgramOptions.host (types.ts:7769).'),
    'P3': at('builder.ts', 1607, 'Exported @internal getBuilderCreationParameters: external oldProgramOrHost in the root-name-array branch can supply the host.'),
    'P4': at('builderPublic.ts', 179, 'Public createSemanticDiagnosticsBuilderProgram implementation third argument; root-name-array overload host at line 171.'),
    'P5': at('builderPublic.ts', 218, 'Public createEmitAndSemanticDiagnosticsBuilderProgram implementation third argument; root-name-array overload host at line 210.'),
    'P6': at('builderPublic.ts', 256, 'Public createAbstractBuilder implementation third argument; root-name-array overload host at line 248.'),
    'P7': at('watchPublic.ts', 136, 'Public createIncrementalProgram input.host, destructured at 140-147; if absent, default is L1.'),
    'P8': at('watch.ts', 986, 'Exported @internal performIncrementalCompilation input.host; declaration watch.ts:978, then forwarded through createIncrementalProgram(input).'),
    'S1': at('tsbuildPublic.ts', 339, 'Public createSolutionBuilder host parameter; no allocation is imposed on caller-supplied object.'),
    'S2': at('tsbuildPublic.ts', 343, 'Public createSolutionBuilderWithWatch host parameter; no allocation is imposed on caller-supplied object.'),
}
unknowns = {
    'U1': at('watchPublic.ts', 151, 'Unknown external CreateProgram callback target when supplied in IncrementalProgramOptions; known default is builderPublic.ts:215.'),
    'U2': at('watchPublic.ts', 696, 'Unknown external host.createProgram callback target, captured at line 455; another call at line 642. Known default and CLI wrapper are followed in report.'),
    'U3': at('tsbuildPublic.ts', 1048, 'Unknown external host.createProgram callback target; known default and CLI wrapper are followed in report.'),
}
# Every line below was selected and its role checked by hand against the AST/source.
paths = {
 'P': [
  ('program.ts',4895,'Saved fileExists from helper parameter host.compilerHost; other reads at 4896-4897.'),
  ('program.ts',1727,'Only internal caller of redirect helper; argument literal field compilerHost: host at 1728, so follow field rather than count wrapper literal.'),
  ('program.ts',1579,'Receiver host joins createProgramOptionsHost and default createCompilerHost.'),
  ('program.ts',1519,'Destructure host: createProgramOptionsHost from options object.'),
  ('program.ts',1518,'Positional arguments are put into options by createCreateProgramOptions; options-object form retains input.'),
  ('program.ts',1484,'createCreateProgramOptions returns host field unchanged.'),
  ('program.ts',390,'createCompilerHost returns createCompilerHostWorker: L1.'),
  ('builder.ts',1620,'Only builder-to-program call: options.host is oldProgramOrHost as CompilerHost at 1623; cast does not allocate.'),
  ('builderPublic.ts',186,'Semantic builder forwards all overload arguments to getBuilderCreationParameters.'),
  ('builderPublic.ts',225,'Emit builder forwards all overload arguments to getBuilderCreationParameters.'),
  ('builderPublic.ts',261,'Abstract builder forwards all overload arguments to getBuilderCreationParameters.'),
  ('executeCommandLine.ts',929,'Only other internal direct call to program.ts createProgram; programOptions.host is local host at 926.'),
  ('executeCommandLine.ts',915,'CLI local host comes from worker: L1.'),
  ('watchPublic.ts',123,'Incremental-host factory calls worker: L1; return host at 128.'),
  ('watchPublic.ts',148,'Incremental program chooses supplied host or incremental-host factory L1.'),
  ('watchPublic.ts',149,'Incremental createProgram chooses supplied callback or default emit builder, through casts that retain function identity.'),
  ('watch.ts',988,'performIncrementalCompilation uses input.host or stores incremental factory L1 in input.host.'),
  ('watch.ts',989,'Forwards same input to createIncrementalProgram.'),
  ('executeCommandLine.ts',949,'CLI incremental path creates L1, puts it in host field at 952, and calls public incremental worker at 951.'),
  ('watchPublic.ts',455,'Watch callback createProgram is destructured from input host; callback entry does not allocate receiver.'),
  ('watchPublic.ts',492,'Watch callback third argument compilerHost is L2.'),
  ('watchPublic.ts',642,'Watch update passes compilerHost as third argument; no new program for undefined rootNames branch of builder.'),
  ('watchPublic.ts',696,'Watch root-name call passes same L2 compilerHost as third argument to builder callback.'),
  ('tsbuildPublic.ts',434,'Solution compilerHost is factory L2.'),
  ('tsbuildPublic.ts',1048,'Solution host.createProgram gets compilerHost as third argument at 1051.'),
  ('watch.ts',861,'Program-host createProgram field defaults to emit builder; custom callback is open.'),
  ('executeCommandLine.ts',982,'CLI wrapper saves compileUsingBuilder from createProgram field.'),
  ('executeCommandLine.ts',988,'CLI wrapper forwards host argument unchanged to saved builder callback.'),
 ],
 'S': [
  ('tsbuildPublic.ts',782,'Restored receiver host is state.host from destructure at 780; same for next four host restores.'),
  ('tsbuildPublic.ts',745,'enableCache takes compilerHost and host from same state.'),
  ('tsbuildPublic.ts',758,'Cache helper receives host as argument at 759, not state.compilerHost.'),
  ('program.ts',527,'Helper saves host.readFile; other original-method reads at 528-531.'),
  ('program.ts',625,'Helper returns original fields in literal; this cache carrier is not original host.'),
  ('tsbuildPublic.ts',766,'state.cache stores returned originals, originalGetSourceFile, originalReadFileWithCache.'),
  ('tsbuildPublic.ts',496,'state.host field initialized with host.'),
  ('tsbuildPublic.ts',429,'host is cast alias of hostOrHostWithWatch, without allocating.'),
  ('tsbuildPublic.ts',2269,'Worker only internal caller of createSolutionBuilderState.'),
  ('tsbuildPublic.ts',340,'Public nonwatch builder forwards host unchanged to worker.'),
  ('tsbuildPublic.ts',344,'Public watch builder forwards host unchanged to worker.'),
  ('executeCommandLine.ts',895,'Only internal nonwatch builder caller passes buildHost.'),
  ('executeCommandLine.ts',885,'Its buildHost comes from createSolutionBuilderHost.'),
  ('executeCommandLine.ts',878,'Only internal watch builder caller passes buildHost.'),
  ('executeCommandLine.ts',855,'Its buildHost comes from createSolutionBuilderWithWatchHost.'),
  ('tsbuildPublic.ts',312,'Nonwatch host factory casts/extends base host; return host at 314.'),
  ('tsbuildPublic.ts',324,'Watch host factory casts/extends base host; copyProperties at 326, return host at 327.'),
  ('tsbuildPublic.ts',295,'Host base uses createProgramHost: L3; return host at 302.'),
 ],
 'C': [
  ('tsbuildPublic.ts',787,'Written compilerHost is state.compilerHost, initialized at 519 with factory result from 434.'),
  ('tsbuildPublic.ts',748,'Saved originalGetSourceFile is compilerHost.getSourceFile on same state field.'),
  ('watchPublic.ts',658,'Watch restores write compilerHost, held const from 492; five sibling writes through 662.'),
  ('watchPublic.ts',636,'Watch helper cache receives compilerHost from same lexical const.'),
  ('watch.ts',783,'Compiler-host factory returns its compilerHost literal L2.'),
 ],
 'D': [
  ('tsbuildPublic.ts',788,'Written auxiliary receiver is state itself.'),
  ('tsbuildPublic.ts',747,'originalReadFileWithCache saved by reading state.readFileWithCache.'),
  ('tsbuildPublic.ts',763,'Temporary replacement only changes state.readFileWithCache, not state identity.'),
  ('tsbuildPublic.ts',551,'State factory returns literal L4.'),
  ('tsbuildPublic.ts',2269,'Worker captures returned state for public builder operations; return object starts at 2270.'),
  ('tsbuildPublic.ts',816,'enableCache caller setupInitialBuild receives same state from builder operations.'),
  ('tsbuildPublic.ts',2050,'Other enableCache caller invalidateProject receives same state.'),
  ('tsbuildPublic.ts',742,'disableCache may run during enableCache on same state.'),
  ('tsbuildPublic.ts',1973,'disableCache caller buildWorker passes same state.'),
  ('tsbuildPublic.ts',2115,'disableCache caller buildNextInvalidatedProjectWorker passes same state.'),
  ('tsbuildPublic.ts',2069,'Timer callback alias is buildNextInvalidatedProject.'),
  ('tsbuildPublic.ts',2072,'Timer captures state as t2; external timer is not given t2 as a supplied callback argument.'),
  ('tsbuildPublic.ts',2074,'Arrow invokes callback(t1,t2,t3); second argument is captured state.'),
 ],
 'W-input': [
  ('watchPublic.ts',419,'Public watch entry can receive any compatible host; it is an input to the L2 factory, not its output identity.'),
  ('executeCommandLine.ts',1031,'Only internal config watch entry caller; watchCompilerHost from 1020.'),
  ('executeCommandLine.ts',1051,'Only internal file-list watch entry caller; watchCompilerHost from 1042.'),
  ('watchPublic.ts',370,'Public watch-host file-list wrapper delegates to internal factory.'),
  ('watchPublic.ts',382,'Public config watch-host wrapper delegates to internal factory.'),
  ('watch.ts',933,'Config watch-host factory uses private watch createWatchCompilerHost.'),
  ('watch.ts',964,'File-list watch-host factory uses private watch createWatchCompilerHost.'),
  ('watch.ts',872,'Private watch-host creator casts/extends createProgramHost L3.'),
  ('watch.ts',873,'Copies watch operations into same L3, not a fresh result.'),
  ('watch.ts',891,'Returns same L3 result.'),
 ]
}
paths = {k: [at(*item) for item in v] for k,v in paths.items()}
index = json.loads((here / 'source-index.json').read_text())
state_functions = {(f['at']['file'], f['at']['line']): f for f in index['functions'] if any('SolutionBuilderState' in p for p in f['parameters'])}
state_calls = []
for call in index['calls']:
    target = call['implementation'] or call['target']
    if target and (target['file'], target['line']) in state_functions:
        fn = state_functions[target['file'], target['line']]
        arg = next(i for i,p in enumerate(fn['parameters']) if 'SolutionBuilderState' in p)
        actual = call['arguments'][arg] if arg < len(call['arguments']) else '<omitted>'
        assert actual in ('state', 't2'), (call, actual)
        state_calls.append({'caller': call['at'], 'target': target, 'state_argument': actual})
assert len(state_functions) == 61 and len(state_calls) == 228
assert not any(f['exported'] for f in state_functions.values())
assert len(index['statePasses']) == 228
assert all(p['target'] and (p['target']['file'],p['target']['line']) in state_functions for p in index['statePasses'])
noncall = [r for r in index['refs'] if (r['target']['file'],r['target']['line']) in state_functions and r['parent'] not in ('CallExpression','FunctionDeclaration')]
assert len(noncall) == 1 and noncall[0]['name'] == 'buildNextInvalidatedProject' and noncall[0]['at']['line'] == 2069
assert allocations['L1']['quote'] == 'const compilerHost: CompilerHost = {'
assert allocations['L2']['quote'] == 'const compilerHost: CompilerHost = {'
assert allocations['L3']['quote'] == 'return {'
assert allocations['L4']['quote'] == 'const state: SolutionBuilderState<T> = {'
sites = json.loads((here / 'sites.json').read_text())['rows']
rows = []
for s in sites:
    file, member = s['site']['file'], s['member']
    group = ('P' if file.endswith('/program.ts') else 'C' if file.endswith('/watchPublic.ts') or member == 'getSourceFile' else 'D' if member == 'readFileWithCache' else 'S')
    literal_ids = {'P': ['L1','L2'], 'S':['L3'], 'C':['L2'], 'D':['L4']}[group]
    boundary_ids = list(k for k in boundaries if k.startswith('P' if group == 'P' else 'S')) if group in ('P','S') else []
    rows.append({'site': s['site'], 'group':group, 'written_literals':literal_ids, 'saved_source_literals':literal_ids, 'outside_boundaries':boundary_ids, 'one_and_same':group in ('C','D')})
result = {'label': "a source-level trace, not 949a2198's reach analysis; it can over- or under-count where lowering's graph differs, and it names each place it can't follow.", 'allocations':allocations,'outside_boundaries':boundaries,'unknown_callback_targets':unknowns,'paths':paths,'rows':rows,'state_functions':61,'state_calls':state_calls}
(here / 'source-trace.json').write_text(json.dumps(result,indent=2)+'\n')
