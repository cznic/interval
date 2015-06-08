// Copyright (c) 2015 The Interval Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package interval handles sets of ordered values laying between two, possibly
// infinite, bounds.
//
// Note: Intervals are usually defined as a set of [extended] real numbers.
// The model of the interval provided by Interface does not know - and thus
// does not care about the type of the bounds, if any. That may lead to correct
// but possibly surprising effects when the bounds domain is not ℝ. An interval
// may be non empty, like for example the open interval (1, 2), but  no integer
// value lies between the bounds.
//
// See also: http://en.wikipedia.org/wiki/Interval_(mathematics)
package interval

import (
	"fmt"
	"math/big"
	"time"
)

var (
	_ Interface = (*BigInt)(nil)
	_ Interface = (*BigRat)(nil)
	_ Interface = (*Byte)(nil)
	_ Interface = (*Duration)(nil)
	_ Interface = (*Float32)(nil)
	_ Interface = (*Float64)(nil)
	_ Interface = (*Int)(nil)
	_ Interface = (*Int16)(nil)
	_ Interface = (*Int32)(nil)
	_ Interface = (*Int64)(nil)
	_ Interface = (*Int8)(nil)
	_ Interface = (*String)(nil)
	_ Interface = (*Time)(nil)
	_ Interface = (*Uint)(nil)
	_ Interface = (*Uint16)(nil)
	_ Interface = (*Uint32)(nil)
	_ Interface = (*Uint64)(nil)
)

// Class represent a type of an interval.
type Class int

// Interval classes.
const (
	Unbounded  Class = iota // (-∞, ∞).
	Empty                   // {}.
	Degenerate              // [a, a] = {a}.

	// Proper and bounded:
	Open       // (a, b) = {x | a < x < b}.
	Closed     // [a, b] = {x | a <= x <= b}.
	LeftOpen   // (a, b] = {x | a < x <= b}.
	LeftClosed // [a, b) = {x | a <= x < b}.

	// Left-bounded and right-unbounded:
	LeftBoundedOpen   // (a, ∞) = {x | x > a}.
	LeftBoundedClosed // [a, ∞) = {x | x >= a}.

	// Left-unbounded and right-bounded:
	RightBoundedOpen   // (-∞, b) = {x | x < b}.
	RightBoundedClosed // (-∞, b] = {x | x <= b}.

	nClasses
)

// Interface represents an unbounded, empty, degenerate, proper, left-bounded
// or right-bounded interval. Where appropriate, the interval bounds are
// defined by ordered values, named by convention a and b, when present.
//
// Proper intervals have an invariant: a < b.
//
// CompareXX return value must obey these rules
//
//	< 0 if interval A or B <  other interval A or B
//	  0 if interval A or B == other interval A or B
//	> 0 if interval A or B >  other interval A or B
type Interface interface {
	// Class returns the interval class.
	Class() Class
	// Clone clones the interval.
	Clone() Interface
	// CompareAA compares interval.A and other.A.
	CompareAA(other Interface) int
	// CompareAB compares interval.A and other.B.
	CompareAB(other Interface) int
	// CompareBB compares interval.B and other.B.
	CompareBB(other Interface) int
	// SetClass sets the interval class.
	SetClass(Class)
	// SetAB sets interval.A using interval.B.
	SetAB()
	// SetB sets interval.B using other.B.
	SetB(other Interface)
	// SetBA sets interval.B using other.A.
	SetBA(other Interface)
}

func compareBA(x, y Interface) int            { return -y.CompareAB(x) }
func setAB(x Interface) Interface             { x.SetAB(); return x }
func setB(x, y Interface) Interface           { x.SetB(y); return x }
func setBA(x, y Interface) Interface          { x.SetBA(y); return x }
func setClass(x Interface, c Class) Interface { x.SetClass(c); return x }

