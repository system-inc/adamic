package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestScoutNullishKeyRepresentations(t *testing.T) {
	t.Parallel()
	for _, key := range []string{"null", "undefined", "null | undefined", "number | null", "string | null | undefined", "number | string | boolean | null | undefined"} {
		t.Run(key, func(t *testing.T) {
			_, err := lowerSource(t, "function run(key: "+key+"): void {const map = new Map<"+key+", number>(); const set = new Set<"+key+">(); map.set(key,1); set.add(key); console.log(String(map.has(key))); console.log(String(set.has(key)));}")
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestScoutNullishCoalesceRepresentationGap(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "function read(value: number | null): number {return value ?? 1;}")
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(err.Error(), "a tagged nullable union coalesced into a different representation") {
		t.Fatalf("want loud conversion gap, got %v", err)
	}
}

func TestScoutNullableUnknownKeyGap(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "function read(key: {readonly name:string} | null): void {const map = new Map<unknown, number>([[key,1]]);}")
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(err.Error(), "a nullable reference viewed as unknown or object") {
		t.Fatalf("want existing boundary gap, got %v", err)
	}
}

func TestScoutNullishLookupPresenceGap(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		`const keys=new Set<unknown>();keys.add(map.get('absent'));`,
		`const keys=new Map<unknown,number>([[map.get('absent'),1]]);`,
		`const key=flag ? map.get('absent') : undefined;`,
	} {
		_, err := lowerSource(t, `function run(map:Map<string,{readonly name:string}|null>,flag:boolean):void{`+body+`}`)
		var gap *NotYet
		if !errors.As(err, &gap) || !strings.Contains(err.Error(), "a nullable lookup boxed as a union without its presence slot") {
			t.Fatalf("want presence conversion gap, got %v", err)
		}
	}
}
