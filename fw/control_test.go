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

func TestLinkStateReportsOnlyOnlineToOfflineTransitions(t *testing.T) {
	var state linkState
	for index, update := range []struct {
		online      bool
		wentOffline bool
	}{
		{online: false, wentOffline: false},
		{online: false, wentOffline: false},
		{online: true, wentOffline: false},
		{online: true, wentOffline: false},
		{online: false, wentOffline: true},
		{online: false, wentOffline: false},
	} {
		if got := state.update(update.online); got != update.wentOffline {
			t.Errorf("update %d went offline = %v, want %v", index, got, update.wentOffline)
		}
	}
}

func TestResetDefaultsRestoresLogicalAndPhysicalStates(t *testing.T) {
	config := spec.Config{Channels: [spec.ChannelCount]spec.ChannelConfig{
		{},
		{Default: true},
		{Inverted: true},
		{Inverted: true, Default: true},
		{Disabled: true, Inverted: true, Default: true},
	}}
	state := newControl(config)
	state.write(spec.RegChannel1, 1, false)
	state.write(spec.RegChannel2, 0, false)
	state.write(spec.RegChannel3, 1, false)
	state.write(spec.RegChannel4, 0, false)

	state.resetDefaults()
	wantValues := [spec.ChannelCount]int32{0, 1, 0, 1, 1}
	wantPinHigh := [spec.ChannelCount]bool{true, false, false, true, true}
	if state.values != wantValues {
		t.Fatalf("values after reset = %v, want %v", state.values, wantValues)
	}
	for index, want := range wantPinHigh {
		if got := state.pinHigh(index); got != want {
			t.Errorf("channel %d pin high after reset = %v, want %v", index+1, got, want)
		}
	}

	state.resetDefaults()
	if state.values != wantValues {
		t.Fatalf("values after repeated reset = %v, want %v", state.values, wantValues)
	}
	if value, null := state.read(spec.RegChannel5); !null || value != 0 {
		t.Errorf("disabled channel read after reset = (%d, %v), want (0, true)", value, null)
	}
}

func TestAnyActiveUsesRegisterState(t *testing.T) {
	tests := map[string]struct {
		config spec.Config
		want   bool
	}{
		"all off": {},
		"normal on": {
			config: spec.Config{Channels: [spec.ChannelCount]spec.ChannelConfig{{Default: true}}},
			want:   true,
		},
		"inverted false is inactive": {
			config: spec.Config{Channels: [spec.ChannelCount]spec.ChannelConfig{{Inverted: true}}},
		},
		"inverted true is active": {
			config: spec.Config{Channels: [spec.ChannelCount]spec.ChannelConfig{{Inverted: true, Default: true}}},
			want:   true,
		},
		"disabled true never active": {
			config: spec.Config{Channels: [spec.ChannelCount]spec.ChannelConfig{{Disabled: true, Default: true}}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			state := newControl(test.config)
			if got := state.anyActive(); got != test.want {
				t.Fatalf("anyActive = %v, want %v", got, test.want)
			}
		})
	}
}

func TestStatusLEDPolicy(t *testing.T) {
	tests := map[string]struct {
		online     bool
		offlineLED bool
		active     bool
		want       int32
	}{
		"online idle":       {online: true, want: ledColorOnline},
		"online active":     {online: true, active: true, want: ledColorActive},
		"offline LED off":   {active: true, want: 0},
		"offline LED on":    {offlineLED: true, want: ledColorOffline},
		"offline overrides": {offlineLED: true, active: true, want: ledColorOffline},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := statusLEDColor(test.online, test.offlineLED, test.active); got != test.want {
				t.Fatalf("statusLEDColor = %#06x, want %#06x", got, test.want)
			}
		})
	}
}

func TestStatusLEDColors(t *testing.T) {
	for name, test := range map[string]struct {
		value int32
		want  [3]byte
	}{
		"online":  {value: ledColorOnline, want: [3]byte{0x10, 0x10, 0x10}},
		"active":  {value: ledColorActive, want: [3]byte{0x40, 0x10, 0x10}},
		"offline": {value: ledColorOffline, want: [3]byte{0x00, 0xFF, 0x00}},
	} {
		t.Run(name, func(t *testing.T) {
			if got := rgbBytes(test.value); got != test.want {
				t.Fatalf("LED bytes = % X, want GRB % X", got, test.want)
			}
		})
	}
}
