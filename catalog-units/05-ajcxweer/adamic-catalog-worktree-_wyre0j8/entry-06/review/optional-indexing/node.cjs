const fs = require('node:fs');
const {stripTypeScriptTypes} = require('node:module');
const source = fs.readFileSync(process.argv[2], 'utf8');
(0,eval)(stripTypeScriptTypes(source));
