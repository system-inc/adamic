import { runProcess, fileStatus } from 'adamic';
import type { Project } from './ownership.ts';
import { absoluteFrom } from './location.ts';
import { clean } from '../gitignore/path.ts';
export type ProjectRun = { output: string; exitCode: number; error: string };
export function projectLabel(project: Project): string {
    return project.configFile === '' ? project.directory : clean(project.directory + '/' + project.configFile);
}
function reason(code: string): string {
    if(code === 'ENOENT') return 'no such file or directory';
    if(code === 'ENOTDIR') return 'not a directory';
    if(code === 'EACCES') return 'permission denied';
    if(code === 'EPERM') return 'operation not permitted';
    if(code === 'ENOEXEC') return 'exec format error';
    return `unrepresented process error ${code}`;
}
function signalName(signal: number): string {
    if(signal === 1) return 'hangup';
    if(signal === 2) return 'interrupt';
    if(signal === 9) return 'killed';
    if(signal === 15) return 'terminated';
    return `unrepresented signal ${signal}`;
}
export function runProject(
    executable: string,
    root: string,
    project: Project,
    arguments_: string[],
    yieldFile: string,
): ProjectRun {
    const directory = absoluteFrom(root, project.directory);
    const prefix = ['--directory', directory];
    if(project.configFile !== '') {
        prefix.push('--tsconfig');
        prefix.push(project.configFile);
    }
    for(const argument of arguments_) prefix.push(argument);
    const environment = [
        'COHERE_VERDICT_FD',
        'COHERE_PROJECT_YIELD',
        `COHERE_PROJECT_ENGINE=${project.engine}`,
        `COHERE_PROJECT_LABEL=${projectLabel(project)}`,
    ];
    if(yieldFile !== '') environment.push(`COHERE_PROJECT_YIELD=${yieldFile}`);
    // Go os.StartProcess first stats Dir. Stat errors say chdir; an existing non-directory
    // passes that check and the child's chdir error instead says fork/exec with the executable.
    const status = fileStatus(directory);
    if(status.kind === 'Error') {
        const reason = status.message.endsWith(': no such file')
            ? 'no such file or directory'
            : status.message.endsWith(': permission denied')
              ? 'permission denied'
              : 'unrepresented filesystem error';
        return { output: '', exitCode: 1, error: `chdir ${directory}: ${reason}` };
    }
    const child = runProcess(executable, prefix, directory, environment);
    let error = '';
    if(child.error !== '') {
        const fields = child.error.split(':');
        error = `fork/exec ${executable}: ${reason(fields[1] ?? '')}`;
    }
    else if(child.signal !== 0) error = `signal: ${signalName(child.signal)}`;
    return { output: child.output, exitCode: error !== '' ? 1 : child.exitCode, error };
}
