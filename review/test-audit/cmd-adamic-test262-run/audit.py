import pathlib,json,subprocess,time,os,difflib
r=pathlib.Path('/workspace/adamic');p=pathlib.Path('/tmp/u014');path='cmd/adamic-test262/run.go';f=r/path;base=f.read_text();(p/'base-source.txt').write_text(base)
changes=[('M1','int64((cpuLimit + time.Second - 1) / time.Second)','int64(cpuLimit / time.Second)','off-by-one rounding'),('M2','status.Signal() == syscall.SIGXCPU','status.Signal() != syscall.SIGXCPU','flip condition'),('M3','return runCommand(2*time.Minute, extra, "/bin/sh", arguments...)','return runCommand(cpuLimit, extra, "/bin/sh", arguments...)','change option'),('M4','if written > room {','if written >= room {','off-by-one bound'),('E_RUN_PROGRAM','func runProgram(cpuLimit time.Duration, extra []string, name string, args ...string) execution {','func runProgram(cpuLimit time.Duration, extra []string, name string, args ...string) execution {\n if true { return execution{} }','empty-answer probe')]
ms=[]
def run(cmd,label,env=None):
 t=time.monotonic()
 with open(p/(label+'.log'),'w') as log:x=subprocess.run(cmd,cwd=r,stdout=log,stderr=subprocess.STDOUT,env=env)
 z=dict(name=label,command=cmd,seconds=time.monotonic()-t,exit=x.returncode,env={k:env[k] for k in ['ADAMIC_MUTANT','ADAMIC_TEST262_MEASURE','ADAMIC_GATE_UNCACHED'] if env and k in env})
 with open(p/'commands.jsonl','a') as log:log.write(json.dumps(z)+'\n')
 return z
for id,old,new,menu in changes:
 assert base.count(old)==1
 line=base[:base.index(old)].count('\n')+1
 ms.append(dict(id=id,file=path,line=line,old=old,new=new,menu=menu,probe=id.startswith('E_')))
(p/'mutants.json').write_text(json.dumps(ms,indent=2))
# Save and compile-check independent changes, always restoring the base.
for m in ms:
 changed=base.replace(m['old'],m['new']);diff=''.join(difflib.unified_diff(base.splitlines(True),changed.splitlines(True),fromfile='a/'+path,tofile='b/'+path));(p/(m['id']+'.diff')).write_text(diff)
 f.write_text(changed);x=run(['go','vet','./cmd/adamic-test262/'],m['id']+'-vet');f.write_text(base);assert x['exit']==0,m['id']
sw=base.replace(changes[0][1],'func() int64 { if os.Getenv("ADAMIC_MUTANT") == "M1" { return int64(cpuLimit / time.Second) }; return int64((cpuLimit + time.Second - 1) / time.Second) }()')
sw=sw.replace(changes[1][1],'(os.Getenv("ADAMIC_MUTANT") == "M2") != (status.Signal() == syscall.SIGXCPU)')
sw=sw.replace(changes[2][1],'wallLimit := 2*time.Minute\n if os.Getenv("ADAMIC_MUTANT") == "M3" { wallLimit = cpuLimit }\n return runCommand(wallLimit, extra, "/bin/sh", arguments...)')
sw=sw.replace(changes[3][1],'if written > room || (os.Getenv("ADAMIC_MUTANT") == "M4" && written == room) {')
sw=sw.replace(changes[4][1],changes[4][2].replace('if true','if os.Getenv("ADAMIC_MUTANT") == "E_RUN_PROGRAM"'))
f.write_text(sw);assert run(['gofmt','-w',path],'switch-format')['exit']==0
assert run(['go','test','-c','-o',str(p/'runner.test'),'./cmd/adamic-test262/'],'switch-build')['exit']==0
(p/'switch.diff').write_text(subprocess.check_output(['git','diff','--',path],cwd=r,text=True))
for m in ms:
 env=os.environ.copy();env.update(ADAMIC_MUTANT=m['id'],ADAMIC_TEST262_MEASURE='1',ADAMIC_GATE_UNCACHED='1')
 run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./cmd/adamic-test262/','-run','.'],m['id']+'-matrix',env)
print('complete')
