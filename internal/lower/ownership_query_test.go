package lower

import (
	"context"
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Compare the complete existing move/concurrency corpus with both admission
// paths. The query remains behind the original move/callback/runtime checks.
func TestOwnershipQueryMoveAgreement(t *testing.T) {
	count, queried := 0, 0
	acceptedMoveRan := false
	refusedMoveRan := false
	queryRefusals := 0
	for _, area := range []string{"concurrency", "moves"} {
		for _, group := range []string{"accepted", "refused"} {
			paths, err := filepath.Glob("../oracle/testdata/" + area + "/" + group + "/*.a")
			if err != nil || len(paths) == 0 {
				t.Fatalf("fixtures %s/%s: %v", area, group, err)
			}
			for _, path := range paths {
				t.Run(area+"/"+group+"/"+filepath.Base(path), func(t *testing.T) {
					count++
					if area == "moves" && group == "accepted" {
						acceptedMoveRan = true
					}
					if area == "moves" && group == "refused" {
						refusedMoveRan = true
					}
					checked, err := load.Load([]string{path})
					if err != nil {
						t.Fatal(err)
					}
					old, oldErr := LowerWithOptions(context.Background(), checked, Options{})
					queryChecked, loadErr := load.Load([]string{path})
					if loadErr != nil {
						t.Fatal(loadErr)
					}
					_, newErr := LowerWithOptions(context.Background(), queryChecked, Options{OwnershipQuery: true, ownershipAgreement: true})
					if (oldErr == nil) != (newErr == nil) {
						t.Fatalf("disagreement: flat=%v query=%v", oldErr, newErr)
					}
					if newErr != nil && strings.Contains(newErr.Error(), "ownership query") {
						queryRefusals++
						t.Logf("independent query: %v", newErr)
					}
					if oldErr == nil {
						for _, answer := range fresh.OwnershipTransfers(old) {
							queried++
							if answer.Answer.Verdict != fresh.Proven {
								t.Fatalf("flat admitted; independent query: %+v", answer)
							}
						}
					}
				})
			}
		}
	}
	t.Logf("agreement: %d fixtures; %d admitted IR transfers; %d independent query refusals", count, queried, queryRefusals)
	if acceptedMoveRan && refusedMoveRan && queryRefusals == 0 {
		t.Fatal("refused fixtures never exercised the independent query")
	}
	if acceptedMoveRan && queried == 0 {
		t.Fatal("query was never exercised")
	}
}
func BenchmarkOwnershipQueryLowering(b *testing.B) {
	path := os.Getenv("ADAMIC_OWNERSHIP_BENCH_SOURCE")
	if path == "" {
		path = "../../stage1/typescript/parser/main.ts"
	}
	for _, enabled := range []bool{false, true} {
		name := "off"
		if enabled {
			name = "on"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				checked, err := load.Load([]string{path})
				if err != nil {
					b.Fatal(err)
				}
				checked.EnableTSGo()
				b.StartTimer()
				program, err := LowerWithOptions(context.Background(), checked, Options{OwnershipQuery: enabled})
				if err != nil {
					b.Fatal(err)
				}
				if i == 0 {
					b.Logf("functions=%d locals=%d", len(program.Functions), len(program.Locals))
				}
			}
		})
	}
}
