import json,pathlib,subprocess,sys
out=pathlib.Path(__file__).parent;name=sys.argv[1]
with (out/(name+'.supervisor.log.txt')).open('w') as f:
 p=subprocess.Popen([sys.executable,str(out/'run.py'),*sys.argv[1:]],stdin=subprocess.DEVNULL,stdout=f,stderr=subprocess.STDOUT,start_new_session=True)
row=dict(runner_pid=p.pid,arguments=sys.argv[1:]);(out/(name+'.launch.json')).write_text(json.dumps(row,indent=2)+'\n');print(json.dumps(row))
