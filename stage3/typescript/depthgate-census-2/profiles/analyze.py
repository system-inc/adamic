"""Run the supplied Go pprof tool; sum CPU sample weights once per filtered stack."""
from pathlib import Path
import subprocess,json,re,hashlib
W=Path('/workspace');GO=W/'adamic-tools/go/bin/go';OUT=W/'depthgate2-profile-analysis';OUT.mkdir(exist_ok=True)
filters={
 'total':[],
 'pair_anchors_excluding_mentionsHeld_and_typeNesting':['-focus=containerWhereOtherGoes|containerPairGoes|containerPairJudged','-ignore=mentionsHeld|typeNesting'],
 'outside_pair_anchors_and_mentionsHeld':['-ignore=containerWhereOtherGoes|containerPairGoes|containerPairJudged|mentionsHeld'],
 'pair_anchors_excluding_mentionsHeld':['-focus=containerWhereOtherGoes|containerPairGoes|containerPairJudged','-ignore=mentionsHeld'],
 'pair_anchors_without_root_excluding_mentionsHeld':['-focus=containerPairGoes|containerPairJudged','-ignore=mentionsHeld|containerWhereOtherGoes'],
 'mentionsHeld':['-focus=mentionsHeld'],
 'walk_excluding_mentionsHeld':['-focus=containerWhereOtherGoes','-ignore=mentionsHeld'],
 'typeNesting_global':['-focus=typeNesting'],
 'walk_excluding_mentionsHeld_and_typeNesting':['-focus=containerWhereOtherGoes','-ignore=mentionsHeld|typeNesting'],
 'outside_walk_and_mentionsHeld':['-ignore=containerWhereOtherGoes|mentionsHeld'],
}
# Intersection for nesting-in-walk is computed from filtered raw stack locations below,
# rather than treating the two regexes as an OR or inventing a pprof AND expression.
def call(cmd,out):
 p=subprocess.run(cmd,capture_output=True,text=True)
 out.write_text(p.stdout);out.with_suffix(out.suffix+'.stderr').write_text(p.stderr)
 if p.returncode:raise RuntimeError((cmd,p.returncode,p.stderr))
 return p.stdout

def samples(text):
 lines=text.splitlines();inside=False;values=[];columns=None
 for line in lines:
  if line=='Samples:':inside=True;continue
  if not inside:continue
  if line.startswith('Locations'):break
  if columns is None and 'cpu/nanoseconds' in line:columns=[x.removesuffix('[dflt]') for x in line.split()];continue
  m=re.match(r'^\s*((?:-?\d+\s+)*-?\d+)\s*:\s*((?:\d+\s*)+)',line)
  if m:
   weights=list(map(int,m.group(1).split()));assert columns and len(weights)==len(columns)
   values.append({'cpu_ns':weights[columns.index('cpu/nanoseconds')],'locations':list(map(int,m.group(2).split()))})
 assert columns,'missing sample columns'
 return values

def nesting_samples(text):
 ids=set();inloc=False;current=None
 for line in text.splitlines():
  if line.startswith('Locations'):inloc=True;continue
  if line.startswith('Mappings'):break
  if not inloc:continue
  m=re.match(r'^\s*(\d+):',line)
  if m:current=int(m.group(1))
  if 'typeNesting' in line:
   assert current is not None
   ids.add(current)
 return sum(s['cpu_ns'] for s in samples(text) if ids.intersection(s['locations']))

report={};commands=[]
for side in ['base','head']:
 p=W/'depthgate2-profile-runs'/f'{side}-utilities-r1'/'cpu.pprof';binary=W/'depthgate2-profiles'/side/'entry-census';out=OUT/side;out.mkdir(exist_ok=True)
 assert p.exists()
 d={'profile_sha256':hashlib.sha256(p.read_bytes()).hexdigest(),'binary_sha256':hashlib.sha256(binary.read_bytes()).hexdigest(),'buckets':{}}
 for mode,flags in [('top',['-top']),('top-cum',['-top','-cum'])]:
  cmd=[str(GO),'tool','pprof',*flags,str(binary),str(p)];commands.append(cmd)
  text=call(cmd,out/(mode+'.txt'));(out/(mode+'-first40.txt')).write_text(''.join(text.splitlines(True)[:40]))
 for name,flags in filters.items():
  cmd=[str(GO),'tool','pprof','-raw',*flags,str(binary),str(p)];commands.append(cmd)
  text=call(cmd,out/(name+'.raw.txt'));ss=samples(text);ns=sum(s['cpu_ns'] for s in ss)
  d['buckets'][name]={'cpu_ns':ns,'cpu_seconds':ns/1e9,'sample_rows':len(ss),'max_locations':max((len(x['locations']) for x in ss),default=0),'cpu_ns_at_64_locations':sum(x['cpu_ns'] for x in ss if len(x['locations'])==64)}
  if name in ['walk_excluding_mentionsHeld','pair_anchors_excluding_mentionsHeld']:
   ns=nesting_samples(text);d['buckets']['typeNesting_in_walk_excluding_mentionsHeld' if name=='walk_excluding_mentionsHeld' else 'typeNesting_in_pair_anchors_excluding_mentionsHeld']={'cpu_ns':ns,'cpu_seconds':ns/1e9,'method':'sum CPU sample weights whose filtered raw location stack includes typeNesting'}
  if name!='total':
   cmd=[str(GO),'tool','pprof','-top','-cum',*flags,str(binary),str(p)];commands.append(cmd);call(cmd,out/(name+'.top-cum.txt'))
 report[side]=d
 b=d['buckets'];assert b['mentionsHeld']['cpu_ns']+b['walk_excluding_mentionsHeld']['cpu_ns']+b['outside_walk_and_mentionsHeld']['cpu_ns']==b['total']['cpu_ns']
 assert b['mentionsHeld']['cpu_ns']+b['pair_anchors_excluding_mentionsHeld']['cpu_ns']+b['outside_pair_anchors_and_mentionsHeld']['cpu_ns']==b['total']['cpu_ns']
 assert b['typeNesting_in_pair_anchors_excluding_mentionsHeld']['cpu_ns']+b['pair_anchors_excluding_mentionsHeld_and_typeNesting']['cpu_ns']==b['pair_anchors_excluding_mentionsHeld']['cpu_ns']
 assert b['typeNesting_in_walk_excluding_mentionsHeld']['cpu_ns']+b['walk_excluding_mentionsHeld_and_typeNesting']['cpu_ns']==b['walk_excluding_mentionsHeld']['cpu_ns']
extra=report['head']['buckets']['total']['cpu_seconds']-report['base']['buckets']['total']['cpu_seconds'];report['delta']={'total_extra_cpu_seconds':extra,'buckets':{}}
for name in report['base']['buckets']:
 delta=report['head']['buckets'][name]['cpu_seconds']-report['base']['buckets'][name]['cpu_seconds']
 report['delta']['buckets'][name]={'extra_cpu_seconds':delta,'percent_of_total_extra_cpu':delta/extra*100 if extra else None}
(OUT/'attribution.json').write_text(json.dumps(report,indent=2)+'\n');(OUT/'pprof-commands.json').write_text(json.dumps(commands,indent=2)+'\n')
print(json.dumps(report['delta'],indent=2))
