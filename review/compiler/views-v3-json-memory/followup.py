from pathlib import Path
import subprocess,os,time,json,hashlib
root=Path('/workspace/v3-json-artifacts/followup')
binaries={'v3':str(root/'v3-sanitized'),'main':'/workspace/v3-json-artifacts/main-killed-native/005/sanitized'}
results=[]
def measure(label,parallel):
 name=label+('-parallel' if parallel else '-alone')
 before=Path('/sys/fs/cgroup/memory.events').read_text(); jobs=[];maximum=0
 for i in (range(5) if parallel else [1]):
  cases=root/(f'chunk-{i}.txt' if parallel else 'exact-694-1389.txt')
  stdout=(root/f'{name}-{i}.stdout').open('wb');stderr=(root/f'{name}-{i}.stderr').open('wb')
  p=subprocess.Popen(['/tmp/v3-json-measure/rss',str(root/f'{name}-{i}.rss'),binaries[label],'--cases',str(cases)],env=dict(os.environ,ASAN_OPTIONS='detect_leaks=0'),stdout=stdout,stderr=stderr)
  jobs.append((i,p,stdout,stderr))
 while any(p.poll() is None for _,p,_,_ in jobs):
  maximum=max(maximum,int(Path('/sys/fs/cgroup/memory.current').read_text()));time.sleep(.05)
 rows=[]
 for i,p,stdout,stderr in jobs:
  p.wait();stdout.close();stderr.close()
  record=dict(line.split('=',1) for line in (root/f'{name}-{i}.rss').read_text().splitlines())
  record.update(chunk=i,stderr=(root/f'{name}-{i}.stderr').read_text(),stdout_sha256=hashlib.file_digest((root/f'{name}-{i}.stdout').open('rb'),'sha256').hexdigest())
  rows.append(record);print(name,i,record,flush=True)
 result={'name':name,'chunks':rows,'memory_events_before':before,'memory_events_after':Path('/sys/fs/cgroup/memory.events').read_text(),'sampled_cgroup_peak_bytes':maximum}
 results.append(result);(root/'measurements.json').write_text(json.dumps(results,indent=2)+'\n')
measure('v3',False)
measure('main',False)
measure('v3',True)
measure('main',True)
