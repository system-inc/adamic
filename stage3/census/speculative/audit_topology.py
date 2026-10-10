"""Real-lowering witness for complete boundaries, canonical sites and equal spans."""
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys

repo,overlay,binary,out=map(lambda s:Path(s).resolve(),sys.argv[1:5])
out.mkdir(parents=True,exist_ok=True)
source=out/'source';source.mkdir(exist_ok=True)
(source/'span.a').write_bytes((Path(__file__).parent/'control/span-identity.a').read_bytes())
def run(name,executable):
    raw=out/(name+'.jsonl')
    with (out/(name+'.log')).open('w') as log:
        subprocess.run([str(executable),str(source),str(raw)],env=dict(os.environ,LATENT_SPECULATIVE='1',LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1'),stdout=log,stderr=subprocess.STDOUT,check=True)
    return raw,[json.loads(line) for line in raw.read_text().splitlines()]
def build(name,old,new):
    folder=out/name;shutil.copytree(overlay,folder,dirs_exist_ok=True)
    path=folder/'internal_lower_latent_speculative.go';text=path.read_text();assert text.count(old)==1
    path.write_text(text.replace(old,new))
    config=json.loads((overlay/'overlay.json').read_text());config['Replace']={key:str(folder/Path(value).name) for key,value in config['Replace'].items()}
    (folder/'overlay.json').write_text(json.dumps(config))
    executable=folder/'census'
    with (folder/'build.log').open('w') as log:
        subprocess.run(['go','build','-buildvcs=false','-overlay='+str(folder/'overlay.json'),'-o',str(executable),'./stage3/census/latent/tool'],cwd=repo,stdout=log,stderr=subprocess.STDOUT,check=True)
    return executable
def canonical(rows):
    assert not rows[0]['diagnostics']
    findings=[f for row in rows[1:] for f in row['findings']]
    assert len(findings)==8
    arrays=[f for f in findings if f['reason']=='an array of T']
    assert {(f['site_kind'],f['depth']) for f in arrays}=={('KindCallExpression',2),('KindArrayLiteralExpression',4),('KindElementAccessExpression',1)},arrays
    value=next(f for f in arrays if f['site_kind']=='KindCallExpression')
    assert value['site_kind']=='KindCallExpression' and value['depth']==2,value
    assert all(row['speculative_coverage']['unvisited_nodes']==0 for row in rows[1:])
def report(name,raw,script=None):
    with (out/(name+'-report.log')).open('w') as log:
        return subprocess.run([sys.executable,str(script or Path(__file__).parent/'report.py'),str(source),str(raw),str(out/'stock.json'),str(out/name)],stdout=log,stderr=subprocess.STDOUT).returncode
raw,rows=run('baseline',binary);canonical(rows)
with (out/'stock.log').open('w') as log:
    subprocess.run(['node',str(Path(__file__).parent/'stock.cjs'),str(source),str(raw),str(out/'stock.json')],stdout=log,stderr=subprocess.STDOUT,check=True)
assert report('baseline',raw)==0
print('baseline: eight checker-clean AST sites, three distinct array failures, equal-span identifier depth 4; stock audit PASS')
executable=build('omit-boundaries-mutant','record.FailedBoundaries = boundaryInventory[record.File]','record.FailedBoundaries = nil')
mutant,_=run('omit-boundaries-mutant',executable)
assert report('omit-boundaries-mutant',mutant)!=0
print('omit-boundaries mutant caught by independent stock ancestry')
executable=build('legacy-site-key-mutant',r'return fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00%d", latentFindingKey(finding), finding.SiteWhere, finding.SiteKind, finding.SiteStart, finding.SiteEnd)', 'return latentFindingKey(finding)')
_,changed=run('legacy-site-key-mutant',executable)
try:canonical(changed)
except AssertionError:print('legacy-site-key mutant caught: distinct call, array and element-access sites collapse')
else:raise AssertionError('legacy-site-key mutant survived')
script=out/'span-only-report.py';text=(Path(__file__).parent/'report.py').read_text()
old="(boundary['start'], boundary['end'], kind_codes[kind])";assert text.count(old)==1
text=text.replace(old,"(boundary['start'], boundary['end'])")
old='(*span, kind) in boundaries';assert text.count(old)==1
script.write_text(text.replace(old,'tuple(span) in boundaries'))
assert report('span-only-mutant',raw,script)!=0
print('span-only mutant caught: parameter and identifier share a span but only identifier failed')
print('PASS: complete typed boundary inventory, distinct AST finding attribution, equal-span kind discrimination; all three mutants caught')
