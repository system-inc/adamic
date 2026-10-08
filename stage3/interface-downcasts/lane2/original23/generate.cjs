// Preserve the private original receiver and its complete imported callback signature.
const fs = require('node:fs');
const path = require('node:path');
const cases = [
 ['good', "[(name: string, event: number, time?: Date): void => { console.log(name); }]", "viewed.callbacks[0]!('a', 0);", 'a\n'],
 ['lazy', '7', "console.log(viewed.modifiedTime === undefined ? 'absent' : 'present');", 'absent\n'],
 ['wrong-array', '7', 'console.log(typeof viewed.callbacks);', 'number\n'],
 ['wrong-element', '[42]', 'console.log(typeof viewed.callbacks[0]);', 'number\n'],
 ['wrong-parameter', "[(name: number, event: number, time?: Date): void => { console.log(`${name}`); }]", 'console.log(typeof viewed.callbacks[0]);', 'function\n'],
 ['lazy-element', "[(name: string, event: number, time?: Date): void => { console.log(name); }, (name: number, event: number, time?: Date): void => { console.log(`${name}`); }]", "viewed.callbacks[0]!('a', 0);", 'a\n'],
 ['replace', "[(name: string, event: number, time?: Date): void => { console.log(name); }]", "viewed.callbacks[0] = (name: string, event: number, time?: Date): void => { console.log(`new:${name}`); }; viewed.callbacks[0]!('a', 0);", 'new:a\n'],
 ['map', "[(name: string, event: number, time?: Date): void => { console.log(name); }]", "viewed.callbacks.map(callback => { callback('a', 0); return 7; });", 'a\n'],
];
cases.push(['short-arity', "[(name: string, event: number): void => { console.log(name); }]", "viewed.callbacks[0]!('a', 0);", 'a\n']);
cases.push(['replace-source-arity', "[(name: string, event: number): void => { console.log(name); }]", "viewed.callbacks[0] = (name: string, event: number, time?: Date): void => { console.log(`new:${name}`); }; viewed.callbacks[0]!('a', 0);", 'new:a\n']);
cases.push(['valued-void-producer', "[(name: string, event: number, time?: Date): number => { console.log(name); return 7; }]", "viewed.callbacks[0]!('a', 0);", 'a\n']);
cases.push(['wrong-invoke', "[(name: number, event: number, time?: Date): void => { console.log(`${name}`); }]", "viewed.callbacks[0]!('a', 0);", 'a\n']);
const probes = [];
for (const [mode,payload,read,source] of cases) {
 const name = 'callbacks-' + mode;
 const body = `import type { FileWatcherWithModifiedTime } from 'original-tsc-watchers';\ninterface Base { readonly modifiedTime: Date | undefined; }\nfunction read(base: Base): void {\n const viewed = base as FileWatcherWithModifiedTime;\n ${read}\n}\nconst raw = { modifiedTime: undefined, callbacks: ${payload} };\nread(raw);\n`;
 fs.writeFileSync(path.join(__dirname,name+'.a'),body);
 probes.push({name,source,diagnostic: ({'wrong-array':'field read failed: viewed.callbacks is not a FileWatcherCallback[]; expected FileWatcherCallback[], found number','wrong-element':'element read failed: viewed.callbacks[0] expected FileWatcherCallback, found number','wrong-invoke':'field read failed: viewed.callbacks[0] expected FileWatcherCallback, found function with incompatible parameter representations','wrong-parameter':'field read failed: viewed.callbacks[0] expected FileWatcherCallback, found function with incompatible parameter representations','replace-source-arity':'field read failed: <array write> expected (name: string, event: number) => void, found function with arity 3'})[mode] || '' ,contracts:['FileWatcherWithModifiedTime']});
}
probes.push({name:'callbacks-frontier',source:'a\n',diagnostic:'',contracts:['FileWatcherWithModifiedTime']});
fs.writeFileSync(path.join(__dirname,'probes.json'),JSON.stringify(probes,null,2)+'\n');
