from pathlib import Path
import json,gzip,hashlib,shutil,difflib,subprocess as s
D=Path('/workspace/scratch/recordreplay');O=Path(__file__).resolve().parent;G=O/'gate';G.mkdir(exist_ok=True)
h=lambda b:hashlib.sha256(b).hexdigest()
for name in ['stock.cjs','gate.py','gate-lower.py','gate2.py','full-pairs.py']:
 shutil.copyfile(D/name,O/name)
shutil.copyfile(D/'run.py',O/'run.py')
resources={}
for prefix in ['control-stock','control-dts','control-types','control-sys-full','candidate-stock','candidate-dts','candidate-types','candidate-sys-full','candidate-ordinary-sys','candidate2-stock','candidate2-dts','candidate2-types','candidate2-sys-full','A-main-sys-full','B-main-sys-full','amendment-apply','amendment-proof']:
 p=D/(prefix+'.json')
 if p.exists():resources[prefix]=json.loads(p.read_text())
 for suffix in ['log','jsonl']:
  p=D/(prefix+'.'+suffix)
  if p.exists():(G/(p.name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
(G/'resources.json').write_text(json.dumps(resources,indent=2)+'\n')
for name in ['control','candidate','candidate2']:
 shutil.copyfile(D/(name+'-stock-data.json'),G/(name+'-stock.json'))
 for public in ['typescript.d.ts','tsserverlibrary.d.ts']:
  (G/(name+'-'+public+'.gz')).write_bytes(gzip.compress((D/(name+'-built')/public).read_bytes(),mtime=0))
 (G/(name+'-sys.ts.txt.gz')).write_bytes(gzip.compress((D/name/'src/compiler/sys.ts').read_bytes(),mtime=0))
 if name!='control':
  old=(D/'control/src/compiler/sys.ts').read_text().splitlines(True);new=(D/name/'src/compiler/sys.ts').read_text().splitlines(True)
  (G/(name+'-source.diff')).write_text(''.join(difflib.unified_diff(old,new,fromfile='control/src/compiler/sys.ts',tofile=name+'/src/compiler/sys.ts')))
(G/'augmentation.d.ts.txt').write_bytes((D/'candidate2/src/recordreplay.d.ts').read_bytes())
shutil.copyfile(D/'head/stage3/adapt/99-wasm-crypto-import/proofs/recordreplay/results.json',G/'amendment-proof.json')
manifest={str(p.relative_to(D/'control')):h(p.read_bytes()) for p in (D/'control/src').rglob('*') if p.is_file()};(G/'control-source-hashes.json').write_text(json.dumps(manifest,indent=2)+'\n')
public={}
for name in ['typescript.d.ts','tsserverlibrary.d.ts']:
 a=(D/'control-built'/name).read_bytes();public[name]={}
 for variant in ['candidate','candidate2']:
  b=(D/(variant+'-built')/name).read_bytes();public[name][variant]={'before_bytes':len(a),'after_bytes':len(b),'before_sha256':h(a),'after_sha256':h(b),'byte_identical':a==b};assert a==b
  (G/(variant+'-'+name+'.diff')).write_text(''.join(difflib.unified_diff(a.decode().splitlines(True),b.decode().splitlines(True))))
(G/'declarations.json').write_text(json.dumps(public,indent=2)+'\n')
(G/'official-apply.log.gz').write_bytes(gzip.compress((D/'apply.log').read_bytes(),mtime=0))
discarded={}
for p in D.glob('*.discarded-overlap.*'):
 if p.suffix=='.json':discarded[p.name]=json.loads(p.read_text())
(G/'discarded-overlap.json').write_text(json.dumps({'reason':'A short ordinary diagnostic census overlapped the second candidate full census. Both observations were discarded and rerun serially after adding a cross-process runner lock. Only final resource/stream files are used.','discarded':discarded},indent=2)+'\n')
(G/'gate.json').write_text(json.dumps({'selected':'candidate2','amendment_commit':s.check_output(['git','rev-parse','HEAD'],cwd=D/'head',text=True).strip(),'candidate':{'stock_diagnostics':0,'adamic_types_exit':0,'javascript_identical':True,'public_declarations_identical':True,'selected':False,'reason':'adds Refused: a global augmentation at sys.ts:1977:1'},'candidate2':{'stock_diagnostics':0,'adamic_types_exit':0,'javascript_identical':True,'public_declarations_identical':True,'selected':True,'next_wall':'sys.ts:1613:46: reading modulePath','new_process_findings':0},'inputs_identical_between_compiler_sides':True,'B_baseline':'main compiler on amended source, required for stage3/census/pair same-input read normalization'},indent=2)+'\n')
