package bluetooth

import (
	"fmt"
	"github.com/muka/go-bluetooth/api"
	"github.com/muka/go-bluetooth/api/service"
	"github.com/muka/go-bluetooth/bluez/profile/adapter"
	"github.com/muka/go-bluetooth/bluez/profile/agent"
	"github.com/muka/go-bluetooth/bluez/profile/gatt"
	"github.com/muka/go-bluetooth/hw"
	"github.com/muka/go-bluetooth/hw/linux/btmgmt"
	log "ledfx/logger"
	"sync"
)

type Server struct {
	waitMu           *sync.Mutex
	adapter          *adapter.Adapter1
	mgmt             *btmgmt.BtMgmt
	app              *service.App
	audioService     *service.Service
	audioServiceChar *service.Char

	appCancel func()
	closeApp  chan struct{}

	adapterId  string
	advertName string
}

func NewServer(advertName string) (s *Server, err error) {
	s = &Server{
		closeApp:   make(chan struct{}),
		waitMu:     &sync.Mutex{},
		advertName: advertName,
	}

	if s.adapter, err = api.GetDefaultAdapter(); err != nil {
		return nil, fmt.Errorf("error getting default Bluetooth adapter: %w", err)
	}

	if s.adapterId, err = s.adapter.GetAdapterID(); err != nil {
		return nil, fmt.Errorf("error getting Bluetooth adapter id: %w", err)
	}

	s.mgmt = hw.NewBtMgmt(s.adapterId)

	if err = s.mgmt.SetPowered(false); err != nil {
		return nil, fmt.Errorf("error setting Bluetooth power state to FALSE: %w", err)
	}
	if err = s.mgmt.SetName(advertName); err != nil {
		return nil, fmt.Errorf("error setting management adapter name to '%s': %w", advertName, err)
	}
	if err = s.mgmt.SetLe(false); err != nil {
		return nil, fmt.Errorf("error setting Bluetooth LE (Low Energy) mode to TRUE: %w", err)
	}
	if err = s.mgmt.SetAdvertising(true); err != nil {
		return nil, fmt.Errorf("error enabling advertising for management adapter: %w", err)
	}
	if err = s.mgmt.SetDiscoverable(true); err != nil {
		return nil, fmt.Errorf("error setting management adapter as discoverable: %w", err)
	}
	if err = s.mgmt.SetPowered(true); err != nil {
		return nil, fmt.Errorf("error setting Bluetooth power state to TRUE: %w", err)
	}

	go func() {
		for {
			s.waitMu.Lock()
			select {
			case <-s.closeApp:
				s.appCancel()
				s.app.Close()
				s.app = nil
				s.waitMu.Unlock()
			}
		}
	}()

	return s, nil
}

func (s *Server) Serve() (err error) {
	if s.app, err = service.NewApp(service.AppOptions{
		AdapterID:         s.adapterId,
		AgentCaps:         agent.CapNoInputNoOutput,
		AgentSetAsDefault: true,
		UUIDSuffix:        "-0000-1000-8000-00805f9b34fb",
		UUID:              "0000",
	}); err != nil {
		return fmt.Errorf("error initializing new bluetooth app: %w", err)
	}
	s.app.SetName(s.advertName)

	if !s.app.Adapter().Properties.Powered {
		if err = s.app.Adapter().SetPowered(true); err != nil {
			return fmt.Errorf("error powering on Bluetooth adapter: %w", err)
		}
	}

	if s.audioService, err = s.app.NewService("110b"); err != nil {
		return fmt.Errorf("error creating new app service: %w", err)
	}
	if s.audioServiceChar, err = s.audioService.NewChar("3344"); err != nil {
		return fmt.Errorf("error initializing new service characteristic: %w", err)
	}

	s.audioServiceChar.Properties.Flags = []string{
		gatt.FlagCharacteristicRead,
		gatt.FlagCharacteristicWrite,
	}

	if err = s.audioService.AddChar(s.audioServiceChar); err != nil {
		return fmt.Errorf("error adding characteristic to service: %w", err)
	}

	descr1, err := s.audioServiceChar.NewDescr("4455")
	if err != nil {
		return fmt.Errorf("error initializing new characteristic descriptor: %w", err)
	}

	descr1.Properties.Flags = []string{
		gatt.FlagDescriptorRead,
		gatt.FlagDescriptorWrite,
	}

	if err = s.audioServiceChar.AddDescr(descr1); err != nil {
		return fmt.Errorf("error adding descriptor to characteristic: %w", err)
	}

	if err = s.app.Run(); err != nil {
		return fmt.Errorf("error starting Bluetooth app: %w", err)
	}

	if s.appCancel, err = s.app.Advertise(uint32(0)); err != nil {
		return fmt.Errorf("error advertising Bluetooth app: %w", err)
	}

	log.Logger.WithField("category", "BLE Server").Infof("Exposed service %s", s.audioService.Properties.UUID)

	return nil
}

func (s *Server) CloseApp() {
	if s.app != nil {
		s.closeApp <- struct{}{}
	}
}

func (s *Server) Wait() {
	s.waitMu.Lock()
	defer s.waitMu.Unlock()
}
