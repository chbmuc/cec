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

import (
	"fmt"
	"github.com/chbmuc/cec"
)

func main() {
	c, err := cec.Open("", "cec.go", true)
	if err != nil {
		fmt.Println(err)
	}
	c.PowerOn(0)
}
```
