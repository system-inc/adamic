package oracle

import "testing"

// A throwaway probe for #0ty4971, never landed: every program reads .length through a Function-typed value.
func TestZZLength(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"p01_plain.ts", "p02_bind_param.ts", "p03_bind_return.ts", "p04_bind_field.ts", "p05_bind_array.ts", "p06_nested.ts", "p07_method.ts", "p08_library.ts", "p09_iterator_next.ts", "p10_collator.ts", "p11_class.ts", "p12_async_generator.ts", "p15_bind_method.ts", "p16_bind_closure_value.ts", "p17_bind_local.ts"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			generatorAgrees(t, "internal/oracle/testdata/zzlen/"+name)
		})
	}
}
