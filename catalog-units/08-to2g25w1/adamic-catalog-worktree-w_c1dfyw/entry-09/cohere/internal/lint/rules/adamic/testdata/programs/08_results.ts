import { panic } from 'adamic';

type Parsed = { readonly kind: 'Ok'; readonly value: number } | { readonly kind: 'Error'; readonly message: string };

function parseAge(input: string): Parsed {
	const trimmed = input.trim();
	if (trimmed.length === 0) {
		return { kind: 'Error', message: 'empty' };
	}
	let value = 0;
	for (const character of trimmed) {
		const digit = character.charCodeAt(0) - 48;
		if (digit < 0 || digit > 9) {
			return { kind: 'Error', message: `'${character}' is not a digit` };
		}
		value = value * 10 + digit;
	}
	if (value > 150) {
		return { kind: 'Error', message: `${value} is too old` };
	}
	return { kind: 'Ok', value };
}

for (const input of ['42', ' 7 ', '', '12a', '200']) {
	const parsed = parseAge(input);
	switch (parsed.kind) {
		case 'Ok':
			console.log(`ok ${parsed.value}`);
			break;
		case 'Error':
			console.log(`error: ${parsed.message}`);
			break;
	}
}

const ages = new Map<string, number>([['Kirk', 42]]);
const age = ages.get('Ahra') ?? panic('no age on record for Ahra');
console.log(`Ahra is ${age}`);
