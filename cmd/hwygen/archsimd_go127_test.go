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

// TestPromotedVectorTracking pins each way a value is recognised as a promoted
// (hwy/asm) vector, whose StoreSlice must keep its name, against an archsimd
// vector, whose StoreSlice becomes Store.
func TestPromotedVectorTracking(t *testing.T) {
	const src = `func f(h asm.Float16x8AVX2, y, dst []uint16, out []float32) {
	a := h.Add(y)
	a.StoreSlice(dst)
	b := asm.LoadFloat16x8Ptr(p)
	b.StoreSlice(dst)
	c := asm.Load[uint16](p)
	c.StoreSlice(dst)
	d := g(h)
	d.StoreSlice(dst)
	e := archsimd.Foo(h)
	e.StoreSlice(out)
	k := local(y)
	k.StoreSlice(out)
	m := mix(y, h)
	m.StoreSlice(dst)
	n := q.Add(y)
	n.StoreSlice(out)
}`
	fn := parseFunc(t, src)
	rewriteArchsimdGo127(fn)
	got := printFunc(t, fn)
	for _, want := range []string{
		"a.StoreSlice(dst)", // method on a promoted receiver, plain argument
		"b.StoreSlice(dst)", // asm package function, plain argument
		"c.StoreSlice(dst)", // generic asm function (not a selector call)
		"d.StoreSlice(dst)", // plain call with a promoted argument
		"e.Store(out)",      // archsimd function never returns a promoted vector
		"k.Store(out)",      // plain call with no promoted argument
		"m.StoreSlice(dst)", // promoted argument not first
		"n.Store(out)",      // unknown receiver
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

// TestRewriteToleratesIncompleteDeclarations covers a declaration without a
// body (an assembly stub) and function types with no parameter list.
func TestRewriteToleratesIncompleteDeclarations(t *testing.T) {
	stub := parseFunc(t, "func stub(h asm.Float16x8AVX2)")
	if stub.Body != nil {
		t.Fatal("expected a bodiless declaration")
	}
	rewriteArchsimdGo127(stub)

	// go/ast cannot walk a function type without a parameter list, so the
	// environment builder is exercised directly.
	if env := buildAsmEnv(&ast.FuncDecl{Name: ast.NewIdent("f"), Type: &ast.FuncType{}, Body: &ast.BlockStmt{}}); len(env) != 0 {
		t.Errorf("unexpected promoted names: %v", env)
	}
	if env := buildAsmEnv(&ast.FuncDecl{Name: ast.NewIdent("f"), Body: &ast.BlockStmt{}}); len(env) != 0 {
		t.Errorf("unexpected promoted names: %v", env)
	}
}

// TestNeonEmitsOneFile checks that a target that is not archsimd gets a single
// file, with no go1.27 constraint and no twin.
func TestNeonEmitsOneFile(t *testing.T) {
	fn := parseFunc(t, `func BaseF(in, out []float32) {
	x := archsimd.LoadFloat32x8Slice(in)
	x.StoreSlice(out)
}`)
	dir := t.TempDir()
	if err := EmitTarget([]*ast.FuncDecl{fn}, NEONTarget(), "p", "f", dir, ContribPackages{HwyCore: true}, nil, nil); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("want one file, got %d", len(entries))
	}
	b, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(entries[0].Name(), "go126") || strings.Contains(string(b), "go1.27") {
		t.Errorf("non-archsimd target was split by toolchain: %s", entries[0].Name())
	}
	if !strings.Contains(string(b), "StoreSlice") {
		t.Errorf("non-archsimd names were rewritten:\n%s", b)
	}
}

// TestEmitTargetReportsTheFirstWriteError makes the go1.26 file impossible to
// write and requires the error to come back, with no go1.27 file after it.
func TestEmitTargetReportsTheFirstWriteError(t *testing.T) {
	fn := parseFunc(t, `func BaseF(in []float32) {
	_ = archsimd.LoadFloat32x8Slice(in)
}`)
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "f_avx2_go126.gen.go"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := EmitTarget([]*ast.FuncDecl{fn}, AVX2Target(), "p", "f", dir, ContribPackages{HwyCore: true}, nil, nil); err == nil {
		t.Fatal("the failed go1.26 write was not reported")
	}
	if _, err := os.Stat(filepath.Join(dir, "f_avx2.gen.go")); err == nil {
		t.Error("the go1.27 file was written after the go1.26 one failed")
	}
}

func TestNeonTargetIsNotRewritten(t *testing.T) {
	if NEONTarget().VecPackage == "archsimd" {
		t.Fatal("NEON must not be an archsimd target")
	}
}
