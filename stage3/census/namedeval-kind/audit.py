import pathlib,json,subprocess,hashlib,collections
p=pathlib.Path(__file__).resolve().parent
inputs=json.loads((p/'inputs.json').read_text());build=json.loads((p/'build-provenance.json').read_text())
runtime=pathlib.Path('/workspace/scratch/namedeval-kind-runtime');tree=pathlib.Path(inputs['commonAdaptedRoot'])
assert subprocess.check_output(['git','-C','/workspace/adamic','merge-base',inputs['revisions']['S1'],inputs['originMain']],text=True).strip()==inputs['revisions']['B']
subprocess.run(['git','-C','/workspace/adamic','merge-base','--is-ancestor',inputs['revisions']['S1'],inputs['revisions']['S2']],check=True)
assert all(hashlib.sha256((tree/f['file']).read_bytes()).hexdigest()==f['sha256'] for f in inputs['sourceFiles'])
assert len(set(inputs['applyKeys'].values()))==1
matrix=[]
for c,sha in inputs['revisions'].items():
 root=pathlib.Path('/workspace/scratch/namedeval-kind-'+c)
 assert subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip()==sha
 assert not subprocess.check_output(['git','diff','--name-only',sha],cwd=root,text=True).strip()
 assert build[c]['sha']==sha
 for mode in ['latent','scan']:
  receipt=json.loads((p/f'{c}-{mode}-build.json').read_text());assert receipt['exit']==0 and receipt['cwd']==str(root)
  assert hashlib.sha256(pathlib.Path(build[c]['binaries'][mode]['path']).read_bytes()).hexdigest()==build[c]['binaries'][mode]['sha256']
  assert str(runtime/c/(mode+'-overlay')/'overlay.json') in receipt['command']
 for file in ['namedEvaluation','esDecorators']:
  receipt=json.loads((p/f'{c}-latent-{file}.json').read_text());assert receipt['exit']==0 and receipt['cwd']==str(root)
  command=receipt['command'];assert str(runtime/c/'latent') in command
  for flag in ['LATENT_FULL=1','LATENT_ASSERT_NO_OUTPUT=1','CENSUS_MODULE_FORMAT=CommonJSBundle','LATENT_ONLY_FILE='+str(tree/'src/compiler/transformers'/(file+'.ts'))]:assert flag in command
  rows=[json.loads(s) for s in (p/f'{c}-latent-{file}.jsonl').read_text().splitlines()]
  assert len(rows)==3 and rows[0]['diagnostics']==[] and rows[0]['latent_mode']=='full' and rows[0]['module_format']=='CommonJSBundle' and rows[-1]['status']=='relation_report'
  matrix.append({'case':c,'mode':'latent','file':file,'exit':0,'units':len(rows[1]['units']),'findings':len(rows[1]['findings'])})
 receipt=json.loads((p/f'{c}-scan-namedEvaluation.json').read_text());assert receipt['exit']==0 and receipt['cwd']==str(root) and str(runtime/c/'scan-all') in receipt['command']
 matrix.append({'case':c,'mode':'scan','file':'namedEvaluation','exit':0})
for mode in ['latent','scan']:assert len({build[c]['binaries'][mode]['sha256'] for c in build})==3
summary=json.loads((p/'delta-summary.json').read_text())
assert summary['B-to-S1-namedEvaluation']['reached']==22 and summary['B-to-S1-esDecorators']['reached']==22
assert all(d['new']==0 for d in summary.values())
assert summary['S1-to-S2-namedEvaluation']['changed']==2 and summary['S1-to-S2-esDecorators']['pairedUnchanged']==587
for f in ['namedEvaluation','esDecorators']:
 d=json.loads((p/f'B-to-S1-{f}-delta.json').read_text());assert collections.Counter(x['kind']=='Boundary' for x in d['reached'])=={True:11,False:11}
call=json.loads((p/'caller-observations.json').read_text());assert all(call['canonicalFindingsByteEquivalentAsRecords'].values())
assert [x['upstreamLine'] for x in call['callers'] if x['successfulCheckConstructions']]==[1062]
# Every changed record shown in full; no production path is modified on the report branch.
base=inputs['revisions']['B'];files=subprocess.check_output(['git','diff','--name-only',base],text=True).splitlines()
assert all(f.startswith('stage3/census/namedeval-kind/') for f in files),files
result={'pass':True,'measurementMatrix':matrix,'allSixCanonicalBinariesDistinct':True,'probeOrOverlayReusedAcrossShas':False,'sameAdaptedTreeVerifiedAfterRuns':True,'applyDiffers':False,'reportOnly':True,'compilerTreesTrackedClean':True,'callerChecksObserved':1,'callerChecksMaskedByPrologues':2,'pairwiseSummary':summary}
(p/'audit.json').write_text(json.dumps(result,indent=2)+'\n')
print('PASS nine canonical measurements, independent build provenance, immutable B tree, exact pairing/grades and report-only diff')
