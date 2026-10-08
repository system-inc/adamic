package lint

import (
	"fmt"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
	"strconv"
	"strings"
	"testing"
)

func checkerReadMismatch(descriptors []registry.Descriptor, declarations string) error {
	masks := map[string]uint64{}
	for _, row := range strings.Split(strings.TrimSpace(declarations), "\n") {
		fields := strings.Split(row, "\t")
		if len(fields) != 2 {
			return fmt.Errorf("invalid upstream reads row %q", row)
		}
		mask, err := strconv.ParseUint(fields[1], 10, 8)
		if err != nil {
			return err
		}
		masks[fields[0]] = mask
	}
	flags := map[string]uint64{"ReadsCompilerOptions": 1, "ReadsDefaultLibrary": 2, "ReadsModuleResolution": 4, "ReadsOtherFiles": 8, "ReadsDesignSystem": 16}
	for _, descriptor := range descriptors {
		if !descriptor.Typed {
			continue
		}
		expected, exists := masks[descriptor.Name]
		if !exists {
			return fmt.Errorf("no upstream reads for %s", descriptor.Name)
		}
		var actual uint64
		for _, read := range descriptor.ProgramReads {
			flag, ok := flags[read]
			if !ok {
				return fmt.Errorf("unknown read %s", read)
			}
			actual |= flag
		}
		if actual != expected {
			return fmt.Errorf("%s declares program reads %d, cohere declares %d", descriptor.Name, actual, expected)
		}
	}
	return nil
}
func TestCheckerProgramReadsMatchCohere(t *testing.T) {
	t.Parallel()
	declarations := execute(t, "", goOracle(t), "--program-reads").output
	descriptors := prepareRegistry(t, ".")
	if err := checkerReadMismatch(descriptors, string(declarations)); err != nil {
		t.Fatal(err)
	}
}
func TestCheckerProgramReadsMutant(t *testing.T) {
	t.Parallel()
	declarations := execute(t, "", goOracle(t), "--program-reads").output
	descriptors := prepareRegistry(t, ".")
	caught := false
	for at, descriptor := range descriptors {
		if descriptor.Name != "no-new-wrappers" {
			continue
		}
		descriptors[at].ProgramReads = []string{"ReadsOtherFiles"}
		if err := checkerReadMismatch(descriptors, string(declarations)); err != nil {
			t.Logf("extra program read caught: %v", err)
			caught = true
		}
	}
	if !caught {
		t.Fatal("extra program read mutant survived")
	}
}
