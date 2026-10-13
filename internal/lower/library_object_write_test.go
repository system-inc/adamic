package lower

import "testing"

func TestFrozenPropertyErrorIsCatchable(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `const object={value:1}; Object.freeze(object); try { object.value=2; } catch { console.log('caught'); }`)
	if err != nil {
		t.Fatal(err)
	}
}
