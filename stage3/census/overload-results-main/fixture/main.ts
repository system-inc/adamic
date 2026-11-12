import { identity, honest } from './helper.js';
export function alpha() { return identity(1); }
export function beta() { return identity('two'); }
export function gamma() { let x = 0; x++; return x; }
export function delta() { return honest('four'); }
export function epsilon() { return new Date(); }
export function zeta() { const callback = honest; return callback('six'); }
