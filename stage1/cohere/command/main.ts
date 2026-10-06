// Host drivers supply argv[0] as the first argument and bind this typed result to their exit status.
// Valid execution requests stop loudly until the checker and process runner have a native port.
import { programArguments } from 'adamic';
import { help, parseArguments } from './flags.ts';
const argumentsFromHost = programArguments();
const program = argumentsFromHost[0] ?? 'cohere';
const renaming = argumentsFromHost[1] === 'rename';
const parsed = parseArguments(program, argumentsFromHost.slice(renaming ? 2 : 1), renaming);
export let commandExitCode = parsed.exitCode;
let stderr = parsed.stderr;
if(!parsed.stopped) {
    if(renaming && parsed.names.length !== 2) {
        stderr =
            help(program, true) + `cohere: expected a target and a new name, got ${parsed.names.length} arguments\n`;
    }
    else if(renaming && parsed.enabled('write') && parsed.enabled('dry-run')) {
        stderr =
            'cohere: --write and --dry-run contradict each other: --dry-run writes nothing, --write applies the rename\n';
    }
    else if(parsed.enabled('verbose') && parsed.enabled('json')) {
        stderr =
            'cohere: --verbose and --json contradict each other: --verbose prints the human account in full, --json prints JSON for a program\n';
    }
    else {
        stderr = 'cohere: stage 1 command execution is not yet ported\n';
    }
    commandExitCode = 1;
}
const lines = stderr.split('\n');
for(let index = 0; index < lines.length - 1; index++) {
    console.error(lines[index] ?? '');
}
