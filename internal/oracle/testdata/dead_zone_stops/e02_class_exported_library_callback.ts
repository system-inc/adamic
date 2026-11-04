import { callBack } from './e02_class_exported_library_callback/back.ts';
function make(): string { return new Late().label; }
let first = '';
try { first = '1'.replace('1', callBack); } catch (error) { first = error instanceof Error ? error.name : 'other'; }
class Late { label = 'late'; }
console.log(first);
console.log('1'.replace('1', callBack));
export { make as other };
