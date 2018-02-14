// Copyright (c) 2015 The Interval Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interval

import (
	"fmt"
	"math/big"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/cznic/mathutil"
)

func caller(s string, va ...interface{}) {
	_, fn, fl, _ := runtime.Caller(2)
	fmt.Fprintf(os.Stderr, "caller: %s:%d: ", path.Base(fn), fl)
	fmt.Fprintf(os.Stderr, s, va...)
	fmt.Fprintln(os.Stderr)
	_, fn, fl, _ = runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "\tcallee: %s:%d: ", path.Base(fn), fl)
	fmt.Fprintln(os.Stderr)
}

func dbg(s string, va ...interface{}) {
	if s == "" {
		s = strings.Repeat("%v ", len(va))
	}
	_, fn, fl, _ := runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "dbg %s:%d: ", path.Base(fn), fl)
	fmt.Fprintf(os.Stderr, s, va...)
	fmt.Fprintln(os.Stderr)
}

func TODO(...interface{}) string {
	_, fn, fl, _ := runtime.Caller(1)
	return fmt.Sprintf("TODO: %s:%d:\n", path.Base(fn), fl)
}

func use(...interface{}) {}

// ============================================================================

var int128 mathutil.Int128

func TestIntersection(t *testing.T) {
	i := 0
	for xa := negInf; xa <= posInf; xa += 10 {
		for xb := xa + 10; xb <= posInf; xb += 10 {
			for _, xc := range classes {
				if (xa < negInf+10 || xb > posInf-10) && xc != Unbounded {
					continue
				}

				for ya := negInf; ya <= posInf; ya += 10 {
					for yb := ya + 10; yb <= posInf; yb += 10 {
						for _, yc := range classes {
							if (ya < negInf+10 || yb > posInf-10) && yc != Unbounded {
								continue
							}

							i++
							x := &interval{xc, xa, xb}
							y := &interval{yc, ya, yb}
							m := map[int]bool{}
							for n := negInf; n <= posInf; n += 5 {
								if x.has(n) && y.has(n) {
									m[n] = true
								}
							}
							result := Intersection(x, y).(*interval)
							for n := negInf; n <= posInf; n += 5 {
								if g, e := result.has(n), m[n]; g != e {
									t.Log(hash(x, y))
									t.Log(result)
									for n := negInf; n <= posInf; n += 5 {
										t.Log(n, x.has(n), y.has(n), result.has(n), m[n])
									}
									t.Fatalf("%v, %d: %v %v %v %v", i, n, x, y, g, e)
								}
							}
							x2 := &Int{xc, xa, xb}
							y2 := &Int{yc, ya, yb}
							result2 := Intersection(x2, y2).(*Int)
							if g, e := result.String(), result2.String(); g != e {
								t.Fatal(x, y, g, e)
							}
						}
					}
				}
			}
		}
	}
	t.Log(i)
}

func TestUnion(t *testing.T) {
	i := 0
	for xa := negInf; xa <= posInf; xa += 10 {
		for xb := xa + 10; xb <= posInf; xb += 10 {
			for _, xc := range classes {
				if (xa < negInf+10 || xb > posInf-10) && xc != Unbounded {
					continue
				}

				x := &interval{xc, xa, xb}
				for ya := negInf; ya <= posInf; ya += 10 {
					for yb := ya + 10; yb <= posInf; yb += 10 {
						for _, yc := range classes {
							if (ya < negInf+10 || yb > posInf-10) && yc != Unbounded {
								continue
							}

							i++
							y := &interval{yc, ya, yb}
							m := map[int]bool{}
							for n := negInf; n <= posInf; n += 5 {
								if x.has(n) || y.has(n) {
									m[n] = true
								}
							}
							result0 := Union(x, y)
							if result0 == nil {
								if g, e := true, isDisjointUnion(x, y); g != e {
									t.Fatal(i, x, y, g, e)
								}

								continue
							}

							result := result0.(*interval)
							for n := negInf; n <= posInf; n += 5 {
								if g, e := result.has(n), m[n]; g != e {
									t.Log(hash(x, y))
									t.Log(result)
									for n := negInf; n <= posInf; n += 5 {
										t.Log(n, x.has(n), y.has(n), result.has(n), m[n])
									}
									t.Fatalf("%v, %d: %v %v %v %v", i, n, x, y, g, e)
								}
							}
						}
					}
				}
			}
		}
	}
	t.Log(i)
}

