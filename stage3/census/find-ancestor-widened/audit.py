from pathlib import Path
import json,gzip,hashlib,html,subprocess
out=Path(__file__).resolve().parent;base=Path('/workspace/scratch/widened')
report=(out/'REPORT.md').read_text();comparisons=json.loads((out/'comparison.json').read_text())
assert len(comparisons)==38
actual={str(p.relative_to('/workspace/scratch/fa-adapted')):hashlib.sha256(p.read_bytes()).hexdigest() for d in ['src/compiler','src/tsc'] for p in sorted((Path('/workspace/scratch/fa-adapted')/d).rglob('*')) if p.is_file() and p.suffix in ['.ts','.a']}
assert actual==json.loads((base/'baseline/source-hashes.json').read_text())==json.loads((out/'source-hashes.json').read_text())
audit=dict(source_hashes_match=True,source_files=len(actual),sites=len(comparisons),sides={})
for side in ['parent','head']:
 raw=gzip.decompress((out/(side+'-stderr.txt.gz')).read_bytes());lines=raw.decode().splitlines()
 expected=(out/(side+'-stderr.sha256')).read_text().split()[0];assert hashlib.sha256(raw).hexdigest()==expected
 indexed=json.loads((out/(side+'-filtered-index.json')).read_text())
 assert all(lines[r['stderr_line']-1]==r['line'] for r in indexed)
 filtered=gzip.decompress((out/(side+'-filtered-probes.txt.gz')).read_bytes()).decode().splitlines();assert filtered==[r['line'] for r in indexed]
 for site in comparisons:
  assert site[side]==[r['line'] for r in indexed if r['site']==site['site']]
  assert all(html.escape(line).replace('|','&#124;') in report for line in site[side])
 stream=(out/(side+'-entry-run.jsonl')).read_bytes();baseline=(base/'baseline'/(side+'-entry-run.jsonl')).read_bytes();assert stream==baseline
 validation=json.loads((out/(side+'-stream-validation.json')).read_text());validation.update(actual_sha256=hashlib.sha256(stream).hexdigest(),baseline_sha256=hashlib.sha256(baseline).hexdigest());(out/(side+'-stream-validation.json')).write_text(json.dumps(validation,indent=2)+'\n')
 assert json.loads((out/(side+'-vet.status.json')).read_text())['exit']==0
 assert json.loads((out/(side+'-entry-run.status.json')).read_text())['exit']==0
 flags=json.loads((out/(side+'-entry-run.pid.json')).read_text())['flags'];assert flags['ADAMIC_PROBE_WIDENED']=='all' and flags['ADAMIC_PROBE_BINDINGS']=='1' and flags['LATENT_ASSERT_NO_OUTPUT']=='1' and flags['GOGC']=='70' and flags['GOMEMLIMIT']=='5GiB';assert 'LATENT_FULL' not in flags and 'LATENT_ONLY_FILE' not in flags
 for source in ['constraint_relation','invariance']:
  pin={'parent':'34035b64','head':'183fc95b'}[side];assert (out/(side+'-'+source+'.go.txt')).read_bytes()==subprocess.check_output(['git','show',pin+':internal/lower/'+source+'.go'])
 audit['sides'][side]=dict(stderr_sha256=expected,stderr_lines=len(lines),selected_occurrences=len(indexed),stream_byte_identical=True,quotes_exact=True,flags_match=True)
assert not list(out.rglob('*.go'))
for p in out.rglob('*'):
 if p.is_file() and p.suffix in ['.md','.py','.cjs','.txt']:
  text=p.read_text();assert '\u2013' not in text and '\u2014' not in text,p
(out/'audit.json').write_text(json.dumps(audit,indent=2)+'\n');print(json.dumps(audit,indent=2))
