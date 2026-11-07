"""Render API-derived counts and complete ledgers, without grep-based counts."""
import pathlib,json,collections,gzip,hashlib,subprocess
repo=pathlib.Path('/workspace/adamic'); scratch=pathlib.Path('/workspace/scratch/sys-timer-slot');out=repo/'stage3/census/sys-timer-slot'
def load(name):return json.loads((scratch/name).read_text())
def at(x):return x['file']+':'+str(x['line'])
def cell(x):return str(x).replace('|','\\|').replace('\n',' ').replace('\r','')
r=load('references.json');notes=json.loads((out/'caller-notes.json').read_text()); result={'provenance':{'main':subprocess.check_output(['git','rev-parse','origin/main'],cwd=repo,text=True).strip(),'upstream':'050880ce59e30b356b686bd3144efe24f875ebc8','typescript':r['typescript'],'srcFiles':r['sourceFiles'],'programFiles':r['programFiles']},'referenceCounts':[],'checks':{}}
lines=['| Caller (adapted file:line) | Receiver / declared slot | Callback parameter signature | Extra arguments with types | Handle destination / use |','| --- | --- | --- | --- | --- |']
for x in r['calls']:
    clear=x['timerName']=='clearTimeout'
    sig='; '.join(x['callbackSignatures'] or []) if not clear else 'N/A (cancellation)'
    args='; '.join(a['text']+': '+a['type'] for a in x['arguments'][2:]) if not clear else 'N/A'
    handle=x['parent']+'; destination type: '+x.get('destinationType','not stored') if not clear else ('cancels '+x['arguments'][0]['text']+': '+x['arguments'][0]['type'] if x['arguments'] else 'no argument; unrelated method')
    handle=notes.get(at(x),handle)
    lines.append('| '+' | '.join(map(cell,[at(x),x['receiverType']+' / '+(at(x['resolvedDeclaration']) if x.get('resolvedDeclaration') else 'unresolved'),sig,args or 'none',handle]))+' |')
lines+=['','Implementing, forwarding, binding and declaration members follow. The kind distinguishes an implementation from a public declaration or local alias. Fits are stock checker assignability to the candidate method type; optional absence is not a callable implementation.','','| Member (file:line) | Kind / owner | Inferred or declared type | Fits A | Fits B | Distinct reference spans (including definition) |','| --- | --- | --- | --- | --- | --- |']
for x in r['members']:
    refs={(a['file'],a['line'],a['column']) for g in x['references'] for a in g['references']}
    result['referenceCounts'].append({'file':x['file'],'line':x['line'],'kind':x['kind'],'owner':x.get('owner'),'referencesIncludingDefinition':len(refs),'fits':x['fits']})
    lines.append('| '+' | '.join(map(cell,[at(x),x['kind']+' / '+str(x.get('owner')),x['type'],x['fits']['CandidateA'],x['fits']['CandidateB'],len(refs)]))+' |')
lines+=['','Timer-bearing object literals, including spread forwarding, checked by the stock API:','','| Object (file:line) | Context | Timer slots fit A | Timer slots fit B | Spread expressions |','| --- | --- | --- | --- | --- |']
for x in r['objects']:
    lines.append('| '+' | '.join(map(cell,[at(x),x.get('contextualType'),x['fits']['CandidateA'],x['fits']['CandidateB'],x['spreads'] or 'none']))+' |')
lines+=['','Inherited hosts and timer neighbors. DirectoryStructureHost and the harness fake System are included even though neither declares a timer:','','| Host (file:line) | Heritage | Timer slots and their declaration locations | Explicit source edit forced |','| --- | --- | --- | --- |']
for x in r['interfaces']:
    if x['file'].startswith('src/lib/'):continue
    values='; '.join(t['name']+' from '+', '.join(at(d) for d in t.get('declarations',[])) for t in x['timers'])
    lines.append('| '+' | '.join(map(cell,[at(x)+' '+str(x.get('name')),x.get('heritage'),values or 'none','No; failures are private storage and Node timer heads, not other public slots']))+' |')
lines+=['','Neighbor declarations (all method/property signatures discovered by the stock AST; indexed System slots inherit their change without a source edit):','','| Neighbor (file:line) | Owner | Slot | Explicit edit forced |','| --- | --- | --- | --- |']
for x in r['members']:
    if x['kind'] not in ['MethodSignature','PropertySignature']:continue
    lines.append('| '+' | '.join(map(cell,[at(x),x.get('owner'),x['text'],'No separate edit; see paired diagnostics']))+' |')
