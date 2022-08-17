package cec

/*
#cgo pkg-config: libcec
//#cgo CFLAGS: -Iinclude
//#cgo LDFLAGS: -lcec
#include <stdio.h>
#include <libcec/cecc.h>

ICECCallbacks g_callbacks;
// callbacks.go exports
void logMessageCallback(void *, const cec_log_message *);

void setupCallbacks(libcec_configuration *conf)
{
	g_callbacks.logMessage = &logMessageCallback;
	g_callbacks.keyPress = NULL;
	g_callbacks.commandReceived = NULL;
	g_callbacks.configurationChanged = NULL;
	g_callbacks.alert = NULL;
	g_callbacks.menuStateChanged = NULL;
	g_callbacks.sourceActivated = NULL;
	(*conf).callbacks = &g_callbacks;
}

void setName(libcec_configuration *conf, char *name)
{
	snprintf((*conf).strDeviceName, 13, "%s", name);
}

*/
import "C"

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"unsafe"
)

// Connection class
type Connection struct {
	connection C.libcec_connection_t
}

type cecAdapter struct {
	Path string
	Comm string
}

func cecInit(deviceName string) (C.libcec_connection_t, error) {
	var connection C.libcec_connection_t
	var conf C.libcec_configuration

	conf.clientVersion = C.uint32_t(C.LIBCEC_VERSION_CURRENT)

	conf.deviceTypes.types[0] = C.CEC_DEVICE_TYPE_RECORDING_DEVICE

	C.setName(&conf, C.CString(deviceName))
	C.setupCallbacks(&conf)

	connection = C.libcec_initialise(&conf)
	if connection == C.libcec_connection_t(nil) {
		return connection, errors.New("Failed to init CEC")
	}
	return connection, nil
}

