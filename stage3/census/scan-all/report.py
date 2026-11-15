"""Run one root closure and write only summary tables to the requested Markdown file."""
import collections, json, os, pathlib, subprocess, sys, tempfile, time
binary, root, destination = sys.argv[1:]
# wait4 reports this child's usage, independent of inherited/cumulative child
# usage across a shell's final exec. Scratch temporary streams avoid pipe blocking.
with tempfile.TemporaryFile() as stdout, tempfile.TemporaryFile() as stderr:
    start = time.monotonic()
    process = subprocess.Popen([binary, root], stdout=stdout, stderr=stderr)
    _, status, usage = os.wait4(process.pid, 0)
    process.returncode = os.waitstatus_to_exitcode(status)
    wall = time.monotonic() - start
    peak = usage.ru_maxrss
    stdout.seek(0); stderr.seek(0)
    if process.returncode:
        raise SystemExit(f'closure process failed ({process.returncode}): '+stderr.read().decode())
    r = json.load(stdout)
if 'Scope' not in r: raise SystemExit(r)
rows = r.get('Stops') or []
# UnderlyingAbort is the actual pipeline stop before a deferred production diagnosis substitutes it.
warning = r['Scope']
lines = ['# scan-all example', '', warning, '', 'Root: `'+root+'`', '', f'Wall: {wall:.4f} s; peak RSS: {peak} KiB (Linux child maximum).', '', 'Completed: '+str(r['Completed'])+'. Boundary: '+r['Boundary'], '', warning, '', '| Observation kind | Rule | Events | Distinct (rule, site, words) triples |', '| --- | --- | ---: | ---: |']
first = next((s for s in rows if s['Kind'] != 'scan stop'), None)
if first:
    lines[10:10] = ['First declaration/body/tail stop: `'+first['Rule']+'` at `'+first['Site']+'`: '+first['Words'], '']
groups = collections.defaultdict(list)
for s in rows: groups[(s['Kind'],s['Rule'])].append(s)
for (kind,rule), stops in sorted(groups.items()):
    lines.append(f'| {kind} | `{rule}` | {len(stops)} | {len(set((s["Rule"],s["Site"],s["Words"]) for s in stops))} |')
lines += ['', warning, '', '| Observation kind | File (from site) | Events | Distinct (rule, site, words) triples |', '| --- | --- | ---: | ---: |']
groups = collections.defaultdict(list)
for s in rows: groups[(s['Kind'],s['Site'].rsplit(':',2)[0])].append(s)
for (kind,file),stops in sorted(groups.items()):
    lines.append(f'| {kind} | `{file}` | {len(stops)} | {len(set((s["Rule"],s["Site"],s["Words"]) for s in stops))} |')
lines += ['', f'Total recorded events: {len(rows)}; distinct (rule, site, words) triples: {len(set((s["Rule"],s["Site"],s["Words"]) for s in rows))}.', '', 'Only the underlying first pipeline stop is counted; a substituted returned diagnosis is not another observation. Later pass results remain unknown.', '']
pathlib.Path(destination).write_text('\n'.join(line.rstrip() for line in lines))
print(f'{destination}: {wall:.4f} s, {peak} KiB, {len(rows)} events')
