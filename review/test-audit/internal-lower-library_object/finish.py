import pathlib,subprocess,json,os,time,shutil
root=pathlib.Path('/workspace/adamic');p=root/'review/test-audit/internal-lower-library_object';os.chdir(root)
# Replay correction: initial switch changed both identically-spelled freeze conditions,
# while M19.diff changes only the first occurrence. Repeat the actual standalone mutant.
shutil.copyfile(p/'M19.log',p/'M19-two-occurrences.log')
meta=json.loads((p/'run-meta.json').read_text());meta['M19-two-occurrences']=meta['M19']
original=subprocess.check_output(['git','show','origin/main:internal/lower/library_object.go']).decode()
(root/'internal/lower/library_object.go').write_text(original.replace('call.Method == "freeze"','call.Method != "freeze"',1))
cmd='timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .';env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u036/cache/M19-corrected';t=time.monotonic()
with (p/'M19.log').open('w') as log:r=subprocess.run(cmd,shell=True,env=env,stdout=log,stderr=subprocess.STDOUT)
meta['M19']=dict(command=cmd,status=r.returncode,wall=time.monotonic()-t,cache=env['ADAMIC_BUILD_CACHE_DIR'],standalone=True)
(root/'internal/lower/library_object.go').write_text(original)
(p/'run-meta.json').write_text(json.dumps(meta,indent=2)+'\n')
# Output-only observation of internal queries, added only after the matrix ended.
# The additional scratch Test is not a mutant catcher and has no verdict.
subprocess.run(['git','apply',str(p/'switch.diff')],check=True)
query=root/'internal/lower/u036_query_witness_test.go';query.write_text((p/'query-witness.go.txt').read_text())
for mid in ['clean','M08','M17']:
 env=os.environ.copy()
 if mid!='clean':env['ADAMIC_MUTANT']=mid
 env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u036/cache/'+mid+'-query'
 with (p/('query-'+mid+'.log')).open('w') as log:
  subprocess.run("timeout 120 go test -v -count=1 -timeout 90s ./internal/lower/ -run '^TestU036QueryWitness$'",shell=True,env=env,stdout=log,stderr=subprocess.STDOUT)
query.unlink()
shutil.copyfile(p/'switch.diff',p/'switch-initial.diff.txt')
f=root/'internal/lower/library_object.go';lines=f.read_text().splitlines(True);hits=[i for i,line in enumerate(lines) if 'ADAMIC_MUTANT' in line and 'M19' in line];assert len(hits)==2
i=hits[-1];indent=lines[i][:len(lines[i])-len(lines[i].lstrip())];lines[i]=indent+'if call, ok := node.(ir.ObjectCall); ok && call.Method == "freeze" {\n';f.write_text(''.join(lines))
subprocess.run(['gofmt','-w',str(f)],check=True)
(p/'switch.diff').write_text(subprocess.check_output(['git','diff','--','internal/lower']).decode())
files=['diagnostics.go','invariance.go','library_object.go','library_regex_callback_shape.go','lower.go','object.go','prelude.go','refusals.go','statements.go']
for f in files:
 s=subprocess.check_output(['git','show','origin/main:internal/lower/'+f]).decode();(root/'internal/lower'/f).write_text(s)
with (p/'final-clean.log').open('w') as log:
 pattern='^('+'|'.join(json.loads((p/'scope.json').read_text()))+')$'
 subprocess.run('timeout 120 go test -count=1 -timeout 90s ./internal/lower/ -run '+"'"+pattern+"'",shell=True,stdout=log,stderr=subprocess.STDOUT)
print('correction, witnesses and restoration complete')
