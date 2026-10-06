// A mixed-project run needs process creation, output collection and exit status.
import { execFileSync } from 'node:child_process';
console.log(execFileSync('/usr/bin/true').length);
