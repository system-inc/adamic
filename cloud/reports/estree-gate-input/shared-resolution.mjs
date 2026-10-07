import {createRequire} from 'node:module';
import assert from 'node:assert/strict';
const root='/tmp/adamic-gate/estree-shared';const require=createRequire(root+'/package.json');
for(const [name,version] of [['prettier','3.9.6'],['yaml','2.9.0'],['yaml-unist-parser','3.2.0'],['typescript','6.0.3'],['@typescript-eslint/typescript-estree','8.65.0']]) {assert.equal(require(name+'/package.json').version,version);console.log(name,version);}
const prettier=await import(root+'/node_modules/prettier/index.mjs');
for(const [parser,text] of [['css','a{color:red}'],['typescript','const answer:number=42'],['yaml','answer: 42\n']])console.log(parser,await prettier.format(text,{parser}));
assert.equal(require('yaml').parse('answer: 42').answer,42);
const unist=await import(root+'/node_modules/yaml-unist-parser/dist/parse.mjs');assert.equal(unist.parse('answer: 42').type,'root');
console.log('PASS: all shared module paths and formatter plugins');
