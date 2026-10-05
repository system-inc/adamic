#!/usr/bin/env python3
# apply.py <repository>: put stream R2's mutants (mutants.go.txt) at the end of mediaquery_test.go's
# mutants list, for one run of the test. `git checkout stage1/cohere/mediaquery/mediaquery_test.go`
# takes them out again.
import pathlib, sys
here = pathlib.Path(__file__).resolve().parent
test = pathlib.Path(sys.argv[1]) / 'stage1' / 'cohere' / 'mediaquery' / 'mediaquery_test.go'
source = test.read_text()
anchor = '\t\tto:   "return text.trimStart();",\n\t},\n'
assert source.count(anchor) == 1, 'the last mutant is not where it was'
source = source.replace(anchor, anchor + (here / 'mutants.go.txt').read_text())
test.write_text(source)
