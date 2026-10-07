#!/usr/bin/env python3
"""Archive the rebased gate streams, fixtures, command records and timing medians."""
import gzip
import hashlib
import json
import os
import pathlib
import statistics
import subprocess

ROOT = pathlib.Path(__file__).resolve().parents[4]
UNIT = pathlib.Path(__file__).resolve().parent
OUT = pathlib.Path(os.environ.get('ADAMIC_WAVE07_LANDING', '/workspace/wave-07-landing'))
EVIDENCE = pathlib.Path(os.environ.get('ADAMIC_WAVE07_EVIDENCE', str(UNIT / 'landing-evidence')))
EVIDENCE.mkdir(exist_ok=True)
index = {}

sources = [('gate',OUT),('original-trio',OUT/'original-trio'),('bench',OUT/'bench'),
    ('timer',pathlib.Path('/workspace/wave-07-timer')),
    ('process',pathlib.Path('/workspace/wave-07-process')),
    ('streams',pathlib.Path('/workspace/wave-07-streams-final')),
    ('rest',pathlib.Path('/workspace/wave-07-next-rest')),
    ('promise',pathlib.Path('/workspace/wave-07-next-promise')),
    ('regex',pathlib.Path('/workspace/wave-07-next-regex')),
    ('continuation-questions',pathlib.Path('/workspace/wave-07-questions')),
    ('regex-question',pathlib.Path('/workspace/wave-07-next-questions')),
    ('react',pathlib.Path('/workspace/wave-07-react-probe')),
    ('listeners',pathlib.Path('/workspace/wave-07-listeners'))]
for label,directory in sources:
    target = EVIDENCE / label
    target.mkdir(exist_ok=True)
    for path in sorted(directory.iterdir()):
        if not path.is_file() or path.stat().st_size > 2_000_000:
            continue
        if path.suffix not in ['.stdout','.stderr','.log','.json','.manifest','.a','.go','.txt']:
            continue
        content = path.read_bytes()
        try:
            content.decode('utf-8')
        except UnicodeDecodeError:
            continue
        if b'\0' in content:
            continue
        with gzip.open(target/(path.name+'.gz'),'wb') as output:
            output.write(content)
        index[label+'/'+path.name] = dict(sha256=hashlib.sha256(content).hexdigest(),bytes=len(content))
for name in ['landing-fetch','landing-rebase','landing-rebase-current','landing-setup','landing-gates','landing-gates-current','speed-fetch','speed-setup','listener-gate','speed-landing-gates']:
    path = pathlib.Path('/workspace/wave-07-artifacts')/(name+'.log')
    if not path.exists():
        continue
    content = path.read_bytes()
    with gzip.open(EVIDENCE/(name+'.log.gz'),'wb') as output:
        output.write(content)
    index[path.name] = dict(sha256=hashlib.sha256(content).hexdigest(),bytes=len(content))
(EVIDENCE/'sha256.json').write_text(json.dumps(index,indent=2)+'\n')
records=json.loads((OUT/'bench/records.json').read_text())
medians=[]
for unit in dict.fromkeys(row['unit'] for row in records):
    for corpus in ['repository','compiler']:
        values={implementation:statistics.median(row['elapsed_ns']/1e9 for row in records
            if row['unit']==unit and row['corpus']==corpus and row['implementation']==implementation)
            for implementation in ['native','go']}
        medians.append(dict(unit=unit,corpus=corpus,**values))
(EVIDENCE/'medians.json').write_text(json.dumps(medians,indent=2)+'\n')
metadata={name:subprocess.check_output(['git',*arguments],cwd=ROOT).decode().strip()
          for name,arguments in [('tested_head',['rev-parse','HEAD']),
                                 ('main',['rev-parse','origin/main']),
                                 ('branch',['branch','--show-current'])]}
metadata.update(original_pushed_head=os.environ.get('ADAMIC_WAVE07_PREVIOUS_HEAD', 'b9e43baf5abcf7dfc8ee8e3322682a426031847a'),
    oracle_cohere=subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT/'cohere').decode().strip(),
    typescript_go=subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT/'cohere/TypeScript').decode().strip(),
    compiler_corpus=subprocess.check_output(['git','rev-parse','HEAD'],cwd='/workspace/wave-07-typescript').decode().strip())
(EVIDENCE/'metadata.json').write_text(json.dumps(metadata,indent=2)+'\n')
print('Archived '+str(len(index))+' streams, fixture files and command records.')
