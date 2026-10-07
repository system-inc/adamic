package load

import (
	"fmt"
	"testing"
)

func TestStarExportCollisionNamesBothModules(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"main.a", "export * from './left.a';\nexport * from './right.a';\n"},
		[2]string{"left.a", "export const shared = 1;\n"},
		[2]string{"right.a", "export const shared = 2;\n"},
	)
	_, err := Load(paths[:1])
	want := fmt.Sprintf("%s:2:1: error TS2308: Adamic 0.1 refuses export * collision for 'shared' between './left.a' and './right.a'; explicitly re-export one binding", paths[0])
	if err == nil || err.Error() != want {
		t.Fatalf("want exact refusal %q, got %v", want, err)
	}
}

func TestExplicitExportResolvesStarCollision(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"main.a", "export * from './left.a';\nexport * from './right.a';\nexport { shared } from './left.a';\n"},
		[2]string{"left.a", "export const shared = 1;\n"},
		[2]string{"right.a", "export const shared = 2;\n"},
	)
	if _, err := Load(paths[:1]); err != nil {
		t.Fatal(err)
	}
}
