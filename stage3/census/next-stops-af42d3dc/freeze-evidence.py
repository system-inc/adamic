import json,hashlib,os,sys,shutil,importlib.util,subprocess
from pathlib import Path
from supervise import run,scratch
repo=Path('/workspace/adamic');out=Path(__file__).resolve().parent;core=repo/'stage3/census/core-sites-af42d3dc';tree=Path('/workspace/after-wave-adapted')
status=json.loads((scratch/'apply-status.json').read_text());assert status['exit_code']==0
# Import apply's exact product-key/cache functions without writing bytecode outside reports.
sys.dont_write_bytecode=True;sys.path.insert(0,str(repo/'stage3'))
import apply
key=apply.product_key();product=Path(os.environ.get('STAGE3_PRODUCT_STORE',str(apply.stage3_cache()/'products')))/(key+'.manifest');assert product.exists()
manifest=[{'file':p.relative_to(tree).as_posix(),'sha256':hashlib.sha256(p.read_bytes()).hexdigest()} for p in sorted((tree/'src').rglob('*')) if p.is_file()]
for d in [out,core]:
 (d/'adapted-source-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n');(d/'adapted-product.manifest.json').write_bytes(product.read_bytes())
 data=json.loads((scratch/'provenance-start.json').read_text());data.update(product_key=key,source_file_count=len(manifest),core_sha256=hashlib.sha256((tree/'src/compiler/core.ts').read_bytes()).hexdigest(),apply_status=status,source_manifest_sha256=hashlib.sha256((d/'adapted-source-manifest.json').read_bytes()).hexdigest(),product_manifest_sha256=hashlib.sha256(product.read_bytes()).hexdigest())
 (d/'provenance.json').write_text(json.dumps(data,indent=2)+'\n')
(core/'core.ts.txt').write_bytes((tree/'src/compiler/core.ts').read_bytes())
(out/'REPORT.md').write_text('# Next stops on af42d3dc\n\nCheckpoint after completed full apply. The 79-root census is pending. Main and TypeScript pins, source hashes and apply status are in provenance.json. Final report branch will contain one commit on main.\n')
(core/'REPORT.md').write_text('# Core sites on af42d3dc\n\nCheckpoint after completed full apply. The guarded full-file census is pending. Both runs use the same frozen adapted tree.\n')
proc=subprocess.run(['git','-C',str(tree),'for-each-ref','--format=%(refname:short)','refs/stage3'],capture_output=True,text=True)
print('git refs child_exit='+str(proc.returncode));proc.check_returncode()
completed=sorted(r.removesuffix('-after').removeprefix('stage3/') for r in proc.stdout.splitlines() if r.endswith('-after'))
expected=sorted(p.name for p in (repo/'stage3/adapt').iterdir() if p.is_dir())
assert completed==expected,(completed,expected)
for d in [out,core]:
 data=json.loads((d/'provenance.json').read_text());data['completed_adapters']=completed;data['adapter_count']=len(completed);data['environment_id']='ccarenv_b64_Y2NhcmVudl9iMTY3MDE5Mzk1YTQ4MTkxYjZiZjhmNWJiY2JlZDllOQ';(d/'provenance.json').write_text(json.dumps(data,indent=2)+'\n')
print('frozen product',key,'source files',len(manifest),'completed adapters',len(completed))
