// Spec: | User-defined type guards `x is T`, and `asserts x is T` |
// Expect: refused a type predicate whose return is not proven
type Pet = { readonly kind: 'Cat'; readonly meow: () => string } | { readonly kind: 'Dog'; readonly bark: () => string };

function isCat(pet: Pet): pet is Extract<Pet, { kind: 'Cat' }> {
	return pet.kind === 'Dog';
}

const pets: readonly Pet[] = [{ kind: 'Dog', bark: () => 'woof' }];
for (const pet of pets) {
	if (isCat(pet)) {
		console.log(pet.meow());
	}
}
