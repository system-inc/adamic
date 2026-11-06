// Package privatefields builds immutable per-type plans for the latent IR copier.
package privatefields

import "reflect"

type Policy uint8

const (
	Nil Policy = iota + 1
	Empty
)

type planField struct {
	policy Policy
	name   string
}

// Plan is built once before publication and can be read concurrently.
// Its zero value copies every field.
type Plan struct{ fields []planField }

func New(owner reflect.Type, policies map[string]Policy) Plan {
	plan := Plan{fields: make([]planField, owner.NumField())}
	for name, policy := range policies {
		f, ok := owner.FieldByName(name)
		if !ok || len(f.Index) != 1 || f.PkgPath == "" || (policy != Nil && policy != Empty) {
			panic("latent state copy: invalid private IR field " + owner.Name() + "." + name)
		}
		plan.fields[f.Index[0]] = planField{policy, owner.Name() + "." + name}
	}
	return plan
}

func (plan Plan) Skip(value reflect.Value, index int) bool {
	if len(plan.fields) == 0 {
		return false
	}
	f := plan.fields[index]
	if f.policy == Nil && !value.Field(index).IsNil() {
		panic("latent state copy: populated IR cache " + f.name + " is unsupported")
	}
	return f.policy != 0
}
