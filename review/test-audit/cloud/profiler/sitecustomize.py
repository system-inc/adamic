import sys, os, atexit, json
seen=set()
def profile(frame,event,arg):
    if event=='call' and frame.f_code.co_filename.endswith('/cloud/lint-wave-check.py'):
        seen.add((frame.f_code.co_qualname,frame.f_code.co_firstlineno))
sys.setprofile(profile)
def save():
    if seen:
        from pathlib import Path
        p=Path(os.environ['U006_REACH_DIR']);p.mkdir(exist_ok=True)
        (p/(str(os.getpid())+'.json')).write_text(json.dumps(sorted(seen)))
atexit.register(save)
