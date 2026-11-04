import { other } from '../l04_patched_prototype.ts';
String.prototype.split = function (): string[] { return [other()]; };
export const ready = true;
