"""Join file schedules only when every original root is present and overlaps agree."""
import argparse,json
from pathlib import Path
p=argparse.ArgumentParser(description=__doc__);p.add_argument('output',type=Path);p.add_argument('inputs',type=Path,nargs='+');a=p.parse_args()
def normalized(row):
    row=dict(row)
    for key in ['findings','diagnostics','diagnostic_sites']:
        if key in row:
            row[key]=sorted(row[key],key=lambda item:json.dumps(item,sort_keys=True))
    return row
header=None;files={};duplicates=0
for source in a.inputs:
    rows=[json.loads(line) for line in source.read_text().splitlines()]
    assert rows and rows[0]['latent_mode']=='full'
    if header is None:header=rows[0]
    else:assert normalized(header)==normalized(rows[0]),'checker program/header mismatch'
    for row in rows[1:]:
        name=row['file']
        if name in files:
            assert normalized(files[name])==normalized(row),('overlap mismatch',name)
            duplicates+=1
        else:files[name]=row
expected=header['root_files']
assert len(set(expected))==len(expected)
assert set(files)==set(expected),('incomplete root coverage',sorted(set(expected)-set(files)),sorted(set(files)-set(expected)))
a.output.write_text(''.join(json.dumps(row)+'\n' for row in [header]+[files[name] for name in sorted(expected)]))
print(f'PASS: {len(files)} original roots, identical checker headers, {duplicates} agreeing overlap records')
