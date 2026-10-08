#!/usr/bin/env python3
from pathlib import Path
root=Path(__file__).resolve().parent/'group1'
root.mkdir(exist_ok=True)
for field in ['getCurrentDirectory','getCommonSourceDirectory']:
 for case in ['good','wrong-value','wrong-arity','wrong-result','generic','callback','stored','direct-good','direct-arity']:
  value="(): string => '/here'"
  if case=='wrong-value': value='7'
  if case in ['wrong-arity','direct-arity']: value="(unused: number): string => '/bad'"
  if case=='wrong-result': value='(): number => 7'
  generic='<T extends Program>' if case=='generic' else ''
  result='string' if case.startswith('direct-') else '() => string'
  read=f'p.{field}()' if case.startswith('direct-') else f'p.{field}'
  text=f'''interface Base {{ readonly kind: 'program' | 'other'; }}
interface Program extends Base {{ readonly kind: 'program'; readonly {field}: () => string; readonly unread: <T>(value: T) => T; }}
function viewed(value: Base): Program {{ return value as Program; }}
function read{generic}(p: Program): {result} {{ return {read}; }}
const raw = {{kind: 'program' as const, {field}: {value}}};
'''
  if case=='generic': text=text.replace('(p: Program)', '(p: T)')
  if case=='callback': text+='const visit = (p: Program): (() => string) => read(p);\nconst callback = visit(viewed(raw));\nconsole.log(callback());\n'
  elif case=='stored': text+='const holder = {callback: read(viewed(raw))};\nconsole.log(holder.callback());\n'
  elif case.startswith('direct-'): text+='console.log(read(viewed(raw)));\n'
  else: text+='const callback = read(viewed(raw));\nconsole.log(callback());\n'
  (root/f'{field}-{case}.a').write_text(text)
