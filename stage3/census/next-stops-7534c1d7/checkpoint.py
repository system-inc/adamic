"""Package completed root children and make one non-forced checkpoint push."""
import gzip,hashlib,json,os,subprocess,sys
from pathlib import Path
base=Path(__file__).resolve().parent
mode=os.environ.get('CENSUS_MODULE_FORMAT','CommonJSBundle');out=base/mode
hashes=json.loads((base/'compression.json').read_text()) if (base/'compression.json').exists() else {}
for p in base.rglob('*.log'):
 data=p.read_bytes();compressed=gzip.compress(data,mtime=0)
 hashes[str(p.relative_to(base))]={'sha256':hashlib.sha256(data).hexdigest(),'bytes':len(data),'compressed_sha256':hashlib.sha256(compressed).hexdigest()}
 if len(compressed)<=20_000_000:p.with_suffix(p.suffix+'.gz').write_bytes(compressed)
 p.unlink()
(base/'compression.json').write_text(json.dumps(hashes,indent=2)+'\n')
files=sorted(out.glob('root-??.jsonl'));rows=[]
for p in files:
 pair=[json.loads(l) for l in p.read_text().splitlines()]
 assert len(pair)==2 and {k:v for k,v in pair[0].items() if k!='seconds'}=={k:v for k,v in pair[1].items() if k!='seconds'}
 rows.append(pair[0])
(out/'partial-census.json').write_text(json.dumps(rows,indent=2)+'\n')
assert not [p for p in base.rglob('*') if p.is_file() and p.stat().st_size>20_000_000]
def run(args):
 rc=subprocess.call(args);print('checkpoint child_exit='+str(rc),flush=True)
 if rc:sys.exit(rc)
run(['git','add',str(base)])
run(['git','commit','-m',f'baobab: {mode} checkpoint {len(rows)}/79 roots'])
run(['git','push','origin','HEAD:refs/heads/typescript/next-stops-7534c1d7-checkpoint'])
