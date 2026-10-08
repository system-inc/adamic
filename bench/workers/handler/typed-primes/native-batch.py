import subprocess,resource,time,sys,json
start=time.perf_counter()
p=subprocess.run(sys.argv[1:],capture_output=True,text=True,check=True)
print(json.dumps({'ms':(time.perf_counter()-start)*1000/20,'rssKiB':resource.getrusage(resource.RUSAGE_CHILDREN).ru_maxrss/(1024 if sys.platform == 'darwin' else 1),'count':int(p.stdout)}))
