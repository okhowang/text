// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

// This program generates tables.go:
//	go run maketables.go | gofmt > tables.go

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func main() {
	fmt.Printf("// generated by go run maketables.go; DO NOT EDIT\n\n")
	fmt.Printf("// Package korean provides Korean encodings such as EUC-KR.\n")
	fmt.Printf("package korean\n\n")

	res, err := http.Get("http://encoding.spec.whatwg.org/index-euc-kr.txt")
	if err != nil {
		log.Fatalf("Get: %v", err)
	}
	defer res.Body.Close()

	mapping := [65536]uint16{}
	reverse := [65536]uint16{}

	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		s := strings.TrimSpace(scanner.Text())
		if s == "" || s[0] == '#' {
			continue
		}
		x, y := uint16(0), uint16(0)
		if _, err := fmt.Sscanf(s, "%d 0x%x", &x, &y); err != nil {
			log.Fatalf("could not parse %q", s)
		}
		if x < 0 || 178*(0xc7-0x81)+(0xfe-0xc7)*94+(0xff-0xa1) <= x {
			log.Fatalf("EUC-KR code %d is out of range", x)
		}
		mapping[x] = y
		if reverse[y] == 0 {
			c0, c1 := uint16(0), uint16(0)
			if x < 178*(0xc7-0x81) {
				c0 = uint16(x/178) + 0x81
				c1 = uint16(x % 178)
				switch {
				case c1 < 1*26:
					c1 += 0x41
				case c1 < 2*26:
					c1 += 0x47
				default:
					c1 += 0x4d
				}
			} else {
				x -= 178 * (0xc7 - 0x81)
				c0 = uint16(x/94) + 0xc7
				c1 = uint16(x%94) + 0xa1
			}
			reverse[y] = c0<<8 | c1
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("scanner error: %v", err)
	}

	fmt.Printf("// eucKRDecode is the decoding table from EUC-KR code to Unicode.\n")
	fmt.Printf("// It is defined at http://encoding.spec.whatwg.org/index-euc-kr.txt\n")
	fmt.Printf("var eucKRDecode = [...]uint16{\n")
	for i, v := range mapping {
		if v != 0 {
			fmt.Printf("\t%d: 0x%04X,\n", i, v)
		}
	}
	fmt.Printf("}\n\n")

	fmt.Printf("// eucKREncode is the encoding table from Unicode to EUC-KR code.\n")
	fmt.Printf("var eucKREncode = [65536]uint16{\n")
	for i, v := range reverse {
		if v != 0 {
			fmt.Printf("\t%d: 0x%04X,\n", i, v)
		}
	}
	fmt.Printf("}\n\n")
}