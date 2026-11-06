// A local that shadows a host binding is judged as the local (performance.ts).
import { declared, run } from './performance.ts';

console.log(run((eventName: string): void => {
    console.log(`fired ${eventName}`);
}, 'one'));
console.log(declared());
