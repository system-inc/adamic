import sys
path, which = sys.argv[1], sys.argv[2]
s = open(path).read()
pairs = {
 'a': ("\t\t<-waited.done\n\t\treturn waited.files\n",
       "\t\t_ = waited\n\t\tfiles, _ := readRepository(directory)\n\t\treturn files\n"),
 'b': ("\tif run.failed {\n\t\trepositories.CompareAndDelete(key, run)\n\t}\n", ""),
 'c': ("func repositoryKey(directory string) string {\n",
       "func repositoryKey(directory string) string {\n\treturn directory\n"),
}
old, new = pairs[which]
assert s.count(old) == 1, which
open(path, 'w').write(s.replace(old, new))
