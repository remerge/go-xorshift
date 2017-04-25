// Copyright 2016 Google Inc.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package uuid

import (
	"fmt"
	"testing"
)

type test struct {
	in      string
	version Version
	variant Variant
	isuuid  bool
}

var constants = []struct {
	c    interface{}
	name string
}{
	{Invalid, "Invalid"},
	{RFC4122, "RFC4122"},
	{Reserved, "Reserved"},
	{Microsoft, "Microsoft"},
	{Future, "Future"},
	{Variant(42), "BadVariant42"},
}

func TestConstants(t *testing.T) {
	for x, tt := range constants {
		v, ok := tt.c.(fmt.Stringer)
		if !ok {
			t.Errorf("%x: %v: not a stringer", x, v)
		} else if s := v.String(); s != tt.name {
			v, _ := tt.c.(int)
			t.Errorf("%x: Constant %T:%d gives %q, expected %q", x, tt.c, v, s, tt.name)
		}
	}
}

func TestRandomUUID(t *testing.T) {
	m := make(map[string]bool)
	for x := 1; x < 32; x++ {
		uuid := New()
		s := uuid.String()
		if m[s] {
			t.Errorf("NewRandom returned duplicated UUID %s", s)
		}
		m[s] = true
		if v := uuid.Version(); v != 4 {
			t.Errorf("Random UUID of version %s", v)
		}
		if uuid.Variant() != RFC4122 {
			t.Errorf("Random UUID is variant %d", uuid.Variant())
		}
	}
}

type badRand struct{}

func (r badRand) Read(buf []byte) (int, error) {
	for i := range buf {
		buf[i] = byte(i)
	}
	return len(buf), nil
}

var asString = "f47ac10b-58cc-0372-8567-0e02b2c3d479"
var asBytes = []byte(asString)

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New()
	}
}
