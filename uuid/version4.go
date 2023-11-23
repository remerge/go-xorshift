// Copyright 2016 Google Inc.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package uuid

import (
	rand "github.com/remerge/go-xorshift"
)

// New is creates a new random UUID or panics.  New is equivalent to
// the expression
//
//	uuid.Must(uuid.NewRandom())
func New() UUID {
	return Must(NewRandom())
}

// NewRandom returns a Random (Version 4) UUID.
func NewRandom() (UUID, error) {
	var uuid UUID
	_, err := rand.Read(uuid[:])
	if err != nil {
		return Nil, err
	}
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10
	return uuid, nil
}
