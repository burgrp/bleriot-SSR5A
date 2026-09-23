//go:build !tinygo

package main

import (
	"github.com/burgrp/bleriot-SSR5A/fw/spec"
	"github.com/burgrp/bleriot/lib/shared/config"
	"github.com/burgrp/bleriot/lib/shared/inventory"
	"github.com/burgrp/bleriot/lib/site/cli"
)

var (
	far          = inventory.Channel{Name: "far", Number: 37, SpreadFactor: config.SpreadFactorS8}
	deviceConfig = spec.Config{
		Channels: [spec.ChannelCount]spec.ChannelConfig{
			{Disabled: false, Inverted: false, Default: false},
			{Disabled: false, Inverted: false, Default: false},
			{Disabled: false, Inverted: false, Default: false},
			{Disabled: false, Inverted: false, Default: false},
			{Disabled: false, Inverted: false, Default: false},
		}}
)

func main() {
	cli.Start(inventory.Inventory{
		{
			Name:    "ssr",
			Address: [4]byte{0xBF, 0xB9, 0xB4, 0x0B},
			Key:     [16]byte{0xF6, 0x96, 0xA8, 0x35, 0x84, 0xE3, 0x64, 0xEC, 0xF5, 0x79, 0x5F, 0x38, 0x34, 0x4E, 0x31, 0xFC},
			Channel: far,
			Type:    spec.Type(deviceConfig),
			Config:  deviceConfig,
		},
	})
}
