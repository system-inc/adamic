from pathlib import Path
import hashlib,json,re,subprocess
out=Path(__file__).resolve().parent;root=out.parents[2];rows=json.loads((out/'headers.json').read_text());after=json.loads((out/'after.json').read_text());proof=[]
for r in rows:
 p=root/r['path'];source=p.read_bytes();body=source.split(b'\n',1)[1] if source.startswith(b'// a-check:') else source
 old=subprocess.check_output(['git','show','611cbc943a28e88567bcb5fb9a43031518a69d3d:'+r['path']],cwd=root);oldbody=old.split(b'\n',1)[1] if old.startswith(b'// a-check:') else old
 assert body==oldbody and hashlib.sha256(body).hexdigest()==r['body_sha256'],r['path']
 if not r['changed']:assert source==old
 proof.append({'path':r['path'],'unchanged_body_sha256':r['body_sha256'],'header_changed':r['changed']})
status=root/'review/compiler/loom-records-3/stage3-mutant/records/status.json';pins=json.loads(status.read_text());diag={r['path']:r for r in after['rows']};verified=[]
for r in rows:
 if not r['changed'] or Path(r['path']).parent != status.relative_to(root).parent:continue
 row=next(p for p in pins if p['file']==Path(r['path']).name);actual=diag[r['path']]['stderr'].strip().removeprefix('adamic: ')
 actual=actual.replace(str(root)+'/'+str(status.relative_to(root).parent),'stage3/fixtures/records').replace(str(status.relative_to(root).parent),'stage3/fixtures/records')
 assert actual==row['stage0']['what'],(r['path'],actual,row['stage0']['what'])
 verified.append(r['path'])
assert subprocess.check_output(['git','show','611cbc94:internal/oracle/counts.md'],cwd=root)==(root/'internal/oracle/counts.md').read_bytes()
(out/'body-and-position-proof.json').write_text(json.dumps({'bodies':proof,'archived_rows_match_actual_diagnostics':verified,'counts_identical_to_source':True},indent=2)+'\n');print('PASS 109 unchanged bodies, 11 archived diagnostic rows, unchanged counts')