func str(c Class, a, b interface{}) string {
	switch c {
	case Unbounded:
		return fmt.Sprintf("(-∞, ∞)")
	case Empty:
		return fmt.Sprintf("{}")
	case Degenerate:
		return fmt.Sprintf("{%v}", a)
	case Open:
		return fmt.Sprintf("(%v, %v)", a, b)
	case Closed:
		return fmt.Sprintf("[%v, %v]", a, b)
	case LeftOpen:
		return fmt.Sprintf("(%v, %v]", a, b)
	case LeftClosed:
		return fmt.Sprintf("[%v, %v)", a, b)
	case LeftBoundedOpen:
		return fmt.Sprintf("(%v, ∞)", a)
	case LeftBoundedClosed:
		return fmt.Sprintf("[%v, ∞)", a)
	case RightBoundedOpen:
		return fmt.Sprintf("(-∞, %v)", b)
	case RightBoundedClosed:
		return fmt.Sprintf("(-∞, %v]", b)
	}
	panic("internal error")
}

func hash(x, y Interface) (Interface, Interface, int) {
	xc := x.Class()
	yc := y.Class()
	if xc > yc {
		x, y = y, x
		xc, yc = yc, xc
	}
	var r int
	switch xc {
	case Degenerate, Open, Closed, LeftOpen, LeftClosed, LeftBoundedOpen, LeftBoundedClosed: // x has A
		switch yc {
		case Degenerate, Open, Closed, LeftOpen, LeftClosed, LeftBoundedOpen, LeftBoundedClosed: // y has A
			r |= enc(x.CompareAA(y))
		}
		switch yc {
		case Open, Closed, LeftOpen, LeftClosed, RightBoundedOpen, RightBoundedClosed: // y has B
			r |= enc(x.CompareAB(y)) << 2
		}
	}
	switch xc {
	case Open, Closed, LeftOpen, LeftClosed, RightBoundedOpen, RightBoundedClosed: // x has B
		switch yc {
		case Degenerate, Open, Closed, LeftOpen, LeftClosed, LeftBoundedOpen, LeftBoundedClosed: // y has A
			r |= enc(compareBA(x, y)) << 4
		}
		switch yc {
		case Open, Closed, LeftOpen, LeftClosed, RightBoundedOpen, RightBoundedClosed: // y has B
			r |= enc(x.CompareBB(y)) << 6
		}
	}
	return x, y, r | (int(xc)*int(nClasses)+int(yc))<<8
}

// -- 00
// -1 01
//  0 10
//  1 11
func enc(n int) int {
	if n < 0 {
		return 1
	}

	if n == 0 {
		return 2
	}

	return 3
}

// Float32 is an interval having float32 bounds.
//
// Note: Using NaNs as bounds has undefined behavior.
type Float32 struct {
	Cls  Class
	A, B float32
}

// String implements fmt.Stringer.
func (i *Float32) String() string { return str(i.Cls, i.A, i.B) }

// Class implements Interface.
func (i *Float32) Class() Class { return i.Cls }

// SetClass implements Interface.
func (i *Float32) SetClass(c Class) { i.Cls = c }

// Clone implements Interface.
func (i *Float32) Clone() Interface { j := *i; return &j }

// CompareAA implements Interface.
func (i *Float32) CompareAA(other Interface) int {
	if i.A < other.(*Float32).A {
		return -1
	}

	if i.A > other.(*Float32).A {
		return 1
	}

	return 0
}

// CompareAB implements Interface.
func (i *Float32) CompareAB(other Interface) int {
	if i.A < other.(*Float32).B {
		return -1
	}

	if i.A > other.(*Float32).B {
		return 1
	}

	return 0
}

// CompareBB implements Interface.
func (i *Float32) CompareBB(other Interface) int {
	if i.A < other.(*Float32).B {
		return -1
	}

	if i.A > other.(*Float32).B {
		return 1
	}

	return 0
}

// SetAB implements Interface.
func (i *Float32) SetAB() { i.A = i.B }

// SetB implements Interface.
func (i *Float32) SetB(other Interface) { i.B = other.(*Float32).B }

// SetBA implements Interface.
func (i *Float32) SetBA(other Interface) { i.B = other.(*Float32).A }

