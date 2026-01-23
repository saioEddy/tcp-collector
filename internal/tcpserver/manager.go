package tcpserver

import (
	"fmt"
	"log"
	"sync"
)

// Manager TCP服务器管理器
type Manager struct {
	servers []*Server
	mu      sync.RWMutex
}

// NewManager 创建新的管理器
func NewManager() *Manager {
	return &Manager{
		servers: make([]*Server, 0),
	}
}

// AddServer 添加服务器
func (m *Manager) AddServer(server *Server) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.servers = append(m.servers, server)
}

// StartAll 启动所有服务器
func (m *Manager) StartAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	log.Printf("[Manager] Starting %d TCP servers", len(m.servers))

	for _, server := range m.servers {
		if err := server.Start(); err != nil {
			// 启动失败,停止已启动的服务器
			m.StopAll()
			return fmt.Errorf("start server error: %w", err)
		}
	}

	log.Printf("[Manager] All TCP servers started")
	return nil
}

// StopAll 停止所有服务器
func (m *Manager) StopAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	log.Printf("[Manager] Stopping %d TCP servers", len(m.servers))

	var wg sync.WaitGroup
	for _, server := range m.servers {
		wg.Add(1)
		go func(s *Server) {
			defer wg.Done()
			if err := s.Stop(); err != nil {
				log.Printf("[Manager] Stop server error: %v", err)
			}
		}(server)
	}

	wg.Wait()
	log.Printf("[Manager] All TCP servers stopped")
	return nil
}
