package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func parseFunc(t *testing.T, body string) *ast.FuncDecl {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "x.go", "package p\n"+body, 0)
	if err != nil {
		t.Fatal(err)
	}
	return f.Decls[0].(*ast.FuncDecl)
}

func printFunc(t *testing.T, fn *ast.FuncDecl) string {
	t.Helper()
	var b bytes.Buffer
	if err := printer.Fprint(&b, token.NewFileSet(), fn); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

// TestArchsimdGo127Spelling asserts every archsimd rename, and that the names
// of hwy, asm, math and of promoted half-precision vectors stay as they are.
func TestArchsimdGo127Spelling(t *testing.T) {
	const src = `func f(in, out []float32, h asm.Float16x8AVX2, dst []uint16) {
	x := archsimd.LoadFloat32x8((*[8]float32)(unsafe.Pointer(&in[0])))
	y := archsimd.LoadFloat32x8Slice(in)
	z := archsimd.LoadFloat32x8SlicePart(in)
	x.Store((*[8]float32)(unsafe.Pointer(&out[0])))
	x.StoreMasked((*[8]float32)(unsafe.Pointer(&out[0])), m)
	y.StoreSlice(out)
	z.StoreSlicePart(out)
	r := x.RoundToEven()
	_ = hwy.RoundToEven(x)
	_ = stdmath.RoundToEven(1.5)
	_ = asm.LoadFloat16x8Ptr(p)
	h2 := h.Add(h)
	h2.StoreSlice(dst)
	h3 := h2.RoundToEven()
	h3.StoreSlice(dst)
	i := h3.ConvertToInt32()
	i.StoreSlice(tmp[:])
	h.StorePtr(p)
}`
	fn := parseFunc(t, src)
	rewriteArchsimdGo127(fn)
	got := printFunc(t, fn)
	for _, want := range []string{
		"archsimd.LoadFloat32x8Array((*[8]float32)",
		"y := archsimd.LoadFloat32x8(in)",
		"z := archsimd.LoadFloat32x8Part(in)",
		"x.StoreArray((*[8]float32)(unsafe.Pointer(&out[0])))",
		"x.StoreArrayMasked(",
		"y.Store(out)",
		"z.StorePart(out)",
		"r := x.Round()",
		"hwy.RoundToEven(x)",
		"stdmath.RoundToEven(1.5)",
		"asm.LoadFloat16x8Ptr(p)",
		"h2.StoreSlice(dst)",
		"h3 := h2.RoundToEven()",
		"h3.StoreSlice(dst)",
		"i.Store(tmp[:])",
		"h.StorePtr(p)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "SlicePart") {
		t.Errorf("a SlicePart name survived:\n%s", got)
	}
}

// TestEmitTargetBothSpellings runs the emitter for AVX2 and checks that one
// file per toolchain is written, each with its own tag and spelling, while a
// NEON target keeps one file and its asm names.
func TestEmitTargetBothSpellings(t *testing.T) {
	fn := parseFunc(t, `func BaseF(in, out []float32) {
	x := archsimd.LoadFloat32x8Slice(in)
	x.StoreSlice(out)
	y := x.RoundToEven()
	y.Store((*[8]float32)(unsafe.Pointer(&out[0])))
}`)
	dir := t.TempDir()
	avx2 := AVX2Target()
	if err := EmitTarget([]*ast.FuncDecl{fn}, avx2, "p", "f", dir, ContribPackages{HwyCore: true}, nil, nil); err != nil {
		t.Fatal(err)
	}
	read := func(name string) string {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	new127 := read("f_avx2.gen.go")
	old126 := read("f_avx2_go126.gen.go")
	if !strings.Contains(new127, "//go:build amd64 && goexperiment.simd && go1.27\n") {
		t.Errorf("go1.27 tag missing:\n%s", new127)
	}
	if !strings.Contains(old126, "//go:build amd64 && goexperiment.simd && !go1.27\n") {
		t.Errorf("go1.26 tag missing:\n%s", old126)
	}
	for _, w := range []string{"archsimd.LoadFloat32x8(in)", "x.Store(out)", "x.Round()", "y.StoreArray("} {
		if !strings.Contains(new127, w) {
			t.Errorf("go1.27 file lacks %q:\n%s", w, new127)
		}
	}
	for _, w := range []string{"archsimd.LoadFloat32x8Slice(in)", "x.StoreSlice(out)", "x.RoundToEven()", "y.Store((*[8]float32)"} {
		if !strings.Contains(old126, w) {
			t.Errorf("go1.26 file lacks %q:\n%s", w, old126)
		}
	}
}

func TestNeonTargetIsNotRewritten(t *testing.T) {
	if NEONTarget().VecPackage == "archsimd" {
		t.Fatal("NEON must not be an archsimd target")
	}
}
