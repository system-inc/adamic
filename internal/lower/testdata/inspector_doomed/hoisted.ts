import 'node:process';
function run(): void {
 function use(): void { session.post('Profiler.enable'); }
 try { use(); } catch { console.log('early'); }
 const inspector: typeof import('inspector') = require('inspector');
 const session = new inspector.Session();
}
try { run(); } catch {}
