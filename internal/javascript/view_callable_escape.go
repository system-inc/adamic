package javascript

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

const viewCallableAdapterRuntime = `
const adamicViewAdapterRoots = new WeakMap();
const adamicViewAdapterKeys = new Map();
const adamicViewAdapterCache = new WeakMap();
const adamicViewAdapterUnderlying = value => adamicViewAdapterRoots.get(value) ?? value;
const adamicViewAdapterKey = id => {
    if (!adamicViewAdapterKeys.has(id)) adamicViewAdapterKeys.set(id, {});
    return adamicViewAdapterKeys.get(id);
};
const adamicViewAdapterIntern = (value, key, make) => {
    const underlying = adamicViewAdapterUnderlying(value);
    if (underlying === undefined || underlying === null) return underlying;
    let views = adamicViewAdapterCache.get(underlying);
    if (views === undefined) { views = new WeakMap(); adamicViewAdapterCache.set(underlying, views); }
    const held = views.get(key)?.deref();
    if (held !== undefined) return held;
    const adapter = make(underlying);
    adamicViewAdapterRoots.set(adapter, underlying);
    views.set(key, new WeakRef(adapter));
    return adapter;
};
` + viewCallableIdentityCollections

func (e *emitter) emitViewCallableEscape(property ir.Property) string {
	target := e.program.ViewContracts[property.ViewEscapeContract-1]
	call := ir.CallClosure{CallContract: property.ViewEscapeContract, CallWhere: "escaping call (read at " + property.ViewWhere + ")", Returns: e.program.ViewContracts[target.Result-1].Of}
	if e.program.ViewContracts[target.Result-1].Name == "void" {
		call.Returns = 0
	}
	invoke := e.viewCallableInvoke(call, property)
	raw := fmt.Sprintf("adamicViewCallablePreparedRead(%s, %s, %s, %t, %t, %t, %s)", e.value(property.Object), quote(property.Name), quote(property.View), property.Method, property.Optional, property.Absent, quote(property.ViewType))
	return fmt.Sprintf("adamicViewAdapterIntern(%s, adamicViewAdapterKey(%d), underlying => new AdamicClosure((self, arguments_) => %s({value: underlying, object: undefined}, arguments_), []))", raw, property.ViewTypeID, invoke)
}
