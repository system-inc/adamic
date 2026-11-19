import os,time,json,sys,traceback
from pathlib import Path
W=Path('/tmp/family5-c02-alternatives')
while not (W/'alternatives-finished.json').exists():time.sleep(10)
pid=json.loads((W/'after-entries-supervisor.json').read_text())['pid']
while Path('/proc/'+str(pid)+'/stat').exists() and Path('/proc/'+str(pid)+'/stat').read_text().split()[2]!='Z':time.sleep(1)
from common import *
try:
 s=(W/'stock.cjs').read_text();needle="if(process.env.FAMILY5_STOCK_PROJECT){options.composite=false;";assert needle in s;s=s.replace(needle,"if(process.env.FAMILY5_STOCK_PROJECT){options.types=['node'];options.composite=false;");(W/'stock-node.cjs').write_text(s)
 data={}
 for name,tree in [('base',Path('/tmp/family5-base-tree')),*[(c,Path('/workspace/family5-alt-'+c)) for c in ['a','b','c']]]:
  if not (tree/'src/compiler/types.ts').exists():continue
  rc=run('node-controls/'+name,['node',W/'stock-node.cjs',tree,W/'node-controls'/(name+'.json')],dict(os.environ,FAMILY5_STOCK_PROJECT='src/services/tsconfig.json'),allow=True)
  data[name]=json.loads((W/'node-controls'/(name+'.json')).read_text());state.setdefault('node_controls',{})[name]={'exit_code':rc,'diagnostic_count':data[name]['diagnostic_count']};save()
 def key(d):return (d['code'],d.get('file'),re.sub(r':\d+:\d+','',d['message']))
 old=collections.Counter(map(key,data['base']['diagnostics']))
 for name in ['a','b','c']:
  if name not in data:continue
  delta=list((collections.Counter(map(key,data[name]['diagnostics']))-old).elements());(W/'node-controls'/(name+'-delta.json')).write_text(json.dumps({'baseline_diagnostics':data['base']['diagnostic_count'],'candidate_diagnostics':data[name]['diagnostic_count'],'introduced_diagnostics':delta},indent=2)+'\n');state['candidates'][name]['services_merge_exit_node_control']=state['node_controls'][name]['exit_code'];state['candidates'][name]['introduced_services_diagnostics_node_control']=len(delta)
 state['node_controls_complete']=True;save()
except Exception as error:
 state['node_controls_stop']=type(error).__name__+': '+str(error);save();(W/'NODE-CONTROLS-STOP.txt').write_text(state['node_controls_stop']+'\n'+traceback.format_exc());print('STOP '+state['node_controls_stop'],flush=True)
finally:(W/'node-controls-finished.json').write_text(json.dumps(state,indent=2)+'\n')
