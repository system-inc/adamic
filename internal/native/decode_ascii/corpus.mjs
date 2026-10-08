import {openSync,writeSync,closeSync} from 'node:fs';
import assert from 'node:assert/strict';
// Binary records: little-endian u16 input length, u16 oracle UTF-8 length, then both byte strings.
// C expands each prefix at every 0..7 allocation offset and ASCII run length 0..64,
// both at the end and with an ASCII suffix. ASCII fences cannot join a UTF-8 sequence.
const out=openSync(process.argv[2],'w');let count=0;
let chunk=Buffer.alloc(1<<20),used=0;
function add(bytes){
 const input=Buffer.from(bytes), expected=Buffer.from(input.toString('utf8'),'utf8');
 // Hold the composition used by the C expansion to Node, including NULs and truncations.
 for(let k=0;k<=64;k++) {
  const fence=Buffer.alloc(k,65), decoded=input.toString('utf8');
  assert.equal(Buffer.concat([fence,input]).toString('utf8'),fence.toString()+decoded);
  assert.equal(Buffer.concat([fence,input,fence]).toString('utf8'),fence.toString()+decoded+fence.toString());
 }
 const size=4+input.length+expected.length;
 if(used+size>chunk.length){writeSync(out,chunk.subarray(0,used));used=0;}
 chunk.writeUInt16LE(input.length,used);chunk.writeUInt16LE(expected.length,used+2);used+=4;
 input.copy(chunk,used);used+=input.length;expected.copy(chunk,used);used+=expected.length;count++;
}
add([]);
for(let a=0;a<256;a++){add([a]);for(let b=0;b<256;b++)add([a,b]);}
const edges=[0,0x7f,0x80,0x8f,0x90,0x9f,0xa0,0xbf,0xc0,0xff];
// Every lead and every subsequent position at decoder decision boundaries.
for(let width=3;width<=4;width++)for(let lead=0;lead<256;lead++)for(let at=1;at<width;at++)for(const byte of edges){
 const seq=Array(width).fill(0x80);seq[0]=lead;seq[at]=byte;add(seq);
}
// Full products for all semantic valid-lead classes; distinguishes E0/ED/F0/F4 bounds.
for(const lead of [0xe0,0xe1,0xed,0xef])for(const b of edges)for(const c of edges)add([lead,b,c]);
for(const lead of [0xf0,0xf1,0xf4])for(const b of edges)for(const c of edges)for(const d of edges)add([lead,b,c,d]);
// Truncations, BOM, NUL and a long all-ASCII run.
for(const seq of [[0xc2,0xa2],[0xe0,0xa0,0x80],[0xed,0xa0,0x80],[0xf0,0x90,0x80,0x80],[0xf4,0x90,0x80,0x80],[0xef,0xbb,0xbf]])for(let k=1;k<=seq.length;k++)add(seq.slice(0,k));
add(Array(257).fill(0));add(Array(4096).fill(65));
let seed=0xdec0de;
for(let i=0;i<10000;i++){const seq=[];seed=(Math.imul(seed,1664525)+1013904223)>>>0;const len=seed%65;for(let j=0;j<len;j++){seed=(Math.imul(seed,1664525)+1013904223)>>>0;seq.push(seed>>>24);}add(seq);}
writeSync(out,chunk.subarray(0,used));closeSync(out);console.log(JSON.stringify({prefixes:count,expandedCases:count*65*8*2,seed:'0xdec0de'}));
