package main

import (
	"testing"

	"github.com/burgrp/bleriot-SSR5A/fw/spec"
)

func TestDefaultsAndInversionDetermineStartupOutputs(t *testing.T) {
	config := spec.Config{Channels: [spec.ChannelCount]spec.ChannelConfig{
		{},
		{Default: true},
		{Inverted: true},
		{Inverted: true, Default: true},
		{Disabled: true, Inverted: true, Default: true},
	}}
	state := newControl(config)
	wantPinHigh := [spec.ChannelCount]bool{true, false, false, true, true}

	for index, want := range wantPinHigh {
		if got := state.pinHigh(index); got != want {
			t.Errorf("channel %d pin high = %v, want %v", index+1, got, want)
		}
	}
	for index := range spec.ChannelCount - 1 {
		want := int32(0)
		if config.Channels[index].Default {
			want = 1
		}
		if value, null := state.read(uint16(index + 1)); null || value != want {
			t.Errorf("channel %d read = (%d, %v), want (%d, false)", index+1, value, null, want)
		}
	}
	if value, null := state.read(spec.RegChannel5); !null || value != 0 {
		t.Errorf("disabled channel read = (%d, %v), want (0, true)", value, null)
	}
}

func TestWritesNormalizeValuesAndDriveConfiguredPolarity(t *testing.T) {
	config := spec.Config{Channels: [spec.ChannelCount]spec.ChannelConfig{
		{},
		{Inverted: true},
	}}
	state := newControl(config)

	if index, changed := state.write(spec.RegChannel1, 42, false); !changed || index != 0 {
		t.Fatalf("channel 1 write = (%d, %v), want (0, true)", index, changed)
	}
	if value, null := state.read(spec.RegChannel1); null || value != 1 {
		t.Fatalf("channel 1 read = (%d, %v), want (1, false)", value, null)
	}
	if state.pinHigh(0) {
		t.Fatal("normal channel pin is high after writing true")
	}

	state.write(spec.RegChannel2, 1, false)
	if !state.pinHigh(1) {
		t.Fatal("inverted channel pin is low after writing true")
	}
}

func TestDisabledNullAndUnknownWritesAreIgnored(t *testing.T) {
	config := spec.Config{Channels: [spec.ChannelCount]spec.ChannelConfig{
		{Disabled: true},
	}}
	state := newControl(config)

	for _, write := range []struct {
		tag   uint16
		value int32
		null  bool
	}{
		{spec.RegChannel1, 1, false},
		{spec.RegChannel2, 1, true},
		{99, 1, false},
	} {
		if _, changed := state.write(write.tag, write.value, write.null); changed {
			t.Errorf("write tag %d unexpectedly changed state", write.tag)
		}
	}
	if !state.pinHigh(0) || !state.pinHigh(1) {
		t.Fatal("ignored write energized an output")
	}
}

func TestStartupLEDIsHalfGreen(t *testing.T) {
	if got, want := rgbBytes(startupLEDGreenRGB), [3]byte{0x80, 0x00, 0x00}; got != want {
		t.Fatalf("startup LED bytes = % X, want GRB % X", got, want)
	}
}
