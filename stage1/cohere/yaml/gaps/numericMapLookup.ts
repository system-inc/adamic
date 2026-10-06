// Same numeric-key lookup checksum with integer or fractional key distributions.
import { programArguments } from 'adamic';
const args = programArguments();
const count = Number(args[0] ?? '185');
const fractional = args[1] === 'fractional';
const rounds = Number(args[2] ?? '10000');
const values = new Map<number, number>();
for(let index = 0; index < count; index++) values.set(fractional ? index / count : index, index);
let checksum = 0;
for(let round = 0; round < rounds; round++) {
    for(let index = 0; index < count; index++) checksum += values.get(fractional ? index / count : index) ?? -1;
}
console.log(String(checksum));
