from pathlib import Path
import json,subprocess,time,hashlib
r=Path('/workspace/scratch/write-file-construct-run');impl=Path('/workspace/scratch/write-file-construct-impl');proof=Path('/workspace/scratch/write-file-construct-final');parent='a77ca4a82271f18c142d8843ea76604b30d3833d';main='89614cc0f9d6aa7b7fe96b56c5c6c1c3eace9825'
def load(name):return json.loads((r/name).read_text())
def complete():
 if not all((r/p).exists() for p in ['root-profile.exit','final-apply.exit','actual-stage-proofs.json','actual-small-proofs.json','built-proofs.json']):return False
 built=load('built-proofs.json');stages=load('actual-stage-proofs.json');small=load('actual-small-proofs.json')
 return len(built)==6 and all(x['name'].startswith('actual') for x in built[:4]) and any(x['name']=='99-actual-source-transpile-check' for x in stages) and len(small)==4
while not complete():time.sleep(5)
for file in ['main-apply.exit','control-apply.exit','lane.exit','root-profile.exit','final-apply.exit','runtime-keys.exit','full-sys-control.exit','full-sys-candidate.exit','99-watch-directory-witness-final.exit','control-build-tests.exit']:
 assert (r/file).read_text().strip()=='0',file
for name in ['actual-stage-proofs.json','built-proofs.json']:
 assert all(x.get('exit',0)==0 for x in load(name)),name
assert [x['exit'] for x in load('actual-small-proofs.json')]==[1,1,0,0]
for x in load('checks-exits.json'):
 if x['name']!='99-watch-directory-witness':assert x['exit']==0,x
pairs=load('pairs.json');assert len(pairs['goneUnique'])==1 and not pairs['newUnique'] and pairs['changedUnique']==0
candidate=r/'lane/adapted-tree';final=r/'final-adapted';control=r/'control-adapted'
def inventory(tree):return {p.relative_to(tree).as_posix():hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src').rglob('*')) if p.is_file()}
measured=inventory(candidate);fresh=inventory(final);assert measured==fresh and len(fresh)==747
(r/'final-source-identity.json').write_text(json.dumps({'sourceFiles':747,'freshFrozenMergeEqualsOnceTestedLaneTree':True,'files':fresh},indent=2)+'\n')
codeFiles=subprocess.check_output(['git','diff','--name-only',parent,'51104a74b'],cwd=impl,text=True).splitlines();codeFiles=[p for p in codeFiles if p.endswith(('.cjs','.json')) or p=='stage3/upstream/LEDGER.md'];code={}
for p in codeFiles:
 assert (impl/p).read_bytes()==(proof/p).read_bytes(),p
 code[p]=hashlib.sha256((proof/p).read_bytes()).hexdigest()
assert not subprocess.check_output(['git','diff','--name-only','6f953033661aac9931a86108ed75b3a40d3b556c','HEAD','--','internal','cmd','cohere','go.mod','go.sum','stage3/census'],cwd=proof,text=True).strip()
freeze=subprocess.check_output(['git','rev-parse','51104a74b'],cwd=impl,text=True).strip();merge=subprocess.check_output(['git','rev-parse','HEAD'],cwd=proof,text=True).strip()
(r/'code-pin-ledger-identity.json').write_text(json.dumps({'frozenCommit':freeze,'merge':merge,'main':main,'finalImplementationMatchesFrozenCodePinsLedger':True,'productionAndCensusGoSourceIdenticalToMeasuredMerge':True,'files':code},indent=2)+'\n')
def roots(name,tree):return {str(Path(x['roots'][0]).relative_to(tree/'src/compiler')):x for x in load('roots-'+name+'.json')}
m=roots('control',control);c=roots('candidate',candidate);assert m.keys()==c.keys() and len(c)==79
loc=lambda row,tree:row.get('where','').replace(str(tree)+'/','')
former=[k for k,x in m.items() if x.get('where','').endswith('/sys.ts:1387:5')];assert len(former)==73
for k in former:assert c[k]['kind']=='NotYet' and loc(c[k],candidate)=='src/compiler/sys.ts:1481:57',k
for k in m.keys()-set(former):assert (m[k]['kind'],loc(m[k],control),m[k].get('reason',''))==(c[k]['kind'],loc(c[k],candidate),c[k].get('reason','')),k
lines=['# All 79 production roots','',f'Control: {parent} merged with main {main}. Candidate: frozen implementation merged with the same main.','', 'Real production Load/Lower; CommonJSBundle. Driver selects entries only; inspect implementation is byte-identical to production.','', '| Root under src/compiler | Control | Candidate |','|---|---|---|']
for k in sorted(c):
 describe=lambda x,t:x['kind']+(' '+loc(x,t) if x.get('where') else '')
 lines.append(f'| `{k}` | {describe(m[k],control)} | {describe(c[k],candidate)} |')
lines+=['', 'All 73 former target roots move to the same clearTimeout boundary; all other six outcomes are unchanged.','']
(r/'production-roots.md').write_text('\n'.join(lines))
subprocess.run(['git','bundle','create',str(r/'frozen-proof.bundle'),'HEAD','^'+parent,'^'+main],cwd=proof,check=True)
summary={'parent':parent,'main':main,'frozenImplementation':freeze,'frozenMerge':merge,'productionRoots':79,'former1387Roots':73,'newProductionStop':'NotYet src/compiler/sys.ts:1481:57','otherProductionOutcomesUnchanged':6,'fullSysUniqueBefore':pairs['main']['uniqueSites'],'fullSysUniqueAfter':pairs['candidate']['uniqueSites'],'gone':1,'new':0,'changed':0,'stockRoots':79,'stockExit':0,'adamicTypesRoots':79,'adamicTypesExit':0,'publicDeclarationDeltaBytes':0,'publicSha256':load('declarations.json')['candidate']['sha256'],'wholeLane':load('lane/verdict.json'),'finalFreshApplyExit':0,'finalSourceFilesIdenticalToLane':747,'proofFiles':['actual-stage-proofs.json','actual-small-proofs.json','built-proofs.json','checks-exits.json','mutants.json','view-cli-mutant.json','runtime-keys.json'],'knownRed93MainCheck':int((r/'93-main-prefix-known-red.exit').read_text()),'ledger':load('ledger-check.json')}
(r/'ACCEPTANCE.json').write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps({k:v for k,v in summary.items() if k not in ['ledger','wholeLane']},indent=2),flush=True)
