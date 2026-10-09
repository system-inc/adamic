package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestStep18DetachedCallableOrigins(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`const host:{label:string;callback:()=>string}={label:'owner',callback:function(this:{readonly label:string}):string{return this.label;}};const callback=host.callback;callback?.();`,
		`class Host{readonly label='owner';get callback():()=>string{return function(this:{readonly label:string}):string{return this.label;};}}const host=new Host();const callback=host.callback;callback?.();`,
		`const host={label:'owner',report():string{return this.label;}};const view:{readonly report:()=>string}=host;const callback=view.report;callback?.();`,
	} {
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(refused.What, "unbound-method") {
			t.Fatalf("got %v, want receiver-dependent copy refusal", err)
		}
	}
}

func TestStep18UnknownDetachedCallable(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`function use(host:{readonly callback:()=>string}):void{const callback=host.callback;callback?.();}`,
		`function use(host:{readonly callback:()=>string}):void{const callback=(host.callback);const copied=callback;copied?.();}`,
	} {
		_, err := lowerSource(t, source)
		var gap *NotYet
		if !errors.As(err, &gap) || gap.What != "a detached callable whose receiver independence is not proven" {
			t.Fatalf("got %v, want unknown origin NotYet", err)
		}
	}
}
