"""Check whole-corpus no-stubs restoration, retaining exact finding identities."""
import json
from pathlib import Path
import sys

paths=list(map(Path,sys.argv[1:4]))
runs=[[json.loads(line) for line in p.read_text().splitlines()] for p in paths]
full,mutant,speculative=runs
assert full==mutant, 'no-stubs mutant differs from current full census'
assert full[0]['diagnostics']==speculative[0]['diagnostics']
assert full[0]['diagnostic_sites']==speculative[0]['diagnostic_sites']
def sites(rows):
    return {tuple(f[k] for k in ('kind','where','reason','text')) + ((f.get('site_where',''),f.get('site_kind',''),f.get('site_start',0),f.get('site_end',0)) if rows[0]['latent_mode']=='speculative' else ()) for row in rows[1:] for f in row['findings'] if f['kind'] in ('NotYet','Refused')}
a,b=sites(full),sites(speculative)
assert len(b)>len(a),(len(a),len(b))
print(json.dumps({'full_by_kind':{k:sum(f[0]==k for f in a) for k in ('NotYet','Refused')},'speculative_by_kind':{k:sum(f[0]==k for f in b) for k in ('NotYet','Refused')},'full_sites':len(a),'no_stubs_sites':len(sites(mutant)),'speculative_sites':len(b),'exact_full_json_restored':True,'checker_diagnostics_identical':True},indent=2))
