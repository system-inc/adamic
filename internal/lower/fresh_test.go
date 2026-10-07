package lower

import (
	"testing"
)

// A readonly field is a slot the cycle finder judges like any other (reviewer R, round 12): its
// constructor writes it, and when this was already put where the value reaches it, that write closes
// a cycle. One whose every write is proven, or whose type can't reach back, still compiles.
func TestReadonlyFieldsAreJudgedByTheirConstructorsWrites(t *testing.T) {
	t.Parallel()
	refused := "class Node {\n\treadonly kids: Node[] = [];\n\treadonly parent: Node | undefined;\n\tconstructor(parent: Node | undefined) {\n\t\tthis.parent = undefined;\n\t\tif (parent !== undefined) {\n\t\t\tparent.kids.push(this);\n\t\t}\n\t\tthis.parent = parent;\n\t}\n}\nconst root = new Node(undefined);\nconst child = new Node(root);\nconsole.log(`${root.kids.length}`);\n"
	program, err := lowerSource(t, refused)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.GraphTypes) == 0 || !program.Classes[0].Graph {
		t.Fatal("readonly constructor cycle has no graph ownership")
	}
	for _, sound := range []struct{ name, source string }{
		{"a readonly field that can't reach back", "interface Label {\n\treadonly text: string;\n}\nclass Box {\n\treadonly label: Label;\n\tconstructor(label: Label) {\n\t\tthis.label = label;\n\t}\n}\nconst box = new Box({ text: `b${1}` });\nconsole.log(box.label.text);\n"},
		{"a readonly parent pointer set before this goes anywhere", "class Node {\n\treadonly name: string;\n\treadonly parent: Node | undefined;\n\tconstructor(name: string, parent: Node | undefined) {\n\t\tthis.name = name;\n\t\tthis.parent = parent;\n\t}\n}\nconst root = new Node('root', undefined);\nconst child = new Node('child', root);\nconsole.log(child.parent === undefined ? 'none' : child.parent.name);\n"},
	} {
		if _, err := lowerSource(t, sound.source); err != nil {
			t.Errorf("%s: want it to compile, got %v", sound.name, err)
		}
	}
}
