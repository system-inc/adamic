package lower

import (
	"os"
	"testing"
)

// These negative witnesses pin admissibility rather than a diagnostic substring.
// Dropping the receiver proof must actually admit the program to fail the test.
func TestMethodDestructuringSafety(t *testing.T) {
	t.Parallel()
	receiver, err := os.ReadFile("testdata/method_destructuring_receiver.a")
	if err != nil {
		t.Fatal(err)
	}
	for name, source := range map[string]string{
		"receiver":            string(receiver),
		"arrow receiver":      `const host = { value: 7, read(): () => number { return () => this.value; } }; const { read } = host; console.log(read()().toString());`,
		"default receiver":    `const host = { value: 7, read(this: { value: number }, value: number = this.value): number { return value; } }; const { read } = host; console.log('done');`,
		"explicit receiver":   `const host = { value: 7, read(this: {value: number} | undefined): number { return this === undefined ? -1 : this.value; } }; const { read } = host; console.log('done');`,
		"structural receiver": `interface Host { read(): number; } function extract({read}: Host): () => number { return read; } const host = { value: 7, read(): number { return this.value; } }; console.log(extract(host)().toString());`,
		"class origin":        `class Host { read(): number { return 7; } } const { read } = new Host(); console.log('done');`,
		"structural class":    `interface View { read(): number; } function extract({read}: View): () => number { return read; } class Host { read(): number { return 7; } } console.log(extract(new Host())().toString());`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := lowerSource(t, source); err == nil {
				t.Fatal("receiver or prototype origin was erased")
			}
		})
	}
}

func TestReceiverIndependentMethodDestructuring(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"const own={next():number{return 1;}};const {next}=own;console.log(`${next()}`);",
		"const own={next():number{return 1;}};const view:{next:()=>number}=own;const {next}=view;console.log(`${next()}`);",
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
}
