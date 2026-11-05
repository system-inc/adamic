import { other } from '../l03_launder_concat.ts';
const source = { other: 1, label: { toString(): string { return other(); } } };
export const view: { other: number } = source;
