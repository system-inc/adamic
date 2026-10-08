package lower

import (
	"errors"
	"strings"
	"testing"
)

func requireCycleRefusal(t *testing.T, err error) {
	t.Helper()
	var refusal *Refused
	if !errors.As(err, &refusal) || !strings.Contains(refusal.Fix, "adamic/cycle-capable") {
		t.Fatalf("want the ruled cycle-capable refusal, got %v", err)
	}
}

func TestLiteralMethodCycleIsRefused(t *testing.T) {
	_, err := lowerSource(t, `function make():void {let holder:{read():number}|undefined;const value={read():number{return holder===undefined?0:1;}};holder=value;console.log('made');}make();`)
	requireCycleRefusal(t, err)
}
