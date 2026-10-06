import { panic, programArguments } from 'adamic';
import { types } from '../facts.ts';
const args = programArguments();
const facts = types(args[0] ?? panic('missing facts'), 'raw-type');
console.log(facts.root().flags.toString());
