from pathlib import Path
import os, subprocess, time, json, re
root = Path(__file__).resolve().parents[2]
out = Path(__file__).resolve().parent
records = []
for mid in ['M1','M2','M3','M4','M5','M6']:
    for scope, pattern in [('target','^TestPredicateSummaryParameterIndex$'), ('package','.')]:
        cmd = ['go','test','-v','-count=1','-timeout','90s','./internal/lower/','-run',pattern]
        env = dict(os.environ, ADAMIC_MUTANT=mid)
        log = out / (mid + '-' + scope + '.log')
        start = time.monotonic()
        with log.open('w') as stream:
            result = subprocess.run(cmd, cwd=root, env=env, stdout=stream, stderr=subprocess.STDOUT)
        seconds = time.monotonic()-start
        content = log.read_text()
        failed = re.findall(r'^--- FAIL: ([^\s/]+)', content, re.M)
        lines = []
        current = ''
        for line in content.splitlines():
            match = re.match(r'^=== (?:RUN|CONT|NAME)\s+(\S+)', line)
            if match:
                current = match.group(1)
            if current.split('/')[0] in failed and re.search(r'_test.go:\d+:', line):
                lines.append(line.strip())
        record = dict(mutant=mid,scope=scope,command='ADAMIC_MUTANT='+mid+' '+' '.join(cmd),seconds=round(seconds,6),exit=result.returncode,failed_tests=failed,failing_lines=lines,timeout='test timed out' in content,log=str(log.relative_to(root)))
        records.append(record)
        (out/'runs.json').write_text(json.dumps(records,indent=2)+'\n')
        print(mid,scope,round(seconds,3),result.returncode,failed,flush=True)
