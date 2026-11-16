"""Recover an orphaned completed group from Linux's retained wait status, without rerunning roots."""
import json,os,signal,sys
from pathlib import Path
from run import out
pid=int(sys.argv[1]);first=int(sys.argv[2]);last=int(sys.argv[3])
raw=Path(f'/proc/{pid}/stat').read_text();fields=raw[raw.rfind(')')+2:].split()
assert fields[0]=='Z','only a completed zombie group can be recovered'
wait_status=int(fields[49]);rc=os.waitstatus_to_exitcode(wait_status);assert rc==0
seconds=float(Path('/proc/uptime').read_text().split()[0])-int(fields[19])/os.sysconf('SC_CLK_TCK')
assert 0<seconds<1700
for i in range(first,last+1):
 s=json.loads((out/f'root-{i:02}-status.json').read_text());assert s['exit_code']==0
 rows=[json.loads(l) for l in (out/f'root-{i:02}.jsonl').read_text().splitlines()]
 assert len(rows)==2 and {k:v for k,v in rows[0].items() if k!='seconds'}=={k:v for k,v in rows[1].items() if k!='seconds'}
try:os.killpg(pid,signal.SIGKILL)
except ProcessLookupError:pass
b=Path(__file__).resolve().parent
status=dict(module_format=out.name,command=[sys.executable,str(b/'group.py'),str(first),str(last)],exit_code=rc,seconds=seconds,stopping_reason=None,process_group_cleaned=True,recovered_after_driver_interruption=True,timing_definition='upper bound from kernel group start through recovery, including zombie time',kernel_stat=raw,kernel_wait_status=wait_status)
(out/f'group-{first}-{last}-status.json').write_text(json.dumps(status,indent=2)+'\n')
print('Recovered group',first,last,'exit',rc,'elapsed upper bound',seconds)
