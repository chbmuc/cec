package cec

// #include <libcec/cecc.h>
import "C"

import (
	"unsafe"

	"github.com/golang/glog"
)

//export logMessageCallback
func logMessageCallback(c unsafe.Pointer, msg C.cec_log_message) C.int {
	glog.V(2).Info(C.GoString(&msg.message[0]))
	return 0
}
