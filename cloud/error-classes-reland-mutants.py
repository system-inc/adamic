#!/usr/bin/env python3
"""Run root-cause mutants without modifying compiler sources."""
import json
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parents[1]
logs = Path('/tmp/adamic-reland-mutants')
logs.mkdir(exist_ok=True)
mutants = [
    ('host-layout-proof-stale', 'internal/native/fields.go',
     '{"name", "message", "cause", "code"}', '{"name", "message", "code"}',
     ['./internal/native'], 'TestRuntimeFieldLayoutsAreIncluded$', ['runtime field "cause" at 2 is absent']),
    ('host-error-identity-removed', 'internal/native/library_node_fs_file.go',
     '\t\te.line("adamic_fs_file_error_classes(&adamic_class_%d, &adamic_class_%d, &adamic_class_%d);", classes["Error"], classes["TypeError"], classes["RangeError"])',
     '\t\te.line("/* host identity omitted */")',
     ['./internal/oracle'], '^TestNodeFSFileAgreesWithNode$/nominal$|TestNativeAgreesWithNode/internal/oracle/testdata/error_host_uncaught.a$', ['stdout differs', 'stderr differs']),
    ('host-error-override-accepted', 'internal/lower/exceptions.go',
     'if err := l.checkHostErrorMethods(); err != nil {',
     'if err := l.checkHostErrorMethods(); err != nil && false {',
     ['./internal/lower'], 'TestHostErrorToStringIsNotYet$', ['want named host Error.toString boundary']),
    ('checked-union-return-unproved', 'internal/lower/exception_paths.go',
     'if checkedReferenceReturn(function, program.Strings) {',
     'if checkedReferenceReturn(function, program.Strings) && false {',
     ['./internal/lower'], 'TestCheckedUnionReferenceCannotThrow$', ['second guard cannot throw']),
    ('string-receiver-typeerror-not-nominal', 'internal/lower/string_receiver.go',
     'l.generatedError("TypeError", "String.prototype."', 'l.generatedError("Error", "String.prototype."',
     ['./internal/oracle'],
     'TestNativeAgreesWithNode/internal/oracle/testdata/error_record_narrowing_identity.a$',
     ['stdout differs']),
    ('direct-construction-loses-message-default', 'internal/lower/exceptions.go',
     'message = l.orDefault(args[0], message, "")', 'message = message',
     ['./internal/oracle'],
     'TestNativeAgreesWithNode/internal/oracle/testdata/coverage_error_optional_messages.a$',
     ['stdout differs', 'member access within null pointer']),
    ('record-typeerror-not-nominal', 'internal/lower/exceptions.go',
     'ok && defined.Throws() {', 'ok && defined.Throws() && false {',
     ['./internal/oracle'],
     'TestNativeAgreesWithNode/internal/oracle/testdata/error_record_narrowing_identity.a$',
     ['stdout differs']),
    ('proven-read-still-throws', 'internal/ir/defined.go',
     '!d.Proven && strings.HasPrefix', 'strings.HasPrefix',
     ['./internal/lower'], 'TestMayThrowPrecision$', ['cannot throw']),
    ('unused-helper-effects', 'internal/lower/closed_frame_inputs.go',
     'if reachable[index] {', 'if reachable[index] || true {',
     ['./internal/lower', './stage3/fixtures'],
     'TestClosedFrameInputRejectsMutation$|TestFixtures/nested-functions/06_parser_token_state.a$',
     ['unused helper invalidated', 'got {Outcome:Refused What:stage3/fixtures/nested-functions/06_parser_token_state.a:6:17']),
    ('cause-descriptors-required-for-boxing', 'internal/lower/unknown.go',
     'if len(args) == 2 && args[1] == options && l.errorType',
     'if false && len(args) == 2 && args[1] == options && l.errorType',
     ['./internal/oracle'],
     'TestNativeAgreesWithNode/internal/oracle/testdata/coverage_error_causes.a$',
     ['dynamic function descriptors']),
    ('opaque-cause-reflection-accepted', 'internal/lower/cycles.go',
     'if err := l.checkErrorCauseReflection(); err != nil {',
     'if err := l.checkErrorCauseReflection(); err != nil && false {',
     ['./internal/lower'], 'TestOpaqueErrorCauseReflectionIsNotYet$',
     ['want opaque Error cause refusal, got <nil>']),
]
with tempfile.TemporaryDirectory(prefix='adamic-reland-overlay-') as directory:
    scratch = Path(directory)
    for name, relative, old, new, packages, test, assertions in mutants:
        source = (root / relative).read_text()
        if source.count(old) != 1:
            raise RuntimeError(f'{name}: expected one target')
        variant = scratch / f'{name}.go'
        variant.write_text(source.replace(old, new))
        overlay = scratch / f'{name}.json'
        overlay.write_text(json.dumps({'Replace': {str(root / relative): str(variant)}}))
        log = logs / f'{name}.log'
        with log.open('w') as output:
            result = subprocess.run(['go', 'test', '-overlay', str(overlay), '-count=1', '-v',
                                     *packages, '-run', test], cwd=root,
                                    stdout=output, stderr=subprocess.STDOUT)
        observed = log.read_text()
        caught = result.returncode != 0 and any(a in observed for a in assertions) and '[build failed]' not in observed
        print(f'{name}: exit {result.returncode}, assertion caught={caught}', flush=True)
        if not caught:
            raise RuntimeError(f'{name} survived or failed before its assertion: {log}')

