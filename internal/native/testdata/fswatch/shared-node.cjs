const fs = require('node:fs');
const path = require('node:path');
const directory = fs.mkdtempSync(path.join(process.env.TMPDIR || '/tmp', 'adamic-shared-'));
const file = path.join(directory, 'file');
fs.writeFileSync(file, 'changed');
const mode = process.argv[2];
if (mode === 'cap') {
 const handles = [];
 for (let i = 0; i < 100; i++) {
  const name = path.join(directory, `file${i}`);
  fs.writeFileSync(name, 'changed'); handles.push(fs.watch(name, () => console.log('unexpected callback')));
 }
 const count = fs.readdirSync('/proc/self/fd').filter(fd => {
  try { return fs.readlinkSync(`/proc/self/fd/${fd}`) === 'anon_inode:inotify'; } catch { return false; }
 }).length;
 if (count !== 1) throw new Error(`inotify count ${count}`);
 console.log('100 watchers, 1 inotify');
 handles.forEach(handle => handle.close());
} else if (mode === 'same') {
 let phase = 0;
 const first = fs.watch(file, () => { if (phase === 0) { console.log('first change'); first.close(); } });
 const second = fs.watch(file, () => {
  if (phase++ === 0) { console.log('second change'); fs.writeFileSync(file, 'changed'); }
  else { console.log('second survives'); second.close(); }
 });
 fs.writeFileSync(file, 'changed');
} else if (mode === 'unref') {
 fs.watch(file, {persistent:false}, () => console.log('unexpected callback'));
 const second = fs.watch(file, () => console.log('unexpected callback'));
 second.close();
 console.log('unref exits');
} else throw new Error(mode);
process.on('exit', () => fs.rmSync(directory, {recursive:true, force:true}));
