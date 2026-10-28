// Spec: A checked predicate used as a value carries its check
// Expect: lowers
// In .ts isCat, unproven, is passed as an argument and checks its own claim at the call through test: the lie
// on the Dog stops there, loudly, before pet.meow is read (#vdt9f1q).
type Pet = { readonly kind: 'Cat'; readonly meow: string } | { readonly kind: 'Dog'; readonly bark: string };

let calls = 0;
function isCat(pet: Pet): pet is Extract<Pet, { kind: 'Cat' }> {
	calls = calls + 1;
	return calls > 0;
}

function visit(pet: Pet, test: (pet: Pet) => pet is Extract<Pet, { kind: 'Cat' }>): string {
	if (test(pet)) {
		return pet.meow;
	}
	return 'not a cat';
}

console.log(visit({ kind: 'Cat', meow: 'mew' }, isCat));
console.log(visit({ kind: 'Dog', bark: 'woof' }, isCat));
