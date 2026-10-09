import pathlib,os,json,subprocess,time
p=pathlib.Path('/workspace/adamic/review/test-defend/internal-native-radix');env=os.environ.copy();env['ADAMIC_RECORD_BENCH']='1';env['TMPDIR']='/workspace/scratch/defend-radix/tmp';env['PATH']='/workspace/scratch/defend-radix/wrappers:'+env['PATH'];env['XDG_CACHE_HOME']='/workspace/scratch/defend-radix/instrumented-cache';env['ADAMIC_BUILD_CACHE_DIR']='/workspace/scratch/defend-radix/cache/coverage';runs=[]
for name in ['TestRecordBenchmark','TestRuntimeStringEquality','TestRuntimeReleasePaths','TestRegExpIteratorResultShape']:
 d=p/'c-coverage'/name;d.mkdir(parents=True,exist_ok=True);env['DEFENSE_C_BIN']=str(d/'bins');env['LLVM_PROFILE_FILE']=str(d/'%p.profraw')
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverpkg=github.com/system-inc/adamic/internal/native','-coverprofile='+str(p/(name+'.cover')),'./internal/native/','-run','^'+name+'$'];start=time.monotonic()
 with (p/(name+'-coverage.log')).open('w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 runs.append(dict(test=name,command=cmd,exit=r.returncode,wall=time.monotonic()-start));(p/'coverage-runs.json').write_text(json.dumps(runs,indent=2));print(name,r.returncode,flush=True)
 if r.returncode:break
 profiles=list(d.glob('*.profraw'));subprocess.run(['/workspace/adamic-tools/llvm/bin/llvm-profdata','merge','-sparse','-o',str(d/'merged.profdata')]+list(map(str,profiles)),check=True)
 bins=list((d/'bins').iterdir());command=['/workspace/adamic-tools/llvm/bin/llvm-cov','export',str(bins[0]),'-instr-profile='+str(d/'merged.profdata')]
 for b in bins[1:]:command+=['-object',str(b)]
 with (d/'export.json').open('w') as f:subprocess.run(command,stdout=f,check=True)
