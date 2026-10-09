#!/usr/bin/env python3
"""Build each input.c mutant in scratch; never mutate the checkout."""
import json, pathlib, shutil, subprocess
root=pathlib.Path.cwd(); scratch=pathlib.Path('/tmp/decode-ascii/mutants');scratch.mkdir(parents=True,exist_ok=True)
runtime=root/'internal/native/runtime'; text=(runtime/'input.c').read_text()
mutants={
 'byte-80-ascii':text.replace('bytes[offset] < 0x80','bytes[offset] <= 0x80'),
 'boundary-one-late':text.replace('\treturn offset;\n}', '\tif (offset < length) { offset++; }\n\treturn offset;\n}',1),
 'drop-run-last-byte':text.replace('size_t size = ascii;','size_t size = ascii == 0 ? 0 : ascii - 1;').replace('memcpy(cursor, bytes, ascii);','memcpy(cursor, bytes, ascii - 1);').replace('cursor += ascii;','cursor += ascii - 1;'),
 'omit-ascii-flag':text.replace('if (ascii == length)', 'if (false)'),
 'false-ascii-flag':text.replace('if (ascii == length)', 'if (true)'),
 'word-past-end':text.replace('length - offset >= sizeof(uint64_t)','offset <= length'),
}
subprocess.run(['node',str(root/'internal/native/decode_ascii/corpus.mjs'),str(scratch/'corpus.bin')],check=True)
results=[]
for name,mutant in mutants.items():
 assert mutant!=text
 directory=scratch/name;directory.mkdir(exist_ok=True)
 for source in runtime.iterdir():
  if source.is_file():shutil.copy2(source,directory/source.name)
 (directory/'input.c').write_text(mutant)
 binary=directory/'probe'; flags=['clang','-std=c11','-Wall','-Wextra','-Werror','-pedantic','-O2','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all','-I',str(directory)]
 probe=directory/'probe.c';probe.write_text((root/'internal/native/decode_ascii/probe.c').read_text().replace('if(indexing_probe()!=0)return 1;', 'if(indexing_probe()!=0)return 1;' if name=='false-ascii-flag' else '(void)indexing_probe;'))
 flags += [str(probe),str(root/'internal/native/decode_ascii/baseline.c')]+[str(directory/p.name) for p in sorted(runtime.glob('*.c'))]+['-lm','-o',str(binary)]
 with (directory/'build.log').open('w') as log:subprocess.run(flags,stdout=log,stderr=subprocess.STDOUT,check=True)
 with (directory/'run.log').open('w') as log:result=subprocess.run([str(binary),str(scratch/'corpus.bin')],stdout=log,stderr=subprocess.STDOUT)
 output=(directory/'run.log').read_text()
 catcher='AddressSanitizer: heap-buffer-overflow' if name=='word-past-end' else ('decode indexing mismatch' if name=='false-ascii-flag' else ('decode cache mismatch' if name in ('byte-80-ascii','boundary-one-late','drop-run-last-byte','omit-ascii-flag') else 'decode mismatch'))
 assert result.returncode!=0 and catcher in output,(name,result.returncode,output)
 if name!='word-past-end':assert 'AddressSanitizer' not in output,output
 results.append(dict(mutant=name,exit=result.returncode,catcher=catcher,log=str(directory/'run.log')));print(results[-1])
(root/'cloud/reports/decode-ascii/cache-mutants.json').write_text(json.dumps(results,indent=2)+'\n')
