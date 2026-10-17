module github.com/go-simd/highway

go 1.26

require (
	golang.org/x/sys v0.41.0
	golang.org/x/text v0.34.0
	golang.org/x/tools v0.42.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-simd/highway/hwy/goat v0.0.2 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/klauspost/asmfmt v1.3.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/samber/lo v1.50.0 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/mod v0.33.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	modernc.org/cc/v4 v4.26.3 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/opt v0.1.4 // indirect
	modernc.org/sortutil v1.2.1 // indirect
	modernc.org/strutil v1.2.1 // indirect
	modernc.org/token v1.1.0 // indirect
)

replace github.com/go-simd/highway/hwy/goat => ./hwy/goat

tool github.com/go-simd/highway/hwy/goat

// v0.0.1 to v0.0.12 are upstream's tags, inherited when this repository was
// forked from github.com/ajroetker/go-highway: their go.mod declares that path,
// not this one, so none of them can be used as github.com/go-simd/highway. The
// module proxy cached them anyway, and @latest pointed at v0.0.12. Retracted so
// that go get skips them; this fork's first usable version is v0.0.13.
retract [v0.0.1, v0.0.12]
