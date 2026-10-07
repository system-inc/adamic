"""Add eight adversarial controls to an exported upstream control directory."""
import json, pathlib, shutil, sys
root = pathlib.Path(sys.argv[1]).resolve()
cases = [(json.loads(p.read_text()), p.parent) for p in root.glob('*/case.json') if not p.parent.name.startswith('extra-')]
process = next((info, path) for info, path in cases if 'stdout_written_directly' in info['test'])
race = next((info, path) for info, path in cases if 'a_handle_kept_in_a_const_nothing_reads' in info['test'])
subjects = [
 ('dynamic-console-key', process, "const log='log';console[log]('x');process.exit(0);"),
 ('template-console-key', process, "console[`log`]('x');process.exit(0);"),
 ('parenthesized-console-key', process, "console[('log')]('x');process.exit(0);"),
 ('unicode-crlf-output', process, "/* 世界🌍 */\r\nconsole.log('🌍');process.exit(0);\r\n"),
 ('race-shorthand-read', race, "export function f(){return Promise.race([work(),new Promise((_r,reject)=>{const timer=setTimeout(reject,100);use({timer});})]);}"),
 ('race-plain-writes', race, "export function f(){return Promise.race([work(),new Promise((_r,reject)=>{let timer;timer=setTimeout(reject,100);timer=0;})]);}"),
 ('race-chained-assignment', race, "export function f(){let a;let b;return Promise.race([work(),new Promise((_r,reject)=>{a=b=setTimeout(reject,100);})]);}"),
 ('race-parenthesized-window', race, '/// <reference lib="dom" />\nexport function f(){return Promise.race([work(),new Promise((_r,reject)=>{window.setTimeout(reject,100);})]);}'),
]
for name, (info, original), source in subjects:
 target = root / ('extra-' + name)
 if target.exists(): shutil.rmtree(target)
 shutil.copytree(original, target)
 subject = info['subject'].lstrip('/')
 prefix = (original / subject).read_text().split('export ')[0]
 (target / subject).write_text(source + '\n' + prefix if name.startswith('race-') else prefix + source + '\n')
 manifest = target / 'roots.manifest'
 manifest.write_text(manifest.read_text().replace(str(original), str(target)))
 (target / 'case.json').write_text(json.dumps(dict(test='Wave13Additional/' + name, subject=info['subject'])))
print('added', len(subjects), 'adversarial controls')
