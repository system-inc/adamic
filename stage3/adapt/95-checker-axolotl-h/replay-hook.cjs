const fs=require('node:fs'),path=require('node:path');
const write=fs.writeFileSync,append=fs.appendFileSync,rows=[];
fs.writeFileSync=function(file,...args){if(typeof file==='string'&&file!==process.env.REPLAY_LOG)rows.push(path.relative(process.env.REPLAY_TREE,file));return write.call(this,file,...args);};
process.on('exit',()=>append(process.env.REPLAY_LOG,JSON.stringify(rows)+'\n'));
