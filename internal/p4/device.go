// SPDX-License-Identifier: Apache-2.0
package p4

import (
	"fmt"
	"log"
	"phase4/internal/app/errors"

	"github.com/gordonklaus/portaudio"
)

func initPA(e *Engine) error {
	if e.audio.initialized {
		log.Print("Engine ➜ PortAudio already initialized")
		return nil
	}

	if err := e.audio.client.Initialize(); err != nil {
		return err
	}
	e.audio.initialized = true
	log.Print("Engine ➜ PortAudio initialized")

	devices, err := e.audio.client.Devices()
	if err != nil {
		exitErr := exitPA(e)
		if exitErr != nil {
			return &errors.FatalError{
				Message: "failed to get audio devices, additionally failed to terminate PortAudio",
				Err:     fmt.Errorf("%w (termination error: %v)", err, exitErr),
			}
		}

		return &errors.FatalError{
			Message: "failed to get audio devices",
			Err:     err,
		}
	}
	if len(devices) == 0 {
		exitErr := exitPA(e)
		if exitErr != nil {
			return &errors.FatalError{
				Message: "no audio devices found, additionally failed to terminate PortAudio",
				Err:     fmt.Errorf("%w (termination error: %v)", fmt.Errorf("no audio devices available"), exitErr),
			}
		}

		return &errors.FatalError{
			Message: "no audio devices found",
			Err:     fmt.Errorf("no audio devices found"),
		}
	}
	e.audio.devices = devices

	return nil
}

func exitPA(e *Engine) error {
	if e.audio.initialized {
		if err := e.audio.client.Terminate(); err != nil {
			return &errors.FatalError{
				Message: "failed to terminate PortAudio",
				Err:     err,
			}
		}
		e.audio.initialized = false
		log.Print("Engine ➜ PortAudio terminated")
	}

	return nil
}

func selectInputDevice(e *Engine) error {
	defaultDeviceID := -1
	deviceID := e.config.Input.Device

	// Fallback (if allowed), when the ID is out of range or the selected
	// device is not an input device. If more input channels have been requested
	// than are available, fallback to the devices max input channels.
	if deviceID == defaultDeviceID || deviceID >= len(e.audio.devices) {
		deviceID = defaultDeviceID
	}
	if deviceID > defaultDeviceID {
		device := e.audio.devices[deviceID]
		if device.MaxInputChannels > 0 {
			if e.config.Input.Channels > device.MaxInputChannels {
				log.Printf("Engine ➜ Warning ➜ Requested %d channels but device only supports %d",
					e.config.Input.Channels, device.MaxInputChannels)
				e.config.Input.Channels = device.MaxInputChannels
			}
			e.audio.inputDevice = device
		} else {
			if e.config.Input.UseDefaultDevice {
				deviceID = defaultDeviceID
			}
		}
	}

	if deviceID == defaultDeviceID && e.config.Input.UseDefaultDevice {
		device, err := e.audio.client.DefaultInputDevice()
		if err != nil {
			return &errors.FatalError{
				Message: "failed to set default PortAudio device",
				Err:     err,
			}
		}
		e.audio.inputDevice = device
	}

	if e.audio.inputDevice == nil {
		return fmt.Errorf("id: %d, useDefaultDevice: %v",
			deviceID, e.config.Input.UseDefaultDevice)
	}

	return nil
}

func printInputDevice(device *portaudio.DeviceInfo) {
	if device == nil {
		log.Print("Engine ➜ No input device selected.")
		return
	}

	log.Printf("Engine ➜ %s ➜ Host API: %s (Type: %s)", device.Name, device.HostApi.Name, device.HostApi.Type)
	log.Printf("Engine ➜ %s ➜ d.Index: %d", device.Name, device.Index)
	log.Printf("Engine ➜ %s ➜ d.MaxInputChannels: %d", device.Name, device.MaxInputChannels)
	log.Printf("Engine ➜ %s ➜ d.DefaultSampleRate: %.2f Hz", device.Name, device.DefaultSampleRate)
	log.Printf("Engine ➜ %s ➜ d.DefaultLowInputLatency: %s", device.Name, device.DefaultLowInputLatency)
	log.Printf("Engine ➜ %s ➜ d.DefaultHighInputLatency: %s", device.Name, device.DefaultHighInputLatency)
}
