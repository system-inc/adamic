'use strict';
const fs=require('fs'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript'),assert=require('assert/strict');
const file='/workspace/scratch/levenshtein/assertion-proposal-owner.ts',candidate=fs.readFileSync('/workspace/scratch/levenshtein/candidate-owner.ts','utf8');let proposed=candidate,reads=0;
for(const expression of ['previous[j - 1]','previous[j]','current[j - 1]','previous[s2.length]']){reads+=proposed.split(expression).length-1;proposed=proposed.replaceAll(expression,expression+'!');}
assert.equal(reads,6);fs.writeFileSync(file,proposed);const options={strict:true,exactOptionalPropertyTypes:true,noUncheckedIndexedAccess:true,noEmit:true,target:ts.ScriptTarget.ES2024,types:[]},p=ts.createProgram([file],options),diagnostics=ts.getPreEmitDiagnostics(p).map(d=>({code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
assert.equal(diagnostics.length,0);console.log(JSON.stringify({status:'measurement only; outside the requested unchanged-body scope',additionalNonNullAssertions:reads,diagnostics},null,2));
