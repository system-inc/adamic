// Spec: | User-defined type guards `x is T`, and `asserts x is T` |
// Expect: lowers
// In .ts the unproven predicate is a checked predicate (R-T80, #7v4rjr7): isCat's answer is checked at
// the call, so the lie stops there, loudly, before pet.meow is read.
type Pet = { readonly kind: 'Cat'; readonly meow: string } | { readonly kind: 'Dog'; readonly bark: string };

function isCat(pet: Pet): pet is Extract<Pet, { kind: 'Cat' }> {
	return pet.kind === 'Dog';
}

const pets: readonly Pet[] = [{ kind: 'Dog', bark: 'woof' }];
for (const pet of pets) {
	if (isCat(pet)) {
		console.log(pet.meow);
	}
}
