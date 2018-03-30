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
		{"1 x 10", 1, 10},
		{"1 x 100", 1, 100},
		{"1 x 1000", 1, 1000},
		{"1 x 10000", 1, 10000},
		{"1 x 100000", 1, 100000},
		{"10 x 10", 10, 10},
		{"10 x 100", 10, 100},
		{"10 x 1000", 10, 1000},
		{"10 x 10000", 10, 10000},
		{"10 x 100000", 10, 100000},
		{"100 x 10", 100, 10},
		{"100 x 100", 100, 100},
		{"100 x 1000", 100, 1000},
	}
	for _, bc := range bcs {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				benchmarkContainerRead(bc.containerCount, bc.itemCount, b)
			}
		})
	}
}

func benchmarkContainerRead(containerCount, itemCount int, b *testing.B) {
	var wg sync.WaitGroup
	now := time.Now()
	for i := 0; i <= containerCount; i++ {
		wg.Add(1)
		go func(now time.Time, itemCount int) {
			container := Container{}
			for j := 0; j < itemCount; j++ {
				container.Add(
					floatTypePool.Get().(DataType),
					stringTypePool.Get().(DataType),
					byteTypePool.Get().(DataType),
					kiloByteTypePool.Get().(DataType),
					megaByteTypePool.Get().(DataType),
				)
			}
			container.Generate(ioutil.Discard, now)
			for _, item := range container.List() {
				switch item.(type) {
				case *FloatType:
					floatTypePool.Put(item)
				case *StringType:
					stringTypePool.Put(item)
				case *ByteType:
					byteTypePool.Put(item)
				case *KiloByteType:
					kiloByteTypePool.Put(item)
				case *MegaByteType:
					megaByteTypePool.Put(item)
				}
			}
			wg.Done()
		}(now, itemCount)
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
	now := time.Now()
	for i := 0; i <= containerCount; i++ {
		wg.Add(1)
		go func() {
			p := new(bytes.Buffer)
			container := Container{}
			for j := 0; j < itemCount; j++ {
				f := floatTypePool.Get().(DataType)
				f.Reset()
				s := stringTypePool.Get().(DataType)
				s.Reset()
				b := byteTypePool.Get().(DataType)
				b.Reset()
				k := kiloByteTypePool.Get().(DataType)
				k.Reset()
				m := megaByteTypePool.Get().(DataType)
				m.Reset()
				container.Add(f, s, b, k, m)
			}
			container.Generate(p, now)
			res, _ := JobResultDataTypes(p.Bytes()[:], mapper)
			res.Generate(ioutil.Discard, now)
			for _, item := range container.List() {
				switch item.(type) {
				case *FloatType:
					floatTypePool.Put(item)
				case *StringType:
					stringTypePool.Put(item)
				case *ByteType:
					byteTypePool.Put(item)
				case *KiloByteType:
					kiloByteTypePool.Put(item)
				case *MegaByteType:
					megaByteTypePool.Put(item)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

var floatTypePool = sync.Pool{
	New: func() interface{} {
		return NewFloatType(randomString(10), rand.Float64()*1000)
	},
}

var stringTypePool = sync.Pool{
	New: func() interface{} {
		return NewStringType(randomString(10), randomString(20))
	},
}

var byteTypePool = sync.Pool{
	New: func() interface{} {
		return NewByteType(randomString(10), rand.Float64()*1000)
	},
}

var kiloByteTypePool = sync.Pool{
	New: func() interface{} {
		return NewKiloByteType(randomString(10), rand.Float64()*1000)
	},
}

var megaByteTypePool = sync.Pool{
	New: func() interface{} {
		return NewMegaByteType(randomString(10), rand.Float64()*1000)
	},
}

func randomString(count int) string {
	result := make([]rune, count)
	for i := range result {
		result[i] = runes[rand.Intn(len(runes))]
	}
	return string(result)
}
