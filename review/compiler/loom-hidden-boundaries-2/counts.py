import os
from verify import run,evidence
result=run('counts-regeneration',['timeout','600','go','test','./internal/oracle','-run','TestCountsAreRecorded','-count=1','-timeout','540s','-parallel','4','-json','-args','-update-counts'])
if result['exit']:raise SystemExit(1)
for label,overlay,baseline in [('counts-audit-current','counts-audit-current-overlay.json',False),('counts-audit-main-abi','counts-audit-main-abi-overlay.json',True),('counts-audit-wrong-abi-mutant','counts-audit-main-abi-overlay.json',False)]:
    os.environ['ADAMIC_COUNTS_AUDIT_BASELINE']='1' if baseline else '0'
    result=run(label,['timeout','180','go','test','-overlay',str(evidence/overlay),'./internal/oracle','-run','^TestAuditClosureCounts','-count=1','-timeout','90s','-parallel','4','-json'])
    if label.endswith('mutant'):
        if result['exit']!=1 or not all(t['action']=='fail' for t in result['tests']):raise SystemExit('wrong ABI not caught')
    elif result['exit']:raise SystemExit(1)
