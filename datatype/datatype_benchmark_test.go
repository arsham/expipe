// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype_test

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/arsham/expipe/datatype"
)

var runes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var (
	floatTypeVar    = newFloatType()
	stringTypeVar   = newStringType()
	byteTypeVar     = newByteType()
	kiloByteTypeVar = newKiloByteType()
	megaByteTypeVar = newMegaByteType()
)

func newFloatType() *datatype.FloatType {
	return datatype.NewFloatType(randomString(10), rand.Float64()*1000)
}
func newStringType() *datatype.StringType {
	return datatype.NewStringType(randomString(10), randomString(20))
}
func newByteType() *datatype.ByteType {
	return datatype.NewByteType(randomString(10), rand.Float64()*1000)
}
func newKiloByteType() *datatype.KiloByteType {
	return datatype.NewKiloByteType(randomString(10), rand.Float64()*1000)
}
func newMegaByteType() *datatype.MegaByteType {
	return datatype.NewMegaByteType(randomString(10), rand.Float64()*1000)
}

func BenchmarkContainerRead(b *testing.B) {
	bcs := []struct {
		name           string
		containerCount int
		itemCount      int
	}{
		{"1 x 10", 1, 10},
		{"1 x 100", 1, 100},
		{"1 x 1000", 1, 1000},
		{"10 x 10", 10, 10},
		{"10 x 100", 10, 100},
		{"10 x 1000", 10, 1000},
		{"100 x 10", 100, 10},
		{"100 x 100", 100, 100},
		{"100 x 1000", 100, 1000},
	}
	for _, bc := range bcs {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				benchmarkContainerRead(b, bc.containerCount, bc.itemCount)
			}
		})
	}
}
func benchmarkContainerRead(b *testing.B, containerCount, itemCount int) {
	var wg sync.WaitGroup
	now := time.Now()
	for i := 0; i <= containerCount; i++ {
		wg.Add(1)
		go func(now time.Time, itemCount int) {
			container := datatype.Container{}
			for j := 0; j < itemCount/5; j++ {
				container.Add(
					floatTypeVar,
					stringTypeVar,
					byteTypeVar,
					kiloByteTypeVar,
					megaByteTypeVar,
				)
			}
			container.Generate(ioutil.Discard, now)
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
	mapper := datatype.DefaultMapper()
	for _, bc := range bcs {
		b.Run(bc.name, func(b *testing.B) {
			benchmarkJobResultDataTypes(b, mapper, bc.containerCount, bc.itemCount)
		})
	}
}

func benchmarkJobResultDataTypes(b *testing.B, mapper datatype.Mapper, containerCount, itemCount int) {
	var wg sync.WaitGroup
	now := time.Now()
	for i := 0; i <= containerCount; i++ {
		wg.Add(1)
		go func() {
			b.StopTimer()
			p := new(bytes.Buffer)
			container := datatype.Container{}
			for j := 0; j < itemCount; j++ {
				container.Add(
					newFloatType(),
					newStringType(),
					newByteType(),
					newKiloByteType(),
					newMegaByteType(),
				)
			}
			container.Generate(p, now)
			b.StartTimer()
			res, _ := datatype.JobResultDataTypes(p.Bytes()[:], mapper)
			res.Generate(ioutil.Discard, now)
			wg.Done()
		}()
	}
	wg.Wait()
}

func randomString(count int) string {
	result := make([]rune, count)
	for i := range result {
		result[i] = runes[rand.Intn(len(runes))]
	}
	return string(result)
}

func BenchmarkContainer(b *testing.B) {
	now := time.Now()
	b.Run("Add", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			container := datatype.Container{}
			container.Add(
				floatTypeVar,
				stringTypeVar,
				byteTypeVar,
				kiloByteTypeVar,
				megaByteTypeVar,
			)
		}
	})

	b.Run("Add More", func(b *testing.B) {
		container := datatype.Container{}
		for i := 0; i < b.N; i++ {
			container.Add(
				floatTypeVar,
				stringTypeVar,
				byteTypeVar,
				kiloByteTypeVar,
				megaByteTypeVar,
			)
		}
	})

	b.Run("Generate", func(b *testing.B) {
		container := datatype.Container{}
		container.Add(
			floatTypeVar,
			stringTypeVar,
			byteTypeVar,
			kiloByteTypeVar,
			megaByteTypeVar,
		)
		for i := 0; i < b.N; i++ {
			container.Generate(ioutil.Discard, now)
		}
	})

	b.Run("Add and Generate", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			container := datatype.Container{}
			container.Add(
				floatTypeVar,
				stringTypeVar,
				byteTypeVar,
				kiloByteTypeVar,
				megaByteTypeVar,
			)
			container.Generate(ioutil.Discard, now)
		}
	})

	b.Run("Add More and Generate", func(b *testing.B) {
		container := datatype.Container{}
		for i := 0; i < b.N; i++ {
			container.Add(
				floatTypeVar,
				stringTypeVar,
				byteTypeVar,
				kiloByteTypeVar,
				megaByteTypeVar,
			)
			container.Generate(ioutil.Discard, now)
		}
	})
}

func BenchmarkStringType(b *testing.B) {
	b.Run("New", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			datatype.NewStringType("wyuNdoEDGYokey", "cAtgiuaBnmZuvalue")
		}
	})

	b.Run("New and Read by buffer", func(b *testing.B) {
		buf := new(bytes.Buffer)
		for i := 0; i < b.N; i++ {
			s := datatype.NewStringType("wyuNdoEDGYokey", "cAtgiuaBnmZuvalue")
			buf.ReadFrom(s)
		}
	})

	b.Run("New and Read by bytes", func(b *testing.B) {
		buf := new(bytes.Buffer)
		s := datatype.NewStringType("wyuNdoEDGYokey", "cAtgiuaBnmZuvalue")
		buf.ReadFrom(s)
		r := make([]byte, len(buf.Bytes())+1)
		for i := 0; i < b.N; i++ {
			s := datatype.NewStringType("wyuNdoEDGYokey", "cAtgiuaBnmZuvalue")
			s.Read(r)
		}
	})

	b.Run("Read by buffer", func(b *testing.B) {
		buf := new(bytes.Buffer)
		s := datatype.NewStringType("wyuNdoEDGYokey", "cAtgiuaBnmZuvalue")
		for i := 0; i < b.N; i++ {
			buf.ReadFrom(s)
		}
	})

	b.Run("Read by bytes", func(b *testing.B) {
		buf := new(bytes.Buffer)
		s := datatype.NewStringType("wyuNdoEDGYokey", "cAtgiuaBnmZuvalue")
		buf.ReadFrom(s)
		r := make([]byte, len(buf.Bytes())+1)
		for i := 0; i < b.N; i++ {
			s.Read(r)
		}
	})
}