// Float64 is an interval having float64 bounds.
//
// Note: Using NaNs as bounds has undefined behavior.
type Float64 struct {
	Cls  Class
	A, B float64
}

// String implements fmt.Stringer.
func (i *Float64) String() string { return str(i.Cls, i.A, i.B) }

// Class implements Interface.
func (i *Float64) Class() Class { return i.Cls }

// SetClass implements Interface.
func (i *Float64) SetClass(c Class) { i.Cls = c }

// Clone implements Interface.
func (i *Float64) Clone() Interface { j := *i; return &j }

// CompareAA implements Interface.
func (i *Float64) CompareAA(other Interface) int {
	if i.A < other.(*Float64).A {
		return -1
	}

	if i.A > other.(*Float64).A {
		return 1
	}

	return 0
}

// CompareAB implements Interface.
func (i *Float64) CompareAB(other Interface) int {
	if i.A < other.(*Float64).B {
		return -1
	}

	if i.A > other.(*Float64).B {
		return 1
	}

	return 0
}

// CompareBB implements Interface.
func (i *Float64) CompareBB(other Interface) int {
	if i.A < other.(*Float64).B {
		return -1
	}

	if i.A > other.(*Float64).B {
		return 1
	}

	return 0
}

// SetAB implements Interface.
func (i *Float64) SetAB() { i.A = i.B }

// SetB implements Interface.
func (i *Float64) SetB(other Interface) { i.B = other.(*Float64).B }

// SetBA implements Interface.
func (i *Float64) SetBA(other Interface) { i.B = other.(*Float64).A }

// Int8 is an interval having int8 bounds.
type Int8 struct {
	Cls  Class
	A, B int8
}

// String implements fmt.Stringer.
func (i *Int8) String() string { return str(i.Cls, i.A, i.B) }

// Class implements Interface.
func (i *Int8) Class() Class { return i.Cls }

// SetClass implements Interface.
func (i *Int8) SetClass(c Class) { i.Cls = c }

// Clone implements Interface.
func (i *Int8) Clone() Interface { j := *i; return &j }

// CompareAA implements Interface.
func (i *Int8) CompareAA(other Interface) int {
	if i.A < other.(*Int8).A {
		return -1
	}

	if i.A > other.(*Int8).A {
		return 1
	}

	return 0
}

// CompareAB implements Interface.
func (i *Int8) CompareAB(other Interface) int {
	if i.A < other.(*Int8).B {
		return -1
	}

	if i.A > other.(*Int8).B {
		return 1
	}

	return 0
}

// CompareBB implements Interface.
func (i *Int8) CompareBB(other Interface) int {
	if i.A < other.(*Int8).B {
		return -1
	}

	if i.A > other.(*Int8).B {
		return 1
	}

	return 0
}

// SetAB implements Interface.
func (i *Int8) SetAB() { i.A = i.B }

// SetB implements Interface.
func (i *Int8) SetB(other Interface) { i.B = other.(*Int8).B }

// SetBA implements Interface.
func (i *Int8) SetBA(other Interface) { i.B = other.(*Int8).A }

// Int16 is an interval having int16 bounds.
type Int16 struct {
	Cls  Class
	A, B int16
}

// String implements fmt.Stringer.
func (i *Int16) String() string { return str(i.Cls, i.A, i.B) }

// Class implements Interface.
func (i *Int16) Class() Class { return i.Cls }

// SetClass implements Interface.
func (i *Int16) SetClass(c Class) { i.Cls = c }

// Clone implements Interface.
func (i *Int16) Clone() Interface { j := *i; return &j }

// CompareAA implements Interface.
func (i *Int16) CompareAA(other Interface) int {
	if i.A < other.(*Int16).A {
		return -1
	}

	if i.A > other.(*Int16).A {
		return 1
	}

	return 0
}

// CompareAB implements Interface.
func (i *Int16) CompareAB(other Interface) int {
	if i.A < other.(*Int16).B {
		return -1
	}

	if i.A > other.(*Int16).B {
		return 1
	}

	return 0
}

