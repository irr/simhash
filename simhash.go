// Copyright 2013 Matthew Fonda. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// simhash package implements Charikar's simhash algorithm to generate a 64-bit
// fingerprint of a given document.
//
// simhash fingerprints have the property that similar documents will have a similar
// fingerprint. Therefore, the hamming distance between two fingerprints will be small
// if the documents are similar
package simhash

import (
    "bytes"
    "regexp"
    "hash/fnv"
)

type Vector [64]int

// Feature consists of a 64-bit hash and a weight
type Feature interface {
    // Sum returns the 64-bit sum of this feature
    Sum() uint64

    // Weight returns the weight of this feature
    Weight() int
}

// FeatureSet represents a set of features in a given document
type FeatureSet interface {
    // GetFeatures returns a []Feature
    GetFeatures() []Feature
}

// Vectorize generates 64 dimension vectors given a set of features.
// Vectors are initialized to zero. The i-th element of the vector is then
// incremented by weight of the i-th feature if the i-th bit of the feature
// is set, and decremented by the weight of the i-th feature otherwise.
func Vectorize(features []Feature) Vector {
    var v Vector
    for _, feature := range features {
        sum := feature.Sum()
        for i := uint8(0); i < 64; i++ {
            bit := ((sum >> i) & 1);
            if bit == 1 {
                v[i] += feature.Weight()
            } else {
                v[i] -= feature.Weight()
            }
        }
    }
    return v
}

// Fingerprint returns a 64-bit fingerprint of the given vector.
// The fingerprint f of a given 64-dimension vector v is defined as follows:
//   f[i] = 1 if v[i] >= 0
//   f[i] = 0 if v[i] < 0
func Fingerprint(v Vector) uint64 {
    var f uint64
    for i := uint8(0); i < 64; i++ {
        if (v[i] >= 0) {
            f |= (1 << i)
        }
    }
    return f
}

type feature struct {
    sum uint64
    weight int
}

// Sum returns the 64-bit hash of this feature
func (f feature) Sum() uint64 {
    return f.sum
}

// Weight returns the weight of this feature
func (f feature) Weight() int {
    return f.weight
}

// Returns a new feature representing the given byte slice, using a weight of 1
func NewFeature(f []byte) feature {
    h := fnv.New64()
    h.Write(f)
    return feature{h.Sum64(), 1}
}

// Returns a new feature representing the given byte slice with the given weight
func NewFeatureWithWeight(f []byte, weight int) feature {
    fw := NewFeature(f)
    fw.weight = weight
    return fw
}

// Compare calculates the Hamming distance between two 64-bit integers
//
// Currently, this is calculated using the Kernighan method [1]. Other methods
// exist which may be more efficient and are worth exploring at some point
//
// [1] http://graphics.stanford.edu/~seander/bithacks.html#CountBitsSetKernighan
func Compare(a uint64, b uint64) uint8 {
    v := a ^ b
    var c uint8
    for c = 0; v != 0; c++ {
          v &= v - 1;
    }
    return c
}

// Returns a 64-bit simhash of the given bytes
func Simhash(fs FeatureSet) uint64 {
    return Fingerprint(Vectorize(fs.GetFeatures()))
}

// WordFeatureSet is a feature set in which each word is a feature,
// all equal weight.
type WordFeatureSet struct {
    b []byte
}

// Returns a []Feature representing each word in the byte slice
func (w WordFeatureSet) GetFeatures() []Feature {
    b := bytes.ToLower(w.b)
    words := regexp.MustCompile(`[\w']+(?:\://[\w\./]+){0,1}`).FindAll(b, -1)
    features := make([]Feature, len(words))
    for i, w := range words {
        features[i] = NewFeature(w)
    }
    return features
}