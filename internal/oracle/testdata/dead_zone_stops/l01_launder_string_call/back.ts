import { other } from '../l01_launder_string_call.ts';
const source = { other: 1, label: { toString(): string { return other(); } } };
export const view: { other: number } = source;
