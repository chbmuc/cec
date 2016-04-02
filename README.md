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
	"flag"
	
	"github.com/chbmuc/cec"
)

func main() {
	flag.Parse()
	cec.Open("", "cec.go")
	cec.PowerOn(cec.TV)
}
```

To see log output from this package, add the '--logtostderr' flag when executing your binary.
To see debug output from libCEC, add the '--logtostderr' and '--v=2' flags when executing your binary.
