"""Collect serial merge evidence after controls and regeneration passed."""
from pathlib import Path
import gzip,json,subprocess,sys
r=Path(__file__).resolve().parent
number,branch,sha,reason=sys.argv[1:5]
assert not list(Path('review').rglob('*.go')), 'review evidence must not be compilable Go'
subprocess.run(['python3',str(r/'record_member.py'),number,branch,sha,reason],check=True)
d=json.loads((r/('member-'+number+'-results.json')).read_text())
leaves=[x for x in d['tests'] if '/' not in x['test'] and x['result']=='pass']
assert all(x['seconds']<60 for x in leaves),leaves
print('Top-level seconds:',[(x['test'],x['seconds']) for x in leaves])
counts=(r/('member-'+number+'-counts.log')).read_text();assert '\tok' in counts or 'ok ' in counts,counts
print(counts.strip())
for p in r.rglob('*.log'):
 if not p.is_file():continue
 with gzip.open(str(p)+'.gz','wb') as f:f.write(p.read_bytes())
report=r/'REPORT.md';report.write_text(report.read_text()+'\n## Member '+number+': '+branch+' '+sha+'\n\n'+reason+'\n\n'+counts.strip()+'\n\nTop-level pass seconds: '+json.dumps({x['test']:x['seconds'] for x in leaves})+'. All named controls pass; count-changes and results JSON retain every row and observation.\n')
