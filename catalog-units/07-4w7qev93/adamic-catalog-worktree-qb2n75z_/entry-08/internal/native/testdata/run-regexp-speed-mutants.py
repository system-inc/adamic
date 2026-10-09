#!/usr/bin/env python3
"""Each optimization must disagree with Node when its safety condition is broken."""
import pathlib
import subprocess
root = pathlib.Path(__file__).resolve().parents[3]
command = ['go', 'test', '-v', '-count=1', '-timeout', '90s', '-run', '^TestRegExpSearchNode$', './internal/native']
mutants = [
 ('ascii-fold-not-canonicalized', 'internal/regexp/native_search.go', 'i.set.contains(canonicalize(c, i.flags))', 'i.set.contains(c)', 1),
 ('anchor-checked-before-surrogate-rewind', 'internal/native/runtime/regexp.c', 'if (p->anchored && at != 0)', 'if (p->anchored && start != 0)', 1),
 ('workspace-heap-block-leaked', 'internal/native/runtime/regexp.c', 'if (frame->allocation_on_heap)\n\t\t\tfree(frame);', 'if (frame->allocation_on_heap)\n\t\t\t(void)frame;', 1),
 ('anchor-ignores-multiline', 'internal/regexp/native_search.go', 'i.assertion == Start && !i.flags.Multiline', 'i.assertion == Start', 1),
 ('unicode-pair-truncated', 'internal/native/runtime/regexp.c', 'if (point > 0xffff) {', 'if (point > 0xffff && false) {', 1),
 ('first-set-drops-alternative', 'internal/regexp/native_search.go', 'a[0] | b[0], a[1] | b[1]', 'a[0] | (b[0] & 0), a[1] | (b[1] & 0)', 2),
 ('prefix-changes-literal', 'internal/regexp/native_search.go', 'prefix = append(prefix, i.set.ranges[0].From)', 'prefix = append(prefix, i.set.ranges[0].From + 1)', 1),
 ('straight-line-capture-offset', 'internal/regexp/native_search.go', 'captures[%d]=at;', 'captures[%d]=at+1;', 1),
 ('saved-captures-not-copied', 'internal/native/runtime/regexp.c', 'memcpy(result.captures, state->captures, 2 * (p->captures + 1) * sizeof *result.captures);', 'memset(result.captures, -1, 2 * (p->captures + 1) * sizeof *result.captures);', 1),
]
for name, file, old, new, count in mutants:
 target=root/file
 original=target.read_text()
 if original.count(old)!=count: raise SystemExit(f'{name}: wrong mutation site count')
 log=pathlib.Path('/tmp')/f'regex-speed-mutant-{name}.log'
 try:
  target.write_text(original.replace(old,new))
  chosen = command if name != 'ascii-fold-not-canonicalized' else [part.replace('^TestRegExpSearchNode$', '^TestRegExpLintPatternsNode$') for part in command]
  with log.open('w') as output: result=subprocess.run(chosen,cwd=root,stdout=output,stderr=subprocess.STDOUT)
  observed=log.read_text()
  witness = 'LeakSanitizer' if name == 'workspace-heap-block-leaked' else 'DISAGREEMENT case='
  if result.returncode==0 or witness not in observed:raise SystemExit(f'{name}: NOT CAUGHT by Node results; see {log}')
  print(f'{name}: caught by {witness}; {log}',flush=True)
 finally:target.write_text(original)
