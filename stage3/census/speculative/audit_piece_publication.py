"""Reject a piece tag on otherwise valid full-file publication inputs.
Usage: audit_piece_publication.py STREAM_AUDIT_OUTPUT OUTPUT
"""
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import sys
source,output=(Path(p).resolve() for p in sys.argv[1:3])
output.mkdir(parents=True,exist_ok=True)
scripts=Path(__file__).resolve().parent
root=source/'source'
raw=output/'piece-tag-mutant.jsonl'
rows=[json.loads(line) for line in (source/'result/speculative.jsonl').read_text().splitlines()]
rows[1]['piece']=dict(id=0,atoms=['synthetic partial-piece tag'],plan_sha256='0'*64)
raw.write_text(''.join(json.dumps(row)+'\n' for row in rows))
run=output/'run'
shutil.copytree(source/'run',run,dirs_exist_ok=True)
record=next((run/'records').glob('*.jsonl'))
records=[json.loads(line) for line in record.read_text().splitlines()]
records[1]['piece']=rows[1]['piece']
record.write_text(''.join(json.dumps(row)+'\n' for row in records))
record.with_suffix('.sha256').write_text(hashlib.sha256(record.read_bytes()).hexdigest()+'\n')
commands=[
 ('finalizer',[sys.executable,str(scripts/'finalize_stream.py'),str(root),str(run),str(output/'finalize')],'partial piece cannot claim complete file coverage'),
 ('reporter',[sys.executable,str(scripts/'report.py'),str(root),str(raw),str(source/'result/stock.json'),str(output/'report')],'piece records cannot claim complete file coverage'),
 ('verifier',[sys.executable,str(scripts/'verify.py'),str(raw),str(source/'result/RESULT.json'),str(root)],'piece records cannot claim complete file coverage')]
for name,command,message in commands:
    with (output/(name+'.log')).open('w') as log:
        code=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=10).returncode
    assert code!=0 and message in (output/(name+'.log')).read_text(),name+' accepted a piece as a complete file'
    print(name+' caught piece-as-complete-file mutant')
