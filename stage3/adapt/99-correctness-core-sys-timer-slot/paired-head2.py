import sys
sys.dont_write_bytecode=True
from run import *
import hashlib,re
rows=[]
def diff(tree,name):
 args=['git','-C',str(tree),'diff','--no-ext-diff','--no-renames','--no-prefix','--unified=0','refs/stage3/'+name+'-before','refs/stage3/'+name+'-after']
 p=subprocess.run(args,stdout=subprocess.PIPE,stderr=subprocess.PIPE);assert p.returncode==0
 text='\n'.join(line for line in p.stdout.decode().splitlines() if not line.startswith('index '))
 return text,p.returncode
for directory in sorted((ROOT/'stage3/adapt').iterdir()):
 if not directory.is_dir():continue
 name=directory.name
 if name=='00-setup':continue
 b,eb=diff(S/'head2',name)
 if name=='99-correctness-core-sys-timer-slot':
  assert sum(x.startswith('+') and not x.startswith('+++') for x in b.splitlines())==4
  rows.append({'adaptation':name,'newDirectory':True,'afterExit':eb,'diffSha256':hashlib.sha256(b.encode()).hexdigest()});continue
 a,ea=diff(S/'head1',name);assert a==b,name
 rows.append({'adaptation':name,'beforeExit':ea,'afterExit':eb,'identicalIncrementalPatch':True,'diffSha256':hashlib.sha256(b.encode()).hexdigest()})
(S/'head2-paired-replay.json').write_text(json.dumps({'allExistingIncrementalPatchesIdentical':True,'newDirectoryOnly':True,'rows':rows},indent=2)+'\n')
print('paired replay child_exit=0 for every recorded git diff; existing adapters identical',flush=True)
