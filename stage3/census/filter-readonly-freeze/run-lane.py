#!/usr/bin/env python3
"""Execute the pinned lane unchanged except a private-tree instrumentation hook after apply.
No source file in stage3/lane or stage3/adapt is edited.
"""
from pathlib import Path
import sys,os
root=Path(__file__).resolve().parents[3]
lane=root/'stage3/lane'
sys.path.insert(0,str(lane))
source=(lane/'run.py').read_text()
needle="    table_error = None\n"
hook="""    if execution['apply_exit'] == 0:
        hook = ROOT / 'stage3/census/filter-readonly-freeze/instrument.cjs'
        execution['instrument_exit'] = run('instrument', ['node', str(hook), str(tree), str(results / 'instrumented-sha256.json')])
        if execution['instrument_exit'] != 0:
            raise RuntimeError('instrumentation failed; refusing to run')
"""
assert source.count(needle)==1
source=source.replace(needle,hook+needle)
# Optional concurrency adjustment only; all runners and test filters remain unchanged.
workers=os.environ.get('FILTER_FREEZE_WORKERS')
if workers:
    assert workers.isdecimal() and int(workers)>0
    target="        env = dict(os.environ)\n"
    assert source.count(target)==1
    source=source.replace(target,"        command.extend(['--workers', '"+workers+"'])\n"+target)
# Give the executed module its original filename; ROOT/LANE and all imports stay pinned.
ns={'__name__':'__main__','__file__':str(lane/'run.py')}
exec(compile(source,str(lane/'run.py'),'exec'),ns)
