#!/usr/bin/env python3
"""Prove native result, budget, ownership and capture-type checks can fail."""
import pathlib
import subprocess
import sys
root = pathlib.Path(__file__).resolve().parents[3]
fixture = ['go', 'test', '-v', '-count=1', '-timeout', '90s', '-run', 'TestNativeAgreesWithNode/internal/oracle/testdata/regexp.a', './internal/oracle']
mutants = [
 ('native-greedy-as-lazy', 'internal/native/runtime/regexp.c', 'i->greedy', '!i->greedy', fixture, 'node:   exit'),
 ('native-captures-not-reset', 'internal/native/runtime/regexp.c', 'state.captures[2 * i->ids[k]] = -1;', 'state.captures[2 * i->ids[k]] = state.captures[2 * i->ids[k]];', fixture, 'node:   exit'),
 ('native-lookbehind-left-to-right', 'internal/regexp/matcher.go', 'direction = -1', 'direction = 1', fixture, 'node:   exit'),
 ('native-case-fold-without-u-distinction', 'internal/native/runtime/regexp.c', 'flags & 4 ?', 'true ?', fixture, 'node:   exit'),
 ('native-matchall-starts-at-zero', 'internal/native/runtime/regexp.c', 'copy->slots[1].number = (double)regex_to_length(regex->slots[1].number);', 'copy->slots[1].number = 0;', fixture, 'node:   exit'),
 ('native-search-does-not-restore-lastindex', 'internal/native/runtime/regexp.c', 'regex->slots[1].number = previous;', 'regex->slots[1].number = previous + 1;', fixture, 'node:   exit'),
 ('native-result-metadata-leaked', 'internal/native/runtime/heap.c', 'let_go(array->properties);', '(void)array->properties;', fixture, 'leaks:'),
 ('native-budget-becomes-failed-match', 'internal/native/runtime/regexp.c', 'static const char message[] = "regexp: instruction step limit exceeded";\n\t\t\tadamic_panic(message, sizeof message - 1);', 'return false;', ['go','test','-v','-count=1','-timeout','60s','-run','^TestRegExpNativeStepLimit$','./internal/native'], 'native catastrophic backtracking'),
 ('native-iterator-done-is-value-slot', 'internal/native/runtime/regexp.c', 'strcmp(object->shape->names[k], "done")', 'strcmp(object->shape->names[k], "value")', ['go','test','-v','-count=1','-timeout','60s','-run','^TestRegExpIteratorResultShape$','./internal/native'], 'unexpected iterator done'),
 ('native-match-index-always-present', 'internal/native/regexp.go', 'missing := array + "->properties == NULL"', 'missing := "false"', fixture, 'node:   exit'),
 ('native-source-escapes-class-slash', 'internal/lower/regexp.go', "c == '/' && !escaped && depth == 0", "c == '/' && !escaped", ['go','test','-v','-count=1','-timeout','30s','-run','^TestRegExpSourceNode$','./internal/lower'], 'source differs'),
 ('native-source-changes-utf8-bytes', 'internal/lower/regexp.go', 'result.WriteByte(c)', 'result.WriteRune(rune(c))', ['go','test','-v','-count=1','-timeout','30s','-run','^TestRegExpSourceNode$','./internal/lower'], 'source differs'),
 ('native-oracle-merges-lone-surrogate-patterns', 'internal/native/regexp_test.go', 'key = \"units:\" + fmt.Sprint(c.PatternUnits) + \"/\" + flags', 'key = \"pattern:\" + c.Pattern + \"/\" + flags', ['go','test','-v','-count=1','-timeout','60s','-run','^TestRegExpBytecodePatternUnits$','./internal/native'], 'DISAGREEMENT case='),
 ('captures-typed-as-always-string', 'internal/load/regexp_library.go', 'interface RegExpExecArray extends Array<string | undefined>', 'interface RegExpExecArray extends Array<string>', ['go','test','-v','-count=1','-timeout','30s','-run','^TestRegExpCaptureTypes$','./internal/load'], 'want a CheckError'),
]
if len(sys.argv) > 1:
    requested = set(sys.argv[1:])
    mutants = [mutant for mutant in mutants if mutant[0] in requested]
    if len(mutants) != len(requested):
        raise SystemExit('unknown mutant requested')
for name, file, old, new, command, witness in mutants:
    target = root / file
    original = target.read_text()
    count = original.count(old)
    expected = 2 if name in ('native-case-fold-without-u-distinction', 'native-greedy-as-lazy') else 1
    if count != expected:
        raise SystemExit(f'{name}: expected {expected} sites, found {count}')
    log = pathlib.Path('/tmp') / (name + '.log')
    try:
        mutated = original.replace(old, new)
        if name == 'native-case-fold-without-u-distinction':
            mutated = mutated.replace('if (flags & 4)\n\t\t\treturn', 'if (true)\n\t\t\treturn')
        if name == 'native-captures-not-reset':
            mutated = mutated.replace('state.captures[2 * i->ids[k] + 1] = -1;', 'state.captures[2 * i->ids[k] + 1] = state.captures[2 * i->ids[k] + 1];')
            mutated = mutated.replace('alternate.captures[2 * i->ids[k]] = -1;', 'alternate.captures[2 * i->ids[k]] = alternate.captures[2 * i->ids[k]];').replace('alternate.captures[2 * i->ids[k] + 1] = -1;', 'alternate.captures[2 * i->ids[k] + 1] = alternate.captures[2 * i->ids[k] + 1];')
        target.write_text(mutated)
        with log.open('w') as output:
            result = subprocess.run(command, cwd=root, stdout=output, stderr=subprocess.STDOUT)
        observed = log.read_text()
        if result.returncode == 0 or witness not in observed:
            raise SystemExit(f'{name}: NOT CAUGHT by its intended check; see {log}')
        print(f'{name}: caught, exit {result.returncode}; {log}', flush=True)
    finally:
        target.write_text(original)
