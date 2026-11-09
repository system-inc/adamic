"""Finalize measurement artifacts without making any checker queries."""
import gzip,json,pathlib,shutil
root=pathlib.Path(__file__).resolve().parent
summary={}
for mode in ['real','census']:
 snapshot=json.loads((root/(mode+'-units.json')).read_text())
 keys={(x['cache'],x['key']) for x in snapshot['pairs']}
 path=root/(mode+'-pairs.json')
 rows=json.loads(path.read_text())
 original=root/(mode+'-pairs-render-output.json')
 if not original.exists():original.write_text(path.read_text())
 else:rows=json.loads(original.read_text())
 selected=[x for x in rows if (x['cache'],x['key']) in keys]
 excluded=[x for x in rows if (x['cache'],x['key']) not in keys]
 path.write_text(json.dumps(selected,indent=2)+'\n')
 summary[mode]={'render_output_pairs':len(rows),'pre_render_flagged_pairs_recovered':len(selected),'render_only_pairs_excluded':len(excluded)}
 raw=root/(mode+'-units.jsonl')
 if raw.exists():
  with raw.open('rb') as inp, (root/(mode+'-units.jsonl.gz')).open('wb') as out:
   with gzip.GzipFile(filename='',mode='wb',fileobj=out,mtime=0) as compressed:shutil.copyfileobj(inp,compressed)
  raw.unlink()
 for suffix in ['.log','.log.json']:
  source=pathlib.Path('/tmp/ny0-'+mode+'-run-corrected'+suffix)
  if source.exists():shutil.copyfile(source,root/(mode+'-run'+suffix))
(root/'pair-recovery.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(summary))
