package lower

import "testing"

const literalReceiverStash = `const stash: { tag: string }[] = [];
const seen: { tag: string | undefined; keep(): void } = { tag: "a", keep(this: { tag: string }) { stash.push(this); } };
seen.keep();
seen.tag = undefined;
`

func TestLiteralReceiverFreshnessRejectsStoredReceiver(t *testing.T) {
	t.Parallel()
	for _, fixture := range []struct {
		name, tail string
	}{
		{"f1", "console.log(String(stash[0]?.tag));"},
		{"f2-guarded", "const first = stash[0];\nif (first !== undefined) console.log(String(first.tag.length));"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(fixture.name+extension, func(t *testing.T) {
				t.Parallel()
				arrowRefusedAt(t, literalReceiverStash+fixture.tail, extension, "2:57")
			})
		}
	}
}

func TestLiteralReceiverFreshnessRejectsWideReadOnlyUse(t *testing.T) {
	t.Parallel()
	const source = `const stash: { tag: string }[] = [];
const seen: { tag: string | undefined; keep(): void } = { tag: "a", keep(this: { tag: string }) { console.log(this.tag); } };
seen.keep();
seen.tag = undefined;
console.log(String(stash[0]?.tag));
`
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			arrowRefusedAt(t, source, extension, "2:57")
		})
	}
}

func TestLiteralReceiverFreshnessStillLowersControls(t *testing.T) {
	t.Parallel()
	for _, fixture := range []struct {
		name, source string
	}{
		{"arrow", `const seen: { tag: string | undefined; keep: () => void } = { tag: "a", keep: () => {} };
seen.keep();
seen.tag = undefined;
console.log(String(seen.tag));
`},
		{"exact", `const seen: { tag: string; keep(): void } = { tag: "a", keep() { console.log(this.tag); } };
seen.keep();
`},
		{"no-methods", `const seen: { tag: string | undefined } = { tag: "a" };
seen.tag = undefined;
console.log(String(seen.tag));
`},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(fixture.name+extension, func(t *testing.T) {
				t.Parallel()
				if _, err := lowerSourceAs(t, fixture.source, extension); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

// literalThisTypeStash is the evidence fixtures' head (stage3/meter/evidence/this-parameter-getter): a
// getter under a contextual ThisType<{ tag: string }> stores this, which Seen's tag then writes undefined
// through (#n7mgy3s).
const literalThisTypeStash = `interface Seen { tag: string | undefined; readonly keep: boolean }
const stash: { tag: string }[] = [];
const seen: Seen & ThisType<{ tag: string }> = { tag: "a", get keep() { stash.push(this); return true; } };
console.log(String(seen.keep));
seen.tag = undefined;
`

func TestLiteralThisTypeFreshnessRejectsStoredReceiver(t *testing.T) {
	t.Parallel()
	for _, fixture := range []struct {
		name, tail string
	}{
		{"g1", "console.log(String(stash[0]?.tag));"},
		{"g2", "const first = stash[0];\nif (first !== undefined) console.log(String(first.tag.length));"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(fixture.name+extension, func(t *testing.T) {
				t.Parallel()
				arrowRefusedAt(t, literalThisTypeStash+fixture.tail, extension, "3:48")
			})
		}
	}
}

func TestLiteralThisTypeFreshnessRejectsWideReadOnlyUse(t *testing.T) {
	t.Parallel()
	const source = `interface Seen { tag: string | undefined; readonly keep: boolean }
const seen: Seen & ThisType<{ tag: string }> = { tag: "a", get keep() { return this.tag.length > 0; } };
seen.tag = undefined;
console.log(String(seen.keep));
`
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			arrowRefusedAt(t, source, extension, "2:48")
		})
	}
}

func TestLiteralThisTypeFreshnessStillLowersControls(t *testing.T) {
	t.Parallel()
	for _, fixture := range []struct {
		name, source string
	}{
		{"own-this", `interface Seen { tag: string | undefined; readonly keep: boolean }
const seen: Seen = { tag: "a", get keep() { return this.tag !== undefined; } };
console.log(String(seen.keep));
seen.tag = undefined;
console.log(String(seen.keep));
`},
		{"arrow", `interface Held { tag: string | undefined; readonly keep: () => boolean }
const seen: Held & ThisType<{ tag: string }> = { tag: "a", keep: (): boolean => true };
console.log(String(seen.keep()));
seen.tag = undefined;
console.log(String(seen.tag));
`},
		{"no-this", `interface Seen { tag: string | undefined; readonly keep: boolean }
const seen: Seen & ThisType<{ tag: string }> = { tag: "a", get keep() { return true; } };
console.log(String(seen.keep));
seen.tag = undefined;
console.log(String(seen.tag));
`},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(fixture.name+extension, func(t *testing.T) {
				t.Parallel()
				if _, err := lowerSourceAs(t, fixture.source, extension); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}
