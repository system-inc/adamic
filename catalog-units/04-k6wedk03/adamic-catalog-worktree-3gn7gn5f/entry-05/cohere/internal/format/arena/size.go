package arena

import "unsafe"

// sizeOf is the size of one element of values.
func sizeOf[T any](values []T) uintptr {
	return unsafe.Sizeof(values[0])
}
