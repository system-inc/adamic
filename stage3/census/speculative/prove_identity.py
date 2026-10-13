"""A token-position attribution mutant changes only the async root's depth."""
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys

repository, overlay, binary, output = map(lambda s:Path(s).resolve(),sys.argv[1:5])
output.mkdir(parents=True,exist_ok=True)
source=output/'source'
source.mkdir(exist_ok=True)
(source/'identity.a').write_bytes((Path(__file__).parent/'control/identity.a').read_bytes())
folder=output/'token-mutant'
shutil.copytree(overlay,folder,dirs_exist_ok=True)
path=folder/'internal_lower_latent_speculative.go'
text=path.read_text()
old='node := sites[latentSpecFindingKey(finding)]'
assert text.count(old)==1
path.write_text(text.replace(old,'node := nodes[finding.Where]'))
configuration=json.loads((overlay/'overlay.json').read_text())
configuration['Replace']={key:str(folder/Path(value).name) for key,value in configuration['Replace'].items()}
(folder/'overlay.json').write_text(json.dumps(configuration,indent=2)+'\n')
mutant=folder/'census'
with (folder/'build.log').open('w') as log:
    subprocess.run(['go','build','-buildvcs=false','-overlay='+str(folder/'overlay.json'),'-o',str(mutant),'./stage3/census/latent/tool'],cwd=repository,stdout=log,stderr=subprocess.STDOUT,check=True)
runs={}
for name, executable in [('speculative',binary),('token-mutant',mutant)]:
    destination=output/(name+'.jsonl')
    with (output/(name+'.log')).open('w') as log:
        subprocess.run([str(executable),str(source),str(destination)],env=dict(os.environ,LATENT_SPECULATIVE='1',LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1'),stdout=log,stderr=subprocess.STDOUT,check=True)
    runs[name]=[json.loads(line) for line in destination.read_text().splitlines()]
def depths(rows):
    return [(f['reason'],f['depth']) for row in rows[1:] for f in row['findings'] if f['kind']=='Refused']
assert depths(runs['speculative'])==[('an async function',0),('await',1)]
assert depths(runs['token-mutant'])==[('an async function',1),('await',1)]
try:
    assert depths(runs['token-mutant'])==[('an async function',0),('await',1)]
except AssertionError:
    print('token-position identity mutant caught: async root incorrectly depth 1; await child remains depth 1')
else:
    raise AssertionError('token-position identity mutant survived')
print('PASS: exact AST identity; failed generic signature continued into await; output guards')
