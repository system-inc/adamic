"""Resolve the observed feature integration conflicts on authorized scratch trees only."""
import pathlib
import re
import subprocess
import sys

repository = pathlib.Path(sys.argv[1]).resolve()
assert str(repository).startswith('/tmp/'), 'scratch trees only'
branch = subprocess.check_output(['git', 'branch', '--show-current'], cwd=repository, text=True).strip()
assert branch.startswith('scratch/latent-'), branch
conflicts = subprocess.check_output(['git', 'diff', '--name-only', '--diff-filter=U'], cwd=repository, text=True).splitlines()
pattern = re.compile(r'^<<<<<<<[^\n]*\n(.*?)^=======\n(.*?)^>>>>>>>[^\n]*\n', re.M | re.S)
for name in conflicts:
    path = repository / name
    original = path.read_text()
    if name == 'internal/oracle/counts.md':
        # Counts do not participate in measurement. Keep current rows and append new feature rows.
        ours = subprocess.check_output(['git', 'show', ':2:' + name], cwd=repository, text=True)
        theirs = subprocess.check_output(['git', 'show', ':3:' + name], cwd=repository, text=True)
        keys = {line.split('|')[1].strip() for line in ours.splitlines() if line.startswith('|')}
        added = [line for line in theirs.splitlines() if line.startswith('|') and line.split('|')[1].strip() not in keys]
        modified = ours.rstrip() + '\n' + '\n'.join(added) + '\n'
    else:
        def resolve(match):
            ours, theirs = match.groups()
            if name == 'internal/lower/lower.go':
                if 'enumInitialization' in ours and 'namespaceInitialization' in theirs:
                    return theirs # namespace initialization already calls enum initialization
                return ours + theirs
            if name == 'internal/lower/enum_initialization.go': return theirs
            if name == 'internal/lower/enums.go':
                if 'ModifierFlagsParameterPropertyModifier' in ours: return theirs
                if not theirs.strip(): return ours # retain flag-enum safeguards
            if name == 'internal/lower/enums_test.go' and not theirs.strip(): return theirs
            if name == 'internal/lower/invariance.go' and not theirs.strip(): return ours
            if name == 'internal/lower/locals.go' and 'NamespaceState' in ours and 'Preallocated' in theirs:
                return theirs.replace('return l.function != nil', 'return l.result.Locals[local].NamespaceState || l.function != nil')
            if name == 'internal/lower/expression.go':
                if 'qualified := l.namespaceMember(callee)' in ours and 'l.nestedSibling(callee)' in theirs:
                    return theirs.rsplit('\tif declaration, isGeneric :=', 1)[0] + ours
                if not ours.strip(): return theirs
                if 'localRead(node, local)' in theirs and 'case ast.KindVoidExpression:' in ours:
                    return theirs + '\tcase ast.KindVoidExpression:' + ours.split('case ast.KindVoidExpression:', 1)[1]

            if name == 'internal/lower/refusals.go':
                if 'ast.KindModuleDeclaration:' in ours and 'ast.KindVoidExpression:' in theirs: return ''
                if 'var assertion *ast.Node' in ours:
                    return ours + '\t\t\treturn true\n\t\t}\n' + theirs
                # Operator/name additions are independent branches at this boundary.
                return ours + theirs
            if name == 'internal/lower/class_inheritance.go':
                if 'enumAssignable(' in theirs and 'nominalMismatch(' in ours:
                    return ours.replace('if !l.checker.IsTypeAssignableTo', 'if !l.enumAssignable(from[index], to[index]) || !l.enumAssignable(to[index], from[index]) || !l.checker.IsTypeAssignableTo', 1)
                if 'classMembersWithParameters' in theirs:
                    return ours.replace('declaration.Members()', 'classMembersWithParameters(declaration)')
                if 'parameterProperty(member)' in theirs:
                    return ours.replace('if member.Kind == ast.KindPropertyDeclaration {', 'if member.Kind == ast.KindPropertyDeclaration || parameterProperty(member) {')
                if not theirs.strip(): return ours
            raise RuntimeError(f'unrecognized conflict in {name}: {ours[:180]!r} / {theirs[:180]!r}')
        modified = pattern.sub(resolve, original)
    assert not re.search(r'^(<<<<<<<|=======|>>>>>>>)', modified, re.M), name
    path.write_text(modified)
    subprocess.run(['git', 'add', name], cwd=repository, check=True)
    print('resolved', name)
subprocess.run(['gofmt', '-w', *[str(repository / name) for name in conflicts if name.endswith('.go')]], check=True)
subprocess.run(['git', 'add', '--update'], cwd=repository, check=True)
