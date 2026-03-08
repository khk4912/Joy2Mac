package joy2mac

import (
	"fmt"
	"sync"
	"time"

	"tinygo.org/x/bluetooth"
)

type AdapterManager struct {
	Adapter        *bluetooth.Adapter
	JoyconSessions map[string]*JoyconSession

	maxJoyconConnections int
	seenDevices          map[string]struct{}
	candidates           []JoyconCandidate

	shuttingDown bool
	mu           sync.Mutex
	connectMu    sync.Mutex
}

func NewAdapterManager(maxJoyconConnections int) *AdapterManager {
	manager := &AdapterManager{
		Adapter:              bluetooth.DefaultAdapter,
		JoyconSessions:       make(map[string]*JoyconSession),
		seenDevices:          make(map[string]struct{}),
		candidates:           make([]JoyconCandidate, 0, maxJoyconConnections),
		maxJoyconConnections: maxJoyconConnections,
	}
	manager.Adapter.SetConnectHandler(manager.onConnect)
	return manager
}

func (am *AdapterManager) AddJoyconSession(session *JoyconSession) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.JoyconSessions[session.address.String()] = session
}

func (am *AdapterManager) ConnectSession(session *JoyconSession) error {
	am.connectMu.Lock()
	device, err := am.Adapter.Connect(session.Address(), bluetooth.ConnectionParams{
		ConnectionTimeout: bluetooth.NewDuration(5 * time.Second),
	})
	am.connectMu.Unlock()
	if err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}

	session.attachDevice(device)
	err = session.setupConnection()

	if err != nil {
		session.resetConnectionState()
		_ = device.Disconnect()
		return err
	}

	session.markConnected()
	session.StartInputNotification(session.inputCh)
	return nil
}

func (am *AdapterManager) Shutdown() {
	am.mu.Lock()
	am.shuttingDown = true
	sessions := make([]*JoyconSession, 0, len(am.JoyconSessions))
	for _, session := range am.JoyconSessions {
		sessions = append(sessions, session)
	}
	am.mu.Unlock()

	for _, session := range sessions {
		_ = session.Disconnect()
	}
}

func (am *AdapterManager) onConnect(device bluetooth.Device, connected bool) {
	am.mu.Lock()
	targetSession := am.JoyconSessions[device.Address.String()]
	shuttingDown := am.shuttingDown
	am.mu.Unlock()

	if shuttingDown || targetSession == nil {
		return
	}

	if !connected {
		if !targetSession.markDisconnected() {
			return
		}
		go targetSession.ReconnectLoop(am)
	}
}