// CompareBB implements Interface.
func (i *Int16) CompareBB(other Interface) int {
	if i.A < other.(*Int16).B {
		return -1
	}

	if i.A > other.(*Int16).B {
		return 1
	}

	return 0
}

// SetAB implements Interface.
func (i *Int16) SetAB() { i.A = i.B }

// SetB implements Interface.
func (i *Int16) SetB(other Interface) { i.B = other.(*Int16).B }

// SetBA implements Interface.
func (i *Int16) SetBA(other Interface) { i.B = other.(*Int16).A }

// Int32 is an interval having int32 bounds.
type Int32 struct {
	Cls  Class
	A, B int32
}

// String implements fmt.Stringer.
func (i *Int32) String() string { return str(i.Cls, i.A, i.B) }

// Class implements Interface.
func (i *Int32) Class() Class { return i.Cls }

// SetClass implements Interface.
func (i *Int32) SetClass(c Class) { i.Cls = c }

// Clone implements Interface.
func (i *Int32) Clone() Interface { j := *i; return &j }

// CompareAA implements Interface.
func (i *Int32) CompareAA(other Interface) int {
	if i.A < other.(*Int32).A {
		return -1
	}

	if i.A > other.(*Int32).A {
		return 1
	}

	return 0
}

// CompareAB implements Interface.
func (i *Int32) CompareAB(other Interface) int {
	if i.A < other.(*Int32).B {
		return -1
	}

	if i.A > other.(*Int32).B {
		return 1
	}

	return 0
}

// CompareBB implements Interface.
func (i *Int32) CompareBB(other Interface) int {
	if i.A < other.(*Int32).B {
		return -1
	}

	if i.A > other.(*Int32).B {
		return 1
	}

	return 0
}

// SetAB implements Interface.
func (i *Int32) SetAB() { i.A = i.B }

// SetB implements Interface.
func (i *Int32) SetB(other Interface) { i.B = other.(*Int32).B }

// SetBA implements Interface.
func (i *Int32) SetBA(other Interface) { i.B = other.(*Int32).A }

// Int64 is an interval having int64 bounds.
type Int64 struct {
	Cls  Class
	A, B int64
}

// String implements fmt.Stringer.
func (i *Int64) String() string { return str(i.Cls, i.A, i.B) }

// Class implements Interface.
func (i *Int64) Class() Class { return i.Cls }

// SetClass implements Interface.
func (i *Int64) SetClass(c Class) { i.Cls = c }

// Clone implements Interface.
func (i *Int64) Clone() Interface { j := *i; return &j }

// CompareAA implements Interface.
func (i *Int64) CompareAA(other Interface) int {
	if i.A < other.(*Int64).A {
		return -1
	}

	if i.A > other.(*Int64).A {
		return 1
	}

	return 0
}

// CompareAB implements Interface.
func (i *Int64) CompareAB(other Interface) int {
	if i.A < other.(*Int64).B {
		return -1
	}

	if i.A > other.(*Int64).B {
		return 1
	}

	return 0
}

// CompareBB implements Interface.
func (i *Int64) CompareBB(other Interface) int {
	if i.A < other.(*Int64).B {
		return -1
	}

	if i.A > other.(*Int64).B {
		return 1
	}

	return 0
}

// SetAB implements Interface.
func (i *Int64) SetAB() { i.A = i.B }

// SetB implements Interface.
func (i *Int64) SetB(other Interface) { i.B = other.(*Int64).B }

// SetBA implements Interface.
func (i *Int64) SetBA(other Interface) { i.B = other.(*Int64).A }

// Int is an interval having int bounds.
type Int struct {
	Cls  Class
	A, B int
}

// String implements fmt.Stringer.
func (i *Int) String() string { return str(i.Cls, i.A, i.B) }

// Class implements Interface.
func (i *Int) Class() Class { return i.Cls }

// SetClass implements Interface.
func (i *Int) SetClass(c Class) { i.Cls = c }

// Clone implements Interface.
func (i *Int) Clone() Interface { j := *i; return &j }

