package client

import "testing"

func TestAlsaDeviceHasTwoChannels(t *testing.T) {
	device, _ := GetDevice()
	if device.Channels != 2 {
		t.Fail()
	}
}

func TestRunReadsFromBuffer(t *testing.T) {
	//_ = Run("localhost")
	t.Fail()
}

func TestPlaybackLoop(t *testing.T) {
	device, _ := GetDevice()
	var input = make(chan []byte, 512)
	go playbackLoop(input, device)
	for {
		input <- []byte{2, 3, 5, 7, 3, 1}
	}
}
