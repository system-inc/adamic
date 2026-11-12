import 'node:process';
function run(): void {
 function use(): void { session.post('Profiler.enable'); }
 use();
 const inspector: typeof import('inspector') = require('inspector');
 const session = new inspector.Session();
}
run();
