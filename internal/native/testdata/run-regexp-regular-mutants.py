#!/usr/bin/env python3
"""Prove eligibility, lazy-cache fallback, context and earliest-start checks."""
import pathlib
import subprocess
root = pathlib.Path(__file__).resolve().parents[3]
mutants = [
 ('caller-prefix-scan-enabled-for-sticky', 'internal/regexp/native.go', '!p.flags.Sticky && len(prefix) != 0', 'len(prefix) != 0', 'TestRegExpSearchNode'),
 ('first-predicate-omitted-without-proof', 'internal/regexp/native_search.go', 'firstProven := !endAnchored && !p.nativeAnchored() && !p.flags.Sticky && filter && (first == "NULL" || p.nativeLiteralWindow(name) == "")', 'firstProven := filter', 'TestRegExpSearchNode'),
 ('one-pass-repeat-skips-body-check', 'internal/regexp/native_search.go', '"if(!(%s_repeat', '"if(false&&!(%s_repeat', 'TestRegExpSearchNode'),
 ('one-pass-repeat-tail-changed', 'internal/regexp/native_search.go', 'tail := p.code[pc]', 'tail := body', 'TestRegExpSearchNode'),
 ('one-pass-repeat-selected-for-stateful', 'internal/regexp/native_search.go', 'if p.flags.Global || p.flags.Sticky {', 'if false && (p.flags.Global || p.flags.Sticky) {', 'TestRegExpSearchNode'),
 ('end-anchored-sticky-start-ignored', 'internal/regexp/native_search.go', 'out.WriteString("if(start!=last)return false;\\n")', 'out.WriteString("start=last;\\n")', 'TestRegExpSearchNode'),
 ('boolean-minimum-width-wrong', 'internal/regexp/native_search.go', 'minimum++', 'minimum += 2', 'TestRegExpSearchNode'),
 ('literal-window-offset-changed', 'internal/regexp/native_regular.go', 'name, len(best), bestOffset)', 'name, len(best), bestOffset-1)', 'TestRegExpSearchNode'),
 ('boolean-fold-omitted', 'internal/regexp/native_search.go', 'fold := canonicalize(c, i.flags)\n\t\t\t\tif i.set.contains(fold)', 'fold := c\n\t\t\t\tif i.set.contains(fold)', 'TestRegExpSearchNode'),
 ('boolean-end-assertion-removed', 'internal/regexp/native_search.go', 'out.WriteString("if(at!=length)goto failed;\\n")', 'out.WriteString("if(at>length)goto failed;\\n")', 'TestRegExpSearchNode'),
 ('boolean-sticky-advances', 'internal/regexp/native_search.go', 'if endAnchored || p.nativeAnchored() || p.flags.Sticky {', 'if endAnchored || p.nativeAnchored() {', 'TestRegExpSearchNode'),
 ('ascii-widening-used-for-non-ascii', 'internal/native/runtime/regexp.c', 'if (input->units == input->length + 1) {', 'if (input->units != 0) {', 'TestRegExpSearchNode'),
 ('ascii-first-filter-omits-candidate', 'internal/regexp/native_regular.go', 'first=hit-input;}\\n", c)', 'first=hit-input;}\\n", c+1)', 'TestRegExpSearchNode'),
 ('search-advance-always-unicode', 'internal/native/runtime/regexp.c', 'start = regex_advance(input, length, (size_t)at, (p->flags & 4) != 0);', 'start = regex_advance(input, length, (size_t)at, true);', 'TestRegExpSearchNode'),
 ('go-latest-start-merged', 'internal/regexp/regular.go', '(next[n.x] < 0 || origin < next[n.x])', '(next[n.x] < 0 || origin > next[n.x])', 'TestRegularPriorityNode'),
 ('go-unbounded-empty-expansion', 'internal/regexp/regular.go', 'if work > regularLimit*8 {', 'if work > regularLimit*8 && false {', 'TestRegularCompileBudget'),
 ('dfa-selected-for-backreference', 'internal/regexp/regular.go', 'i.op == opReference || i.op == opLook', 'i.op == opLook', 'TestRegExpRegularEngineNode'),
 ('latest-start-merged', 'internal/native/runtime/regexp_regular.c', 'origin < origins[pc]', 'origin > origins[pc]', 'TestRegExpRegularEngineNode'),
 ('first-accept-is-leftmost', 'internal/native/runtime/regexp_regular.c', 'best = origin;', 'return origin;', 'TestRegExpRegularEngineNode'),
 ('dfa-overflow-as-no-match', 'internal/native/runtime/regexp_regular.c', 'if (count == DFA_LIMIT)\n\t\t\t\t\treturn -1;', 'if (count == DFA_LIMIT)\n\t\t\t\t\treturn 0;', 'TestRegExpRegularEngineNode'),
 ('dfa-word-context-forgotten', 'internal/native/runtime/regexp_regular.c', '((context & 1) != 0) != word', 'false != word', 'TestRegExpSearchNode'),
 ('lookbehind-context-changed', 'internal/regexp/native_search.go', 'reverse = append(reverse, sub.set.ranges[0].From)', 'reverse = append(reverse, sub.set.ranges[0].From + 1)', 'TestRegExpSearchNode'),
]
for name, file, old, new, test in mutants:
 target = root/file
 original = target.read_text()
 if original.count(old) != 1: raise SystemExit(f'{name}: wrong mutation site count: {original.count(old)}')
 log = pathlib.Path('/tmp')/f'regex-regular-mutant-{name}.log'
 try:
  target.write_text(original.replace(old,new))
  with log.open('w') as output:
   package = './internal/regexp' if name.startswith('go-') else './internal/native'
   timeout = '3s' if name == 'go-unbounded-empty-expansion' else '90s'
   result = subprocess.run(['go','test','-v','-count=1','-timeout',timeout,'-run',f'^{test}$',package],cwd=root,stdout=output,stderr=subprocess.STDOUT,timeout=120)
  witness = 'test timed out after' if name == 'go-unbounded-empty-expansion' else 'DISAGREEMENT'
  if result.returncode == 0 or witness not in log.read_text():
   raise SystemExit(f'{name}: NOT CAUGHT by Node results; {log}')
  print(f'{name}: caught by {witness}; {log}',flush=True)
 finally:
  target.write_text(original)
