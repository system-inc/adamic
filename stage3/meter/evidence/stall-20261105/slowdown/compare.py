from pathlib import Path
import json,csv,sys
r=Path('/tmp/meter-stall-20261105/slowdown'); prefix='/tmp/meter-stall-20261105/adapted/'
def read(label):
 result={}
 for line in (r/label/'times.jsonl').read_text().splitlines():
  x=json.loads(line);key=x['where'].removeprefix(prefix)
  if key in result:raise ValueError('duplicate identity: '+key)
  result[key]=x
 return result
a,b=read('good'),read('bad');rows=[]
for key in sorted(a.keys()&b.keys()):
 x,y=a[key],b[key];g,h=x['wall_seconds'],y['wall_seconds'];rows.append(dict(unit=key,good=g,bad=h,increase=h-g,ratio=h/g,good_snapshot=x['snapshot_wall_seconds'],bad_snapshot=y['snapshot_wall_seconds'],good_cpu=x['cpu_seconds'],bad_cpu=y['cpu_seconds'],good_lowering_residual=g-x['snapshot_wall_seconds'],bad_lowering_residual=h-y['snapshot_wall_seconds']))
rows.sort(key=lambda x:x['increase'],reverse=True)
with (r/'common-units.csv').open('w') as f:
 w=csv.DictWriter(f,fieldnames=rows[0].keys(),lineterminator='\n');w.writeheader();w.writerows(rows)
summary=dict(common_units=len(rows),good_completed=len(a),bad_completed=len(b),good_common_wall_seconds=sum(x['good'] for x in rows),bad_common_wall_seconds=sum(x['bad'] for x in rows),good_common_snapshot_seconds=sum(x['good_snapshot'] for x in rows),bad_common_snapshot_seconds=sum(x['bad_snapshot'] for x in rows),top30=rows[:30])
summary.update(good_common_cpu_seconds=sum(x['good_cpu'] for x in rows),bad_common_cpu_seconds=sum(x['bad_cpu'] for x in rows))
summary['common_wall_ratio']=summary['bad_common_wall_seconds']/summary['good_common_wall_seconds']
summary['common_wall_increase_seconds']=summary['bad_common_wall_seconds']-summary['good_common_wall_seconds']
summary['snapshot_increase_seconds']=summary['bad_common_snapshot_seconds']-summary['good_common_snapshot_seconds']
summary['snapshot_share_of_increase']=summary['snapshot_increase_seconds']/summary['common_wall_increase_seconds']
(r/'comparison.json').write_text(json.dumps(summary,indent=2)+'\n')
lines=['| Unit | Good s | Bad s | Increase s | Ratio | Good snapshot s | Bad snapshot s |','|---|---:|---:|---:|---:|---:|---:|']
for x in rows[:30]:lines.append(f"| `{x['unit']}` | {x['good']:.3f} | {x['bad']:.3f} | {x['increase']:.3f} | {x['ratio']:.2f}x | {x['good_snapshot']:.3f} | {x['bad_snapshot']:.3f} |")
(r/'top30.md').write_text('\n'.join(lines)+'\n')
print(json.dumps({k:v for k,v in summary.items() if k!='top30'}))
