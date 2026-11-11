import json,resource,subprocess,sys,time,pathlib
prefix=sys.argv[1]; start=time.monotonic()
with open(prefix+'.log','w') as log:
 p=subprocess.Popen(sys.argv[2:],stdout=log,stderr=subprocess.STDOUT)
 _,status,usage=__import__('os').wait4(p.pid,0)
 p.returncode=__import__('os').waitstatus_to_exitcode(status)
result=dict(command=sys.argv[2:],exit=p.returncode,wall_seconds=time.monotonic()-start,peak_rss_kib=usage.ru_maxrss,user_seconds=usage.ru_utime,system_seconds=usage.ru_stime)
pathlib.Path(prefix+'.time.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result),flush=True)
sys.exit(p.returncode)
