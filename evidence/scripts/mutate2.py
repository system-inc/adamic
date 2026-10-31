import sys
path, which = sys.argv[1], sys.argv[2]
s = open(path).read()
pairs = {
 # reviewer's must: the run freed before its reading
 'early-free': [("\t\tclose(run.done)\n\t}()\n\trun.files, run.failed = readRepository(directory)\n",
                 "\t}()\n\tclose(run.done)\n\trun.files, run.failed = readRepository(directory)\n")],
 # reviewer's should: the free not deferred, so it runs only on a normal return
 'no-defer': [("\tdefer func() {\n\t\tif !returned {", "\tfree := func() {\n\t\tif !returned {"),
              ("\t\tclose(run.done)\n\t}()\n", "\t\tclose(run.done)\n\t}\n"),
              ("\treturned = true\n\treturn run.files\n", "\treturned = true\n\tfree()\n\treturn run.files\n")],
 # extra: a waiter takes a panicked run's answer instead of asking again
 'take-panicked': [("\t\tif !waited.panicked {\n\t\t\treturn waited.files\n\t\t}\n", "\t\treturn waited.files\n")],
}
for old, new in pairs[which]:
    assert s.count(old) == 1, (which, old)
    s = s.replace(old, new)
open(path, 'w').write(s)
