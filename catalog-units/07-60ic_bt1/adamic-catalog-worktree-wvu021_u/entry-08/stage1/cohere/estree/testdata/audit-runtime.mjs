// Successful operations use the ordinary Node oracle. Audit refusals are catchable per file.
export * from '../../../../oracle/adamic.mjs';
export function panic(message) { throw new Error(message); }
