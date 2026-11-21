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
buffer = 'internal/native_runtime/c/node_buffer.c'
crypto = 'internal/native_runtime/c/node_crypto.c'
typed = 'internal/native_runtime/c/typed_array.c'
lower = 'internal/lower/typed_arrays.go'
buffer_lower = 'internal/lower/library_node_buffer.go'
mutants = [
    ('from_utf8_surrogate', buffer, '(unsigned)adamic_utf8_at(text, (double)at)', '(unsigned char)text->bytes[at]', 'encodings'),
    ('from_numeric_bytes', typed, '((uint8_t *)array->data)[index] = (uint8_t)modulo;', '((uint8_t *)array->data)[index] = (uint8_t)modulo ^ 1;', 'writes'),
    ('utf8_reprocess', buffer, 'if (old != 12) { at--; }', 'if (old != 12) { at += 0; }', 'utf8'),
    ('utf16_surrogate', buffer, 'written += put_point((unsigned char *)text->bytes + written, point);', 'written += put_point((unsigned char *)text->bytes + written, point >= 0xd800 && point <= 0xdfff ? 0xfffd : point);', 'utf16'),
    ('utf16_odd_length', buffer, 'if (encoding == 1) {\n  adamic_string *text', 'if (encoding == 1 && length % 2 != 0) { return adamic_retain(&adamic_string_empty); }\n if (encoding == 1) {\n  adamic_string *text', 'utf16'),
    ('base64_padding', buffer, " : '=';", " : 'A';", 'encodings'),
    ('base64_whitespace', buffer, 'if (lo < 0) { continue; }', 'if (lo < 0) { break; }', 'encodings'),
    ('hex_case', buffer, 'static const char hex[] = "0123456789abcdef";', 'static const char hex[] = "0123456789ABCDEF";', 'encodings'),
    ('byte_swap_store', buffer, 'adamic_typed_array_write_converted(buffer, (size_t)index, value);', 'adamic_typed_array_write_converted(buffer, (size_t)index, ((uint8_t *)buffer->data)[(size_t)index]);', 'bom'),
    ('buffer_length', lower, 'return ir.Length{Array: array, Optional: access.QuestionDotToken != nil}, true, nil', 'if l.nodeBufferType(receiver, "Buffer") { return ir.Binary{Operator:ir.Add, Left:ir.Length{Array:array}, Right:ir.NumberConstant{Value:1}}, true, nil }; return ir.Length{Array: array, Optional: access.QuestionDotToken != nil}, true, nil', 'utf16'),
    ('buffer_index', lower, 'read := ir.Expression(ir.ArrayIndex{Array: array, Index: index, Element: ir.Number})', 'if l.nodeBufferType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression)), "Buffer") { index = ir.Binary{Operator:ir.Add,Left:index,Right:ir.NumberConstant{Value:1}} }; read := ir.Expression(ir.ArrayIndex{Array: array, Index: index, Element: ir.Number})', 'writes'),
    ('hash_initial_state', crypto, '0x6a09e667,0xbb67ae85', '0x6a09e666,0xbb67ae85', 'crypto'),
    ('hash_update_bytes', crypto, 'for (size_t i=0;i<input->length;i++)', 'for (size_t i=1;i<input->length;i++)', 'crypto'),
    ('hash_digest_hex', crypto, 'static const char hex[]="0123456789abcdef";', 'static const char hex[]="0123456789ABCDEF";', 'crypto'),
    ('hash_finalization', crypto, 'if (hash->slots[1].number != 0)', 'if (hash->slots[1].number == -1)', 'finalized'),
    ('alloc_unsafe_truncate_first', buffer, 'if (!(size >= 0) || size > 9007199254740991.0) {', 'size = trunc(size); if (!(size >= 0) || size > 9007199254740991.0) {', 'alloc_unsafe'),
    ('alloc_unsafe_separator', buffer, 'bool grouped = isfinite(value) &&', 'bool grouped = false && isfinite(value) &&', 'alloc_unsafe'),
    ('alloc_unsafe_failure_panics', buffer, 'adamic_throw_coded("RangeError", "ERR_MEMORY_ALLOCATION_FAILED",', 'adamic_panic(message, sizeof message - 1); adamic_throw_coded("RangeError", "ERR_MEMORY_ALLOCATION_FAILED",', 'alloc_unsafe'),
    ('slice_copy', buffer_lower, 'return ir.TypedArraySubarray{Array: array, Arguments: lowered}, true, nil', 'return ir.NodeBufferCall{Function: "buffer_copy", Arguments: []ir.Expression{ir.TypedArraySubarray{Array: array, Arguments: lowered}}, Returns: ir.Uint8Array}, true, nil', 'slice'),
    ('slice_view_offset', typed, 'view->data = (unsigned char *)array->data + from * width(array->kind);', 'view->data = array->data;', 'slice'),
    ('ascii_unmasked', buffer, 'out[at] = (char)(bytes[start + at] & 0x7f);', 'out[at] = (char)bytes[start + at];', 'ascii'),
    ('ascii_mask_ff', buffer, 'out[at] = (char)(bytes[start + at] & 0x7f);', 'out[at] = (char)(bytes[start + at] & 0xff);', 'ascii'),
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
        # TestNativeAgreesWithNode is a family of shards (families_test.go), so the fixture's shard is any of them.
        pattern = r'TestNativeAgreesWithNode_[0-9]+/internal/oracle/testdata/node_buffer_' + fixture + r'\.a$'
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
