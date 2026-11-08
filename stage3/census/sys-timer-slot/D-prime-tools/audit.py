import sys
sys.dont_write_bytecode=True
from run import *
import hashlib,os
base='d7af5ed347d9b76eb418c65a8a0f875ae89bfc46';out=ROOT/'stage3/census/sys-timer-slot'
def capture(args):
 p=subprocess.run(args,cwd=ROOT,stdout=subprocess.PIPE,stderr=subprocess.PIPE);assert p.returncode==0,(args,p.stderr.decode());return p.stdout
old=capture(['git','show',base+':stage3/census/sys-timer-slot/REPORT.md']);assert (out/'REPORT.md').read_bytes().startswith(old)
a=json.loads(capture(['git','show',base+':stage3/census/sys-timer-slot/REPORT.json']));b=json.loads((out/'REPORT.json').read_text());assert all(b[k]==v for k,v in a.items())
unchanged=0
for line in capture(['git','ls-tree','-r',base,'stage3/census/sys-timer-slot']).decode().splitlines():
 meta,name=line.split('\t')
 if name in ['stage3/census/sys-timer-slot/REPORT.md','stage3/census/sys-timer-slot/REPORT.json']:continue
 raw=(ROOT/name).read_bytes();assert hashlib.sha1(b'blob '+str(len(raw)).encode()+b'\0'+raw).hexdigest()==meta.split()[2],name;unchanged+=1
result=json.loads((out/'D-prime.json').read_text());assert result['stock']['adamic']==[] and result['stock']['solution']==[] and result['stock']['solutionExit']==0
assert result['census']['providerUnit']['status']=='attempted';assert len(result['census']['gone'])==2 and result['census']['new']==[]
assert hashlib.sha256((out/'D-prime-evidence.tar.gz').read_bytes()).hexdigest()==(out/'D-prime-evidence.sha256').read_text().split()[0]
for name,h in result['publishedAndTrackedHostInputsUnchanged'].items():assert hashlib.sha256(pathlib.Path(name).read_bytes()).hexdigest()==h
exclude={os.getpid()};pid=os.getppid()
while pid>1:
 exclude.add(pid)
 try:pid=int(pathlib.Path('/proc',str(pid),'stat').read_text().split(') ',1)[1].split()[1])
 except OSError:break
live=[]
for p in pathlib.Path('/proc').iterdir():
 if not p.name.isdigit() or int(p.name) in exclude:continue
 try:
  state=(p/'stat').read_text().split(') ',1)[1].split()[0];args=(p/'cmdline').read_bytes()
  if state!='Z' and b'/workspace/scratch/sys-timer-property-prime' in args:live.append(int(p.name))
 except OSError:pass
assert live==[],live
r={'status':'pass','existingEvidenceFilesUnchanged':unchanged,'oldMarkdownPrefixByteIdentical':True,'oldJsonValuesUnchanged':True,'trackedHostAndPublishedTimerDeclarationsUnchanged':True,'taskOwnedLiveProcesses':live,'captureChildExits':0}
(out/'D-prime-audit.json').write_text(json.dumps(r,indent=2)+'\n');print(json.dumps(r,indent=2))
