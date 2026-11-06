const fs = require('node:fs');
const path = require('node:path');
const directory = fs.mkdtempSync(path.join(process.env.TMPDIR || '/tmp', 'adamic-shared-'));
const file = path.join(directory, 'file');
fs.writeFileSync(file, 'changed');
const mode = process.argv[2];
if (mode === 'cap' || mode === 'cap-failure') {
 const names = Array.from({length:100}, (_, i) => path.join(directory, `file${i}`));
 names.forEach(name => fs.writeFileSync(name, 'changed'));
 if (mode === 'cap-failure') {
  // Measure Node's watch error with no descriptor slots left. File creation is over.
  const descriptors = [];
  try { for (;;) descriptors.push(fs.openSync('/dev/null', 'r')); }
  catch (error) { if (error.code !== 'EMFILE' || error.syscall !== 'open') throw error; }
  let failure;
  try { fs.watch(names[0], () => {}); }
  catch (error) { failure = error; }
  descriptors.forEach(fd => fs.closeSync(fd));
  if (!failure || failure.code !== 'EMFILE' || failure.syscall !== 'watch') throw new Error('expected EMFILE at watch');
  console.log(`watch failure code=${failure.code} syscall=${failure.syscall}`);
 } else {
  const handles = names.map(name => fs.watch(name, (type) => {
   if (type !== 'change') throw new Error(`expected change, got ${type}`);
   console.log('cap change'); handles.forEach(handle => handle.close());
  }));
  const count = fs.readdirSync('/proc/self/fd').filter(fd => {
   try { return fs.readlinkSync(`/proc/self/fd/${fd}`) === 'anon_inode:inotify'; } catch { return false; }
  }).length;
  if (count !== 1) throw new Error(`inotify count ${count}`);
  console.log('100 watchers, 1 inotify');
  fs.writeFileSync(names[99], 'changed');
 }
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
