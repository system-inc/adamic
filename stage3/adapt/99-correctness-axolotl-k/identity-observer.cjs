const fs = require('node:fs');
const path = require('node:path');
const {createRequire} = require('node:module');
if (!process.send && process.env.SAMEMAP_IDENTITIES && path.basename(process.argv[1] || '') === 'run.js') {
 const Mocha = createRequire(path.join(process.cwd(), 'package.json'))('mocha');
 const original = Mocha.Runner.prototype.emit;
 const rows = [];
 Mocha.Runner.prototype.emit = function (...args) {
  const [event,test] = args;
  if (['pass','fail','pending'].includes(event)) rows.push([event,test.titlePath().filter(Boolean).join(' ')]);
  return Reflect.apply(original,this,args);
 };
 process.on('exit',()=>fs.writeFileSync(process.env.SAMEMAP_IDENTITIES,JSON.stringify(rows.sort())));
}
