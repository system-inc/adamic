import json,os,pathlib,subprocess
out=pathlib.Path('review/compiler/chain-slice-5/delta-negatives');rows=[]
for round in range(3):
 for side in (['before','after'] if round%2==0 else ['after','before']):
  stdout=out/'logs'/f'parser-{side}-{round}.out';stderr=out/'logs'/f'parser-{side}-{round}.err'
  with stdout.open('w') as output,stderr.open('w') as errors:
   p=subprocess.Popen(['timeout','45',f'/tmp/delta-negatives-parser-{side}','--manifest',str(out/'parser-manifest.txt'),'--whole','--count'],stdout=output,stderr=errors)
   pid,status,usage=os.wait4(p.pid,0);p.returncode=os.waitstatus_to_exitcode(status)
  rows.append({'side':side,'round':round,'exit':p.returncode,'stdout':stdout.read_text(),'stderr':stderr.read_text(),'peak_rss_kib':usage.ru_maxrss});assert p.returncode==0,rows[-1]
(out/'parser-rss.json').write_text(json.dumps(rows,indent=2)+'\n');print(json.dumps(rows,indent=2))
