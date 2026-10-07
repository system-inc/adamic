package fresh

import (
	"iter"
	"slices"
)

// facts is a persistent search tree. A state snapshot shares its root; changing one
// fact copies only its search path. Values stored in it must also be immutable.
// Priorities depend only on keys, so independent states keep compatible trees.
type facts[K ~int | ~int32, V any] struct{ root *fact[K, V] }
type fact[K ~int | ~int32, V any] struct {
	key         K
	value       V
	left, right *fact[K, V]
	count       int
}

func priority[K ~int | ~int32](key K) uint64 {
	x := uint64(int64(key)) + 0x9e3779b97f4a7c15
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	return x ^ (x >> 31)
}
func count[K ~int | ~int32, V any](n *fact[K, V]) int {
	if n == nil {
		return 0
	}
	return n.count
}
func branch[K ~int | ~int32, V any](key K, value V, left, right *fact[K, V]) *fact[K, V] {
	return &fact[K, V]{key: key, value: value, left: left, right: right, count: 1 + count(left) + count(right)}
}
func (s facts[K, V]) node(key K) *fact[K, V] {
	for n := s.root; n != nil; {
		switch {
		case key < n.key:
			n = n.left
		case key > n.key:
			n = n.right
		default:
			return n
		}
	}
	return nil
}
func (s facts[K, V]) lookup(key K) (V, bool) {
	if n := s.node(key); n != nil {
		return n.value, true
	}
	var zero V
	return zero, false
}
func (s facts[K, V]) get(key K) V         { v, _ := s.lookup(key); return v }
func (s facts[K, V]) size() int           { return count(s.root) }
func (s *facts[K, V]) set(key K, value V) { s.root = insert(s.root, key, value) }
func insert[K ~int | ~int32, V any](n *fact[K, V], key K, value V) *fact[K, V] {
	if n == nil {
		return branch(key, value, nil, nil)
	}
	if key == n.key {
		return branch(key, value, n.left, n.right)
	}
	if key < n.key {
		left := insert(n.left, key, value)
		if priority(left.key) < priority(n.key) {
			return branch(left.key, left.value, left.left, branch(n.key, n.value, left.right, n.right))
		}
		return branch(n.key, n.value, left, n.right)
	}
	right := insert(n.right, key, value)
	if priority(right.key) < priority(n.key) {
		return branch(right.key, right.value, branch(n.key, n.value, n.left, right.left), right.right)
	}
	return branch(n.key, n.value, n.left, right)
}
func (s *facts[K, V]) remove(key K) { s.root = remove(s.root, key) }
func remove[K ~int | ~int32, V any](n *fact[K, V], key K) *fact[K, V] {
	if n == nil {
		return nil
	}
	if key < n.key {
		left := remove(n.left, key)
		if left == n.left {
			return n
		}
		return branch(n.key, n.value, left, n.right)
	}
	if key > n.key {
		right := remove(n.right, key)
		if right == n.right {
			return n
		}
		return branch(n.key, n.value, n.left, right)
	}
	return combine(n.left, n.right)
}
func combine[K ~int | ~int32, V any](left, right *fact[K, V]) *fact[K, V] {
	if left == nil {
		return right
	}
	if right == nil {
		return left
	}
	if priority(left.key) < priority(right.key) {
		return branch(left.key, left.value, left.left, combine(left.right, right))
	}
	return branch(right.key, right.value, combine(left, right.left), right.right)
}
func (s facts[K, V]) all() iter.Seq2[K, V] { return s.changed(facts[K, V]{}) }

// changed omits only subtrees shared by identity. Every fact that might have
// changed is still joined by the analysis's original union operation.
func (s facts[K, V]) changed(previous facts[K, V]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		var walk func(*fact[K, V]) bool
		walk = func(n *fact[K, V]) bool {
			if n == nil || previous.node(n.key) == n {
				return true
			}
			return walk(n.left) && yield(n.key, n.value) && walk(n.right)
		}
		walk(s.root)
	}
}

type objectFacts struct{ facts[object, bool] }

func (s *objectFacts) add(o object) bool {
	if s.get(o) {
		return false
	}
	s.set(o, true)
	return true
}
func (s objectFacts) sorted() []object {
	list := make([]object, 0, s.size())
	for o := range s.all() {
		list = append(list, o)
	}
	slices.Sort(list)
	return list
}
