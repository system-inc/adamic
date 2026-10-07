// Direct Go CSS printer boundaries, not the surrounding whole-file Prettier API.
import { format } from '../print.ts';
import { quote } from '../../selector/nodes.ts';
for(const text of ['\ufeffa{b:c}', '// x\ra{}', '\u00a0', '---\na:     b\n---\na{}']) {
    const result = format(text);
    console.log(result.kind === 'Refused' ? 'error ' + result.message : quote(result.text));
}
