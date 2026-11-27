package main

import (
	"go/ast"
	"strings"
)

// go1.27 renamed the simd/archsimd loads and stores. The transformer builds the
// go1.26 spelling; for the go1.27 file the names are rewritten structurally on
// the finished AST, never as text, because an old and a new name of one form
// collide (Store meant "array pointer" and now means "slice").
//
//	go1.26                          go1.27
//	archsimd.LoadFloat32x8(ptr)     archsimd.LoadFloat32x8Array(ptr)
//	archsimd.LoadFloat32x8Slice(s)  archsimd.LoadFloat32x8(s)
//	archsimd.LoadFloat32x8SlicePart archsimd.LoadFloat32x8Part
//	v.Store(ptr)                    v.StoreArray(ptr)
//	v.StoreMasked(ptr, m)           v.StoreArrayMasked(ptr, m)
//	v.StoreSlice(s)                 v.Store(s)
//	v.StoreSlicePart(s)             v.StorePart(s)
//	v.RoundToEven()                 v.Round()
//
// Reference: go1.27.4 simd/archsimd (slice_gen_amd64.go, types_amd64.go).
//
// Only archsimd names change. A call through another package (hwy, asm, math)
// is left alone, and so is a method of an asm promoted half-precision vector
// (Float16x8AVX2.StoreSlice, RoundToEven, ...), which hwy/asm defines itself.

// archsimdGo127Tag and archsimdPreGo127Tag constrain the two spellings of an
// AVX file.
func archsimdGo127Tag(base string) string    { return base + " && go1.27" }
func archsimdPreGo127Tag(base string) string { return base + " && !go1.27" }

// archsimdPkgFuncGo127 maps the go1.26 name of a package-level archsimd
// function to its go1.27 name.
func archsimdPkgFuncGo127(name string) string {
	switch {
	case strings.Contains(name, "SlicePart"):
		return strings.Replace(name, "SlicePart", "Part", 1)
	case strings.HasPrefix(name, "Load") && strings.HasSuffix(name, "Slice"):
		return strings.TrimSuffix(name, "Slice")
	case strings.HasPrefix(name, "Load") && !strings.HasSuffix(name, "Array") && !strings.HasSuffix(name, "Part"):
		return name + "Array"
	}
	return name
}

// archsimdMethodGo127 maps the go1.26 name of an archsimd vector method to its
// go1.27 name.
func archsimdMethodGo127(name string) string {
	switch name {
	case "Store":
		return "StoreArray"
	case "StoreMasked":
		return "StoreArrayMasked"
	case "StoreSlice":
		return "Store"
	case "StoreSlicePart":
		return "StorePart"
	}
	if strings.HasPrefix(name, "RoundToEven") {
		return "Round" + strings.TrimPrefix(name, "RoundToEven")
	}
	return name
}

// notReceiver lists identifiers that are packages, never vector values.
var notReceiver = map[string]bool{
	"hwy": true, "asm": true, "math": true, "stdmath": true, "bits": true,
	"unsafe": true, "vec": true, "algo": true, "archsimd": true, "sort": true,
}

// asmResultFree lists methods of an asm promoted vector that return an
// archsimd value (or a scalar) rather than another promoted vector.
var asmResultFree = map[string]bool{
	"ConvertToInt32": true, "AsInt32x8": true, "AsFloat32x8": true,
	"AsInt32x16": true, "AsFloat32x16": true, "Data": true,
	"Less": true, "Greater": true, "Equal": true, "NotEqual": true,
	"LessEqual": true, "GreaterEqual": true,
	"ReduceSum": true, "ReduceMax": true, "ReduceMin": true,
}

// mentionsAsm reports whether a type or value expression names the asm package.
func mentionsAsm(e ast.Expr) bool {
	found := false
	ast.Inspect(e, func(n ast.Node) bool {
		if s, ok := n.(*ast.SelectorExpr); ok {
			if id, ok := s.X.(*ast.Ident); ok && id.Name == "asm" {
				found = true
			}
		}
		return !found
	})
	return found
}

// asmEnv records which local identifiers hold an asm promoted vector.
type asmEnv map[string]bool

func (env asmEnv) isAsm(e ast.Expr) bool {
	switch x := e.(type) {
	case *ast.Ident:
		return env[x.Name]
	case *ast.ParenExpr:
		return env.isAsm(x.X)
	case *ast.StarExpr:
		return env.isAsm(x.X)
	case *ast.UnaryExpr:
		return env.isAsm(x.X)
	case *ast.IndexExpr:
		return env.isAsm(x.X)
	case *ast.SliceExpr:
		return env.isAsm(x.X)
	case *ast.CompositeLit:
		return x.Type != nil && mentionsAsm(x.Type)
	case *ast.CallExpr:
		sel, ok := x.Fun.(*ast.SelectorExpr)
		if !ok {
			if mentionsAsm(x.Fun) {
				return true
			}
			return env.anyArgAsm(x)
		}
		if id, ok := sel.X.(*ast.Ident); ok && notReceiver[id.Name] {
			if id.Name == "asm" {
				return true
			}
			if id.Name == "archsimd" {
				return false
			}
			// A package function over promoted vectors returns one.
			return env.anyArgAsm(x)
		}
		if asmResultFree[sel.Sel.Name] {
			return false
		}
		return env.isAsm(sel.X)
	}
	return false
}

func (env asmEnv) anyArgAsm(call *ast.CallExpr) bool {
	for _, a := range call.Args {
		if env.isAsm(a) {
			return true
		}
	}
	return false
}

// buildAsmEnv scans a function in source order and records its promoted vectors.
func buildAsmEnv(fn *ast.FuncDecl) asmEnv {
	env := asmEnv{}
	params := func(ft *ast.FuncType) {
		if ft == nil || ft.Params == nil {
			return
		}
		for _, f := range ft.Params.List {
			if mentionsAsm(f.Type) {
				for _, n := range f.Names {
					env[n.Name] = true
				}
			}
		}
	}
	params(fn.Type)
	if fn.Body == nil {
		return env
	}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.FuncLit:
			params(s.Type)
		case *ast.AssignStmt:
			for i, l := range s.Lhs {
				id, ok := l.(*ast.Ident)
				if !ok {
					continue
				}
				if len(s.Lhs) == len(s.Rhs) {
					env[id.Name] = env.isAsm(s.Rhs[i])
				} else {
					env[id.Name] = false
				}
			}
		case *ast.ValueSpec:
			typed := s.Type != nil && mentionsAsm(s.Type)
			for i, id := range s.Names {
				switch {
				case typed:
					env[id.Name] = true
				case len(s.Names) == len(s.Values):
					env[id.Name] = env.isAsm(s.Values[i])
				default:
					env[id.Name] = false
				}
			}
		case *ast.RangeStmt:
			for _, e := range []ast.Expr{s.Key, s.Value} {
				if id, ok := e.(*ast.Ident); ok {
					env[id.Name] = false
				}
			}
		}
		return true
	})
	return env
}

// rewriteArchsimdGo127 renames, in place, the archsimd names of one function
// from the go1.26 spelling to the go1.27 one.
func rewriteArchsimdGo127(fn *ast.FuncDecl) {
	env := buildAsmEnv(fn)
	ast.Inspect(fn, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if id, ok := sel.X.(*ast.Ident); ok && notReceiver[id.Name] {
			if id.Name == "archsimd" {
				sel.Sel.Name = archsimdPkgFuncGo127(sel.Sel.Name)
			}
			return true
		}
		if env.isAsm(sel.X) {
			return true
		}
		sel.Sel.Name = archsimdMethodGo127(sel.Sel.Name)
		return true
	})
}
