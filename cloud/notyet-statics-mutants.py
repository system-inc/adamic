#!/usr/bin/env python3
"""Independently break structural dispatch rules through Go source overlays."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parent.parent
logs = Path('/tmp/notyet-statics-mutants')
logs.mkdir(exist_ok=True)
class_path = root / 'internal/lower/class.go'
static_path = root / 'internal/lower/class_static.go'
expression_path = root / 'internal/lower/expression.go'
class_source = class_path.read_text()
static_source = static_path.read_text()
expression_source = expression_path.read_text()
javascript_path = root / 'internal/javascript/javascript.go'
javascript_source = javascript_path.read_text()
guard = '\t// A structural signature identifies neither the static side nor an instance.'
old_guard = '''\tif len(l.staticGlobals) > 0 && method != nil && len(method.Declarations) > 0 && method.Declarations[0].Kind == ast.KindMethodSignature {
        return nil, l.notYet(node, "a method call through a structural signature in a program with statics; use typeof the declaring class")
    }
'''
call = '\treturn ir.CallClosure{Closure: closure, Arguments: arguments, Returns: returns}, nil'
guess = '''    if property, ok := closure.(ir.Property); ok && property.Method {
        for _, side := range l.statics {
            if function, exists := side.methods[property.Name]; exists {
                return ir.Call{Function: function, Arguments: append([]ir.Expression{property.Object}, arguments...), Returns: returns}, nil
            }
        }
    }
'''
mutants = [
    ('blanket-guard', 'instance', class_path, class_source, guard, old_guard + guard, 'Lower:'),
    ('missing-static-map', 'static', static_path, static_source, ', Methods: lowered.methodList()', '', 'native'),
    ('guess-static-side', 'mixed', expression_path, expression_source, call, guess + call, 'stdout'),
    ('missing-inherited-map', 'inherited', static_path, static_source,
     'lowered.methods[name] = function', 'if name != "read" { lowered.methods[name] = function }', 'native'),
    ('lose-optional-short-circuit', 'optional', expression_path, expression_source,
     'property.Method = true', 'property.Method = true; property.Optional = false', 'native'),
    ('missing-generic-static-map', 'generic', static_path, static_source, ', Methods: lowered.methodList()', '', 'native'),
    ('javascript-loses-static-map', 'static', javascript_path, javascript_source,
     'adamicStaticMethods.set(value, Object.getPrototypeOf(storage)); ', '', 'JavaScript backend'),
    ('javascript-late-main-maps', 'static', javascript_path, javascript_source,
     'emitter.statements(program.Main)\n\tfor _, prototype := range emitter.prototypes {\n\t\tbuilder.WriteString(prototype)\n\t}',
     'for _, prototype := range emitter.prototypes {\n\t\tbuilder.WriteString(prototype)\n\t}\n\temitter.statements(program.Main)', 'JavaScript backend'),
]
for name, fixture, path, original, before, after, catcher in mutants:
    assert original.count(before) == 1, (name, original.count(before))
    with tempfile.TemporaryDirectory(prefix='notyet-statics-mutant-') as directory:
        scratch = Path(directory)
        changed = scratch / path.name
        changed.write_text(original.replace(before, after))
        overlay = scratch / 'overlay.json'
        overlay.write_text(json.dumps({'Replace': {str(path): str(changed)}}))
        arguments = ['go', 'test', '-overlay=' + str(overlay), './internal/oracle', '-run',
                     'TestNativeAgreesWithNode/internal/oracle/testdata/structural_statics_' + fixture + r'\.a$',
                     '-count=1', '-timeout', '10m']
        logfile = logs / (name + '.log')
        with logfile.open('w') as output:
            result = subprocess.run(arguments, cwd=root, env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'), stdout=output, stderr=subprocess.STDOUT)
        evidence = logfile.read_text()
        if result.returncode == 0 or '[build failed]' in evidence or catcher not in evidence:
            raise SystemExit(f'{name}: not killed by intended check; exit {result.returncode}, see {logfile}')
        print(f'{name}: caught, exit {result.returncode}, {catcher}; {logfile}', flush=True)
