//go:build cohere_poison

package arena

// poisonReleased: under cohere_poison, Reset overwrites what it releases with the arena's Poison.
const poisonReleased = true
