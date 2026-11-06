#!/usr/bin/env python3
"""Certify the meter's artifact predicate without an unrelated body census."""
import ast
import gzip
import tarfile
import hashlib
import json
from pathlib import Path
import re
import sys

out = Path(__file__).resolve().parent
repo = out.parents[2]
sys.dont_write_bytecode=True
sys.path.insert(0,str(repo/'stage3/meter'))
import report

def evidence(name):
    path=out/name
    if path.exists():
        return path.read_text()
    if Path(str(path)+'.gz').exists():
        return gzip.open(str(path)+'.gz','rt').read()
    with tarfile.open(out/'auxiliary-entry-attempts.tar.gz','r:gz') as archive:
        stream=archive.extractfile(name)
        assert stream is not None
        return stream.read().decode()

provenance=json.loads((out/'provenance.json').read_text())
for name in ('census','entry-census'):
    assert hashlib.sha256((Path('/tmp/next-stops-f6c44ef5')/name).read_bytes()).hexdigest()==provenance['binaries'][name]
entry_symbols=evidence('entry-symbols.txt')
ordinary_symbols=evidence('census-symbols.txt')
assert 'namespaceInitialization' not in entry_symbols
assert 'namespaceInitialization' in ordinary_symbols
entry_bytes=Path('/tmp/next-stops-f6c44ef5/entry-census').read_bytes()
ordinary_bytes=Path('/tmp/next-stops-f6c44ef5/census').read_bytes()
reason_literals={reason:{'entry':reason.encode() in entry_bytes,'ordinary':reason.encode() in ordinary_bytes} for reason in report.READINESS_REASONS}
assert all(not r['entry'] and r['ordinary'] for r in reason_literals.values())

replacements=json.loads(Path('/tmp/next-stops-f6c44ef5/entry-overlay/overlay.json').read_text())['Replace']
package=repo/'internal/lower'
targets={str(p) for p in package.glob('*.go') if not p.name.endswith('_test.go')}
targets.update(k for k in replacements if Path(k).parent==package and not k.endswith('_test.go'))
active={name:Path(replacements.get(name,name)).read_text() for name in targets}
refs=[]
emitters=[]
for name,text in active.items():
    for line_no,line in enumerate(text.splitlines(),1):
        if 'namespaceInitialization' in line and not line.lstrip().startswith('//'):
            refs.append({'file':str(Path(name).relative_to(repo)),'line':line_no,'text':line.strip()})
        for reason in report.READINESS_REASONS:
            if reason in line:
                emitters.append({'file':str(Path(name).relative_to(repo)),'line':line_no,'reason':reason})
assert len(refs)==1 and re.match(r'func \(l \*lowering\) namespaceInitialization\(',refs[0]['text'])
assert len(emitters)==4 and {e['file'] for e in emitters}=={'internal/lower/namespaces.go'}
text=active[str(package/'namespaces.go')]
start=text.index('func (l *lowering) namespaceInitialization(')
end=text.index('\n}\n\n// Mutable aliases',start)+2
for reason in report.READINESS_REASONS:
    assert all(start<=m.start()<end for m in re.finditer(re.escape(reason),text))

units=Path(replacements[str(package/'latent_units.go')]).read_text()
timer=units.index('defer latentUnitTimer.Start(record.File, unit.Where)()')
assert 'Status: "attempted"' in units[:timer]
assert set(re.findall(r'unit\.Status\s*=\s*"([^"]+)"',units[timer:]))=={'panic'}
assert units.index('unit.Status = "deferred_generic_instantiation"')<timer
assert units.index('unit.Status = "split_checker_body"')<timer

records=[json.loads(x) for x in (out/'census.jsonl').read_text().splitlines()]
potential=[r for r in records if len(r['roots'])==1 and r['kind'] in ('NotYet','Refused') and r.get('reason') in report.READINESS_REASONS]
required={report.stop_file(r['where']) for r in potential}
assert len(potential)==3 and len(required)==3
attempted=set()
witnesses={}
headers=[]
for directory in [out/'tsc-full-parts/part-1',out/'tsc-full-parts/part-2']:
    header=json.loads(evidence(str((directory/'latent.jsonl').relative_to(out))).splitlines()[0])
    assert header['latent_mode']=='full' and not header['checker_rejected'] and not header['diagnostics']
    assert header['roots']==['/tmp/next-stops-f6c44ef5/adapted/src/tsc/tsc.ts']
    headers.append(header)
    rows=[json.loads(x) for x in evidence(str((directory/'times.jsonl').relative_to(out))).splitlines()]
    for index,row in enumerate(rows,1):
        name=str(Path(row['file']).resolve())
        attempted.add(name)
        if name in required and name not in witnesses:
            witnesses[name]={'stream':str((directory/'times.jsonl').relative_to(out)),'line':index,'row':row}
