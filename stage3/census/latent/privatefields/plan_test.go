package privatefields

import (
	"reflect"
	"sync"
	"testing"
)

type planFixture20pb346 struct {
	Exported int
	cache    *int
	rebuild  map[string]int
	other    string
}

func TestLatentPrivateIRPlanPreservesUnlistedFields20pb346(t *testing.T) {
	t.Parallel()
	owner := reflect.TypeFor[planFixture20pb346]()
	policies := map[string]Policy{"cache": Nil, "rebuild": Empty}
	plan := New(owner, policies)
	// The caller's table is not retained by the published plan.
	delete(policies, "cache")
	value := reflect.ValueOf(planFixture20pb346{Exported: 7, rebuild: map[string]int{"x": 1}, other: "keep"})
	var readers sync.WaitGroup
	for range 8 {
		readers.Go(func() {
			for i := range owner.NumField() {
				want := i == 1 || i == 2
				if got := plan.Skip(value, i); got != want {
					t.Errorf("field %s: skip=%v, want %v", owner.Field(i).Name, got, want)
				}
				if (Plan{}).Skip(value, i) {
					t.Errorf("zero plan skipped field %d", i)
				}
			}
		})
	}
	readers.Wait()
}

func TestLatentPrivateIRPlanRejectsPopulatedNilCache20pb346(t *testing.T) {
	t.Parallel()
	plan := New(reflect.TypeFor[planFixture20pb346](), map[string]Policy{"cache": Nil})
	number := 7
	defer func() {
		want := "latent state copy: populated IR cache planFixture20pb346.cache is unsupported"
		if got := recover(); got != want {
			t.Fatalf("panic=%v, want %q", got, want)
		}
	}()
	plan.Skip(reflect.ValueOf(planFixture20pb346{cache: &number}), 1)
}