// CompareAA implements Interface.
func (i *Int) CompareAA(other Interface) int {
	if i.A < other.(*Int).A {
		return -1
	}

	if i.A > other.(*Int).A {
		return 1
	}

	return 0
}

// CompareAB implements Interface.
func (i *Int) CompareAB(other Interface) int {
	if i.A < other.(*Int).B {
		return -1
	}

	if i.A > other.(*Int).B {
		return 1
	}

	return 0
}

// CompareBB implements Interface.
func (i *Int) CompareBB(other Interface) int {
	if i.A < other.(*Int).B {
		return -1
	}

	if i.A > other.(*Int).B {
		return 1
	}

	return 0
}

// SetAB implements Interface.
func (i *Int) SetAB() { i.A = i.B }

// SetB implements Interface.
func (i *Int) SetB(other Interface) { i.B = other.(*Int).B }

// SetBA implements Interface.
func (i *Int) SetBA(other Interface) { i.B = other.(*Int).A }

// Byte is an interval having byte bounds.
type Byte struct {
	Cls  Class
	A, B byte
}

// String implements fmt.Stringer.
func (i *Byte) String() string { return str(i.Cls, i.A, i.B) }

// Class implements Interface.
func (i *Byte) Class() Class { return i.Cls }

// SetClass implements Interface.
func (i *Byte) SetClass(c Class) { i.Cls = c }

// Clone implements Interface.
func (i *Byte) Clone() Interface { j := *i; return &j }

// CompareAA implements Interface.
func (i *Byte) CompareAA(other Interface) int {
	if i.A < other.(*Byte).A {
		return -1
	}

	if i.A > other.(*Byte).A {
		return 1
	}

	return 0
}

// CompareAB implements Interface.
func (i *Byte) CompareAB(other Interface) int {
	if i.A < other.(*Byte).B {
		return -1
	}

	if i.A > other.(*Byte).B {
		return 1
	}

	return 0
}

// CompareBB implements Interface.
func (i *Byte) CompareBB(other Interface) int {
	if i.A < other.(*Byte).B {
		return -1
	}

	if i.A > other.(*Byte).B {
		return 1
	}

	return 0
}

// SetAB implements Interface.
func (i *Byte) SetAB() { i.A = i.B }

// SetB implements Interface.
func (i *Byte) SetB(other Interface) { i.B = other.(*Byte).B }

// SetBA implements Interface.
func (i *Byte) SetBA(other Interface) { i.B = other.(*Byte).A }

// Uint16 is an interval having uint16 bounds.
type Uint16 struct {
	Cls  Class
	A, B uint16
}

// String implements fmt.Stringer.
func (i *Uint16) String() string { return str(i.Cls, i.A, i.B) }

// Class implements Interface.
func (i *Uint16) Class() Class { return i.Cls }

// SetClass implements Interface.
func (i *Uint16) SetClass(c Class) { i.Cls = c }

// Clone implements Interface.
func (i *Uint16) Clone() Interface { j := *i; return &j }

// CompareAA implements Interface.
func (i *Uint16) CompareAA(other Interface) int {
	if i.A < other.(*Uint16).A {
		return -1
	}

	if i.A > other.(*Uint16).A {
		return 1
	}

	return 0
}

// CompareAB implements Interface.
func (i *Uint16) CompareAB(other Interface) int {
	if i.A < other.(*Uint16).B {
		return -1
	}

	if i.A > other.(*Uint16).B {
		return 1
	}

	return 0
}

// CompareBB implements Interface.
func (i *Uint16) CompareBB(other Interface) int {
	if i.A < other.(*Uint16).B {
		return -1
	}

	if i.A > other.(*Uint16).B {
		return 1
	}

	return 0
}

// SetAB implements Interface.
func (i *Uint16) SetAB() { i.A = i.B }

// SetB implements Interface.
func (i *Uint16) SetB(other Interface) { i.B = other.(*Uint16).B }

// SetBA implements Interface.
func (i *Uint16) SetBA(other Interface) { i.B = other.(*Uint16).A }

// Uint32 is an interval having uint32 bounds.
type Uint32 struct {
	Cls  Class
	A, B uint32
}

