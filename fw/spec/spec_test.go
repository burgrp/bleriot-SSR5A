package spec

import "testing"

func TestTypeProvidesAllChannelsByDefault(t *testing.T) {
	deviceType := Type(Config{})
	if err := deviceType.Validate(); err != nil {
		t.Fatal(err)
	}

	if len(deviceType.Registers) != ChannelCount {
		t.Fatalf("register count = %d, want %d", len(deviceType.Registers), ChannelCount)
	}
	for index, register := range deviceType.Registers {
		if register.Tag != uint16(index+1) || register.Name != channelNames[index] {
			t.Errorf("register %d = (%d, %q), want (%d, %q)", index, register.Tag, register.Name, index+1, channelNames[index])
		}
	}
}

func TestTypeOmitsDisabledChannelsWithoutRenumbering(t *testing.T) {
	config := Config{Channels: [ChannelCount]ChannelConfig{
		{},
		{Disabled: true},
		{},
		{Disabled: true},
		{},
	}}
	registers := Type(config).Registers

	wantTags := []uint16{RegChannel1, RegChannel3, RegChannel5}
	if len(registers) != len(wantTags) {
		t.Fatalf("register count = %d, want %d", len(registers), len(wantTags))
	}
	for index, wantTag := range wantTags {
		if registers[index].Tag != wantTag {
			t.Errorf("register %d tag = %d, want %d", index, registers[index].Tag, wantTag)
		}
	}
}
