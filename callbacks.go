package cec

// #include <libcec/cecc.h>
import "C"

import (
        "log"
        "unsafe"
)

//export logMessageCallback
func logMessageCallback(c unsafe.Pointer, msg C.cec_log_message) C.int {
	log.Println(C.GoString(&msg.message[0]))

        return 0;
}
