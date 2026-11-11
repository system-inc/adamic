import pathlib,subprocess,json,hashlib,difflib,sys,os,signal,time,shutil
sys.dont_write_bytecode=True
from measure import run as measure
out=pathlib.Path(__file__).resolve().parents[1];repo=out.parents[2];base=pathlib.Path('/workspace/scratch/baobab-adapted');follow=out/'continuation'
def invoke(command,log,bound=None):
 log.parent.mkdir(parents=True,exist_ok=True);start=time.monotonic();timed_out=False
 with log.open('wb') as dest:
  child=subprocess.Popen(command,stdout=dest,stderr=subprocess.STDOUT,cwd=repo,start_new_session=True)
  try:child.wait(timeout=bound)
  except subprocess.TimeoutExpired:
   timed_out=True;os.killpg(child.pid,signal.SIGTERM)
   try:child.wait(timeout=10)
   except subprocess.TimeoutExpired:os.killpg(child.pid,signal.SIGKILL);child.wait()
 record=dict(command=command,child_exit=child.returncode,wall_seconds=time.monotonic()-start,timed_out=timed_out)
 log.with_suffix(log.suffix+'.status.json').write_text(json.dumps(record,indent=2)+'\n');print(json.dumps(record),flush=True)
 if child.returncode:raise SystemExit(child.returncode)
def checkpoint(label):invoke(['python3',str(out/'checkpoint.py'),label],follow/'preparation'/(label+'-checkpoint.log'))
def core(label,root):
 target=follow/label;target.mkdir(parents=True,exist_ok=True)
 invoke(['/workspace/scratch/baobab-first-stop',str(root/'src/compiler/core.ts'),str(target/'core-first-stop.jsonl')],target/'core-first-stop.log',1500)
def make(label,annotation):
 root=pathlib.Path('/workspace/scratch/baobab-road2-'+label)
 target=follow/label;target.mkdir(parents=True,exist_ok=True)
 if not root.exists():invoke(['bash',str(repo/'stage3/apply.sh'),str(root)],target/'apply.log',2400)
 p=root/'src/compiler/scanner.ts';original=(base/'src/compiler/scanner.ts').read_bytes();changed=original
 for name in ['getLineAndCharacterOfPosition','getLineStarts']:
  before=('export function '+name+'(sourceFile: SourceFileLike,').encode() if name=='getLineAndCharacterOfPosition' else b'export function getLineStarts(sourceFile: SourceFileLike)'
  after=before.replace(b'SourceFileLike',annotation.encode());assert changed.count(before)==1,name;changed=changed.replace(before,after,1)
 p.write_bytes(changed)
 files={}
 for source in sorted((base/'src').rglob('*')):
  if not source.is_file():continue
  rel=source.relative_to(base);other=root/rel;assert other.is_file(),str(rel)
  old,new=hashlib.sha256(source.read_bytes()).hexdigest(),hashlib.sha256(other.read_bytes()).hexdigest()
  if old!=new:files[str(rel)]={'before':old,'after':new}
 assert set(files)=={'src/compiler/scanner.ts'},files
 assert not [p for p in (root/'src').rglob('*') if p.is_file() and not (base/p.relative_to(root)).exists()]
 (target/'source-diff.json').write_text(json.dumps({'annotation':annotation,'changed_files':files,'only_two_parameter_types':p.read_bytes()==changed,'baseline_main_sha':'d690f276749f91e6b36fe3a9726a5a453fcd7d39'},indent=2)+'\n')
 (target/'scanner.patch').write_text(''.join(difflib.unified_diff(original.decode().splitlines(True),changed.decode().splitlines(True),fromfile='baseline/src/compiler/scanner.ts',tofile=label+'/src/compiler/scanner.ts')))
 return root
if __name__=='__main__':
 expected=['checker.ts','emitter.ts','parser.ts','program.ts','scanner.ts','utilities.ts']
 assert all((follow/'baseline/raw'/(n.removesuffix('.ts')+'.status.json')).exists() for n in expected),'finish six baseline attempts before variants'
 assert json.loads((follow/'original/summary.json').read_text())['whole_src_diagnostics']==0
 sites=json.loads((out/'sites.json').read_text());eligible={s['file'] for s in sites}|{name for s in sites for name in s.get('observed_in',[])};durations={s['file']:s['wall_seconds'] for s in json.loads((out/'coverage.json').read_text())};changed_files={s['file'] for s in sites if s['callee'] in ['getLineStarts','getLineAndCharacterOfPosition']};files=sorted(eligible,key=lambda name:(0 if name not in expected else 1 if name in changed_files else 2,durations.get(name,0),name));(follow/'variant-files.json').write_text(json.dumps(files,indent=2)+'\n')
 core('original',base)
 prepared={}
 for label,annotation in [('a','{ readonly text: string; lineMap?: readonly number[] }'),('b','{ readonly text: string; lineMap: readonly number[] | undefined }')]:
  root=make(label,annotation)
  invoke(['node',str(follow/'source-audit.cjs'),str(root),label],follow/label/'source-audit.log',600)
  invoke(['node',str(follow/'emit-public.cjs'),str(root),label],follow/label/'public-emit.log',600)
  core(label,root)
  checkpoint('road2-'+label+'-prepared')
  prepared[label]=root
 from concurrent.futures import ThreadPoolExecutor
 for name in files:
  def measure_variant(item):
   label,root=item;measure(label,root/'src/compiler',[name],1500,False,concurrency=2)
  with ThreadPoolExecutor(max_workers=2) as pool:list(pool.map(measure_variant,prepared.items()))
  checkpoint('road2-pair-'+name.replace('/','__'))
