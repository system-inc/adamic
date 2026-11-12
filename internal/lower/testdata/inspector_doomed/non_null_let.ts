import 'node:process';
import type { Session } from 'inspector';
function run(): string {
 let activeSession: Session | undefined;
 function enable(): void {
  const inspector: typeof import('inspector') = require('inspector');
  const session = new inspector.Session();
  activeSession = session;
 }
 function disable(): string {
  if (activeSession) {
   const s = activeSession;
   activeSession.post('Profiler.stop', () => {
    activeSession = undefined;
    s.disconnect();
   });

   return 'stopped';
  }
  return 'disabled';
 }
 activeSession!.post('Profiler.stop'); return 'unreachable';
}
console.log(run());
