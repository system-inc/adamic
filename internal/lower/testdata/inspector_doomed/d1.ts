import 'node:process';
let leak = () => {};
try {
 leak = () => { session.post('Profiler.enable'); };
 const inspector: typeof import('inspector') = require('inspector');
 const session = new inspector.Session();
} catch { console.log('require threw'); }
try { leak(); } catch (error) {
 if (error instanceof Error) console.log('caught ' + error.name);
}
