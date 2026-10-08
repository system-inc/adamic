"""Require an exact Git-tree path set at every pin; retain parse-error counts."""
import collections, json, pathlib, subprocess, sys
pins, census, output = map(pathlib.Path, sys.argv[1:])
groups = collections.defaultdict(list)
with census.open() as source:
    for line in source:
        row = json.loads(line)
        groups[(row['Repo'], row['SHA'])].append(row)
summary = []
for pin in json.loads(pins.read_text()):
    tree = subprocess.check_output(['git','-C',pin['root'],'ls-tree','-rz',pin['sha']])
    expected = {entry.split(b'\t',1)[1].decode() for entry in tree.split(b'\0') if entry and entry.split(b'\t',1)[1].endswith((b'.ts',b'.tsx'))}
    rows = groups.pop((pin['repo'],pin['sha']))
    paths = {row['Path'] for row in rows}
    if not expected or paths != expected or len(paths) != len(rows):
        raise RuntimeError('missing, extra or duplicate source: '+pin['repo'])
    summary.append(dict(repo=pin['repo'], sha=pin['sha'], files=len(rows),
        declarations=sum(row['Declaration'] for row in rows),
        syntax_error_files=sum(row['ParseDiagnostics']>0 for row in rows),
        options=sum(row['Options'] for row in rows),
        type_shape=sum(row['TypeShape'] for row in rows),
        type_files=sum(row['TypeShape']>0 for row in rows),
        max_type_shape=max(row['TypeShape'] for row in rows)))
if groups: raise RuntimeError('unrequested corpus pins')
output.write_text(json.dumps(summary,indent=2)+'\n')
