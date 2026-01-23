package tcpserver

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// DataHandler 数据处理函数类型
type DataHandler func(deviceID string, data []byte, timestamp int64) error

// Server TCP服务器
type Server struct {
	port         int
	deviceID     string
	frameLength  int
	readTimeout  time.Duration
	idleTimeout  time.Duration
	dataHandler  DataHandler
	listener     net.Listener
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewServer 创建新的TCP服务器
func NewServer(port int, deviceID string, frameLength int, readTimeout, idleTimeout int, handler DataHandler) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		port:         port,
		deviceID:     deviceID,
		frameLength:  frameLength,
		readTimeout:  time.Duration(readTimeout) * time.Second,
		idleTimeout:  time.Duration(idleTimeout) * time.Second,
		dataHandler:  handler,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf("0.0.0.0:%d", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s error: %w", addr, err)
	}

	s.listener = listener
	log.Printf("[TCP] Server listening on %s, device: %s", addr, s.deviceID)

	s.wg.Add(1)
	go s.acceptLoop()

	return nil
}

// Stop 停止服务器
func (s *Server) Stop() error {
	log.Printf("[TCP] Stopping server on port %d", s.port)
	s.cancel()

	if s.listener != nil {
		s.listener.Close()
	}

	s.wg.Wait()
	log.Printf("[TCP] Server stopped on port %d", s.port)
	return nil
}

// acceptLoop 接受连接循环
func (s *Server) acceptLoop() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		// 设置Accept超时,以便能响应ctx.Done()
		if tcpListener, ok := s.listener.(*net.TCPListener); ok {
			tcpListener.SetDeadline(time.Now().Add(1 * time.Second))
		}

		conn, err := s.listener.Accept()
		if err != nil {
			// 判断是否是超时错误
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}

			// 判断是否是服务器关闭
			select {
			case <-s.ctx.Done():
				return
			default:
				log.Printf("[TCP] Accept error on port %d: %v", s.port, err)
				continue
			}
		}

		log.Printf("[TCP] New connection from %s on port %d", conn.RemoteAddr(), s.port)

		// 处理连接
		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

// handleConnection 处理单个连接
func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	log.Printf("[TCP] Handling connection from %s", remoteAddr)

	buffer := make([]byte, s.frameLength)
	lastActivityTime := time.Now()

	for {
		select {
		case <-s.ctx.Done():
			log.Printf("[TCP] Connection closed by server shutdown: %s", remoteAddr)
			return
		default:
		}

		// 检查空闲超时
		if time.Since(lastActivityTime) > s.idleTimeout {
			log.Printf("[TCP] Connection idle timeout: %s", remoteAddr)
			return
		}

		// 设置读取超时
		conn.SetReadDeadline(time.Now().Add(s.readTimeout))

		// 读取完整的数据帧
		n, err := io.ReadFull(conn, buffer)
		if err != nil {
			if err == io.EOF {
				log.Printf("[TCP] Connection closed by client: %s", remoteAddr)
				return
			}

			// 判断是否是超时错误
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// 读取超时,继续循环检查空闲超时
				continue
			}

			log.Printf("[TCP] Read error from %s: %v", remoteAddr, err)
			return
		}

		// 更新最后活动时间
		lastActivityTime = time.Now()

		// 验证数据长度
		if n != s.frameLength {
			log.Printf("[TCP] Invalid frame length from %s: expected %d, got %d", remoteAddr, s.frameLength, n)
			continue
		}

		// 获取时间戳(毫秒)
		timestamp := time.Now().UnixMilli()

		// 处理数据
		if err := s.dataHandler(s.deviceID, buffer, timestamp); err != nil {
			log.Printf("[TCP] Data handler error: %v", err)
			// 不断开连接,继续处理下一帧
		}
	}
}
