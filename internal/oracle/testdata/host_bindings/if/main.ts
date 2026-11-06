// tsc's guarded read of onProfilerEvent, in the then-branch of if (performance.ts).
import { counted, mark, markUnlessUndefined, profiler } from './performance.ts';

mark('beforeParse');
markUnlessUndefined('afterParse');
console.log(String(counted()));
console.log(profiler());