func isDisjointUnion(x, y *interval) bool {
	const (
		initial = iota
		in
		after
	)
	var state int
	for n := negInf; n <= posInf; n += 5 {
		if x.has(n) || y.has(n) {
			switch state {
			case initial:
				state = in
			case in:
				// nop
			case after:
				return true
			}

			continue
		}

		switch state {
		case initial:
			// nop
		case in:
			state = after
		case after:
			// nop
		}

	}
	return false
}

func ExampleBigInt() {
	x := &BigInt{LeftOpen, big.NewInt(1), big.NewInt(2)}
	y := &BigInt{LeftClosed, big.NewInt(2), big.NewInt(3)}
	fmt.Printf("x %v, y %v: x ∩ y %v, x ∪ y %v", x, y, Intersection(x, y), Union(x, y))
	// Output:
	// x (1, 2], y [2, 3): x ∩ y {2}, x ∪ y (1, 3)
}

func ExampleBigRat() {
	x := &BigRat{LeftOpen, big.NewRat(1, 1), big.NewRat(2, 1)}
	y := &BigRat{LeftClosed, big.NewRat(2, 1), big.NewRat(3, 1)}
	fmt.Printf("x %v, y %v: x ∩ y %v, x ∪ y %v", x, y, Intersection(x, y), Union(x, y))
	// Output:
	// x (1/1, 2/1], y [2/1, 3/1): x ∩ y {2/1}, x ∪ y (1/1, 3/1)
}

func ExampleByte() {
	x := &Byte{LeftOpen, 1, 2}
	y := &Byte{LeftClosed, 2, 3}
	fmt.Printf("x %v, y %v: x ∩ y %v, x ∪ y %v", x, y, Intersection(x, y), Union(x, y))
	// Output:
	// x (1, 2], y [2, 3): x ∩ y {2}, x ∪ y (1, 3)
}

func ExampleDuration() {
	x := &Duration{LeftOpen, 1, 2}
	y := &Duration{LeftClosed, 2, 3}
	fmt.Printf("x %v, y %v: x ∩ y %v, x ∪ y %v", x, y, Intersection(x, y), Union(x, y))
	// Output:
	// x (1ns, 2ns], y [2ns, 3ns): x ∩ y {2ns}, x ∪ y (1ns, 3ns)
}

func ExampleFloat32() {
	x := &Float32{LeftOpen, 1, 2}
	y := &Float32{LeftClosed, 2, 3}
	fmt.Printf("x %v, y %v: x ∩ y %v, x ∪ y %v", x, y, Intersection(x, y), Union(x, y))
	// Output:
	// x (1, 2], y [2, 3): x ∩ y {2}, x ∪ y (1, 3)
}

func ExampleFloat64() {
	x := &Float32{LeftOpen, 1, 2}
	y := &Float32{LeftClosed, 2, 3}
	fmt.Printf("x %v, y %v: x ∩ y %v, x ∪ y %v", x, y, Intersection(x, y), Union(x, y))
	// Output:
	// x (1, 2], y [2, 3): x ∩ y {2}, x ∪ y (1, 3)
}

func ExampleInt() {
	x := &Int{LeftOpen, 1, 2}
	y := &Int{LeftClosed, 2, 3}
	fmt.Printf("x %v, y %v: x ∩ y %v, x ∪ y %v", x, y, Intersection(x, y), Union(x, y))
	// Output:
	// x (1, 2], y [2, 3): x ∩ y {2}, x ∪ y (1, 3)
}

func ExampleInt8() {
	x := &Int8{LeftOpen, 1, 2}
	y := &Int8{LeftClosed, 2, 3}
	fmt.Printf("x %v, y %v: x ∩ y %v, x ∪ y %v", x, y, Intersection(x, y), Union(x, y))
	// Output:
	// x (1, 2], y [2, 3): x ∩ y {2}, x ∪ y (1, 3)
}

func ExampleInt16() {
	x := &Int16{LeftOpen, 1, 2}
	y := &Int16{LeftClosed, 2, 3}
	fmt.Printf("x %v, y %v: x ∩ y %v, x ∪ y %v", x, y, Intersection(x, y), Union(x, y))
	// Output:
	// x (1, 2], y [2, 3): x ∩ y {2}, x ∪ y (1, 3)
}

