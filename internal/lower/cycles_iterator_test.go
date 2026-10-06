package lower

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func TestIteratorInterfaceSlotsRefuseClosingWrite(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"cyc_iface.a", "cyc_iterable.a"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join("testdata", "cycles", name)
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			_, err = Lower(context.Background(), program)
			var refusal *Refused
			if !errors.As(err, &refusal) {
				t.Fatalf("want cycle refusal, got %v", err)
			}
			for _, want := range []string{"Holder.cursor", "adamic/cycle-capable", "Weak<", name + ":5:1"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("want %q in refusal: %v", want, err)
				}
			}
			t.Log(err)
		})
	}
}

func TestIteratorInterfaceSlotsAllowAcyclicControls(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"cyc_iface_control.a"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join("..", "oracle", "testdata", name)
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			if _, err = Lower(context.Background(), program); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Array iterator loops are lowered, but storing array.values() in a slot is not supported yet.
func TestArrayIteratorInterfaceSlotGap(t *testing.T) {
	t.Parallel()
	program, err := load.Load([]string{filepath.Join("testdata", "cycles", "cyc_array_iface_control.a")})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Lower(context.Background(), program)
	var refusal *Refused
	if !errors.As(err, &refusal) || !strings.Contains(err.Error(), "inherited library member values") {
		t.Fatalf("want the stored array iterator's explicit gap, got %v", err)
	}
}

func TestIteratorEmptyCollectionProofRejectsOtherUses(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, setup string }{
		{"alias", "const alias = map; alias.set('a'.repeat(2), holder);"},
		{"helper", "function keep(value: Map<string, Holder>, item: Holder): void { value.set('aa', item); } keep(map, holder);"},
		{"later insertion", "holder.cursor = map.keys(); map.set('a'.repeat(2), holder);"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, `type Holder = { cursor: Iterator<string> | undefined };
const map = new Map<string, Holder>();
const holder: Holder = { cursor: undefined };
`+probe.setup+`
holder.cursor = map.keys();
console.log('done');`)
			var refusal *Refused
			if !errors.As(err, &refusal) || !strings.Contains(err.Error(), "adamic/cycle-capable") {
				t.Fatalf("want cycle refusal after collection escaped or changed, got %v", err)
			}
		})
	}
}

func TestIteratorCapturesReachOtherLibraryInterfaces(t *testing.T) {
	t.Parallel()
	for _, view := range []string{"Iterable<string>", "IteratorObject<string>"} {
		t.Run(view, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, `type Holder = { cursor: `+view+` | undefined };
const map = new Map<string, Holder>();
const holder: Holder = { cursor: undefined };
map.set('a'.repeat(2), holder);
holder.cursor = map.keys();`)
			var refusal *Refused
			if !errors.As(err, &refusal) || !strings.Contains(err.Error(), "adamic/cycle-capable") {
				t.Fatalf("want assignable iterator capture refusal, got %v", err)
			}
		})
	}
}
