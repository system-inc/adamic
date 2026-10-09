import { DocNode } from '/tmp/defend-css/profile-C1/source/css/print_doc.ts';
const doc = new DocNode('text', [], '🐈abc');
let total = 0;
for(let i = 0; i < 20; i++) total += doc.width();
console.log(total);
