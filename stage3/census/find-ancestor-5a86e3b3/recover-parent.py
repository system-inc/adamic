"""Recover the completed retained parent's own exit from Linux's dead-task record."""
from pathlib import Path
import hashlib,json,os,time
out=Path(__file__).resolve().parent;pid=7838;proc=Path('/proc')/str(pid)
raw=(proc/'stat').read_text();fields=raw[raw.rfind(')')+2:].split();assert fields[0]=='Z';wait_status=int(fields[49]);code=os.waitstatus_to_exitcode(wait_status);assert code==0
rows=[json.loads(s) for s in (out/'parent-run.jsonl').read_text().splitlines()];assert rows[0]['latent_mode']=='first-error' and rows[0]['status']=='measurement';assert len(rows)==1+len(rows[0]['sources'])==82;assert {r['file'] for r in rows[1:]}==set(rows[0]['sources'])
for file,sha in json.loads((out/'source-hashes.json').read_text()).items():assert hashlib.sha256((Path('/tmp/fa-adapted')/file).read_bytes()).hexdigest()==sha
hz=os.sysconf('SC_CLK_TCK');uptime=float(Path('/proc/uptime').read_text().split()[0]);boot_epoch=time.time()-uptime;start=boot_epoch+int(fields[19])/hz;end=(out/'parent-run.jsonl').stat().st_mtime
note='The obsolete controller cancellation also closed its resource wrapper. The compiler remained running in its own session and completed. Its actual own exit is recovered from /proc/PID/stat field 52; no peak RSS is recoverable after exit. Wall time is estimated from stream final-write time minus kernel process start.'
(out/'parent-kernel-exit.json').write_text(json.dumps(dict(pid=pid,proc_stat=raw,proc_status=(proc/'status').read_text(),clock_ticks_per_second=hz,exit_wait_status=wait_status,exit=code,stream_complete=True,note=note),indent=2)+'\n')
cmd=['env','-u','LATENT_FULL','-u','LATENT_SPECULATIVE','-u','LATENT_ONLY_FILE','LATENT_ASSERT_NO_OUTPUT=1','LATENT_TIMES=/tmp/fa-report/stage3/census/find-ancestor-e6887750/parent-run-units.jsonl','/tmp/fa-parent-census','/tmp/fa-adapted/src/tsc/tsc.ts','/tmp/fa-report/stage3/census/find-ancestor-e6887750/parent-run.jsonl']
(out/'parent-run.status.json').write_text(json.dumps(dict(command=cmd,cwd='/tmp/fa-report',exit=code,wall_seconds=round(end-start,3),wall_estimated=True,stopped=None,exit_source='parent-kernel-exit.json',resource_wrapper_interrupted=True,note=note),indent=2)+'\n')
(out/'parent-run.time.json').write_text(json.dumps(dict(peak_rss_kib=None,user_seconds=int(fields[11])/hz,system_seconds=int(fields[12])/hz,peak_memory_unavailable=note),indent=2)+'\n')
print('retained parent child_exit='+str(code)+'; 81 complete source records; peak RSS unavailable',flush=True)