// String implements fmt.Stringer.
func (i *Uint32) String() string { return str(i.Cls, i.A, i.B) }

// Class implements Interface.
func (i *Uint32) Class() Class { return i.Cls }

// SetClass implements Interface.
func (i *Uint32) SetClass(c Class) { i.Cls = c }

// Clone implements Interface.
func (i *Uint32) Clone() Interface { j := *i; return &j }

// CompareAA implements Interface.
func (i *Uint32) CompareAA(other Interface) int {
	if i.A < other.(*Uint32).A {
		return -1
	}

	if i.A > other.(*Uint32).A {
		return 1
	}

	return 0
}

// CompareAB implements Interface.
func (i *Uint32) CompareAB(other Interface) int {
	if i.A < other.(*Uint32).B {
		return -1
	}

	if i.A > other.(*Uint32).B {
		return 1
	}

	return 0
}

// CompareBB implements Interface.
func (i *Uint32) CompareBB(other Interface) int {
	if i.A < other.(*Uint32).B {
		return -1
	}

	if i.A > other.(*Uint32).B {
		return 1
	}

	return 0
}

// SetAB implements Interface.
func (i *Uint32) SetAB() { i.A = i.B }

// SetB implements Interface.
func (i *Uint32) SetB(other Interface) { i.B = other.(*Uint32).B }

// SetBA implements Interface.
func (i *Uint32) SetBA(other Interface) { i.B = other.(*Uint32).A }

// Uint64 is an interval having uint64 bounds.
type Uint64 struct {
	Cls  Class
	A, B uint64
}

// String implements fmt.Stringer.
func (i *Uint64) String() string { return str(i.Cls, i.A, i.B) }

// Class implements Interface.
func (i *Uint64) Class() Class { return i.Cls }

// SetClass implements Interface.
func (i *Uint64) SetClass(c Class) { i.Cls = c }

// Clone implements Interface.
func (i *Uint64) Clone() Interface { j := *i; return &j }

// CompareAA implements Interface.
func (i *Uint64) CompareAA(other Interface) int {
	if i.A < other.(*Uint64).A {
		return -1
	}

	if i.A > other.(*Uint64).A {
		return 1
	}

	return 0
}

// CompareAB implements Interface.
func (i *Uint64) CompareAB(other Interface) int {
	if i.A < other.(*Uint64).B {
		return -1
	}

	if i.A > other.(*Uint64).B {
		return 1
	}

	return 0
}

// CompareBB implements Interface.
func (i *Uint64) CompareBB(other Interface) int {
	if i.A < other.(*Uint64).B {
		return -1
	}

	if i.A > other.(*Uint64).B {
		return 1
	}

	return 0
}

// SetAB implements Interface.
func (i *Uint64) SetAB() { i.A = i.B }

// SetB implements Interface.
func (i *Uint64) SetB(other Interface) { i.B = other.(*Uint64).B }

// SetBA implements Interface.
func (i *Uint64) SetBA(other Interface) { i.B = other.(*Uint64).A }

// Uint is an interval having uint bounds.
type Uint struct {
	Cls  Class
	A, B uint
}

// String implements fmt.Stringer.
func (i *Uint) String() string { return str(i.Cls, i.A, i.B) }

// Class implements Interface.
func (i *Uint) Class() Class { return i.Cls }

// SetClass implements Interface.
func (i *Uint) SetClass(c Class) { i.Cls = c }

// Clone implements Interface.
func (i *Uint) Clone() Interface { j := *i; return &j }

// CompareAA implements Interface.
func (i *Uint) CompareAA(other Interface) int {
	if i.A < other.(*Uint).A {
		return -1
	}

	if i.A > other.(*Uint).A {
		return 1
	}

	return 0
}

// CompareAB implements Interface.
func (i *Uint) CompareAB(other Interface) int {
	if i.A < other.(*Uint).B {
		return -1
	}

	if i.A > other.(*Uint).B {
		return 1
	}

	return 0
}

