from pathlib import Path
import subprocess,os,time,json
s=Path('/workspace/loose-slot-scratch/eleven');out=Path(__file__).resolve().parent
for which in ['head','base']:
 r=s/which
 while not (r/'cohere/TypeScript/tsc/go.mod').exists():time.sleep(3)
 # Wait for the setup command's exact pin checkout, not merely clone creation.
 pin=subprocess.check_output(['git','ls-tree','HEAD','cohere'],cwd=r,text=True).split()[2]
 while subprocess.check_output(['git','rev-parse','HEAD'],cwd=r/'cohere',text=True).strip()!=pin:time.sleep(3)
 with (out/'evidence'/f'npm-{which}.log').open('wb')as log:subprocess.run(['npm','ci','--prefix','stage3/api'],cwd=r,stdout=log,stderr=log,check=True)
 with (out/'evidence'/f'apply-{which}.log').open('wb')as log:subprocess.run(['./stage3/apply.sh',str(s/('adapted-'+which))],cwd=r,stdout=log,stderr=log,check=True)
 (s/(which+'-input.ready')).write_text('complete\n')
 print(which,'adaptation complete',flush=True)
