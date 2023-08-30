package main

import (
	"golang.org/x/sys/unix"
	"log"
	"net"
	"syscall"
)

const BACKLOG int = 64

func Accept(fd int) (int, error) {
	nfd, _, err := syscall.Accept(fd)
	return nfd, err
}

func Close(fd int) {
	syscall.Close(fd)
}

func Read(fd int, buf []byte) (int, error) {
	return syscall.Read(fd, buf)
}

func Write(fd int, buf []byte) (int, error) {
	return syscall.Write(fd, buf)
}

func Connect(host [4]byte, port int) (int, error) {
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		log.Printf("int socket err: %v\n", err)
		return -1, err
	}
	var addr syscall.SockaddrInet4
	addr.Addr = host
	addr.Port = port
	err = syscall.Connect(s, &addr)
	if err != nil {
		log.Printf("connect err: %v\n", err)
		return -1, err
	}
	return s, err
}

func TCPServer(port int) (int, error) {
	// use syscall, open a socket
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		log.Printf("int socket err: %v\n", err)
		return -1, err
	}

	err = syscall.SetsockoptInt(s, syscall.SOL_SOCKET, unix.SO_REUSEPORT, port)
	if err != nil {
		log.Printf("set SO_REUSEPORT err: %v\n", err)
		syscall.Close(s)
		return -1, err
	}

	// ip address translation
	var addr [4]byte
	copy(addr[:], net.ParseIP("127.0.0.1").To4())
	err = syscall.Bind(s, &syscall.SockaddrInet4{
		Port: port,
		Addr: addr,
	})
	if err != nil {
		log.Printf("bind addr err: %v\n", err)
		syscall.Close(s)
		return -1, err
	}

	// start listening
	err = syscall.Listen(s, BACKLOG)
	if err != nil {
		log.Printf("listen socket err: %v\n", err)
		syscall.Close(s)
		return -1, err
	}

	return s, nil
}
