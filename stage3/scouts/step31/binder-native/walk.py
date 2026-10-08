#!/usr/bin/env python3
"""Bounded discovery walk; throwing copies never enter an acceptance slot."""
import difflib,gzip,json,os,re,shutil,subprocess,time
from pathlib import Path
P=Path(__file__).resolve().parent
ROOT=Path(os.environ.get('WALK_ROOT','/workspace/cache/step31-binder-walk-v3'));ROOT.mkdir(exist_ok=True)
TREE=ROOT/'tree'
if not TREE.exists():shutil.copytree('/workspace/cache/step31-binder-driver-slice-v2',TREE)
BINARY='/workspace/cache/step31-binder-area-adamic';CWD='/workspace/cache/step31-binder-area';rows=json.loads((ROOT/'stops.json').read_text()) if (ROOT/'stops.json').exists() else []
# Resume only after completing the explicit patch for the final recorded stop.
for order in range(len(rows)+1,16):
 start=time.monotonic();stem=ROOT/f'stop-{order:02d}'
 with stem.with_suffix('.stdout').open('wb') as out,stem.with_suffix('.stderr').open('wb') as err:
  code=subprocess.run([BINARY,'build',str(TREE/'main.a'),'-o',str(ROOT/'discovery-never-accept')],cwd=CWD,env=dict(os.environ,ADAMIC_NATIVE_SPLIT='0'),stdout=out,stderr=err).returncode
 text=stem.with_suffix('.stderr').read_text();m=re.search(r'(?:adamic: )?(/[^\n]+?):(\d+):(\d+): (.*)',text)
 if code==0 or not m:
  (ROOT/'terminal.json').write_text(json.dumps({'order':order,'exit':code,'message':text}));break
 file,line,col,message=m.groups();
 if rows and rows[-1]['file']==str(Path(file).relative_to(TREE)) and rows[-1]['line']==int(line) and rows[-1]['column']==int(col) and rows[-1]['message']==message:
  print('unchanged stop; explicit placeholder required',flush=True);break
 row={'order':order,'file':str(Path(file).relative_to(TREE)),'line':int(line),'column':int(col),'message':message,'exit':code,'wallSeconds':time.monotonic()-start};rows.append(row)
 (ROOT/'stops.json').write_text(json.dumps(rows,indent=2)+'\n');print(json.dumps(row),flush=True)
 before=Path(file).read_text()
 with gzip.GzipFile(filename=str(ROOT/f'stop-{order:02d}.source.gz'),mode='wb',mtime=0) as source:source.write(Path(file).read_bytes())
 result=subprocess.run(['node',str(P/'stub.cjs'),file,line,col,str(order)],text=True,capture_output=True)
 after=Path(file).read_text()
 (ROOT/f'patch-{order:02d}.diff').write_text(''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/'+row['file'],tofile='b/'+row['file'])))
 (ROOT/f'patch-{order:02d}.json').write_text(result.stdout);(ROOT/f'patch-{order:02d}.stderr').write_text(result.stderr)
 if result.returncode:
  print('explicit placeholder required',flush=True);break
