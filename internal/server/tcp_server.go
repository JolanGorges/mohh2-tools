package server

import (
	"log/slog"
	"net"
	"os"
)

func StartTCPServer(port string) {
	ln, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		slog.Error("Could not start TCP server", "addr", ln.Addr(), "err", err)
		os.Exit(1)
	}

	slog.Info("TCP server listening", "addr", ln.Addr())

	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Error("Could not accept TCP connection", "addr", ln.Addr(), "err", err)
			continue
		}

		slog.Info("TCP connection accepted", "localAddr", ln.Addr(), "remoteAddr", conn.RemoteAddr())

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	checkForSSL := true
	buffer := make([]byte, 1024)
	parser := NewAriesParser(conn)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err.Error() == "EOF" {
				slog.Info("TCP connection closed by client", "localAddr", conn.LocalAddr(), "remoteAddr", conn.RemoteAddr())
				return
			}
			slog.Error("Could not read from TCP connection", "localAddr", conn.LocalAddr(), "remoteAddr", conn.RemoteAddr(), "err", err)
			return
		}

		if checkForSSL {
			if n > 5 && buffer[0] == 0x16 && buffer[1] == 3 && buffer[2] == 0 {
				slog.Error("The game must be patched to not use SSL", "localAddr", conn.LocalAddr(), "remoteAddr", conn.RemoteAddr())

				err := conn.Close()
				if err != nil {
					slog.Error("Could not close TCP connection", "localAddr", conn.LocalAddr(), "remoteAddr", conn.RemoteAddr(), "err", err)
					return
				}

				slog.Info("TCP connection closed", "localAddr", conn.LocalAddr(), "remoteAddr", conn.RemoteAddr())
				return
			}

			checkForSSL = false
		}

		parser.parse(buffer[:n])
	}
}
