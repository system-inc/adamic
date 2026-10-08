#!/usr/bin/env node
'use strict';
const fs = require('node:fs');
const path = require('node:path');
if (process.argv.length !== 4) throw new Error('usage: node family-mutant.cjs <family> <scratch-tree>');
const family = process.argv[2];
if (!['diagnostics', 'tuples', 'empty-arrays'].includes(family)) throw new Error('unknown family');
const spec = require(`./${family}.json`);
const edit = spec.edits[spec.undo_index];
const file = path.join(path.resolve(process.argv[3]), edit.file);
const text = fs.readFileSync(file, 'utf8');
if (text.split(edit.after).length !== 2) throw new Error('undo target missing or ambiguous');
fs.writeFileSync(file, text.replace(edit.after, edit.before));
console.log(JSON.stringify({ family, file: edit.file, restored: edit.before }, null, 2));