func ExampleInt32() {
	x := &Int32{LeftOpen, 1, 2}
	y := &Int32{LeftClosed, 2, 3}
	fmt.Printf("x %v, y %v: x ∩ y %v, x ∪ y %v", x, y, Intersection(x, y), Union(x, y))
	// Output:
	// x (1, 2], y [2, 3): x ∩ y {2}, x ∪ y (1, 3)
}

func ExampleInt64() {
	x := &Int64{LeftOpen, 1, 2}
	y := &Int64{LeftClosed, 2, 3}
	fmt.Printf("x %v, y %v: x ∩ y %v, x ∪ y %v", x, y, Intersection(x, y), Union(x, y))
	// Output:
	// x (1, 2], y [2, 3): x ∩ y {2}, x ∪ y (1, 3)
}

func ExampleInt128() {
	x := &Int128{LeftOpen, int128.SetInt64(1), int128.SetInt64(2)}
	y := &Int128{LeftClosed, int128.SetInt64(2), int128.SetInt64(3)}
	fmt.Printf("x %v, y %v: x ∩ y %v, x ∪ y %v", x, y, Intersection(x, y), Union(x, y))
	// Output:
	// x (1, 2], y [2, 3): x ∩ y {2}, x ∪ y (1, 3)
}

func ExampleUint() {
	x := &Uint{LeftOpen, 1, 2}
	y := &Uint{LeftClosed, 2, 3}
	fmt.Printf("x %v, y %v: x ∩ y %v, x ∪ y %v", x, y, Intersection(x, y), Union(x, y))
	// Output:
	// x (1, 2], y [2, 3): x ∩ y {2}, x ∪ y (1, 3)
}

func ExampleUint16() {
	x := &Uint16{LeftOpen, 1, 2}
	y := &Uint16{LeftClosed, 2, 3}
	fmt.Printf("x %v, y %v: x ∩ y %v, x ∪ y %v", x, y, Intersection(x, y), Union(x, y))
	// Output:
	// x (1, 2], y [2, 3): x ∩ y {2}, x ∪ y (1, 3)
}

func ExampleUint32() {
	x := &Uint32{LeftOpen, 1, 2}
	y := &Uint32{LeftClosed, 2, 3}
	fmt.Printf("x %v, y %v: x ∩ y %v, x ∪ y %v", x, y, Intersection(x, y), Union(x, y))
	// Output:
	// x (1, 2], y [2, 3): x ∩ y {2}, x ∪ y (1, 3)
}

func ExampleUint64() {
	x := &Uint64{LeftOpen, 1, 2}
	y := &Uint64{LeftClosed, 2, 3}
	fmt.Printf("x %v, y %v: x ∩ y %v, x ∪ y %v", x, y, Intersection(x, y), Union(x, y))
	// Output:
	// x (1, 2], y [2, 3): x ∩ y {2}, x ∪ y (1, 3)
}

func ExampleString() {
	x := &String{LeftOpen, "aqua", "bar"}
	y := &String{LeftClosed, "bar", "closed"}
	fmt.Printf("x %v, y %v: x ∩ y %v, x ∪ y %v", x, y, Intersection(x, y), Union(x, y))
	// Output:
	// x (aqua, bar], y [bar, closed): x ∩ y {bar}, x ∪ y (aqua, closed)
}

func ExampleTime() {
	x := &Time{LeftOpen, time.Unix(1, 0), time.Unix(2, 0)}
	y := &Time{LeftClosed, time.Unix(2, 0), time.Unix(3, 0)}
	fmt.Printf("x %v, y %v: x ∩ y %v, x ∪ y %v", x, y, Intersection(x, y), Union(x, y))
	// Output:
	// x (1970-01-01 01:00:01 +0100 CET, 1970-01-01 01:00:02 +0100 CET], y [1970-01-01 01:00:02 +0100 CET, 1970-01-01 01:00:03 +0100 CET): x ∩ y {1970-01-01 01:00:02 +0100 CET}, x ∪ y (1970-01-01 01:00:01 +0100 CET, 1970-01-01 01:00:03 +0100 CET)
}
