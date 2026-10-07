import fs from 'node:fs';
import zlib from 'node:zlib';
import readline from 'node:readline';
const records = Object.create(null);
const strings = new Map();
const input = fs.createReadStream(process.env.ADAMIC_CONTEXT_RECORDS).pipe(zlib.createGunzip());
for await (const line of readline.createInterface({input,crlfDelay:Infinity})) {
 const [key,wire] = JSON.parse(line);
 const held = strings.get(wire);
 if (held === undefined) {strings.set(wire,wire);records[key]=wire;}else {records[key]=held;}
}
strings.clear();
const live = new Set();
export function panic(message) {if(process.env.ADAMIC_CONTEXT_TRACE) {console.error(new Error(message).stack); }process.stderr.write(`adamic: panic: ${message}\n`);process.exit(70);}
export function programArguments() {return process.argv.slice(2);}
export function readTextFile(path) {try {return {kind:'Ok',text:fs.readFileSync(path,'utf8')};}catch(error) {return {kind:'Error',message:String(error)};}}
export function writeTextFile(path,text) {try {fs.writeFileSync(path,text);return {kind:'Ok'};}catch(error) {return {kind:'Error',message:String(error)};}}
export function tsgoProgram() {const program=live.size+1;live.add(program);return program;}
export function tsgoRelease(program) {if(!live.delete(program)) {panic('invalid or released checker handle');}}
export function tsgoInspect(program,path,start,end,kind,question) {
 if(!live.has(program)) {panic('invalid or released checker handle');}
 const key=`${path}\t${start}\t${end}\t${kind}\t${question}`;
 if(!Object.hasOwn(records,key)) {panic(`missing recorded fact: ${key}`);}
 return records[key];
}

export function tsgoTypeParts() {panic("unexpected type-parts query in raw-fact comparison");}
