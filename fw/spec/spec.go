package spec

import (
	"github.com/burgrp/bleriot/lib/shared/inventory"
	"github.com/burgrp/bleriot/lib/shared/puya"
)

const ChannelCount = 5

type ChannelConfig struct {
	Disabled bool
	Inverted bool
	Default  bool
}

type Config struct {
	Channels [ChannelCount]ChannelConfig
}

const (
	RegChannel1 uint16 = iota + 1
	RegChannel2
	RegChannel3
	RegChannel4
	RegChannel5
)

var (
	Chip         = puya.PY32F030x8
	channelNames = [ChannelCount]string{
		"channel.1",
		"channel.2",
		"channel.3",
		"channel.4",
		"channel.5",
	}
)

func Type(config Config) inventory.DeviceType {
	deviceType := inventory.DeviceType{
		Name:      "ssr5a",
		Chip:      Chip,
		Registers: make([]inventory.Register, 0, ChannelCount),
	}
	for index, channel := range config.Channels {
		if channel.Disabled {
			continue
		}
		deviceType.Registers = append(deviceType.Registers, inventory.Register{
			Tag:  uint16(index + 1),
			Name: channelNames[index],
			Type: inventory.TypeBool,
		})
	}
	return deviceType
}
