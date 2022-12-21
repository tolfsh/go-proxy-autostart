package main

import (
	"context"
	"log"
	"net"
	"os"
	"sync"
	"time"
	"io"

	mapset "github.com/deckarep/golang-set/v2"
	dockerTypes "github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
)

var (
	infoLogger       = log.New(log.Writer(), "[INFO] ", log.Ldate|log.Ltime)
	errorLogger      = log.New(log.Writer(), "[ERROR] ", log.Ldate|log.Ltime)
	containerReady = true
	containerHealth  string
	containerState	 string
	clients          mapset.Set[*net.Conn]
	server_ip        string
	server_port      string
	containerName    string
	servicePort      string
)

func main() {
	getEnvParam()
	clients = mapset.NewSet[*net.Conn]()
	listenAddress := net.JoinHostPort(server_ip, server_port)
	listener, err := net.Listen("tcp", listenAddress)
	go monitorContainer(containerName)
	if err != nil {
		errorLogger.Fatalf("Cannot open listener: %s", err.Error())
	}
	infoLogger.Printf("Accepting connections on %s", listenAddress)
	for {
		conn, err := listener.Accept()
		if err != nil {
			errorLogger.Fatalf("Cannot open listener: %s", err.Error())
		}
		infoLogger.Printf("New Connection from %s", conn.RemoteAddr())
		if containerReady {
			go handleConnection(conn)
		} else {
			err := startContainer(containerName)
			if err != nil {
				errorLogger.Printf("Cannot start %s: %s", containerName, err.Error())
			} else {
				go handleConnection(conn)
			}
		}
	}
}

func getEnvParam() {
	server_ip = os.Getenv("LISTEN_IP")
	if server_ip == "" {
		server_ip = "0.0.0.0"
	}
	server_port = os.Getenv("LISTEN_PORT")
	if server_port == "" {
		server_port = "25565"
	}
	containerName = os.Getenv("CONTAINER_NAME")
	if containerName == "" {
		errorLogger.Fatalf("CONTAINER_NAME is mandatory")
	}
	servicePort = os.Getenv("SERVICE_PORT")
	if servicePort == "" {
		servicePort = "25565"
	}
}

func startContainer(containerName string) error {
	cli, err := docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()
	infoLogger.Printf("Starting container %s", containerName)
	cli.ContainerStart(context.Background(), containerName, dockerTypes.ContainerStartOptions{})
	return nil
}

func monitorContainer(containerName string) {
	cli, err := docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())
	if err != nil {
		errorLogger.Panicf("Cannot create Docker client: %s", err.Error())
	}
	defer cli.Close()
	for {
		resp, err := cli.ContainerInspect(context.Background(), containerName)
		if err != nil {
			errorLogger.Panicf("Cannot get container state: %s", err.Error())
		}
		if resp.State.Health.Status != containerHealth {
			infoLogger.Printf("Container %s is %s", containerName, resp.State.Health.Status)
			containerHealth = resp.State.Health.Status
			if resp.State.Health.Status == "healthy" {
				containerReady = true
			} else {
				containerReady = false
			}
		}
		if resp.State.Status != containerState {
			infoLogger.Printf("")
		}
		time.Sleep(time.Second * 5)
	}
}

func handleConnection(clientConn net.Conn) {
	defer func() {
		clients.Remove(&clientConn)
		infoLogger.Printf("Client %s left", conn.RemoteAddr())
	}()
	clients.Add(&clientConn)
	
	if !containerReady {
		infoLogger.Printf("Waiting for container %s to be ready", containerName)
		for i := 0; i<30; i++ {
			if containerReady {
				break
			}
			time.Sleep(time.Second)
		}
		if !containerReady {
			errorLogger.Printf("Container %s has not started in 30s. Aborting", containerName)
			return
		}
	}

	serviceConnectionString := net.JoinHostPort(containerName, servicePort)
	serviceConn, err := net.DialTimeout("tcp", serviceConnectionString, time.Second*10)
	if err != nil {
		errorLogger.Printf("Cannot connect to %s", serviceConnectionString)
	}
	
	var waitGroup sync.WaitGroup
	waitGroup.Add(2)

	go readAndTransfer(clientConn, serviceConn, &waitGroup)
	go readAndTransfer(serviceConn, clientConn, &waitGroup)

	waitGroup.Wait()
}

func readAndTransfer(src net.Conn, dst net.Conn, wait *sync.WaitGroup) {
	defer dst.Close()
	defer wait.Done()
	for {
		buf := make([]byte, 8192)
		count, err := src.Read(buf)
		if err != nil && err != io.EOF {
			errorLogger.Printf("Error: %s", err.Error())
		}
		if count != 0 {
			dst.Write(buf)
		} else {
			return
		}
	}
}
