#!/usr/bin/env python3
import hashlib, json, pathlib, subprocess, sys, os
root=pathlib.Path.cwd(); scratch=pathlib.Path('/tmp/decode-ascii'); stage=sys.argv[1]
out=scratch/('objects-'+stage); out.mkdir(exist_ok=True)
for source in sorted((root/'internal/native/runtime').glob('*.c')):
    subprocess.run(['clang','-std=c11','-Wall','-Wextra','-Werror','-pedantic','-O2','-ffp-contract=off','-fno-optimize-sibling-calls','-c',str(source),'-o',str(out/(source.stem+'.o'))],check=True)
if stage=='after':
    rows=[]
    for obj in sorted(out.glob('*.o')):
        before=(scratch/'objects-before'/obj.name).read_bytes(); after=obj.read_bytes()
        rows.append(dict(object=obj.name,identical=before==after,before=hashlib.sha256(before).hexdigest(),after=hashlib.sha256(after).hexdigest()))
    assert [r['object'] for r in rows if not r['identical']]==['input.o']
    (root/os.environ.get('ADAMIC_DECODE_OBJECT_RESULTS','cloud/reports/decode-ascii/objects.json')).write_text(json.dumps(rows,indent=2)+'\n')
    print(f"{len(rows)-1}/{len(rows)-1} non-input objects byte-identical; input.o differs")
