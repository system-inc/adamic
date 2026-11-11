import subprocess, resource, time, json, sys
start=time.monotonic()
code=subprocess.run(sys.argv[2:]).returncode
r=resource.getrusage(resource.RUSAGE_CHILDREN)
open(sys.argv[1],'w').write(json.dumps({'child_exit':code,'wall_seconds':time.monotonic()-start,'peak_rss_kib':r.ru_maxrss,'user_seconds':r.ru_utime,'system_seconds':r.ru_stime},indent=2)+'\n')
print('child_exit='+str(code),flush=True)
sys.exit(code if code>=0 else 128-code)
