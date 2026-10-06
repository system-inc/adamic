package lower

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestInputTupleSpreadsAreNotYet(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`import { writeTextFile } from 'adamic'; const pair: [string, string] = ['/tmp/adamic-review-spread.txt', 'text']; const result = writeTextFile(...pair); console.log(result.kind);`,
		`import { readTextFile } from 'adamic'; const one: [string] = ['/tmp/file']; const result = readTextFile(...one); console.log(result.kind);`,
	} {
		_, err := lowerSource(t, source)
		var gap *NotYet
		if !errors.As(err, &gap) || !strings.HasSuffix(gap.Where, fmt.Sprintf("main.a:1:%d", strings.Index(source, "...")+1)) || !strings.Contains(gap.What, "SpreadElement") {
			t.Errorf("want NotYet with where and SpreadElement, got %v", err)
		}
	}
}
