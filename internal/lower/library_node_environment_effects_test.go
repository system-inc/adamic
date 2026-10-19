package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestProcessEnvironmentMutationsAreRefusedInParallelWork(t *testing.T) {
	t.Parallel()
	for _, one := range []struct{ name, body string }{
		{"write", `process.env.ADAMIC_PARALLEL_ENV = key;`},
		{"element write", `process.env[key] = 'a';`},
		{"write value", `const value = (process.env.ADAMIC_PARALLEL_ENV = key);`},
		{"delete", `delete process.env.ADAMIC_PARALLEL_ENV;`},
		{"element delete", `delete process.env[key];`},
		{"delete value", `const ok = delete process.env.ADAMIC_PARALLEL_ENV;`},
	} {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, `import process from 'node:process'; import {parallelMap} from 'adamic';
const items: readonly string[] = ['ADAMIC_PARALLEL_ENV'];
parallelMap(items, (key: string): string => {`+one.body+` return key;});`)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, "parallelMap work writes state another task can reach") || !strings.Contains(refused.What, "process environment") {
				t.Fatalf("want process environment effect refusal, got %v", err)
			}
		})
	}
}
