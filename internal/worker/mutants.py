#!/usr/bin/env python3
"""Run the required source mutants, restoring each file even when a check fails."""
from pathlib import Path
import subprocess
import sys

root = Path(__file__).resolve().parents[2]
mutants = [
    ('drop-last-header', 'internal/worker/worker.go',
     'for (const field of output.headers)', 'for (const field of output.headers.slice(0, -1))',
     'TestWorkers/echo', 'header append order'),
    ('reverse-headers', 'internal/worker/worker.go',
     'for (const field of output.headers)', 'for (const field of [...output.headers].reverse())',
     'TestWorkers/echo', 'header append order'),
    ('decode-path', 'internal/worker/worker.go',
     'path: url.pathname,', 'path: decodeURIComponent(url.pathname),',
     'TestWorkers/echo', 'echo 0'),
    ('wrong-handle-export', 'internal/javascript/javascript.go',
     'name = functionName(program, exported.Function)',
     'name = functionName(program, exported.Function); if exported.Name == "handle" { for function, value := range program.Functions { if value.Name == "other" { name = functionName(program, function) } } }',
     'TestWorkers/echo', 'oracle mismatch'),
    ('panic-200', 'internal/worker/worker.go',
     "new Response('internal error', { status: 500 })", "new Response('internal error', { status: 200 })",
     'TestWorkers/panic', 'panic and recovery'),
    ('panic-kills-next', 'internal/worker/runtime.mjs',
     'throw new AdamicPanic(message);', 'globalThis.process.exit(70);',
     'TestWorkers/panic', 'Node: exit status 70'),
    ('accept-string-result', 'internal/worker/worker.go',
     'types.GetReturnTypeOfSignature(signature) != responseType',
     '(types.GetReturnTypeOfSignature(signature) != responseType && types.GetReturnTypeOfSignature(signature).Flags()&checker.TypeFlagsString == 0)',
     'TestSignature/string', 'accepted=false: <nil>'),

    ('utf8-code-units', 'internal/worker/utf8.mjs',
     'return encoded(text).length;', 'return text.length;',
     'TestRuntime', 'AssertionError'),
    ('accept-lookalike-request', 'internal/worker/worker.go',
     'types.GetTypeOfSymbolAtLocation(signature.Parameters()[0], declaration) != requestType',
     '(types.GetTypeOfSymbolAtLocation(signature.Parameters()[0], declaration) != requestType && types.TypeToString(types.GetTypeOfSymbolAtLocation(signature.Parameters()[0], declaration)) != types.TypeToString(requestType))',
     'TestSignature/lookalike', 'accepted=false: <nil>'),
    ('swallow-handler-bug', 'internal/worker/worker.go',
     'if (!(error instanceof AdamicPanic)) throw error;', 'if (false) throw error;',
     'TestBridgeRethrows', 'Missing expected rejection'),
    ('drop-generic-export', 'internal/worker/worker.go',
     'declaration.Kind == ast.KindFunctionDeclaration && len(declaration.TypeParameters()) != 0',
     'false && declaration.Kind == ast.KindFunctionDeclaration && len(declaration.TypeParameters()) != 0',
     'TestGenericExportRefused', 'generic export was silently lost'),
]
failed = False
for name, file, before, after, test, witness in mutants:
    path = root / file
    original = path.read_bytes()
    source = original.decode()
    if source.count(before) != 1:
        raise RuntimeError(f'{name}: mutation anchor not unique')
    log = Path('/tmp') / f'workers-mutant-{name}.log'
    try:
        path.write_text(source.replace(before, after, 1))
        with log.open('wb') as output:
            result = subprocess.run(['go', 'test', '-count=1', '-run', '^' + test + '$', './internal/worker'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
        report = log.read_text()
        caught = result.returncode != 0 and witness in report and '[build failed]' not in report
        print(f'{name}: {"CAUGHT" if caught else "NOT CAUGHT"} by {test} ({witness}); log={log}', flush=True)
        failed |= not caught
    finally:
        path.write_bytes(original)
sys.exit(1 if failed else 0)
