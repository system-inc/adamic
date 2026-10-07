"""Resolve the measured main/feature conflicts in never-pushed worktrees."""
import re
import subprocess
import tempfile
from pathlib import Path


def hunk(file, ours, theirs):
    if 'ast.KindModuleDeclaration:' in ours and 'ast.KindVoidExpression:' in theirs:
        return ''
    if file.endswith('counts.md'):
        return ours + theirs
    if file.endswith('enums_test.go'):
        return theirs
    if not ours:
        return theirs
    if not theirs:
        return ours
    if '!l.checker.IsTypeAssignableTo(from[index]' in ours:
        if '!l.enumAssignable(from[index]' in ours:
            return ours
        return ours.replace('if !l.checker.', 'if !l.enumAssignable(from[index], to[index]) || !l.enumAssignable(to[index], from[index]) || !l.checker.', 1)
    if 'lowering.noteAccessorNames(modules)' in ours:
        if 'if err := lowering.enumInitialization(modules); err != nil {' in ours and not ours.endswith('\t}\n'):
            return ours + '\t\treturn nil, err\n\t}\n' + theirs
        return ours + theirs
    if 'var assertion *ast.Node' in ours:
        return ours + '\t\t\treturn true\n\t\t}\n' + theirs
    if 'func (l *lowering) checkMemberOverrides' in ours:
        return ours.replace('range declaration.Members()', 'range classMembersWithParameters(declaration)')
    if 'accessorMember(member)' in ours:
        return ours.replace('if member.Kind == ast.KindPropertyDeclaration {', 'if member.Kind == ast.KindPropertyDeclaration || parameterProperty(member) {')
    if 'range module.Statements.Nodes' in ours and 'namespaceDeclarations' in theirs:
        return theirs
    if file.endswith('enums.go') and 'node.Parent.Kind != ast.KindSourceFile' in ours:
        return theirs
    if 'return l.localRead(node, local)' in theirs and 'ir.Read' in ours:
        void = ours[ours.index('\tcase ast.KindVoidExpression:'):] if '\tcase ast.KindVoidExpression:' in ours else ''
        return theirs + void
    if 'qualified := l.namespaceMember(callee)' in ours and 'l.nestedSibling(callee)' in theirs:
        return theirs[:theirs.rindex('\tif declaration, isGeneric')] + ours
    if file.endswith('locals.go') and 'NamespaceState' in ours and 'Preallocated' in theirs:
        return '\treturn l.result.Locals[local].NamespaceState || l.function != nil && (l.result.Locals[local].Global || (l.result.Locals[local].Preallocated && l.result.Locals[local].Captured))\n'
    raise RuntimeError(f'unknown conflict in {file}:\n{ours}\nVERSUS\n{theirs}')


def merge(file, ours, base, theirs):
    with tempfile.TemporaryDirectory(prefix='parser-merge-') as directory:
        paths = []
        for number, text in enumerate([ours, base, theirs]):
            path = Path(directory) / str(number)
            path.write_text(text)
            paths.append(str(path))
        result = subprocess.run(['git', 'merge-file', '-p', *paths], capture_output=True, text=True)
        if result.returncode > 127:
            raise RuntimeError(result.stderr)
        source = re.sub(r'^<<<<<<<[^\n]*\n(.*?)^=======\n(.*?)^>>>>>>>[^\n]*\n', lambda m: hunk(file, m.group(1), m.group(2)), result.stdout, flags=re.M | re.S)
        if '<<<<<<<' in source:
            raise RuntimeError('unresolved conflict')
        return source


def index_conflicts(work, files):
    for file in files:
        blobs = []
        for stage in [2, 1, 3]:
            result = subprocess.run(['git', 'show', f':{stage}:{file}'], cwd=work, capture_output=True, text=True)
            blobs.append(result.stdout)
        (work / file).write_text(merge(file, *blobs))
