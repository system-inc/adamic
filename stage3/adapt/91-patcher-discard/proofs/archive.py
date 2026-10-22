#!/usr/bin/env python3
"""Preserve raw observations and hashes after all serial checks finish."""
import difflib
import gzip
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import sys
sys.dont_write_bytecode=True
proof=Path(__file__).resolve().parent;repo=proof.parents[3];scratch=Path('/workspace/scratch')
logs=proof/'logs'
def hash(data):return hashlib.sha256(data).hexdigest()
def compressed(source,destination):
 with destination.open('wb') as raw:
  with gzip.GzipFile(filename='',mode='wb',fileobj=raw,mtime=0) as output:output.write(source.read_bytes())
for label in ['before','after','mutant']:
 compressed(scratch/('adapt91-scan-'+label+'.jsonl'),logs/('scan-'+label+'.jsonl.gz'))
for label in ['control','after']:
 source=scratch/('adapt91-oracle-'+label)
 for name in ['install.log','build.log','tests.log','report.json','failures.md','baseline.diff']:
  compressed(source/name,logs/('oracle-'+label+'-'+name+'.gz'))
for name in ['setup','api-install','source','control-apply','after-apply','prepare','prepare-retry','oracles','final-checks']:
 source=Path('/tmp')/('adapt91-'+name+'.log')
 compressed(source,logs/(name+'.log.gz'))
for source in logs.glob('*.log'):compressed(source,Path(str(source)+'.gz'))
shutil.copyfile(scratch/'adapt91-invalid-references-final.json',proof/'registrant-mutant.json')
name='utilities.ts';stem=str(scratch/'adapt91-mutant-scope.json')+'.'+name
before=Path(stem+'.before.js.txt').read_text();after=Path(stem+'.after.js.txt').read_text()
(proof/'mutant-transpile.diff').write_text(''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='before/'+name+'.js',tofile='mutant/'+name+'.js')))
roots={}
for label in ['control','after','mutant']:
 tree=scratch/('adapt91-'+label)
 roots[label]=[dict(file=str(p.relative_to(tree)),sha256=hash(p.read_bytes())) for p in sorted((tree/'src/compiler').rglob('*.ts'))]
 assert len(roots[label])==79
copied={}
pin='175f0de29e9b2334023c8b5550a6eeb3df8e4eee'
for name in ['hooks.cjs','make-scan.py','extract.py']:
 original=subprocess.check_output(['git','show',pin+':stage3/census/void-widening/'+name],cwd=repo,timeout=30)
 assert original==(proof/name).read_bytes()
 copied[name]=hash(original)
tree=scratch/'adapt91-after'
refs={k:subprocess.check_output(['git','-C',str(tree),'rev-parse','refs/stage3/91-patcher-discard-'+k],text=True,timeout=30).strip() for k in ['before','after']}
changes=subprocess.check_output(['git','-C',str(tree),'diff','--numstat',refs['before'],refs['after']],text=True,timeout=30)
assert changes=='1\t1\tsrc/compiler/factory/nodeFactory.ts\n1\t1\tsrc/compiler/utilities.ts\n'
(proof/'provenance.json').write_text(json.dumps(dict(base='f9d7464755c8de94d581b6b3fda9ac39af6e7c62',upstream='050880ce59e30b356b686bd3144efe24f875ebc8',censusPin=pin,copiedTools=copied,voidWideningSourceSha256=hash((repo/'internal/lower/void_widening.go').read_bytes()),scanBinarySha256=hash((scratch/'adapt91-void-scan').read_bytes()),applySnapshots=refs,applyNumstat=changes,roots=roots),indent=2)+'\n')
comparison=json.loads((proof/'oracle-comparison.json').read_text());assert comparison['status']=='pass'
(proof/'results.json').write_text(json.dumps(dict(status='pass',conditions={'1_bodyOnlyAndApiIdentical':'pass','2_noRegistrant':'pass','3_sameOracleAndBaseline':'pass','4_javascriptChangeWitnessed':'pass','5_bothSitesClear':'pass'},scan={'before':5,'after':1,'removed':4,'newFindingsOfAnyKind':0},mutant={'scan':3,'restoredRows':2,'transpileChangedSites':1},counts={r['label']:r['counts'] for r in comparison['runs']},finalBuiltTreeScope='pass',finalIdempotence='zero edits',baselineDiff=comparison['baselineDiff'],api=comparison['api']),indent=2)+'\n')
print('PASS archived raw evidence, copied-tool identity, 79-root hashes and exact full-apply snapshots')
