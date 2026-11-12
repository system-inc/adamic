#!/usr/bin/env python3
"""Freeze completed child streams in deterministic gzip; never hide an incomplete run."""
import gzip,hashlib,json,pathlib,shutil,sys
scratch=pathlib.Path(sys.argv[1]);out=pathlib.Path(sys.argv[2]);out.mkdir(parents=True,exist_ok=True)
def freeze(src,dst):
 status=json.loads((src/'status.json').read_text());assert status['child_exit_code']==0 and not status['deadline_expired'],src
 dst.mkdir(parents=True,exist_ok=True);shutil.copyfile(src/'status.json',dst/'status.json')
 for name in ['census.jsonl','passes.jsonl','units.jsonl','stderr.log']:
  p=src/name
  if p.exists():(dst/(name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
for side in ['base','head']:
 for start in ['0','1500','3000']:
  dst=out/'census'/side/start;freeze(scratch/'runs'/side/start,dst)
  p=scratch/(side+'-'+start+'-trace.jsonl')
  (dst/'trace.jsonl.gz').write_bytes(gzip.compress(p.read_bytes() if p.exists() else b'',mtime=0))
shutil.copyfile(scratch/'inputs.json',out/'inputs.json')
provenance={'head':'efd9bff92ce62bc92463917c6b72882adf32280e','base':'4bba06cb2dc4a0046c11fc97e9953ec85ed0e638','prototype':'4cadb42b6f48af2e4a99f334da775f308ea88818','historical_inventory_commit':'78933e52b0abd0da2fce91e73b4b71c1b36dc4e4','merged_tree':(scratch/'merged-tree.txt').read_text().strip(),'binaries':{}}
for side in ['base','head']:
 p=scratch/(side+'-census');provenance['binaries'][side]={'sha256':hashlib.sha256(p.read_bytes()).hexdigest(),'bytes':p.stat().st_size}
 for name in ['hook.go','predicates_checked.go','trace.go']:
  p=scratch/(side+'-final')/name;dst=out/'overlays'/side/(name+'.txt');dst.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(p,dst)
 for name in [side+'-build-final.log',side+'-overlay.log']:
  p=scratch/name;(out/(name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
(out/'provenance.json').write_text(json.dumps(provenance,indent=2)+'\n')
(out/'apply.log.gz').write_bytes(gzip.compress((scratch/'apply.log').read_bytes(),mtime=0))
manifest={str(p.relative_to(out)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(out.rglob('*')) if p.is_file() and p.name!='sha256.json'}
(out/'sha256.json').write_text(json.dumps(manifest,indent=2)+'\n')
