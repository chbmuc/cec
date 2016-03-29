cec.go - a golang binding for libcec
====================================

`cec.go` is a Go interface to [LibCEC](http://libcec.pulse-eight.com/).

## Install

Make sure you have libcec and it's header files installed (`apt-get install libcec-dev`)

    go get github.com/chbmuc/cec

## Getting Started

A simple example to turn on the TV:

```go
package main

import "github.com/chbmuc/cec"

func main() {
	c := cec.Open("", "cec.go")
	c.PowerOn(0)
}
```