// CompareBB implements Interface.
func (i *Uint) CompareBB(other Interface) int {
	if i.A < other.(*Uint).B {
		return -1
	}

	if i.A > other.(*Uint).B {
		return 1
	}

	return 0
}

// SetAB implements Interface.
func (i *Uint) SetAB() { i.A = i.B }

// SetB implements Interface.
func (i *Uint) SetB(other Interface) { i.B = other.(*Uint).B }

// SetBA implements Interface.
func (i *Uint) SetBA(other Interface) { i.B = other.(*Uint).A }

// String is an interval having string bounds.
type String struct {
	Cls  Class
	A, B string
}

// String implements fmt.Stringer.
func (i *String) String() string { return str(i.Cls, i.A, i.B) }

// Class implements Interface.
func (i *String) Class() Class { return i.Cls }

// SetClass implements Interface.
func (i *String) SetClass(c Class) { i.Cls = c }

// Clone implements Interface.
func (i *String) Clone() Interface { j := *i; return &j }

// CompareAA implements Interface.
func (i *String) CompareAA(other Interface) int {
	if i.A < other.(*String).A {
		return -1
	}

	if i.A > other.(*String).A {
		return 1
	}

	return 0
}

// CompareAB implements Interface.
func (i *String) CompareAB(other Interface) int {
	if i.A < other.(*String).B {
		return -1
	}

	if i.A > other.(*String).B {
		return 1
	}

	return 0
}

// CompareBB implements Interface.
func (i *String) CompareBB(other Interface) int {
	if i.A < other.(*String).B {
		return -1
	}

	if i.A > other.(*String).B {
		return 1
	}

	return 0
}

// SetAB implements Interface.
func (i *String) SetAB() { i.A = i.B }

// SetB implements Interface.
func (i *String) SetB(other Interface) { i.B = other.(*String).B }

// SetBA implements Interface.
func (i *String) SetBA(other Interface) { i.B = other.(*String).A }

// Time is an interval having time.Time bounds.
type Time struct {
	Cls  Class
	A, B time.Time
}

// String implements fmt.Stringer.
func (i *Time) String() string { return str(i.Cls, i.A, i.B) }

// Class implements Interface.
func (i *Time) Class() Class { return i.Cls }

// SetClass implements Interface.
func (i *Time) SetClass(c Class) { i.Cls = c }

// Clone implements Interface.
func (i *Time) Clone() Interface { j := *i; return &j }

// CompareAA implements Interface.
func (i *Time) CompareAA(other Interface) int {
	if i.A.Before(other.(*Time).A) {
		return -1
	}

	if i.A.After(other.(*Time).A) {
		return 1
	}

	return 0
}

// CompareAB implements Interface.
func (i *Time) CompareAB(other Interface) int {
	if i.A.Before(other.(*Time).B) {
		return -1
	}

	if i.A.After(other.(*Time).B) {
		return 1
	}

	return 0
}

// CompareBB implements Interface.
func (i *Time) CompareBB(other Interface) int {
	if i.A.Before(other.(*Time).B) {
		return -1
	}

	if i.A.After(other.(*Time).B) {
		return 1
	}

	return 0
}

// SetAB implements Interface.
func (i *Time) SetAB() { i.A = i.B }

// SetB implements Interface.
func (i *Time) SetB(other Interface) { i.B = other.(*Time).B }

// SetBA implements Interface.
func (i *Time) SetBA(other Interface) { i.B = other.(*Time).A }

// Duration is an interval having time.Duration bounds.
type Duration struct {
	Cls  Class
	A, B time.Duration
}

// String implements fmt.Stringer.
func (i *Duration) String() string { return str(i.Cls, i.A, i.B) }

// Class implements Interface.
func (i *Duration) Class() Class { return i.Cls }

// SetClass implements Interface.
func (i *Duration) SetClass(c Class) { i.Cls = c }

// Clone implements Interface.
func (i *Duration) Clone() Interface { j := *i; return &j }

