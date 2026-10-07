#!/usr/bin/env python3
"""Run independent mutations of the class exemption, restoring production each time."""
from pathlib import Path
import subprocess

source = Path('internal/lower/optional_widening.go')
original = source.read_text()
evidence = Path('docs/optional-widening/evidence')
mutants = [
    ('skip-compound-generics', 'TestOptionalWideningRefused/class_generic_compound$',
     'if l.optionalUnresolvedArgument(pattern, map[*checker.Type]bool{}) {\n\t\treturn true',
     'if l.optionalUnresolvedArgument(pattern, map[*checker.Type]bool{}) {\n\t\treturn false'),
    ('direct-subclasses-only', 'TestOptionalWideningRefused/class_transitive$',
     'if found := l.optionalClassBase(base, source, seen); found != nil {',
     'if found := base; found.Symbol() == source {'),
    ('skip-class-expressions', 'TestOptionalWideningRefused/class_expression$',
     'node.Kind == ast.KindClassDeclaration || node.Kind == ast.KindClassExpression',
     'node.Kind == ast.KindClassDeclaration'),
    ('skip-imports', 'TestOptionalWideningWholeProgram/true$',
     'for _, file := range modules {',
     'modules = []*ast.SourceFile{root}\n\t\tfor _, file := range modules {'),
    ('skip-extra-roots', 'TestOptionalWideningWholeProgram/false$',
     'range l.program.Files() {', 'range l.program.Files()[:1] {'),
    ('skip-generic-substitution', 'TestOptionalWideningAllowed/generic_subclass$',
     'candidate = instantiateType(l.checker, candidate, newTypeMapper(from, to))',
     'candidate = candidate'),
    ('exempt-structural-devtools', 'TestOptionalWideningRefused/widening_',
     'if !isClassInstance(source) {\n\t\t\t\t\treturn found',
     'if !isClassInstance(source) {\n\t\t\t\t\treturn nil'),
]
try:
    for name, test, before, after in mutants:
        assert original.count(before) == 1, (name, original.count(before))
        source.write_text(original.replace(before, after))
        log = evidence / ('class-mutant-' + name + '.log')
        with log.open('w') as output:
            result = subprocess.run(['go', 'test', './internal/lower', '-run', test,
                                     '-count=1', '-timeout', '30m'], stdout=output,
                                    stderr=subprocess.STDOUT)
        text = log.read_text()
        assert result.returncode == 1 and '--- FAIL:' in text, (name, text)
        assert 'build failed' not in text and 'panic:' not in text, (name, text)
        print(name, 'exit=1', flush=True)
        source.write_text(original)
finally:
    source.write_text(original)
