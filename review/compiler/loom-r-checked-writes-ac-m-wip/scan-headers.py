import pathlib,subprocess,json,re,concurrent.futures
out=pathlib.Path('review/compiler/loom-r-checked-writes-ac-m-wip');paths=sorted(set(subprocess.check_output(['git','diff','--name-only','--diff-filter=AM','origin/main','--','*.a'],text=True).splitlines()+subprocess.check_output(['git','ls-files','--others','--exclude-standard','--','*.a'],text=True).splitlines()));(out/'paths.json').write_text(json.dumps(paths,indent=2)+'\n')
def one(pair):
 i,p=pair;r=subprocess.run(['timeout','30',str((out/'adamic-acheck').resolve()),'c',p],capture_output=True,text=True,timeout=35);(out/f'compile-{i:02d}.stderr.log').write_text(r.stderr);(out/f'compile-{i:02d}.stdout.log').write_text(r.stdout);err=r.stderr;status='checked';reason=None
 if 'Adamic 0.1 refuses ' in err:status='refused';reason=err.split('Adamic 0.1 refuses ',1)[1].splitlines()[0]
 elif re.search(r'error TS\d+',err):status='type error';reason=re.search(r'error (TS\d+)',err)[1]
 elif "can't lower"in err and ' yet'in err:status='NotYet';reason=err.strip()
 elif r.returncode:status='unclassified';reason=err
 return dict(path=p,status=status,reason=reason,exit=r.returncode,first_line=pathlib.Path(p).read_bytes().partition(b'\n')[0].decode())
with concurrent.futures.ThreadPoolExecutor(max_workers=4)as pool:rows=list(pool.map(one,enumerate(paths)))
(out/'status-before.json').write_text(json.dumps(rows,indent=2)+'\n');assert all(x['status']!='unclassified'for x in rows),rows
