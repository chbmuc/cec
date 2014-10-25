package cec

/* 
#cgo pkg-config: libcec 
#include <stdio.h>
#include <libcec/cecc.h>

ICECCallbacks g_callbacks;
// callbacks.go exports
int logMessageCallback(void *, const cec_log_message);

void setupCallbacks(libcec_configuration *conf)
{
	g_callbacks.CBCecLogMessage = &logMessageCallback;
	(*conf).callbacks = &g_callbacks;
}

void setName(libcec_configuration *conf, char *name)
{
	snprintf((*conf).strDeviceName, 13, "%s", name);
}

static void clearLogicalAddresses(cec_logical_addresses* addresses)
{
	int i;

	addresses->primary = CECDEVICE_UNREGISTERED;
	for (i = 0; i < 16; i++)
		addresses->addresses[i] = 0;
}

void setLogicalAddress(cec_logical_addresses* addresses, cec_logical_address address)
{
	if (addresses->primary == CECDEVICE_UNREGISTERED)
		addresses->primary = address;

	addresses->addresses[(int) address] = 1;
}

*/ 
import "C"

import (
	"errors"
	"encoding/hex"
	"strings"
	"log"
	"fmt"
)

type CECConfiguration struct { 
	DeviceName string 
} 

 
type CECAdapter struct { 
	Path string 
	Comm string 
} 

func cecInit(config CECConfiguration) error { 
	var conf C.libcec_configuration 

	conf.clientVersion = C.uint32_t(C.CEC_CLIENT_VERSION_CURRENT)
	conf.serverVersion = C.uint32_t(C.CEC_SERVER_VERSION_CURRENT)

	for i:=0; i<5; i++ {
		conf.deviceTypes.types[i] = C.CEC_DEVICE_TYPE_RESERVED
        }
	conf.deviceTypes.types[0] = C.CEC_DEVICE_TYPE_RECORDING_DEVICE

	C.setName(&conf, C.CString(config.DeviceName))
	C.setupCallbacks(&conf) 

	result := C.cec_initialise(&conf) 
	if result < 1 { 
		return errors.New("Failed to init CEC") 
	}
	return nil 
} 

func getAdapter(name string) (CECAdapter, error) { 
	var adapter CECAdapter 

	var deviceList [10]C.cec_adapter  
	devicesFound := int(C.cec_find_adapters(&deviceList[0], 10, nil))

	for i:=0; i < devicesFound; i++ {
		device := deviceList[i] 
		adapter.Path = C.GoStringN(&device.path[0], 1024) 
		adapter.Comm = C.GoStringN(&device.comm[0], 1024) 

		if strings.Contains(adapter.Path, name) || strings.Contains(adapter.Comm, name) {
			return adapter, nil 
		}
	}

	return adapter, errors.New("No Device Found") 
}

func openAdapter(adapter CECAdapter) error { 
        C.cec_init_video_standalone()

	result := C.cec_open(C.CString(adapter.Comm), C.CEC_DEFAULT_CONNECT_TIMEOUT) 
	if result < 1 { 
		return errors.New("Failed to open adapter") 
	} 

	return nil 
} 

func Transmit(command string) {
        var cec_command C.cec_command

        cmd, err := hex.DecodeString(removeSeparators(command))
        if err != nil {
                log.Fatal(err)
        }
        cmd_len := len(cmd)

        if (cmd_len > 0) {
                cec_command.initiator = C.cec_logical_address((cmd[0] >> 4) & 0xF)
                cec_command.destination = C.cec_logical_address(cmd[0] & 0xF)
                if (cmd_len > 1) {
                        cec_command.opcode_set = 1
                        cec_command.opcode = C.cec_opcode(cmd[1])
                } else {
                        cec_command.opcode_set = 0
                }
                if (cmd_len > 2) {
                        cec_command.parameters.size = C.uint8_t(cmd_len-2)
                        for i := 0; i < cmd_len-2; i++ {
                                cec_command.parameters.data[i] = C.uint8_t(cmd[i+2])
                        }
                } else {
                        cec_command.parameters.size = 0
                }
        }

        C.cec_transmit((*C.cec_command)(&cec_command))
}

func Destroy() {
	C.cec_destroy()
}

func PowerOn(address int) error {
	if C.cec_power_on_devices(C.cec_logical_address(address)) != 0 {
		return errors.New("Error in cec_power_on_devices")
	}
	return nil
}

func Standby(address int) error {
	if C.cec_standby_devices(C.cec_logical_address(address)) != 0 {
		return errors.New("Error in cec_standby_devices")
	}
	return nil
}

func VolumeUp() error {
	if C.cec_volume_up(1) != 0 {
		return errors.New("Error in cec_volume_up")
	}
	return nil
}

func VolumeDown() error {
	if C.cec_volume_down(1) != 0 {
		return errors.New("Error in cec_volume_down")
	}
	return nil
}

func Mute() error {
	if C.cec_mute_audio(1) != 0 {
		return errors.New("Error in cec_mute_audio")
	}
	return nil
}

func KeyPress(address int, key int) error {
	if C.cec_send_keypress(C.cec_logical_address(address), C.cec_user_control_code(key), 1) != 1 {
		return errors.New("Error in cec_send_keypress")
	}
	return nil
}

func KeyRelease(address int) error {
	if C.cec_send_key_release(C.cec_logical_address(address), 1) != 1 {
		return errors.New("Error in cec_send_key_release")
	}
	return nil
}

func GetActiveDevices() [16]bool {
	var devices [16]bool
	result := C.cec_get_active_devices()

	for i:=0; i < 16; i++ {
		if int(result.addresses[i]) > 0 {
			devices[i] = true
		}
	}

	return devices
}

func GetDeviceOSDName(address int) string {
	result := C.cec_get_device_osd_name(C.cec_logical_address(address))

	return C.GoString(&result.name[0])
}

func IsActiveSource(address int) bool {
	result := C.cec_is_active_source(C.cec_logical_address(address))

	if int(result) != 0 {
		return true
	} else {
		return false
	}
}

func GetDeviceVendorId(address int) uint64 {
	result := C.cec_get_device_vendor_id(C.cec_logical_address(address))

	return uint64(result)
}

func GetDevicePhysicalAddress(address int) string {
	result := C.cec_get_device_physical_address(C.cec_logical_address(address))

	return fmt.Sprintf("%x.%x.%x.%x", (uint(result) >> 12) & 0xf, (uint(result) >> 8) & 0xf, (uint(result) >> 4) & 0xf, uint(result) & 0xf)
}

func GetDevicePowerStatus(address int) string {
	result := C.cec_get_device_power_status(C.cec_logical_address(address))

	// C.CEC_POWER_STATUS_UNKNOWN == error

	if int(result) == C.CEC_POWER_STATUS_ON {
		return "on"
	} else if int(result) == C.CEC_POWER_STATUS_STANDBY {
		return "standby"
	} else if int(result) == C.CEC_POWER_STATUS_IN_TRANSITION_STANDBY_TO_ON {
		return "starting"
	} else if int(result) == C.CEC_POWER_STATUS_IN_TRANSITION_ON_TO_STANDBY {
		return "shutting down"
	} else {
		return ""
	}
}
