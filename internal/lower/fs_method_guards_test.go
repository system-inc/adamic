package lower

import (
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestParseIntMapUsesIndexRadix(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNode(test, "console.log(['10','10','10'].map(Number.parseInt).join(','));")
}

// Absolute paths isolate parallel rows while both Node runs use the same path.
// Each successful row removes its files before the next execution.
func fileSystemAgreementSource(test *testing.T, source string) string {
	test.Helper()
	for _, name := range []string{"x", "missing"} {
		source = strings.ReplaceAll(source, "'"+name+"'", strconv.Quote(filepath.Join(test.TempDir(), name)))
	}
	return source
}

func TestFileSystemOpenStringFlagsLower(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNode(
		test,
		fileSystemAgreementSource(
			test,
			`import {writeFileSync,openSync,closeSync,readFileSync,unlinkSync} from 'node:fs';
 writeFileSync('x','opened bytes'); const fd=openSync('x','r'); closeSync(fd);
 console.log(readFileSync('x','utf8')); unlinkSync('x');`,
		),
	)
}

func TestFileSystemRemoveDefaultRetryDelayLowers(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNode(
		test,
		fileSystemAgreementSource(test, `import {writeFileSync,existsSync,rmSync} from 'node:fs';
 writeFileSync('x','remove me'); console.log(existsSync('x')?'present':'absent');
 rmSync('x',{retryDelay:100}); console.log(existsSync('x')?'present':'absent');`),
	)
}

func TestFileSystemExistsOperation(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNode(
		test,
		fileSystemAgreementSource(test, `import {existsSync,writeFileSync,unlinkSync} from 'node:fs';
 console.log(existsSync('x')?'present':'absent'); writeFileSync('x','exists');
 console.log(existsSync('x')?'present':'absent'); unlinkSync('x');
 console.log(existsSync('x')?'present':'absent');`),
	)
}

func TestFileSystemStatThrowsByDefault(test *testing.T) {
	test.Parallel()
	// Missing-file behavior distinguishes throwIfNoEntry's default: Node exits
	// before the final print; returning undefined instead continues and exits zero.
	lowersAndAgreesWithNodeExit(test, fileSystemAgreementSource(test, `import {existsSync,statSync} from 'node:fs';
 console.log(existsSync('missing')?'present':'absent'); statSync('missing');
 console.log('stat returned instead of throwing');`), 70)
}
