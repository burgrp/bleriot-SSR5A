//go:build tinygo

package main

import (
	"machine"
	"time"

	"github.com/burgrp/bleriot-SSR5A/fw/spec"
	"github.com/burgrp/bleriot/lib/node"
	"github.com/burgrp/bleriot/lib/node/pan211x"
	"github.com/burgrp/tinygo-drivers/ws2812"
)

const (
	pinSmartLED = machine.PA6

	pinSpiCsn  = machine.PA4
	pinSpiSck  = machine.PA5
	pinSpiData = machine.PA7
)

var channelPins = [spec.ChannelCount]machine.Pin{
	machine.PA11,
	machine.PA12,
	machine.PA15,
	machine.PB3,
	machine.PB4,
}

type Device struct {
	control     controlState
	led         ws2812.Device
	ledBytes    [3]byte
	online      bool
	statusDirty bool
}

func bleriotMain(provisioning node.Provisioning, config spec.Config) {
	device := newDevice(config)

	bleNode, err := pan211x.StartNode(provisioning, pinSpiSck, pinSpiData, pinSpiCsn, device)
	if err != nil {
		halt("failed to start BleRiot node: " + err.Error())
	}

	var link linkState
	for {
		online, led, changed := bleNode.PollWithStatus()
		device.online = online
		if link.update(online) {
			device.resetDefaults()
		}
		if changed {
			device.statusDirty = true
		}
		if device.statusDirty {
			if err := device.setStatusLED(led); err == nil {
				device.statusDirty = false
			}
		}
	}
}

func newDevice(config spec.Config) *Device {
	for _, pin := range channelPins {
		configureActiveLowOutput(pin)
	}

	device := &Device{control: newControl(config), statusDirty: true}
	for index, pin := range channelPins {
		pin.Set(device.control.pinHigh(index))
	}

	pinSmartLED.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	pinSmartLED.Low()
	pinSmartLED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	device.led = ws2812.NewWS2812(pinSmartLED)
	return device
}

func configureActiveLowOutput(pin machine.Pin) {
	pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	pin.High()
	pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
}

func (device *Device) Read(tag uint16) (value int32, null bool) {
	return device.control.read(tag)
}

func (device *Device) Write(tag uint16, value int32, null bool) {
	index, changed := device.control.write(tag, value, null)
	if changed {
		channelPins[index].Set(device.control.pinHigh(index))
		device.statusDirty = true
	}
}

func (device *Device) resetDefaults() {
	device.control.resetDefaults()
	for index, pin := range channelPins {
		pin.Set(device.control.pinHigh(index))
	}
	device.statusDirty = true
}

func (device *Device) setStatusLED(offlineLED bool) error {
	return device.setLED(statusLEDColor(device.online, offlineLED, device.control.anyActive()))
}

func (device *Device) setLED(value int32) error {
	device.ledBytes = rgbBytes(value)
	_, err := device.led.Write(device.ledBytes[:])
	return err
}

func halt(message string) {
	for _, pin := range channelPins {
		pin.High()
	}
	pinSmartLED.Low()
	println(message)
	for {
		time.Sleep(time.Second)
	}
}
