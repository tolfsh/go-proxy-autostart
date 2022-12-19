package main

import (
	"log"
	"net"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
)

const (
	server_ip     = "127.0.0.1"
	server_port   = "25565"
	servicePort   = "25566"
	containerName = "minecraft"
)

var (
	infoLogger       = log.New(log.Writer(), "[INFO]", log.Ldate|log.Ltime)
	errorLogger      = log.New(log.Writer(), "[ERROR]", log.Ldate|log.Ltime)
	containerStarted = true
	clients          mapset.Set[*net.Conn]
)

func main() {
	clients = mapset.NewSet[*net.Conn]()
	listener, err := net.Listen("tcp", net.JoinHostPort(server_ip, server_port))
	go monitorContainer(containerName)
	if err != nil {
		errorLogger.Fatalf("Cannot open listener: %s", err.Error())
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			errorLogger.Fatalf("Cannot open listener: %s", err.Error())
		}
		if containerStarted {
			go handleConnection(conn)
		} else {
			startContainer()
		}
	}
}

func startContainer() {
	infoLogger.Printf("Container %s has been started", containerName)
	// TODO add real implémtentation
}

func monitorContainer(containerName string) bool {
	containerStarted = true
	return true
	//TODO add real implémentation
}

func handleConnection(clientConn net.Conn) {
	defer clientConn.Close()
	defer func() {
		clients.Remove(&clientConn)
	}()

	clients.Add(&clientConn)

	infoLogger.Printf("New Connection from %s", clientConn.RemoteAddr())
	serviceConnectionString := net.JoinHostPort(containerName, servicePort)
	serviceConn, err := net.DialTimeout("tcp", serviceConnectionString, time.Second*10)
	if err != nil {
		errorLogger.Panicf("Cannot connect to %s", serviceConnectionString)
	}

	var waitGroup sync.WaitGroup
	waitGroup.Add(2)
	go func() {
		defer waitGroup.Done()
		go readAndTransfer(&clientConn, &serviceConn)
		go readAndTransfer(&serviceConn, &clientConn)
	}()

	waitGroup.Wait()
}

func readAndTransfer(src *net.Conn, dst *net.Conn) {
	for {
		var buf [8192]byte
		count, err := src.Read(buf)
	}
}
