package native

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// numberHarness formats doubles, given as hexadecimal bit patterns on stdin, one per line.
const numberHarness = `#include "adamic.h"

#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main(void) {
	char line[64];
	while (fgets(line, sizeof line, stdin) != NULL) {
		uint64_t bits = strtoull(line, NULL, 16);
		double value;
		memcpy(&value, &bits, sizeof value);
		char buffer[ADAMIC_NUMBER_FORMAT_MAX];
		size_t length = adamic_number_format(value, buffer);
		fwrite(buffer, 1, length, stdout);
		fputc('\n', stdout);
	}
	return 0;
}
`

// numberOracle is the same on Node: String(value), JavaScript's own answer.
const numberOracle = `const view = new DataView(new ArrayBuffer(8));
const lines = require('node:fs').readFileSync(0, 'utf8').trim().split('\n');
const out = lines.map((line) => { view.setBigUint64(0, BigInt('0x' + line)); return String(view.getFloat64(0)); });
process.stdout.write(out.join('\n') + '\n');
`

// sweep is the doubles the formatter is held to: every boundary ECMAScript's rules have, the values
// people print most, and random bit patterns from a fixed seed, so a failure reproduces.
func sweep() []float64 {
	values := []float64{
		0, math.Copysign(0, -1), math.NaN(), math.Inf(1), math.Inf(-1),
		1, -1, 0.1, 0.2, 0.1 + 0.2, 0.5, 1.5, 2.5, 100, 1234.5678,
		1e21, 1e21 - 65536, 999999999999999900000, 123456789012345680000,
		1e-6, 1e-7, 1.5e-7, 0.000001234, 0.0000001234,
		math.MaxFloat64, math.SmallestNonzeroFloat64, 2.2250738585072014e-308,
		1 << 53, 1<<53 + 2, -(1 << 53), math.Pi, math.E, 26.568583470577035, 1.4142135623730951,
	}
	for integer := -1000; integer <= 2000; integer++ {
		values = append(values, float64(integer))
	}
	// Whole numbers are written directly below 2^53 and searched for from there: both sides of that
	// edge, every power of two up to 2^60 and the integers beside it, and every 10^k - 1 and 10^k.
	for power := 0; power <= 60; power++ {
		two := math.Ldexp(1, power)
		values = append(values, two, two-1, two+1, -two, -(two - 1))
	}
	nine := 0.0
	for digits := 1; digits <= 22; digits++ {
		nine = nine*10 + 9
		values = append(values, nine, nine+1, -nine)
	}
	values = append(values, 9007199254740991, 9007199254740990, 9007199254740989, 4503599627370495.5)
	for exponent := -330; exponent <= 310; exponent++ {
		power := math.Pow(10, float64(exponent))
		values = append(values, power, math.Nextafter(power, 0), math.Nextafter(power, math.Inf(1)), -power)
	}
	random := rand.New(rand.NewPCG(1984, 628))
	for range 200_000 {
		values = append(values, math.Float64frombits(random.Uint64()))
	}
	for range 50_000 {
		// Short decimals, as programs write them: an integer over a power of ten.
		values = append(values, float64(random.IntN(1_000_000))/math.Pow(10, float64(random.IntN(12))))
	}
	return values
}

func TestNumbersFormatExactlyAsJavaScriptDoes(t *testing.T) {
	t.Parallel()
	values := sweep()
	var input strings.Builder
	for _, value := range values {
		fmt.Fprintf(&input, "%016x\n", math.Float64bits(value))
	}

	binary := filepath.Join(t.TempDir(), "harness")
	if err := Build(numberHarness, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	native := runWithInput(t, input.String(), binary)
	oracle := runWithInput(t, input.String(), "node", "--eval", numberOracle)

	nativeLines := strings.Split(strings.TrimSuffix(native, "\n"), "\n")
	oracleLines := strings.Split(strings.TrimSuffix(oracle, "\n"), "\n")
	if len(nativeLines) != len(values) || len(oracleLines) != len(values) {
		t.Fatalf("got %d native lines and %d from Node for %d values", len(nativeLines), len(oracleLines), len(values))
	}
	mismatches := 0
	for index, value := range values {
		if nativeLines[index] != oracleLines[index] {
			mismatches++
			if mismatches <= 10 {
				t.Errorf("%016x: native %q, Node %q", math.Float64bits(value), nativeLines[index], oracleLines[index])
			}
		}
	}
	if mismatches > 0 {
		t.Errorf("%d of %d values format differently", mismatches, len(values))
	}
	t.Logf("%d values, %d mismatches", len(values), mismatches)
}

// runWithInput runs a command on input and returns its stdout. Like the oracle's, the command can't
// outlive its test: a deadline, its own process group, and the group killed when either ends.
func runWithInput(t *testing.T, input string, name string, arguments ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)
	command := exec.CommandContext(ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		return syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	}
	command.WaitDelay = 5 * time.Second
	command.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("%s: %v\n%s", name, err, stderr.String())
	}
	return stdout.String()
}
