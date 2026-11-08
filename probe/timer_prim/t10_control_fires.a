// Control: the id is computed but never handed to clearTimeout, so the timer fires on Node.
import 'node:timers';

const t = setTimeout(() => { console.log('fired'); }, 5);
const id = +t;
console.log(id > 0 ? 'positive' : 'not');
