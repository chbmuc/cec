package cec

import(
	"log"
	"time"
	"strings"
)

type Device struct {
	OSDName string
	Vendor string
	LogicalAddress int
	ActiveSource bool
	PowerStatus string
	PhysicalAddress string
}

var logicalNames = []string{ "TV", "Recording", "Recording2", "Tuner",
	"Playback","Audio", "Tuner2", "Tuner3",
	"Playback2", "Recording3", "Tuner4", "Playback3",
	"Reserved", "Reserved2", "Free", "Broadcast" }

var vendorList = map[uint64]string{ 0x000039:"Toshiba", 0x0000F0:"Samsung",
	0x0005CD:"Denon", 0x000678:"Marantz", 0x000982:"Loewe", 0x0009B0:"Onkyo",
	0x000CB8:"Medion", 0x000CE7:"Toshiba", 0x001582:"Pulse Eight",
	0x0020C7:"Akai", 0x002467:"Aoc", 0x008045:"Panasonic", 0x00903E:"Philips",
	0x009053:"Daewoo", 0x00A0DE:"Yamaha", 0x00D0D5:"Grundig",
	0x00E036:"Pioneer", 0x00E091:"LG", 0x08001F:"Sharp", 0x080046:"Sony",
	0x18C086:"Broadcom", 0x6B746D:"Vizio", 0x8065E9:"Benq",
	0x9C645E:"Harman Kardon" }

func Open(name string, deviceName string) {
	var config CECConfiguration
	config.DeviceName = deviceName

	if er := cecInit(config); er != nil {
		log.Println(er)
		return	
	}

	adapter, er := getAdapter(name)
	if er != nil {
		log.Println(er)
		return
	}

	er = openAdapter(adapter)
	if er != nil {
		log.Println(er)
		return
	}
}

func Key(address int, key int) {
	er := KeyPress(address, key)
	if er != nil {
		log.Println(er)
		return
	}
	time.Sleep(10 * time.Millisecond)
	er = KeyRelease(address)
	if er != nil {
		log.Println(er)
		return
	}
}

func List() map[string]Device {
	devices := make(map[string]Device)

	active_devices := GetActiveDevices()

	for address, active := range active_devices {
		if (active) {
			var dev Device

			dev.LogicalAddress = address
			dev.PhysicalAddress = GetDevicePhysicalAddress(address)
			dev.OSDName = GetDeviceOSDName(address)
			dev.PowerStatus = GetDevicePowerStatus(address)
			dev.ActiveSource = IsActiveSource(address)
			dev.Vendor = GetVendorById(GetDeviceVendorId(address))

			devices[logicalNames[address]] = dev
		}
	}
	return devices
}

func removeSeparators(in string) string {
        // remove separators (":", "-", " ", "_")
        out := strings.Map(func(r rune) rune {
                if strings.IndexRune(":-_ ", r) < 0 {
                        return r
                }
                return -1
        }, in)

	return(out)
}

func GetLogicalAddressByName(name string) int {
	name = removeSeparators(name)
	l := len(name)

	if name[l-1] == '1' {
		name = name[:l-1]
	}

	for i:=0; i<16; i++ {
		if logicalNames[i] == name {
			return i
		}
	}

	if name == "Unregistered" {
		return 15
	}

	return -1
}

func GetLogicalNameByAddress(addr int) string {
	return logicalNames[addr]
}

func GetVendorById(id uint64) string {
	return vendorList[id]
}