lines+=['','Counts and method','','All counted references come from TypeScript 6.0.3 LanguageService.findReferences over the whole adapted src, including compiler, services, server, harness, testRunner/unittests and installers. Every discovered named timer member and binding is queried, including System itself. Each per-member count deduplicates (file,line,column) across reference groups and includes declarations. Counts overlap between related symbols and must not be summed. The retained references.json includes every raw group, declaration and use context. Text search was used for discovery of tooling and source reading, never as a reference count.','','The first table intentionally includes neighboring and global timers too, with their resolved declaration location. This avoids silently dropping calls forwarded through hosts. Raw handle references and callback declaration text are retained in references.json.']
for label in ['main','A','B']:
    data=load(label+'-stock.json');result['checks'][label]=data
base=result['checks']['main']
lines+=['','Stock diagnostic comparison','','| Tree | Adamic-options diagnostics | Upstream solution diagnostics | Solution exit | Compiler/test build exits |','| --- | --- | --- | --- | --- |']
cmds=load('commands-builds.json')
for label in ['main','A','B']:
    data=result['checks'][label]
    def added(a,b):
        key=lambda d:json.dumps(d,sort_keys=True)
        old=collections.Counter(map(key,b));new=collections.Counter(map(key,a));return [json.loads(k) for k,n in (new-old).items() for _ in range(n)]
    additions={'adamic':added(data['adamic'],base['adamic']),'solution':added(data['solution'],base['solution'])}
    result['checks'][label]['additions']=additions
    exits={x['label']:x['exit'] for x in cmds if x['label'].startswith(label+'-')}
    lines.append('| '+' | '.join(map(cell,[label,len(data['adamic']),len(data['solution']),data['solutionExit'],exits]))+' |')
lines+=['','Each added diagnostic follows; repeated diagnostics across the two checker modes are intentional.']
for label in ['A','B']:
    for mode,ds in result['checks'][label]['additions'].items():
        for d in ds:lines.append(label+' '+mode+': `'+at(d)+'` TS'+str(d['code'])+': '+cell(d['message']))
lines+=['','Artifact comparison','','Compiler and test builds use upstream tasks with --lkg, selecting pinned stock TypeScript 6.0.3. API checks use the separately pinned 6.0.3 compiler API. Adamic-options checks retain all 79 compiler roots, the prelude, Node types and the loader library overlays, rather than relaxing strictFunctionTypes. The stock API solution builder force-builds src/tsconfig.json with every original project option and reference. The actual upstream compiler/test tasks also retain normal typechecking. Main is zero in both modes; candidate failures are preserved.','','The typed A/B compiler builds each emitted eight unchanged JavaScript files and skipped tsserverlibrary.js after errors. finish-shim.mjs invokes the unchanged upstream lssl writer callback solely to finish the ninth file. The declaration bundler reads the .d.ts files actually emitted by stock tsc despite diagnostics. These artifact-only completions do not turn either failed build green. See declarations.json and A-declarations.diff.txt / B-declarations.diff.txt for exact built, internal and reference-baseline changes. Baseline edits are drafts derived from actual emitted declaration lines; no reference baseline is written in a candidate. Existing unrelated main-to-upstream API differences are excluded from the incremental diff.']
if (out/'declarations.json').exists():
    decl=json.loads((out/'declarations.json').read_text());result['artifacts']=decl
    for label,row in decl.items():lines+=['',label+': compiler JavaScript '+str(row['javascript']['mainCount'])+' vs '+str(row['javascript']['candidateCount'])+' files, byte-identical: '+str(row['javascript']['identical'])+'. Internal declaration identical: '+str(row['typescript.internal.d.ts'].get('identical'))+'.']
lines+=['','Latent census','','LATENT_FULL=1 and LATENT_ONLY_FILE select the adapted sys.ts, while all compiler roots remain loaded. The standard scratch overlay disables production Lower. LATENT_ASSERT_NO_OUTPUT=1 checks the disabled ordinary Load and Lower paths. The cited report was read from typescript/next-stops-after-wave; this uses pinned main\'s make_overlay.py and full-mode driver, with only physical Go evidence filenames changed to .go.txt. No speculative continuation is enabled.']
for label in ['main','A','B']:
    p=scratch/(label+'-latent.jsonl')
    if p.exists():
        rows=[json.loads(l) for l in p.read_text().splitlines() if l];result.setdefault('latent',{})[label]=rows
        lines+=['',label+': raw full census retained in '+label+'-latent.jsonl.gz.']
(out/'REPORT.json').write_text(json.dumps(result,indent=2)+'\n');(out/'REPORT.md').write_text('\n'.join(lines)+'\n')
for name in ['references.json','main-stock.json','A-stock.json','B-stock.json','main-public-timers.json','A-public-timers.json','B-public-timers.json']:
    shutil=__import__('shutil');shutil.copy2(scratch/name,out/name)
for p in scratch.glob('*'):
    if p.is_file() and (p.suffix in ['.log','.jsonl'] or p.name.startswith('commands-')):
        if p.suffix in ['.log','.jsonl']:
            with gzip.open(out/(p.name+'.gz'),'wb') as f:f.write(p.read_bytes())
        else:__import__('shutil').copy2(p,out/p.name)
