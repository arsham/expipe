// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package datatype

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"
	"time"
)

var runes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func BenchmarkContainerRead(b *testing.B) {
	bcs := []struct {
		containerCount int
		itemCount      int
	}{
		{10, 10},
		{100, 10},
		{1000, 10},
		{10, 100},
		{100, 100},
		{1000, 100},
	}
	for _, bc := range bcs {
		name := fmt.Sprintf("Benchmak_%d_%d", bc.containerCount, bc.itemCount)
		b.Run(name, func(b *testing.B) {
			benchmarkContainerRead(bc.containerCount, bc.itemCount, b)
		})
	}
}

func benchmarkContainerRead(containerCount, itemCount int, b *testing.B) {
	for i := 0; i <= containerCount; i++ {
		container := Container{}
		for j := 0; j <= itemCount; j++ {
			container.Add(randomFloatType(), randomStringType(), randomByteType(), randomKiloByteType(), randomMegaByteType())
		}
		fmt.Fprint(ioutil.Discard, container.String(time.Now()))
	}
}

func BenchmarkJobResultDataTypes(b *testing.B) {
	bcs := []struct {
		containerCount int
		itemCount      int
	}{
		{100, 10},
		{1000, 10},
		{100, 100},
		{1000, 100},
		{10000, 30},
	}
	mapper := DefaultMapper()
	for _, bc := range bcs {
		name := fmt.Sprintf("Benchmak_%d_%d", bc.containerCount, bc.itemCount)
		b.Run(name, func(b *testing.B) {
			benchmarkJobResultDataTypes(mapper, bc.containerCount, bc.itemCount, b)
		})
	}
}

func benchmarkJobResultDataTypes(mapper Mapper, containerCount, itemCount int, b *testing.B) {
	for i := 0; i <= containerCount; i++ {
		container := Container{}
		buf := new(bytes.Buffer)
		for j := 0; j <= itemCount; j++ {
			container.Add(randomFloatType(), randomStringType(), randomByteType(), randomKiloByteType(), randomMegaByteType())
		}
		buf.WriteString(container.String(time.Now()))
		res := JobResultDataTypes(buf, mapper)
		if len(res.String(time.Now())) == 0 {
			fmt.Println(0)
		}
	}
}

func randomFloatType() FloatType {
	return FloatType{Key: randomString(10), Value: rand.Float64() * 1000}
}
func randomStringType() StringType {
	return StringType{Key: randomString(10), Value: randomString(20)}
}
func randomByteType() ByteType {
	return ByteType{Key: randomString(10), Value: rand.Float64() * 1000}
}
func randomKiloByteType() KiloByteType {
	return KiloByteType{Key: randomString(10), Value: rand.Float64() * 1000}
}
func randomMegaByteType() MegaByteType {
	return MegaByteType{Key: randomString(10), Value: rand.Float64() * 1000}
}

func randomString(count int) string {
	result := make([]rune, count)
	for i := range result {
		result[i] = runes[rand.Intn(len(runes))]
	}
	return string(result)
}
