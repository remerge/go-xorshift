// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorshift

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"testing"
	"testing/iotest"
)

const (
	numTestSamples = 10000
)

type statsResults struct {
	mean        float64
	stddev      float64
	closeEnough float64
	maxError    float64
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func nearEqual(a, b, closeEnough, maxError float64) bool {
	absDiff := math.Abs(a - b)
	// Necessary when one value is zero and one value is close to zero.
	if absDiff < closeEnough {
		return true
	}
	return absDiff/max(math.Abs(a), math.Abs(b)) < maxError
}

var testSeeds = []int64{1, 1754801282, 1698661970, 1550503961}

// checkSimilarDistribution returns success if the mean and stddev of the
// two statsResults are similar.
func (sr *statsResults) checkSimilarDistribution(
	expected *statsResults,
) error {
	if !nearEqual(sr.mean, expected.mean,
		expected.closeEnough, expected.maxError) {
		s := fmt.Sprintf("mean %v != %v (allowed error %v, %v)", sr.mean,
			expected.mean, expected.closeEnough, expected.maxError)
		fmt.Println(s)
		return errors.New(s)
	}
	if !nearEqual(sr.stddev, expected.stddev, 0, expected.maxError) {
		s := fmt.Sprintf("stddev %v != %v (allowed error %v, %v)", sr.stddev,
			expected.stddev, expected.closeEnough, expected.maxError)
		fmt.Println(s)
		return errors.New(s)
	}
	return nil
}

func getStatsResults(samples []float64) *statsResults {
	res := new(statsResults)
	var sum, squaresum float64
	for _, s := range samples {
		sum += s
		squaresum += s * s
	}
	res.mean = sum / float64(len(samples))
	res.stddev = math.Sqrt(squaresum/float64(len(samples)) - res.mean*res.mean)
	return res
}

func checkSampleDistribution(
	t *testing.T,
	samples []float64,
	expected *statsResults,
) {
	actual := getStatsResults(samples)
	err := actual.checkSimilarDistribution(expected)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func checkSampleSliceDistributions(
	t *testing.T,
	samples []float64,
	nslices int,
	expected *statsResults,
) {
	chunk := len(samples) / nslices
	for i := 0; i < nslices; i++ {
		low := i * chunk
		var high int
		if i == nslices-1 {
			high = len(samples) - 1
		} else {
			high = (i + 1) * chunk
		}
		checkSampleDistribution(t, samples[low:high], expected)
	}
}

//
// Normal distribution tests
//

func generateNormalSamples(
	nsamples int,
	mean, stddev float64,
	seed int64,
) []float64 {
	r := New(NewSource(seed))
	samples := make([]float64, nsamples)
	for i := range samples {
		samples[i] = r.NormFloat64()*stddev + mean
	}
	return samples
}

func testNormalDistribution(
	t *testing.T,
	nsamples int,
	mean, stddev float64,
	seed int64,
) {
	samples := generateNormalSamples(nsamples, mean, stddev, seed)
	errorScale := max(1.0, stddev) // Error scales with stddev
	expected := &statsResults{mean, stddev,
		0.10 * errorScale, 0.08 * errorScale}

	// Make sure that the entire set matches the expected distribution.
	checkSampleDistribution(t, samples, expected)

	// Make sure that each half of the set matches the expected distribution.
	checkSampleSliceDistributions(t, samples, 2, expected)

	// Make sure that each 7th of the set matches the expected distribution.
	checkSampleSliceDistributions(t, samples, 7, expected)
}

// Actual tests

func TestStandardNormalValues(t *testing.T) {
	for _, seed := range testSeeds {
		testNormalDistribution(t, numTestSamples, 0, 1, seed)
	}
}

func TestNonStandardNormalValues(t *testing.T) {
	sdmax := 1000.0
	mmax := 1000.0
	if testing.Short() {
		sdmax = 5
		mmax = 5
	}
	for sd := 0.5; sd < sdmax; sd *= 2 {
		for m := 0.5; m < mmax; m *= 2 {
			for _, seed := range testSeeds {
				testNormalDistribution(t, numTestSamples, m, sd, seed)
				if testing.Short() {
					break
				}
			}
		}
	}
}

//
// Exponential distribution tests
//

func generateExponentialSamples(
	nsamples int,
	rate float64,
	seed int64,
) []float64 {
	r := New(NewSource(seed))
	samples := make([]float64, nsamples)
	for i := range samples {
		samples[i] = r.ExpFloat64() / rate
	}
	return samples
}

func testExponentialDistribution(
	t *testing.T,
	nsamples int,
	rate float64,
	seed int64,
) {
	mean := 1 / rate
	stddev := mean

	samples := generateExponentialSamples(nsamples, rate, seed)
	errorScale := max(1.0, 1/rate) // Error scales with the inverse of the rate
	expected := &statsResults{mean, stddev,
		0.10 * errorScale, 0.20 * errorScale}

	// Make sure that the entire set matches the expected distribution.
	checkSampleDistribution(t, samples, expected)

	// Make sure that each half of the set matches the expected distribution.
	checkSampleSliceDistributions(t, samples, 2, expected)

	// Make sure that each 7th of the set matches the expected distribution.
	checkSampleSliceDistributions(t, samples, 7, expected)
}

// Actual tests

func TestStandardExponentialValues(t *testing.T) {
	for _, seed := range testSeeds {
		testExponentialDistribution(t, numTestSamples, 1, seed)
	}
}

func TestNonStandardExponentialValues(t *testing.T) {
	for rate := 0.05; rate < 10; rate *= 2 {
		for _, seed := range testSeeds {
			testExponentialDistribution(t, numTestSamples, rate, seed)
			if testing.Short() {
				break
			}
		}
	}
}

func TestFloat32(t *testing.T) {
	// For issue 6721, the problem came after 7533753 calls, so check 10e6.
	num := int(10e6)
	// But do the full amount only on builders (not locally).
	// But ARM5 floating point emulation is slow (Issue 10749), so
	// do less for that builder:

	r := New(NewSource(1))
	for ct := 0; ct < num; ct++ {
		f := r.Float32()
		if f >= 1 {
			t.Fatal("Float32() should be in range [0,1). ct:", ct, "f:", f)
		}
	}
}

func TestReadUniformity(t *testing.T) {
	testBufferSizes := []int{
		2, 4, 7, 64, 1024, 1 << 16, 1 << 20,
	}
	for _, seed := range testSeeds {
		for _, n := range testBufferSizes {
			func(t *testing.T, n int, seed int64) {
				t.Run(fmt.Sprintf("%d %d", n, seed), func(t *testing.T) {
					r := New(NewSource(seed))
					buf := make([]byte, n)
					nRead, err := r.Read(buf)
					if err != nil {
						t.Errorf("Read err %v", err)
					}
					if nRead != n {
						t.Errorf("Read returned unexpected n; %d != %d", nRead, n)
					}

					// Expect a uniform distribution of byte values, which lie in [0, 255].
					var (
						mean       = 255.0 / 2
						stddev     = math.Sqrt(255.0 * 255.0 / 12.0)
						errorScale = stddev / math.Sqrt(float64(n))
					)

					expected := &statsResults{mean, stddev,
						0.10 * errorScale, 0.08 * errorScale}

					// Cast bytes as floats to use the common distribution-validity checks.
					samples := make([]float64, n)
					for i, val := range buf {
						samples[i] = float64(val)
					}
					// Make sure that the entire set matches the expected distribution.
					checkSampleDistribution(t, samples, expected)
				})
			}(t, n, seed)
		}
	}
}

func TestReadEmpty(t *testing.T) {
	r := New(NewSource(1))
	buf := make([]byte, 0)
	n, err := r.Read(buf)
	if err != nil {
		t.Errorf("Read err into empty buffer; %v", err)
	}
	if n != 0 {
		t.Errorf("Read into empty buffer returned unexpected n of %d", n)
	}
}

func TestReadByOneByte(t *testing.T) {
	r := New(NewSource(1))
	b1 := make([]byte, 100)
	_, err := io.ReadFull(iotest.OneByteReader(r), b1)
	if err != nil {
		t.Errorf("read by one byte: %v", err)
	}
	r = New(NewSource(1))
	b2 := make([]byte, 100)
	_, err = r.Read(b2)
	if err != nil {
		t.Errorf("read: %v", err)
	}
	if !bytes.Equal(b1, b2) {
		t.Errorf("read by one byte vs single read:\n%x\n%x", b1, b2)
	}
}

func TestReadSeedReset(t *testing.T) {
	r := New(NewSource(42))
	b1 := make([]byte, 128)
	_, err := r.Read(b1)
	if err != nil {
		t.Errorf("read: %v", err)
	}
	r.Seed(42)
	b2 := make([]byte, 128)
	_, err = r.Read(b2)
	if err != nil {
		t.Errorf("read: %v", err)
	}
	if !bytes.Equal(b1, b2) {
		t.Errorf("mismatch after re-seed:\n%x\n%x", b1, b2)
	}
}

// Benchmarks

func BenchmarkInt63Threadsafe(b *testing.B) {
	for n := b.N; n > 0; n-- {
		Int63()
	}
}

func BenchmarkInt63Unthreadsafe(b *testing.B) {
	r := New(NewSource(1))
	for n := b.N; n > 0; n-- {
		r.Int63()
	}
}

func BenchmarkIntn1000(b *testing.B) {
	r := New(NewSource(1))
	for n := b.N; n > 0; n-- {
		r.Intn(1000)
	}
}

func BenchmarkInt63n1000(b *testing.B) {
	r := New(NewSource(1))
	for n := b.N; n > 0; n-- {
		r.Int63n(1000)
	}
}

func BenchmarkInt31n1000(b *testing.B) {
	r := New(NewSource(1))
	for n := b.N; n > 0; n-- {
		r.Int31n(1000)
	}
}

func BenchmarkFloat32(b *testing.B) {
	r := New(NewSource(1))
	for n := b.N; n > 0; n-- {
		r.Float32()
	}
}

func BenchmarkFloat64(b *testing.B) {
	r := New(NewSource(1))
	for n := b.N; n > 0; n-- {
		r.Float64()
	}
}

func BenchmarkPerm3(b *testing.B) {
	r := New(NewSource(1))
	for n := b.N; n > 0; n-- {
		r.Perm(3)
	}
}

func BenchmarkPerm30(b *testing.B) {
	r := New(NewSource(1))
	for n := b.N; n > 0; n-- {
		r.Perm(30)
	}
}

func BenchmarkRead3(b *testing.B) {
	r := New(NewSource(1))
	buf := make([]byte, 3)
	b.ResetTimer()
	for n := b.N; n > 0; n-- {
		r.Read(buf)
	}
}

func BenchmarkRead64(b *testing.B) {
	r := New(NewSource(1))
	buf := make([]byte, 64)
	b.ResetTimer()
	for n := b.N; n > 0; n-- {
		r.Read(buf)
	}
}

func BenchmarkRead1000(b *testing.B) {
	r := New(NewSource(1))
	buf := make([]byte, 1000)
	b.ResetTimer()
	for n := b.N; n > 0; n-- {
		r.Read(buf)
	}
}
