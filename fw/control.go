package main

import "github.com/burgrp/bleriot-SSR5A/fw/spec"

const startupLEDGreenRGB = int32(0x008000)

type controlState struct {
	config spec.Config
	values [spec.ChannelCount]int32
}

func newControl(config spec.Config) controlState {
	state := controlState{config: config}
	for index, channel := range config.Channels {
		if channel.Default {
			state.values[index] = 1
		}
	}
	return state
}

func (state *controlState) read(tag uint16) (value int32, null bool) {
	index, valid := channelIndex(tag)
	if !valid || state.config.Channels[index].Disabled {
		return 0, true
	}
	return state.values[index], false
}

func (state *controlState) write(tag uint16, value int32, null bool) (index int, changed bool) {
	index, valid := channelIndex(tag)
	if null || !valid || state.config.Channels[index].Disabled {
		return 0, false
	}

	value = normalizeBool(value)
	if state.values[index] == value {
		return index, false
	}
	state.values[index] = value
	return index, true
}

func (state *controlState) pinHigh(index int) bool {
	channel := state.config.Channels[index]
	if channel.Disabled {
		return true
	}
	energized := state.values[index] != 0
	if channel.Inverted {
		energized = !energized
	}
	return !energized
}

func channelIndex(tag uint16) (int, bool) {
	if tag < spec.RegChannel1 || tag > spec.RegChannel5 {
		return 0, false
	}
	return int(tag - spec.RegChannel1), true
}

func normalizeBool(value int32) int32 {
	if value == 0 {
		return 0
	}
	return 1
}

func rgbBytes(value int32) [3]byte {
	return [3]byte{
		byte(value >> 8),
		byte(value >> 16),
		byte(value),
	}
}