// CompareAA implements Interface.
func (i *Duration) CompareAA(other Interface) int {
	if i.A < other.(*Duration).A {
		return -1
	}

	if i.A > other.(*Duration).A {
		return 1
	}

	return 0
}

// CompareAB implements Interface.
func (i *Duration) CompareAB(other Interface) int {
	if i.A < other.(*Duration).B {
		return -1
	}

	if i.A > other.(*Duration).B {
		return 1
	}

	return 0
}

// CompareBB implements Interface.
func (i *Duration) CompareBB(other Interface) int {
	if i.A < other.(*Duration).B {
		return -1
	}

	if i.A > other.(*Duration).B {
		return 1
	}

	return 0
}

// SetAB implements Interface.
func (i *Duration) SetAB() { i.A = i.B }

// SetB implements Interface.
func (i *Duration) SetB(other Interface) { i.B = other.(*Duration).B }

// SetBA implements Interface.
func (i *Duration) SetBA(other Interface) { i.B = other.(*Duration).A }

// BigInt is an interval having math/big.Int bounds.
type BigInt struct {
	Cls  Class
	A, B *big.Int
}

// String implements fmt.Stringer.
func (i *BigInt) String() string { return str(i.Cls, i.A, i.B) }

// Class implements Interface.
func (i *BigInt) Class() Class { return i.Cls }

// SetClass implements Interface.
func (i *BigInt) SetClass(c Class) { i.Cls = c }

// Clone implements Interface.
func (i *BigInt) Clone() Interface {
	j := &BigInt{Cls: i.Cls}
	if i.A != nil {
		j.A = big.NewInt(0).Set(i.A)
	}
	if i.B != nil {
		j.B = big.NewInt(0).Set(i.B)
	}
	return j
}

// CompareAA implements Interface.
func (i *BigInt) CompareAA(other Interface) int {
	return i.A.Cmp(other.(*BigInt).A)
}

// CompareAB implements Interface.
func (i *BigInt) CompareAB(other Interface) int {
	return i.A.Cmp(other.(*BigInt).B)
}

// CompareBB implements Interface.
func (i *BigInt) CompareBB(other Interface) int {
	return i.B.Cmp(other.(*BigInt).B)
}

// SetAB implements Interface.
func (i *BigInt) SetAB() { i.A.Set(i.B) }

// SetB implements Interface.
func (i *BigInt) SetB(other Interface) { i.B.Set(other.(*BigInt).B) }

// SetBA implements Interface.
func (i *BigInt) SetBA(other Interface) { i.B.Set(other.(*BigInt).A) }

// BigRat is an interval having math/big.Rat bounds.
type BigRat struct {
	Cls  Class
	A, B *big.Rat
}

// String implements fmt.Stringer.
func (i *BigRat) String() string { return str(i.Cls, i.A, i.B) }

// Class implements Interface.
func (i *BigRat) Class() Class { return i.Cls }

// SetClass implements Interface.
func (i *BigRat) SetClass(c Class) { i.Cls = c }

// Clone implements Interface.
func (i *BigRat) Clone() Interface {
	j := &BigRat{Cls: i.Cls}
	if i.A != nil {
		j.A = big.NewRat(1, 1).Set(i.A)
	}
	if i.B != nil {
		j.B = big.NewRat(1, 1).Set(i.B)
	}
	return j
}

// CompareAA implements Interface.
func (i *BigRat) CompareAA(other Interface) int {
	return i.A.Cmp(other.(*BigRat).A)
}

// CompareAB implements Interface.
func (i *BigRat) CompareAB(other Interface) int {
	return i.A.Cmp(other.(*BigRat).B)
}

// CompareBB implements Interface.
func (i *BigRat) CompareBB(other Interface) int {
	return i.B.Cmp(other.(*BigRat).B)
}

// SetAB implements Interface.
func (i *BigRat) SetAB() { i.A.Set(i.B) }

// SetB implements Interface.
func (i *BigRat) SetB(other Interface) { i.B.Set(other.(*BigRat).B) }

// SetBA implements Interface.
func (i *BigRat) SetBA(other Interface) { i.B.Set(other.(*BigRat).A) }