assert headers[0]==headers[1] and len(headers[0]['sources'])==81
graph=json.loads((out/'graph-order.json').read_text())
assert len(graph)==81 and set(graph)==set(headers[0]['sources']) and graph[0]==headers[0]['roots'][0]
assert required<=attempted

source=(repo/'stage3/meter/report.py').read_text()
module=ast.parse(source)
function=next(n for n in module.body if isinstance(n,ast.FunctionDef) and n.name=='root_order_checks')
allowed_statuses=set()
for comparison in ast.walk(function):
    if isinstance(comparison,ast.Compare) and isinstance(comparison.left,ast.Subscript) and isinstance(comparison.left.value,ast.Name) and comparison.left.value.id=='unit' and isinstance(comparison.left.slice,ast.Constant) and comparison.left.slice.value=='status':
        for value in comparison.comparators:
            if isinstance(value,ast.Tuple):
                allowed_statuses.update(n.value for n in value.elts if isinstance(n,ast.Constant))
assert {'attempted','panic'}<=allowed_statuses
assignment=next(n for n in function.body if isinstance(n,ast.Assign) and any(isinstance(t,ast.Name) and t.id=='candidates' for t in n.targets))
expression=ast.Expression(body=assignment.value)
candidates=eval(compile(expression,str(repo/'stage3/meter/report.py'),'eval'),{
    '__builtins__':{},'len':len,'records':records,'READINESS_REASONS':report.READINESS_REASONS,
    'stop_file':report.stop_file,'attempted':attempted,'readiness':set()})
assert len(candidates)==3
nodes={}
for path in ('initial-node-probes/results.json','types-node-probe/results.json'):
    nodes.update(json.loads((out/path).read_text()))
checks=[]
for row in candidates:
    root=str(Path(row['roots'][0]).resolve())
    node=nodes[root]
    assert node['program_files_complete'] and node['build_exit']==0 and node['node_version'].startswith('v24.') and not node['timeout']
    checks.append({'root':report.site_where(root),'where':report.site_where(row['where']),'reason':row['reason'],
                   'node':node,'classification':'root-order artifact' if node['throws'] else 'counts against root file',
                   'attempt_witness':witnesses[report.stop_file(row['where'])]})
artifacts=[row for row in checks if row['node']['throws']]
assert len(artifacts)==3
proof={
 'compiler_commit':provenance['compiler_commit'],'source_files':79,
 'certification':'exact meter candidate expression, evaluated with witnessed attempted membership and statically certified empty readiness-finding set',
 'entry_root':headers[0]['roots'][0],'entry_sources_sha256':hashlib.sha256(json.dumps(sorted(graph)).encode()).hexdigest(),
 'candidate_expression':ast.get_source_segment(source,assignment.value),
 'unknown_attempted_files_irrelevant': 'Every ordinary readiness stop file is already witnessed; additional attempted files cannot change candidate membership.',
 'overlay_source_hashes':{str(Path(name).relative_to(repo)):hashlib.sha256(active[name].encode()).hexdigest() for name in [str(package/'lower.go'),str(package/'namespaces.go'),str(package/'latent_units.go')]},
 'readiness_emitters':emitters,'active_namespace_initialization_references':refs,
 'namespace_initialization_symbol':{'ordinary':True,'guarded_entry':False},
 'reason_literals_in_binaries':reason_literals,
 'timer_status_contract':{'source':'internal/lower/latent_units.go in the compiled meter overlay, generated from stage3/census/latent/units.go.txt','statuses':['attempted','panic'],'both_allowed_by_meter':True},
 'attempt_witnesses':witnesses,'readiness_node_checks':checks,'root_order_artifacts':artifacts,
 'ordinary_lowering':{'count':2,'roots':['src/compiler/corePublic.ts','src/compiler/hostErrors.ts']},
 'entry_program_lowering_certified':False,
 'no_complete_entry_body_census_claimed':True
}
(out/'artifact-proof.json').write_text(json.dumps(proof,indent=2)+'\n')
print('Certified exact meter labels: 3 root-order artifacts; all original first stops retained.')
