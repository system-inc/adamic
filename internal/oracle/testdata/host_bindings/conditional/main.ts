// The true arm of ?: under onProfilerEvent's typeof guard (performance.ts).
import { arity, named } from './performance.ts';

console.log(String(arity()));
console.log(named());
