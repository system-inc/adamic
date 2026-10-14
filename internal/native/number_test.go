package native

import (
	"bytes"
	"context"
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

// runWithInput runs a command on input and returns its stdout. Like the oracle's, the command can't
// outlive its test: a deadline, its own process group, and the group killed when either ends.
func runWithInput(t *testing.T, input string, name string, arguments ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)
	command := programCommand(ctx, "", name, arguments...)
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
