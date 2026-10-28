// Spec: A checked predicate used as a value carries its check
// Expect: lowers
// In .ts the alias test carries isCat's check, so the call through it checks the claim as a direct call would (#vdt9f1q).
type Pet = { readonly kind: 'Cat'; readonly meow: string } | { readonly kind: 'Dog'; readonly bark: string };

let calls = 0;
function isCat(pet: Pet): pet is Extract<Pet, { kind: 'Cat' }> {
	calls = calls + 1;
	return calls > 0;
}

const test = isCat;
const pets: readonly Pet[] = [{ kind: 'Cat', meow: 'mew' }, { kind: 'Dog', bark: 'woof' }];
for (const pet of pets) {
	if (test(pet)) {
		console.log(pet.meow);
	}
}