func getAdapter(connection C.libcec_connection_t, name string) (cecAdapter, error) {
	var adapter cecAdapter

	var deviceList [10]C.cec_adapter
	devicesFound := int(C.libcec_find_adapters(connection, &deviceList[0], 10, nil))

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

func openAdapter(connection C.libcec_connection_t, adapter cecAdapter) error {
	C.libcec_init_video_standalone(connection)

	result := C.libcec_open(connection, C.CString(adapter.Comm), C.CEC_DEFAULT_CONNECT_TIMEOUT)
	if result < 1 {
		return errors.New("Failed to open adapter")
	}

	return nil
}

// Transmit CEC command - command is encoded as a hex string with
// colons (e.g. "40:04")
func (c *Connection) Transmit(command string) {
	var cecCommand C.cec_command

	cmd, err := hex.DecodeString(removeSeparators(command))
	if err != nil {
		log.Fatal(err)
	}
	cmdLen := len(cmd)

	if cmdLen > 0 {
		cecCommand.initiator = C.cec_logical_address((cmd[0] >> 4) & 0xF)
		cecCommand.destination = C.cec_logical_address(cmd[0] & 0xF)
		if cmdLen > 1 {
			cecCommand.opcode_set = 1
			cecCommand.opcode = C.cec_opcode(cmd[1])
		} else {
			cecCommand.opcode_set = 0
		}
		if cmdLen > 2 {
			cecCommand.parameters.size = C.uint8_t(cmdLen - 2)
			for i := 0; i < cmdLen-2; i++ {
				cecCommand.parameters.data[i] = C.uint8_t(cmd[i+2])
			}
		} else {
			cecCommand.parameters.size = 0
		}
	}

	C.libcec_transmit(c.connection, (*C.cec_command)(&cecCommand))
}

// Destroy - destroy the cec connection
func (c *Connection) Destroy() {
	C.libcec_destroy(c.connection)
}

// PowerOn - power on the device with the given logical address
func (c *Connection) PowerOn(address int) error {
	if C.libcec_power_on_devices(c.connection, C.cec_logical_address(address)) != 0 {
		return errors.New("Error in cec_power_on_devices")
	}
	return nil
}

// Standby - put the device with the given address in standby mode
func (c *Connection) Standby(address int) error {
	if C.libcec_standby_devices(c.connection, C.cec_logical_address(address)) != 0 {
		return errors.New("Error in cec_standby_devices")
	}
	return nil
}

// VolumeUp - send a volume up command to the amp if present
func (c *Connection) VolumeUp() error {
	if C.libcec_volume_up(c.connection, 1) != 0 {
		return errors.New("Error in cec_volume_up")
	}
	return nil
}

// VolumeDown - send a volume down command to the amp if present
func (c *Connection) VolumeDown() error {
	if C.libcec_volume_down(c.connection, 1) != 0 {
		return errors.New("Error in cec_volume_down")
	}
	return nil
}

// Mute - send a mute/unmute command to the amp if present
func (c *Connection) Mute() error {
	if C.libcec_mute_audio(c.connection, 1) != 0 {
		return errors.New("Error in cec_mute_audio")
	}
	return nil
}

// KeyPress - send a key press (down) command code to the given address
func (c *Connection) KeyPress(address int, key int) error {
	if C.libcec_send_keypress(c.connection, C.cec_logical_address(address), C.cec_user_control_code(key), 1) != 1 {
		return errors.New("Error in cec_send_keypress")
	}
	return nil
}

// KeyRelease - send a key releas command to the given address
func (c *Connection) KeyRelease(address int) error {
	if C.libcec_send_key_release(c.connection, C.cec_logical_address(address), 1) != 1 {
		return errors.New("Error in cec_send_key_release")
	}
	return nil
}

// GetActiveDevices - returns an array of active devices
func (c *Connection) GetActiveDevices() [16]bool {
	var devices [16]bool
	result := C.libcec_get_active_devices(c.connection)

	for i := 0; i < 16; i++ {
		if int(result.addresses[i]) > 0 {
			devices[i] = true
		}
	}

	return devices
}

// GetDeviceOSDName - get the OSD name of the specified device
func (c *Connection) GetDeviceOSDName(address int) string {
	name := make([]byte, 14)
	C.libcec_get_device_osd_name(c.connection, C.cec_logical_address(address), (*C.char)(unsafe.Pointer(&name[0])))

	return string(name)
}

// IsActiveSource - check if the device at the given address is the active source
func (c *Connection) IsActiveSource(address int) bool {
	result := C.libcec_is_active_source(c.connection, C.cec_logical_address(address))

	if int(result) != 0 {
		return true
	}

	return false
}

// GetDeviceVendorID - Get the Vendor-ID of the device at the given address
func (c *Connection) GetDeviceVendorID(address int) uint64 {
	result := C.libcec_get_device_vendor_id(c.connection, C.cec_logical_address(address))

	return uint64(result)
}

// GetDevicePhysicalAddress - Get the physical address of the device at
// the given logical address
func (c *Connection) GetDevicePhysicalAddress(address int) string {
	result := C.libcec_get_device_physical_address(c.connection, C.cec_logical_address(address))

	return fmt.Sprintf("%x.%x.%x.%x", (uint(result)>>12)&0xf, (uint(result)>>8)&0xf, (uint(result)>>4)&0xf, uint(result)&0xf)
}

// GetDevicePowerStatus - Get the power status of the device at the
// given address
func (c *Connection) GetDevicePowerStatus(address int) string {
	result := C.libcec_get_device_power_status(c.connection, C.cec_logical_address(address))

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

func (c *Connection) GetDeviceCecVersion(address int) string {
	result := int(C.libcec_get_device_cec_version(c.connection, C.cec_logical_address(address)))

	if result == C.CEC_VERSION_1_2 {
		return "1.2"
	} else if result == C.CEC_VERSION_1_2A {
		return "1.2a"
	} else if result == C.CEC_VERSION_1_3 {
		return "1.3"
	} else if result == C.CEC_VERSION_1_3A {
		return "1.3a"
	} else if result == C.CEC_VERSION_1_4 {
		return "1.4"
	} else if result == C.CEC_VERSION_2_0 {
		return "2.0"
	} else if result == C.CEC_VERSION_UNKNOWN {
		return "unknown"
	}

	return ""
}
