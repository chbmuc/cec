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
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
)

type CECConfiguration struct {
	DeviceName string
}

type CECAdapter struct {
	Path string
	Comm string
}

var conn C.libcec_connection_t

func cecInit(config CECConfiguration) error {
	var conf C.libcec_configuration

	//TODO(bminor13): These constants aren't present in cecc.h - are they necessary?
	//conf.clientVersion = C.uint32_t(C.CEC_CLIENT_VERSION_CURRENT)
	//conf.serverVersion = C.uint32_t(C.CEC_SERVER_VERSION_CURRENT)

	for i := 0; i < 5; i++ {
		conf.deviceTypes.types[i] = C.CEC_DEVICE_TYPE_RESERVED
	}
	conf.deviceTypes.types[0] = C.CEC_DEVICE_TYPE_RECORDING_DEVICE

	C.setName(&conf, C.CString(config.DeviceName))
	C.setupCallbacks(&conf)

	conn = C.libcec_initialise(&conf)
	// TODO(bminor13): Need the correct error check here.
	// if int(conn) < 1 {
	// 	conn = 0
	// 	return errors.New("Failed to init CEC")
	// }
	return nil
}

func getAdapter(name string) (CECAdapter, error) {
	var adapter CECAdapter

	var deviceList [10]C.cec_adapter
	devicesFound := int(C.libcec_find_adapters(conn, &deviceList[0], 10, nil))

	for i := 0; i < devicesFound; i++ {
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
	C.libcec_init_video_standalone(conn)

	result := C.libcec_open(conn, C.CString(adapter.Comm), C.CEC_DEFAULT_CONNECT_TIMEOUT)
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

	if cmd_len > 0 {
		cec_command.initiator = C.cec_logical_address((cmd[0] >> 4) & 0xF)
		cec_command.destination = C.cec_logical_address(cmd[0] & 0xF)
		if cmd_len > 1 {
			cec_command.opcode_set = 1
			cec_command.opcode = C.cec_opcode(cmd[1])
		} else {
			cec_command.opcode_set = 0
		}
		if cmd_len > 2 {
			cec_command.parameters.size = C.uint8_t(cmd_len - 2)
			for i := 0; i < cmd_len-2; i++ {
				cec_command.parameters.data[i] = C.uint8_t(cmd[i+2])
			}
		} else {
			cec_command.parameters.size = 0
		}
	}

	C.libcec_transmit(conn, (*C.cec_command)(&cec_command))
}

func Destroy() {
	C.libcec_destroy(conn)
	// TODO(bminor13): Clear conn here?
}

func PowerOn(address int) error {
	if C.libcec_power_on_devices(conn, C.cec_logical_address(address)) != 0 {
		return errors.New("Error in cec_power_on_devices")
	}
	return nil
}

func Standby(address int) error {
	if C.libcec_standby_devices(conn, C.cec_logical_address(address)) != 0 {
		return errors.New("Error in libcec_standby_devices")
	}
	return nil
}

func VolumeUp() error {
	if C.libcec_volume_up(conn, 1) != 0 {
		return errors.New("Error in libcec_volume_up")
	}
	return nil
}

func VolumeDown() error {
	if C.libcec_volume_down(conn, 1) != 0 {
		return errors.New("Error in libcec_volume_down")
	}
	return nil
}

func Mute() error {
	if C.libcec_mute_audio(conn, 1) != 0 {
		return errors.New("Error in libcec_mute_audio")
	}
	return nil
}

func KeyPress(address int, key int) error {
	if C.libcec_send_keypress(conn, C.cec_logical_address(address), C.cec_user_control_code(key), 1) != 1 {
		return errors.New("Error in libcec_send_keypress")
	}
	return nil
}

func KeyRelease(address int) error {
	if C.libcec_send_key_release(conn, C.cec_logical_address(address), 1) != 1 {
		return errors.New("Error in libcec_send_key_release")
	}
	return nil
}

func GetActiveDevices() [16]bool {
	var devices [16]bool
	result := C.libcec_get_active_devices(conn)

	for i := 0; i < 16; i++ {
		if int(result.addresses[i]) > 0 {
			devices[i] = true
		}
	}

	return devices
}

func GetDeviceOSDName(address int) string {
	result := C.libcec_get_device_osd_name(conn, C.cec_logical_address(address))

	return C.GoString(&result.name[0])
}

func IsActiveSource(address int) bool {
	result := C.libcec_is_active_source(conn, C.cec_logical_address(address))

	if int(result) != 0 {
		return true
	} else {
		return false
	}
}

func GetDeviceVendorId(address int) uint64 {
	result := C.libcec_get_device_vendor_id(conn, C.cec_logical_address(address))

	return uint64(result)
}

func GetDevicePhysicalAddress(address int) string {
	result := C.libcec_get_device_physical_address(conn, C.cec_logical_address(address))

	return fmt.Sprintf("%x.%x.%x.%x", (uint(result)>>12)&0xf, (uint(result)>>8)&0xf, (uint(result)>>4)&0xf, uint(result)&0xf)
}

func GetDevicePowerStatus(address int) string {
	result := C.libcec_get_device_power_status(conn, C.cec_logical_address(address))

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
