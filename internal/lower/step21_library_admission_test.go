package lower

import "testing"

// These were NotYet on main. The corresponding step21_library_failures,
// step21_library_types and step21_library_host fixtures hold the failures to Node.
func TestStep21LibraryAdmission(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"try { console.log('a'.replaceAll(/a/, 'b')); } catch {}",
		"const v = new Uint8Array(2); try { v.set(new Uint8Array(3)); } catch { console.log('caught'); }",
		"function size(): number { return -1; } try { const v = new Uint8Array(size()); } catch { console.log('caught'); }",
		"const object={value:1}; Object.freeze(object); try { object.value=2; } catch { console.log('caught'); }",
		"import { createHash } from 'node:crypto'; try { createHash('sha256').update('a').digest('hex'); } catch {}",
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Errorf("catchable library call: %s: %v", source, err)
		}
	}
}
