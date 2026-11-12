package lower

import (
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestDateISOStringOtherMethodStopsByName(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, `console.log(new Date(0).toUTCString());`, ".ts")
	var missing *NotYet
	if !errors.As(err, &missing) || missing.What != "toUTCString: Date methods other than getTime/valueOf in the fs file host" {
		t.Fatalf("Date method boundary must retain named NotYet: %v", err)
	}
}

func TestDateISOStringTscShapeLowers(t *testing.T) {
	t.Parallel()
	if _, err := lowerSourceAs(t, `console.log((new Date()).toISOString().replace(/:/g, "-"));`, ".ts"); err != nil {
		t.Fatal(err)
	}
}

func dateISOStringMethodMutant(t *testing.T) string {
	t.Helper()
	return buildcache.GoTestMutated(t, "", "date-iso-method-lower.test", "./internal/lower", nil, []buildcache.Mutation{{
		File:   "internal/lower/library_node_fs_file.go",
		Before: `if symbol.Name != "getTime" && symbol.Name != "valueOf" && symbol.Name != "toISOString" {`,
		After:  `if false {`,
	}})
}
func TestProduct_DateISOStringMethodMutant(t *testing.T) { t.Parallel(); dateISOStringMethodMutant(t) }
func TestDateISOStringMethodMutantIsCaught(t *testing.T) {
	t.Parallel()
	binary := dateISOStringMethodMutant(t)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Minute)
	defer cancel()
	stdout, stderr, err := childguard.Capture(ctx, exec.CommandContext(ctx, binary, "-test.run", "^TestDateISOStringOtherMethodStopsByName$", "-test.count=1", "-test.v"))
	var exited *exec.ExitError
	if !errors.As(err, &exited) || exited.ExitCode() != 1 || !strings.Contains(string(stdout), "Date method boundary must retain named NotYet") {
		t.Fatalf("Date arm mutant must fail boundary pin: err=%v stdout=%s stderr=%s", err, stdout, stderr)
	}
	t.Logf("Date arm mutant caught: %s", stdout)
}
