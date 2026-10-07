import json
from pathlib import Path
root=Path(__file__).parent
def check(args,mode):
 assert '-DADAMIC_COUNT' not in args, 'allocation counting enabled'
 assert [a for a in args if a.startswith('-O')]==(['-O2'] if mode=='release' else ['-O1']), 'wrong optimization flags'
 assert ('-g' in args)==(mode=='sanitized'), 'wrong debug flag'
 assert [a for a in args if a.startswith('-fsanitize=')]==([] if mode=='release' else ['-fsanitize=address,undefined']), 'wrong sanitizer flags'
for suite in ['coverage','volume']:
 for mode in ['release','sanitized']:
  check(json.loads((root/f'{suite}-{mode}.clang.json').read_text()),mode)
  print(suite,mode,'PASS')
release=json.loads((root/'coverage-release.clang.json').read_text())
for name,flag in [('wrong optimization','-O1'),('sanitized release','-fsanitize=address,undefined'),('counting build','-DADAMIC_COUNT')]:
 try: check(release+[flag],'release')
 except AssertionError as e: print('mutant',name,'caught:',e)
 else: raise AssertionError('mutant escaped: '+name)
