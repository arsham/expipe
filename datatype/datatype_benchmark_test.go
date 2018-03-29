// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"sync"
	"testing"
	"time"
)

var runes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func BenchmarkContainerRead(b *testing.B) {
	bcs := []struct {
		name           string
		containerCount int
		itemCount      int
	}{
		{"10 x 10", 10, 10},
		{"10 x 100", 10, 100},
		{"10 x 1000", 10, 1000},
		{"10 x 10000", 10, 10000},
		{"10 x 100000", 10, 100000},
		{"100 x 10", 100, 10},
		{"1000 x 10", 1000, 10},
		{"100 x 100", 100, 100},
		{"1000 x 100", 1000, 100},
	}
	for _, bc := range bcs {
		b.Run(bc.name, func(b *testing.B) {
			benchmarkContainerRead(bc.containerCount, bc.itemCount, b)
		})
	}
}

func benchmarkContainerRead(containerCount, itemCount int, b *testing.B) {
	b.StopTimer()
	rFT := randomFloatType(itemCount)
	rST := randomStringType(itemCount)
	rBT := randomByteType(itemCount)
	rKBT := randomKiloByteType(itemCount)
	rMBT := randomMegaByteType(itemCount)
	b.StartTimer()
	var wg sync.WaitGroup
	for i := 0; i <= containerCount; i++ {
		wg.Add(1)
		go func() {
			b.StopTimer()
			container := Container{}
			for j := 0; j < itemCount; j++ {
				container.Add(rFT[j], rST[j], rBT[j], rKBT[j], rMBT[j])
			}
			b.StartTimer()
			container.Generate(ioutil.Discard, time.Now())
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkJobResultDataTypes(b *testing.B) {
	bcs := []struct {
		name           string
		containerCount int
		itemCount      int
	}{
		{"10 x 10", 10, 10},
		{"10 x 100", 10, 100},
		{"10 x 1000", 10, 1000},
		{"10 x 10000", 10, 10000},
		{"100 x 10", 100, 10},
		{"1000 x 10", 1000, 10},
		{"100 x 100", 100, 100},
		{"1000 x 100", 1000, 100},
		{"10000 x 30", 10000, 30},
	}
	mapper := DefaultMapper()
	for _, bc := range bcs {
		b.Run(bc.name, func(b *testing.B) {
			benchmarkJobResultDataTypes(mapper, bc.containerCount, bc.itemCount, b)
		})
	}
}

func benchmarkJobResultDataTypes(mapper Mapper, containerCount, itemCount int, b *testing.B) {
	var wg sync.WaitGroup
	b.StopTimer()
	rFT := randomFloatType(itemCount)
	rST := randomStringType(itemCount)
	rBT := randomByteType(itemCount)
	rKBT := randomKiloByteType(itemCount)
	rMBT := randomMegaByteType(itemCount)
	b.StartTimer()
	for i := 0; i <= containerCount; i++ {
		wg.Add(1)
		go func() {
			b.StopTimer()
			p := new(bytes.Buffer)
			container := Container{}
			for j := 0; j < itemCount; j++ {
				container.Add(rFT[j], rST[j], rBT[j], rKBT[j], rMBT[j])
			}
			container.Generate(p, time.Now())
			b.StartTimer()
			res, _ := JobResultDataTypes(p.Bytes(), mapper)
			res.Generate(ioutil.Discard, time.Now())
			wg.Done()
		}()
	}
	wg.Wait()
}

func randomFloatType(size int) []FloatType {
	r := make([]FloatType, size)
	for i := 0; i < size; i++ {
		r[i] = FloatType{Key: randomString(10), Value: rand.Float64() * 1000}
	}
	return r
}

func randomStringType(size int) []StringType {
	r := make([]StringType, size)
	for i := 0; i < size; i++ {
		r[i] = StringType{Key: randomString(10), Value: randomString(20)}
	}
	return r
}

func randomByteType(size int) []ByteType {
	r := make([]ByteType, size)
	for i := 0; i < size; i++ {
		r[i] = ByteType{Key: randomString(10), Value: rand.Float64() * 1000}
	}
	return r
}

func randomKiloByteType(size int) []KiloByteType {
	r := make([]KiloByteType, size)
	for i := 0; i < size; i++ {
		r[i] = KiloByteType{Key: randomString(10), Value: rand.Float64() * 1000}
	}
	return r
}
func randomMegaByteType(size int) []MegaByteType {
	r := make([]MegaByteType, size)
	for i := 0; i < size; i++ {
		r[i] = MegaByteType{Key: randomString(10), Value: rand.Float64() * 1000}
	}
	return r
}

func randomString(count int) string {
	result := make([]rune, count)
	for i := range result {
		result[i] = runes[rand.Intn(len(runes))]
	}
	return string(result)
}
