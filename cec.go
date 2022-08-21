package cec

import (
	"encoding/hex"
	"log"
	"strings"
	"time"
)

// Device structure
type Device struct {
	OSDName         string
	Vendor          string
	LogicalAddress  int
	ActiveSource    bool
	CECVersion      string
	PowerStatus     string
	PhysicalAddress string
	Language        string
}

var logicalNames = []string{"TV", "Recording", "Recording2", "Tuner",
	"Playback", "Audio", "Tuner2", "Tuner3",
	"Playback2", "Recording3", "Tuner4", "Playback3",
	"Reserved", "Reserved2", "Free", "Broadcast"}

var vendorList = map[uint64]string{0x000039: "Toshiba", 0x0000F0: "Samsung",
	0x0005CD: "Denon", 0x000678: "Marantz", 0x000982: "Loewe", 0x0009B0: "Onkyo",
	0x000CB8: "Medion", 0x000CE7: "Toshiba", 0x001582: "Pulse Eight",
	0x0020C7: "Akai", 0x002467: "Aoc", 0x008045: "Panasonic", 0x00903E: "Philips",
	0x009053: "Daewoo", 0x00A0DE: "Yamaha", 0x00D0D5: "Grundig",
	0x00E036: "Pioneer", 0x00E091: "LG", 0x08001F: "Sharp", 0x080046: "Sony",
	0x18C086: "Broadcom", 0x6B746D: "Vizio", 0x8065E9: "Benq",
	0x9C645E: "Harman Kardon"}

var keyList = map[int]string{0x00: "Select", 0x01: "Up", 0x02: "Down", 0x03: "Left",
	0x04: "Right", 0x05: "RightUp", 0x06: "RightDown", 0x07: "LeftUp",
	0x08: "LeftDown", 0x09: "RootMenu", 0x0A: "SetupMenu", 0x0B: "ContentsMenu",
	0x0C: "FavoriteMenu", 0x0D: "Exit", 0x20: "0", 0x21: "1", 0x22: "2", 0x23: "3",
	0x24: "4", 0x25: "5", 0x26: "6", 0x27: "7", 0x28: "8", 0x29: "9", 0x2A: "Dot",
	0x2B: "Enter", 0x2C: "Clear", 0x2F: "NextFavorite", 0x30: "ChannelUp",
	0x31: "ChannelDown", 0x32: "PreviousChannel", 0x33: "SoundSelect",
	0x34: "InputSelect", 0x35: "DisplayInformation", 0x36: "Help",
	0x37: "PageUp", 0x38: "PageDown", 0x40: "Power", 0x41: "VolumeUp",
	0x42: "VolumeDown", 0x43: "Mute", 0x44: "Play", 0x45: "Stop", 0x46: "Pause",
	0x47: "Record", 0x48: "Rewind", 0x49: "FastForward", 0x4A: "Eject",
	0x4B: "Forward", 0x4C: "Backward", 0x4D: "StopRecord", 0x4E: "PauseRecord",
	0x50: "Angle", 0x51: "SubPicture", 0x52: "VideoOnDemand",
	0x53: "ElectronicProgramGuide", 0x54: "TimerProgramming",
	0x55: "InitialConfiguration", 0x60: "PlayFunction", 0x61: "PausePlay",
	0x62: "RecordFunction", 0x63: "PauseRecordFunction",
	0x64: "StopFunction", 0x65: "Mute",
	0x66: "RestoreVolume", 0x67: "Tune", 0x68: "SelectMedia",
	0x69: "SelectAvInput", 0x6A: "SelectAudioInput", 0x6B: "PowerToggle",
	0x6C: "PowerOff", 0x6D: "PowerOn", 0x71: "Blue", 0X72: "Red", 0x73: "Green",
	0x74: "Yellow", 0x75: "F5", 0x76: "Data", 0x91: "AnReturn",
	0x96: "Max"}

// Open - open a new connection to the CEC device with the given name
func Open(name string, deviceName string, printLogs bool) (*Connection, error) {
	c := new(Connection)

	var err error

	c.connection, err = cecInit(deviceName, printLogs)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	adapter, err := getAdapter(c.connection, name)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = openAdapter(c.connection, adapter)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return c, nil
}

// Key - send key press and release commands (hold key for 10ms) to the device
// at the given address, the key code can be specified as a hex-code or by
// its name
func (c *Connection) Key(address int, key interface{}) {
	var keycode int

	switch key := key.(type) {
	case string:
		if key[:2] == "0x" && len(key) == 4 {
			keybytes, err := hex.DecodeString(key[2:])
			if err != nil {
				log.Println(err)
				return
			}
			keycode = int(keybytes[0])
		} else {
			keycode = GetKeyCodeByName(key)
		}
	case int:
		keycode = key
	default:
		log.Println("Invalid key type")
		return
	}
	er := c.KeyPress(address, keycode)
	if er != nil {
		log.Println(er)
		return
	}
	time.Sleep(10 * time.Millisecond)
	er = c.KeyRelease(address)
	if er != nil {
		log.Println(er)
		return
	}
}

// List - list active devices (returns a map of Devices)
func (c *Connection) List() map[string]Device {
	devices := make(map[string]Device)

	activeDevices := c.GetActiveDevices()

	for address, active := range activeDevices {
		if active {
			var dev Device

			dev.LogicalAddress = address
			dev.PhysicalAddress = c.GetDevicePhysicalAddress(address)
			dev.OSDName = c.GetDeviceOSDName(address)
			dev.PowerStatus = c.GetDevicePowerStatus(address)
			dev.ActiveSource = c.IsActiveSource(address)
			dev.CECVersion = c.GetDeviceCecVersion(address)
			dev.Language = c.GetDeviceMenuLanguage(address)
			dev.Vendor = GetVendorByID(c.GetDeviceVendorID(address))

			devices[logicalNames[address]] = dev
		}
	}
	return devices
}

// removeSeparators - remove separators (":", "-", " ", "_")
func removeSeparators(in string) string {
	out := strings.Map(func(r rune) rune {
		if strings.IndexRune(":-_ ", r) < 0 {
			return r
		}
		return -1
	}, in)

	return (out)
}

// GetKeyCodeByName - get the keycode by its name
func GetKeyCodeByName(name string) int {
	name = removeSeparators(name)
	name = strings.ToLower(name)

	for code, value := range keyList {
		if strings.ToLower(value) == name {
			return code
		}
	}

	return -1
}

// GetLogicalAddressByName - get logical address by its name
func GetLogicalAddressByName(name string) int {
	name = removeSeparators(name)
	l := len(name)

	if name[l-1] == '1' {
		name = name[:l-1]
	}

	name = strings.ToLower(name)

	for i := 0; i < 16; i++ {
		if strings.ToLower(logicalNames[i]) == name {
			return i
		}
	}

	if name == "unregistered" {
		return 15
	}

	return -1
}

// GetLogicalNameByAddress - get logical name by address
func GetLogicalNameByAddress(addr int) string {
	return logicalNames[addr]
}

// GetVendorByID - Get vendor by ID
func GetVendorByID(id uint64) string {
	return vendorList[id]
}
