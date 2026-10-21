package worker

import (
	"fmt"
	"os"
	"testing"

	"github.com/system-inc/adamic/internal/oracletest"
)

func TestMain(main *testing.M) {
	// Mutant children inherit the parent's live descriptor tool; rebuilding the
	// same helper in every child is setup work inside the owning shard's clock.
	if os.Getenv("ADAMIC_WORKER_MUTANT_CHILD") == "1" {
		info, err := os.Stat(os.Getenv("ADAMIC_JSON_TYPES"))
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
			fmt.Fprintln(os.Stderr, "Worker mutant child requires the parent's JSON descriptor tool")
			os.Exit(2)
		}
		os.Exit(main.Run())
	}

	os.Exit(oracletest.Run(main.Run))
}
