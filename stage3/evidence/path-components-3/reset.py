import subprocess,sys
from pathlib import Path
tree=sys.argv[1]
for command in [['git','-C',tree,'reset','--hard','HEAD'],['git','-C',tree,'clean','-fdx','--','src']]:
 r=subprocess.run(command);print('child_exit='+str(r.returncode),flush=True)
 if r.returncode:sys.exit(r.returncode)
