"""Package completed proof artifacts, without substituting summaries for raw data."""
from pathlib import Path
import json,hashlib,shutil,tarfile,io,subprocess
repo=Path('/workspace/adamic');s=Path('/workspace/scratch/debug-messages');raw=s/'raw';u=repo/'stage3/adapt/99-correctness-core-debug-messages'
load=lambda n:json.loads((raw/n).read_text())
lane=load('whole-lane-reconciled.json');replay=load('paired-replay-comparison.json');census=load('census-comparison.json');build=load('build-comparison.json');witness=load('final-witness.json');guards=load('adapter-guards.json')
assert lane['wholeOracleCompleted'] and lane['acceptedUnitsExactlyOnce']==398 and all(v==0 for v in lane['counts'].values())
assert replay['complete'] and replay['identicalOtherAdapters']==95
assert census['added']==[] and len(census['removed'])==4 and len(census['restoringMutantSite'])==1
assert witness['allPassed'] and len(witness['cases'])==32
assert all(p['byteIdentical'] for p in build['publicDeclarations'])
for side in ['before','after']:
 assert all(v==1 for v in load(side+'-counter-control.json')[0]['counts'].values())
for path in sorted((s/'evidence').iterdir()):
 if path.is_file():shutil.copyfile(path,u/path.name)
for name in ['paired-replay-comparison.json','paired-replay.diff']:
 shutil.copyfile(raw/name,u/name)
# Exact raw Go overlay sources have .go.txt physical names. No production Go source changed.
with tarfile.open(u/'guarded-overlay.tar.gz','w:gz') as tf:tf.add(s/'overlay',arcname='overlay')
for member in tarfile.open(u/'guarded-overlay.tar.gz').getnames():assert not member.endswith('.go'),member
members={str(p.relative_to(s)):p.read_bytes() for p in sorted(raw.rglob('*')) if p.is_file()}
for side in ['before','after','mutant']:members['sources/'+side+'-debug.ts.txt']=(s/side/'src/compiler/debug.ts').read_bytes()
for name in ['finish-apply.py','start-after-replay.py','make-pins.cjs']:members['controllers/'+name]=(s/name).read_bytes()
manifest={name:hashlib.sha256(data).hexdigest() for name,data in members.items()};members['MANIFEST.json']=(json.dumps(manifest,indent=2)+'\n').encode()
with tarfile.open(u/'raw-proof.tar.gz','w:gz') as tf:
 for name,data in members.items():
  info=tarfile.TarInfo(name);info.size=len(data);tf.addfile(info,io.BytesIO(data))
base='d15620712c459baf72b3d359c1829df8d0808fda';checkpoint='9fe9d71cbb885469409efbf39672f73c31bae200'
implementation={}
for name in ['adapt.cjs','pins.json']:
 now=(u/name).read_bytes();old=subprocess.check_output(['git','show',checkpoint+':'+str((u/name).relative_to(repo))],cwd=repo);assert now==old;implementation[name]=hashlib.sha256(now).hexdigest()
proof={'task':'b5w3ycg','ruling':'Adamic, Nov 7 19:16Z, road (b), JavaScript-changing','base':base,'typescript':'050880ce59e30b356b686bd3144efe24f875ebc8','postApplyCheckpoint':checkpoint,'implementationMatchesPostApplyCheckpoint':True,'implementationSha256':implementation,'ownershipDecision':'new directory; no active whole-owner or exact-message pin','gates':{'wholeLaneZeroCounts':True,'fourIndependentCounterControls':True,'onlyFailureTextChanged':True,'restoringCensusMutant':True,'getFunctionNameUnchanged':True,'ledgerLine':True,'bothFullApplies':True,'pairedReplayOnlyOursDiffers':True,'both79RootTypesZeroDiagnostics':True,'allEmittedJavaScriptHunksReviewed':True,'bothPublicDeclarationsIdentical':True,'fourCensusSitesGoneNoneAdded':True},'wholeLane':lane,'nodeWitnessCases':32,'normalBuilds':{side:load(side+'-build.json') for side in ['before','after']},'fullApply':{'beforeExit':0,'beforeMode':'verified keyed product restored by stage3/apply.sh','after':load('after-apply.json'),'controllerTransfer':load('apply-controller-transfer.json')},'types':{side:dict(exit=load(side+'-types.json')['exit'],roots=79,diagnostics=0,seconds=load(side+'-types.json')['seconds']) for side in ['before','after']},'census':{'flags':{'LATENT_FULL':'1','LATENT_ONLY_FILE':'<tree>/src/compiler/debug.ts','LATENT_ASSERT_NO_OUTPUT':'1'},'unitsPerRun':105,'runResults':{side:load(side+'-census.json') for side in ['before','after','mutant']},'removed':census['removed'],'added':[],'restoringMutantSite':census['restoringMutantSite'],'getFunctionNameRecordingUnits':sorted({r['recordingUnit'] for r in census['getFunctionNameObservations']})},'replay':replay,'publicDeclarations':build['publicDeclarations'],'changedJavaScript':[r for r in build['emittedJavaScript'] if r['changed']]}
(u/'PROOF.json').write_text(json.dumps(proof,indent=2)+'\n')
archives=sorted(u.glob('*.tar.gz'));assert all(p.stat().st_size<20_000_000 for p in archives)
(u/'ARCHIVES.json').write_text(json.dumps([{'file':p.name,'bytes':p.stat().st_size,'sha256':hashlib.sha256(p.read_bytes()).hexdigest()} for p in archives],indent=2)+'\n')
print('packaged',len(archives),'archives; largest',max(p.stat().st_size for p in archives),'bytes')
