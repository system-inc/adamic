'use strict';
const {fs,path,assert} = require('./engine.cjs');
const pin = require('./inventory.json').sites[0];
const file = path.join(path.resolve(process.argv[2]),'src/compiler',pin.file);
const text = fs.readFileSync(file,'utf8');
assert(text.includes(pin.after), 'adapted site absent');
fs.writeFileSync(file,text.replace(pin.after,pin.before));
console.log(JSON.stringify({file:pin.file,owner:pin.owner,restored:pin.before}));
