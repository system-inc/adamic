#!/usr/bin/env python3
# apply.py: put stream R2's mutants (mutants.go.txt) at the end of gitignore_test.go's mutants list, for
# one run of the test. `git checkout stage1/cohere/gitignore/gitignore_test.go` takes them out again.
import pathlib
here = pathlib.Path(__file__).resolve().parent
test = here.parents[2] / 'stage1' / 'cohere' / 'gitignore' / 'gitignore_test.go'
source = test.read_text()
anchor = '\t\tto:   "\'0123456789ABCDEF\'",\n\t},\n'
assert source.count(anchor) == 1, 'the last mutant is not where it was'
source = source.replace(anchor, anchor + (here / 'mutants.go.txt').read_text())
test.write_text(source)
