package oracle

import (
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRegExpProtocolHostIdentities(t *testing.T) {
	source := "import {readdirSync, closeSync, rmSync} from 'node:fs';\nimport {performance} from 'node:perf_hooks';\ntry { readdirSync('walking/missing'); } catch (error) {\n if (error instanceof Error) console.log(`directory ${error.constructor === Error}`);\n}\ntry { readdirSync('\\0'); } catch (error) {\n if (error instanceof Error) {\n  console.log(`path ${error.constructor === TypeError} ${error instanceof TypeError}`);\n  error.name = 'RangeError';\n  console.log(`renamed ${error.constructor === TypeError} ${error instanceof RangeError}`);\n }\n}\ntry { closeSync(2147483647); } catch (error) { if (error instanceof Error) console.log(`file-system ${error.constructor === Error}`); }\ntry { closeSync(-1); } catch (error) {\n if (error instanceof Error) console.log(`file ${error.constructor === RangeError}`);\n}\ntry { rmSync('walking'); } catch (error) {\n if (error instanceof Error) console.log(`system ${error.constructor === Error} ${error instanceof TypeError}`);\n}\ntry { process.exitCode = 0.5; } catch (error) {\n if (error instanceof Error) console.log(`exit ${error.constructor === RangeError}`);\n}\ntry { process.chdir('l2-identity-missing-directory'); } catch (error) {\n if (error instanceof Error) console.log(`directory-process ${error.constructor === Error}`);\n}\ntry { performance.measure('missing', 'l2-identity-missing-mark'); } catch (error) {\n if (error instanceof Error) console.log(`performance ${error.constructor === Error} ${error.constructor === SyntaxError} ${error instanceof SyntaxError}`);\n}"
	path := filepath.Join(t.TempDir(), "host-identities.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "native")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	how := inputRun{directory: filepath.Join(repository, "internal/oracle/testdata")}
	environment := []string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"}
	if runtime.GOOS == "linux" {
		environment[0] = "ASAN_OPTIONS=detect_leaks=1"
	}
	actual := executeInput(t, how, environment, binary)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("native did not finish cleanly: %d %s", actual.exitCode, actual.stderr)
	}
	truth := onNodeWith(t, how, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		t.Fatalf("Node did not finish cleanly: %d %s", truth.exitCode, truth.stderr)
	}
	if difference := disagreement(truth, actual); difference != "" {
		t.Fatalf("independent Node comparison: %s\nNode %s\nnative %s", difference, truth.stdout, actual.stdout)
	}
	t.Log("both exit zero, empty stderr, native ASan/UBSan clean; LeakSanitizer enabled only on Linux")
}
