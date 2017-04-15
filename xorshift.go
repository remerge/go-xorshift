package xorshift

import "math/rand"

type xorshiftSource struct {
	state [2]uint64
}

func NewSource(seed int64) rand.Source64 {
	var rng xorshiftSource
	rng.Seed(seed)
	return &rng
}

func New(source rand.Source) *rand.Rand {
	return rand.New(source)
}

func (src *xorshiftSource) Seed(seed int64) {
	rand.Seed(seed)
	for i := 0; i < len(src.state); i++ {
		src.state[i] = rand.Uint64()
	}
}

func (src *xorshiftSource) Int63() int64 {
	return int64(src.Uint64()) & ((1 << 63) - 1)
}

// this is xorshift+ https://en.wikipedia.org/wiki/Xorshift
func (src *xorshiftSource) Uint64() uint64 {
	s1 := src.state[0]
	s0 := src.state[1]
	src.state[0] = s0
	s1 ^= s1 << 23
	src.state[1] = (s1 ^ s0 ^ (s1 >> 17) ^ (s0 >> 26))
	return src.state[1] + s0
}
