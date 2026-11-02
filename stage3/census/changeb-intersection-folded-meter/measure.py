"""One child process tree, wall duration and Linux wait4 resource accounting."""
import datetime,json,os,pathlib,subprocess,sys,time
resource_path=pathlib.Path(sys.argv[1])
start=time.monotonic(); utc=datetime.datetime.now(datetime.timezone.utc).isoformat()
child=subprocess.Popen(sys.argv[2:])
pid,status,usage=os.wait4(child.pid,0)
code=os.waitstatus_to_exitcode(status)
child.returncode=code
resource_path.write_text(json.dumps({'command':sys.argv[2:],'started_utc':utc,'ended_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'wall_seconds':time.monotonic()-start,'peak_rss_kib':usage.ru_maxrss,'user_seconds':usage.ru_utime,'system_seconds':usage.ru_stime,'exit_code':code,'accounting':'Linux wait4 for timeout process and its reaped descendants; ru_maxrss in KiB'},indent=2)+'\n')
sys.exit(code if code>=0 else 128-code)
