"""Audit complete Go test JSON logs; enforce timing only on the reference box."""
import argparse
import json
import os
import sys
from pathlib import Path
import re

LIMIT = 30.0

def leaves(path):
 terminal = {}
 for line in path.read_text().splitlines():
  try: event=json.loads(line)
  except ValueError: continue
  if event.get('Test') and event.get('Action') in {'pass','fail','skip'}:
   terminal[event['Test']] = event
 return {name:event for name,event in terminal.items() if not any(other.startswith(name+'/') for other in terminal)}

def hold_or_log(message):
 if os.environ.get('ADAMIC_UNIT_BUDGET') == '1': raise ValueError(message)
 print(message, file=sys.stderr)

def check(events):
 failures = [name for name,event in events.items() if event['Action']=='fail']
 if failures: raise ValueError(json.dumps(failures))
 slow = [dict(test=name,seconds=event.get('Elapsed',0)) for name,event in events.items()
  if event.get('Elapsed',0)>LIMIT]
 if slow: hold_or_log(json.dumps(slow))

if __name__ == '__main__':
 parser=argparse.ArgumentParser()
 parser.add_argument('directory',type=Path)
 parser.add_argument('--output',type=Path)
 args=parser.parse_args()
 rows=[]
 for line in (args.directory/'results.jsonl').read_text().splitlines():
  row=json.loads(line);name=row['test'];events=leaves(args.directory/(row.get('log') or name+'.jsonl'))
  assert row['exit']==0 and events, (name,row)
  check(events)
  if (row.get('shard') is not None or row.get('unit_selector')) and row['wall'] > LIMIT:
   hold_or_log(f"{name} selector {row.get('shard') or name} invocation took {row['wall']:.3f} seconds")
  rows.append(dict(package=row['package'],test=name,shard=row.get('shard'),invocation_seconds=row['wall'],group_seconds=row['result'].get('Elapsed',0),
   max_unit_seconds=max(e.get('Elapsed',0) for e in events.values()),units=len(events),
   selectors=[dict(test=n,action=e['Action'],seconds=e.get('Elapsed',0),
    run='/'.join('^'+re.escape(part)+'$' for part in n.split('/'))) for n,e in sorted(events.items())]))
 result=dict(limit_seconds=LIMIT,budget_enforced=os.environ.get('ADAMIC_UNIT_BUDGET')=='1',tests=rows,definition='A unit is an independently selectable leaf. Parent groups and shared builds keyed by inputs are not timed units. Unit execution and comparison outcomes are never cached; decoder Node goldens are prepared inputs.')
 text=json.dumps(result,indent=2)+'\n'
 if args.output:args.output.write_text(text)
 else: print(text,end='')
