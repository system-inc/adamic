import sys,subprocess,time,resource,json
log,stats,*cmd=sys.argv[1:]
t=time.monotonic()
with open(log,'wb') as out:
 p=subprocess.Popen(cmd,stdout=out,stderr=subprocess.STDOUT)
 rc=p.wait()
r=resource.getrusage(resource.RUSAGE_CHILDREN)
d=dict(command=cmd,exit_code=rc,wall_seconds=time.monotonic()-t,user_seconds=r.ru_utime,system_seconds=r.ru_stime,peak_rss_kib=r.ru_maxrss)
open(stats,'w').write(json.dumps(d,indent=2)+'\n')
print('child_exit='+str(rc),flush=True)
sys.exit(rc if rc>=0 else 128-rc)
