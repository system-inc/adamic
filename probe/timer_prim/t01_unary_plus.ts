// #995spev host row: a Timeout's Symbol.toPrimitive fills Node's knownTimersById, so clearTimeout of the id it gave cancels the timer.
import 'node:timers';

const t = setTimeout(() => { console.log('fired'); }, 5);
const id = +t; clearTimeout(id);
console.log('cleared');
