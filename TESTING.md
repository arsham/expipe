# About

In this document I show you how to run tests and benchmarks.

## Testing

```bash
go test $(glide nv)
```

## Coverage

Use this [gist](https://gist.github.com/arsham/f45f7e7eea7e18796bc1ed5ced9f9f4a). Then run:

```bash
goverall
```

## Benchmarks

To run all benchmarks:

```bash
go test $(glide nv) -run=^$ -bench=.
```

For showing the memory and cpu profiles, on each folder run:

```bash
BASENAME=$(basename $(pwd))
go test -run=^$ -bench=. -cpuprofile=cpu.out -benchmem -memprofile=mem.out
go tool pprof -pdf $BASENAME.test cpu.out > cpu.pdf && open cpu.pdf
go tool pprof -pdf $BASENAME.test mem.out > mem.pdf && open mem.pdf
```
