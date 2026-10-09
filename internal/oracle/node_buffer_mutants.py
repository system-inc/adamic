#!/usr/bin/env python3
"""Run independent host mutants, restoring the exact original after each run.

Run from the repository root after sourcing cloud/setup.sh's environment file.
Every test run goes to its own log. A mutant counts only when the Node comparison
fails, without a compiler error, sanitizer report or leak finding.
"""
from pathlib import Path
import re
import subprocess
import sys

root = Path(__file__).resolve().parents[2]
logs = Path('/tmp/host-buffer-mutants')
logs.mkdir(exist_ok=True)
buffer = 'internal/native/runtime/node_buffer.c'
crypto = 'internal/native/runtime/node_crypto.c'
lower = 'internal/lower/object.go'
mutants = [
    ('fallback_identity', 'internal/lower/library_node_crypto_fallback.go', 'return ir.Conditional{Condition: condition, WhenTrue: fallback, WhenNot: fallback}, true, nil', 'if read, ok := fallback.(ir.Read); ok { for _, statement := range l.forwarderValues { if declared, ok := statement.(ir.Declare); ok && declared.Local == read.Local { fallback = declared.Value } } }; return ir.Conditional{Condition: condition, WhenTrue: fallback, WhenNot: fallback}, true, nil', 'fallback'),
    ('crypto_namespace_absent', 'internal/lower/library_node_buffer.go', 'return ir.ObjectLiteral{}, true, nil', 'return ir.Undefined{Of: ir.Object}, true, nil', 'namespace'),
    ('from_utf8_surrogate', buffer, '(unsigned)adamic_utf8_at(text, (double)at)', '(unsigned char)text->bytes[at]', 'encodings'),
    ('from_numeric_bytes', buffer, 'return (unsigned)byte;', 'return (unsigned)byte ^ 1;', 'writes'),
    ('utf8_reprocess', buffer, 'if (old != 12) { at--; }', 'if (old != 12) { at += 0; }', 'utf8'),
    ('utf16_surrogate', buffer, 'written += put_point((unsigned char *)text->bytes + written, point);', 'written += put_point((unsigned char *)text->bytes + written, point >= 0xd800 && point <= 0xdfff ? 0xfffd : point);', 'utf16'),
    ('utf16_odd_length', buffer, 'if (encoding == 1) {\n  adamic_string *text', 'if (encoding == 1 && length % 2 != 0) { return adamic_retain(&adamic_string_empty); }\n if (encoding == 1) {\n  adamic_string *text', 'utf16'),
    ('base64_padding', buffer, " : '=';", " : 'A';", 'encodings'),
    ('base64_whitespace', buffer, 'if (lo < 0) { continue; }', 'if (lo < 0) { break; }', 'encodings'),
    ('hex_case', buffer, 'static const char hex[] = "0123456789abcdef";', 'static const char hex[] = "0123456789ABCDEF";', 'encodings'),
    ('byte_swap_store', buffer, 'if (slot != NULL) { slot->number = (double)byte_value(value); }', 'if (slot != NULL) { slot->number = slot->number + 0 * byte_value(value); }', 'bom'),
    ('fs_byte_swap_store', buffer, 'if (slot != NULL) { slot->number = (double)byte_value(value); }', 'if (slot != NULL) { slot->number = slot->number + 0 * byte_value(value); }', 'fs_decode'),
    ('buffer_length', lower, 'return ir.Length{Array: object}, nil', 'if l.nodeBufferType(l.checker.GetTypeAtLocation(access.Expression), "Buffer") { return ir.Binary{Operator:ir.Add, Left:ir.Length{Array:object}, Right:ir.NumberConstant{Value:1}},nil }; return ir.Length{Array: object}, nil', 'utf16'),
    ('buffer_index', lower, 'return l.defined(node, ir.ArrayIndex{Array: object, Index: position, Element: element}), nil', 'if l.nodeBufferType(l.checker.GetTypeAtLocation(access.Expression), "Buffer") { position = ir.Binary{Operator:ir.Add,Left:position,Right:ir.NumberConstant{Value:1}} }; return l.defined(node, ir.ArrayIndex{Array: object, Index: position, Element: element}), nil', 'writes'),
    ('hash_initial_state', crypto, '0x6a09e667,0xbb67ae85', '0x6a09e666,0xbb67ae85', 'crypto'),
    ('hash_update_bytes', crypto, 'for (size_t i=0;i<input->length;i++)', 'for (size_t i=1;i<input->length;i++)', 'crypto'),
    ('hash_digest_hex', crypto, 'static const char hex[]="0123456789abcdef";', 'static const char hex[]="0123456789ABCDEF";', 'crypto'),
    ('hash_finalization', crypto, 'if (hash->slots[1].number != 0)', 'if (hash->slots[1].number == -1)', 'finalized'),
]
selection = sys.argv[1:]
summary = []
if selection and (logs / 'summary.log').exists():
    summary = [line for line in (logs / 'summary.log').read_text().splitlines() if not any(line.startswith(name + ':') for name in selection)]
for name, filename, before, after, fixture in mutants:
    if selection and name not in selection:
        continue
    path = root / filename
    original = path.read_text()
    logfile = logs / (name + '.log')
    try:
        # Token matching keeps the mutants reproducible after C/Go formatting.
        tokens = re.findall(r"0[xX][0-9a-fA-F]+|[A-Za-z_][A-Za-z_0-9]*|[0-9]+|[^\s]", before)
        pattern = r'\s*'.join(re.escape(token) for token in tokens)
        changed, count = re.subn(pattern, lambda match: after, original, count=1)
        assert count == 1, name
        path.write_text(changed)
        pattern = 'TestNativeAgreesWithNode/internal/oracle/testdata/node_buffer_' + fixture + r'\.a$'
        with logfile.open('wb') as output:
            result = subprocess.run(['go', 'test', './internal/oracle', '-run', pattern, '-count=1', '-timeout', '30m'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
    finally:
        path.write_text(original)
    output = logfile.read_text()
    differences = sorted(set(re.findall(r'(?:JavaScript backend: )?(?:stdout differs|stderr differs|exit codes differ)', output)))
    # Native finalization mistakes also change exit status; require source comparison
    # output to be present and no unrelated execution/compiler failure.
    caught = result.returncode != 0 and ('node:' in output) and bool(differences)
    other = any(marker in output for marker in ['AddressSanitizer:', 'LeakSanitizer:', 'runtime error:', 'compiling runtime', 'undefined reference', 'build failed', 'leaks:'])
    verdict = 'CAUGHT' if caught and not other else 'INVALID'
    line = f'{name}: {verdict}; fixture={fixture}; comparisons={", ".join(differences)}; log={logfile}'
    print(line, flush=True)
    summary.append(line)
    (logs / 'summary.log').write_text('\n'.join(summary) + '\n')
    if verdict != 'CAUGHT':
        raise SystemExit('Mutant not proven by Node comparison: ' + name)