# Isolate handler mutants in copied runtimes: no concurrent source mutation.
for name, replacement, test_name, assertions in [
    ('undefined-panic-class', 'const message = error instanceof AdamicPanic ? error.message : String(error);',
     'TestOracleUncaughtExceptionNames$', ['uncaught exception identity', 'AdamicPanic is not defined']),
    ('synthetic-error-format-bypassed', 'const message = String(error);',
     'TestOracleUncaughtSyntheticErrorFormatting$', ['synthetic exception formatting']),
    ('host-error-format-erased', 'const message = Error.prototype.toString.call(error);',
     'TestOracleUncaughtHostErrorFormatting$', ['host exception formatting']),
]:
    source = (root / 'oracle/adamic.mjs').read_text()
    old = """const message = error instanceof Error ? String(error)
\t\t: error !== null && typeof error === 'object' && typeof error.name === 'string' && typeof error.message === 'string'
\t\t\t? Error.prototype.toString.call(error) : String(error);"""
    if source.count(old) != 1:
        raise RuntimeError(name + ': expected one target')
    with tempfile.TemporaryDirectory(prefix='adamic-reland-runner-') as directory:
        scratch = Path(directory)
        variant = scratch / 'adamic.mjs'
        variant.write_text(source.replace(old, replacement))
        relative = 'cmd/adamic-test262/oracle_exceptions_test.go'
        test = (root / relative).read_text()
        old = 'filepath.Abs("../../oracle/adamic.mjs")'
        if test.count(old) != 1:
            raise RuntimeError('uncaught-error test: expected one runtime path')
        test_variant = scratch / 'oracle_exceptions_test.go'
        test_variant.write_text(test.replace(old, 'filepath.Abs(' + json.dumps(str(variant)) + ')'))
        overlay = scratch / 'test.json'
        overlay.write_text(json.dumps({'Replace': {str(root / relative): str(test_variant)}}))
        log = logs / (name + '.log')
        with log.open('w') as output:
            result = subprocess.run(['go', 'test', '-overlay', str(overlay), '-count=1',
                                     './cmd/adamic-test262', '-run', test_name],
                                    cwd=root, stdout=output, stderr=subprocess.STDOUT)
        observed = log.read_text()
        caught = result.returncode != 0 and all(a in observed for a in assertions)
        print(f'{name}: exit {result.returncode}, assertion caught={caught}', flush=True)
        if not caught:
            raise RuntimeError(name + ' survived')
