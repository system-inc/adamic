#!/usr/bin/env python3
"""Make an output-recording Go overlay; the ordinary oracle checks remain intact."""
import json,pathlib,sys
root=pathlib.Path.cwd();out=pathlib.Path(sys.argv[1]).resolve();out.mkdir(parents=True,exist_ok=True)
source=root/'internal/oracle/cache_test.go'
text=source.read_text()
needle='func rememberRun(t *testing.T, result run) {'
assert text.count(needle)==1
replacement='\nvar observationNumber = map[string]int{}\nfunc rememberRun(t *testing.T, result run) {\n if directory := os.Getenv("ADAMIC_SPLIT_OBSERVATIONS"); directory != "" {\n  observationLock.Lock()\n  number := observationNumber[t.Name()]\n  observationNumber[t.Name()] = number + 1\n  value, err := json.Marshal(struct { Test string; Number int; Run recordedRun }{t.Name(), number, record(result)})\n  if err == nil { err = os.WriteFile(filepath.Join(directory, fmt.Sprintf("%x-%05d.json", sha256.Sum256([]byte(t.Name())), number)), value, 0600) }\n  observationLock.Unlock()\n  if err != nil { t.Fatal(err) }\n }\n'
replacement='var observationLock sync.Mutex'+replacement
changed=out/'cache_test.go';changed.write_text(text.replace(needle,replacement))
(out/'overlay.json').write_text(json.dumps({'Replace':{str(source):str(changed)}}))
