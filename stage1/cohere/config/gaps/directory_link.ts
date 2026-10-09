// A matched source directory link needs canonical visitation, which today's file API cannot supply.
import { panic, programArguments } from 'adamic';
import { readProjectConfig } from '../tsconfig.ts';
const path = programArguments()[0] ?? panic('usage: directory_link.ts <tsconfig>');
const result = readProjectConfig(path);
if (result.kind === 'Error') { console.log(result.message); }
else { for (const file of result.project.files) { console.log(file); } }
