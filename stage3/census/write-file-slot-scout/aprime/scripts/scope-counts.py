from pathlib import Path
import json
r=Path('/workspace/scratch/write-file-slot-aprime');rows=[]
for f in sorted((r/'namespace-counts').glob('*.json')):
 x=json.loads(f.read_text());classes=[]
 for site in x['sites']:
  source='\n'.join(site)
  if '/adapted-tree/built/local/' in source and '/node_modules/typescript/' not in source.split('\n')[0]:classes.append('adapted')
  elif '/node_modules/typescript/' in source:classes.append('toolchain')
  else:classes.append('other')
 unique=set(classes)
 if x['reads']:
  assert x['installs']==len(classes),'positive read with truncated site inventory'
  assert len(unique)==1,'positive read in mixed-source process; requires a more detailed recount'
 rows.append(dict(file=f.name,**x,siteClasses=classes,classification=next(iter(unique)) if len(unique)==1 else 'mixed' if unique else 'none'))
summary={'processes':len(rows),'allInstalls':sum(x['installs'] for x in rows),'allNaturalReads':sum(x['reads'] for x in rows),'allControlReads':sum(x['controls'] for x in rows),'adaptedReads':sum(x['reads'] for x in rows if x['classification']=='adapted'),'toolchainReads':sum(x['reads'] for x in rows if x['classification']=='toolchain'),'otherReads':sum(x['reads'] for x in rows if x['classification'] not in ['adapted','toolchain']),'adaptedControls':sum(x['controls'] for x in rows if x['classification']=='adapted'),'adaptedInstallSites':sum(c=='adapted' for x in rows for c in x['siteClasses']),'lanePass':(r/'baseline-lane/verdict.txt').exists() and (r/'baseline-lane/verdict.txt').read_text().startswith('PASS'),'processesWithReads':[x for x in rows if x['reads']]}
(r/'namespace-summary-scoped.json').write_text(json.dumps(summary,indent=2)+'\n');(r/'namespace-processes.json').write_text(json.dumps(rows,indent=2)+'\n');print(json.dumps(summary,indent=2))
assert summary['lanePass'] and summary['adaptedReads']==0 and summary['otherReads']==0 and summary['adaptedControls']>0
