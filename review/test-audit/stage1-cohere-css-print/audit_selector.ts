import { readTextFile } from 'adamic';
const selected = readTextFile('/tmp/u079/selector');
export const auditMutation = selected.kind === 'Error' ? '' : selected.text.trim();
