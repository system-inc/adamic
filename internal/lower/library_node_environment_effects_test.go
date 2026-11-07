package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestProcessEnvironmentMutationsAreRefusedInParallelWork(test *testing.T) {
	test.Parallel()
	for _, one := range []struct{ name, body string }{
		{"write", `process.env.ADAMIC_PARALLEL_ENV = key;`},
		{"element write", `process.env[key] = 'a';`},
		{"write value", `const value = (process.env.ADAMIC_PARALLEL_ENV = key);`},
		{"delete", `delete process.env.ADAMIC_PARALLEL_ENV;`},
		{"element delete", `delete process.env[key];`},
		{"delete value", `const ok = delete process.env.ADAMIC_PARALLEL_ENV;`},
	} {
		test.Run(one.name, func(test *testing.T) {
			test.Parallel()
			_, lowerSourceError := lowerSource(
				test,
				`import process from 'node:process'; import {parallelMap} from 'adamic';
const items: readonly string[] = ['ADAMIC_PARALLEL_ENV'];
parallelMap(items, (key: string): string => {`+one.body+` return key;});`,
			)
			var refused *Refused
			if !errors.As(lowerSourceError, &refused) ||
				!strings.Contains(refused.What, "parallelMap work writes state another task can reach") ||
				!strings.Contains(refused.What, "process environment") {
				test.Fatalf("want process environment effect refusal, got %v", lowerSourceError)
			}
		})
	}
}
