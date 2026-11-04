import subprocess,resource,json,sys,time
start=time.monotonic()
p=subprocess.run(sys.argv[2:])
r=resource.getrusage(resource.RUSAGE_CHILDREN)
with open(sys.argv[1],'w') as f: json.dump(dict(returncode=p.returncode,peak_rss_kib=r.ru_maxrss,wall_seconds=time.monotonic()-start,user_seconds=r.ru_utime,system_seconds=r.ru_stime),f,indent=2)
sys.exit(p.returncode if p.returncode>=0 else 128-p.returncode)
