const knownUnits: readonly string[] = 'em rem px'.split(' ');
export class Printer { readonly units: number; constructor(units: readonly string[]) { this.units = units.length; } }
export function format(): string { return String(new Printer(knownUnits).units); }
console.log(format());
