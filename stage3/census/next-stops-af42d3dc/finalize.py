"""Prepare and validate the final report without creating or pushing its commit."""
import json,os,sys
from pathlib import Path
from supervise import run,scratch
repo=Path('/workspace/adamic');out=Path(__file__).resolve().parent;core=repo/'stage3/census/core-sites-af42d3dc'
assert json.loads((scratch/'groups-status.json').read_text())['exit_code']==0
assert sum(len(json.loads(p.read_text())) for p in out.glob('group-*-rows.json'))==79
for d in [out,core]:
 data=json.loads((d/'provenance.json').read_text());data.update(setup_status=json.loads((scratch/'setup-status.json').read_text()),core_census_status=json.loads((scratch/'core-census-status.json').read_text()),load_end=os.getloadavg(),last_checkpoint=json.loads((out/'CHECKPOINT.json').read_text()))
 (d/'provenance.json').write_text(json.dumps(data,indent=2)+'\n')
for name,cmd in [
 ('final-summary',[sys.executable,str(out/'summarize.py')]),
 ('source-excerpts',['node',str(out/'source-excerpts.cjs')]),
 ('new-stop-map',[sys.executable,str(out/'map-new-stops.py')]),
 ('core-render',[sys.executable,str(core/'write-report.py')]),
 ('next-render',[sys.executable,str(out/'write-report.py')]),
 ('archive',[sys.executable,str(out/'archive-evidence.py')]),
 ('validation',[sys.executable,str(out/'validate.py')])]:
 result=run(name,cmd,limit=1800)
 if result['exit_code']:sys.exit(1)
print('Final report is prepared and validated; no commit created by this script.',flush=True)
