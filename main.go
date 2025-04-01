package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

type TunnelConfig struct {
	LocalAddr   string
	RemoteAddr  string
	SSHAddr     string
	SSHUser     string
	SSHKeyPath  string
	SSHPassword string
	Type        string
}

func main() {
	log.Println("Starting SSH Tunnel...")

	for _, server := range Config.Server {
		for _, service := range server.Services {
			log.Printf("Starting tunnel for %s -> %s", service.Local, service.Remote)

			if server.User == "" {
				log.Fatal("Please provide SSH username")
			}

			if server.Password == "" && server.Key == "" {
				log.Fatal("Please provide SSH password or private key path")
			}

			config := &TunnelConfig{
				LocalAddr:   service.Local,
				RemoteAddr:  service.Remote,
				SSHAddr:     fmt.Sprintf("%s:%d", server.Host, server.Port),
				SSHUser:     server.User,
				SSHKeyPath:  server.Key,
				SSHPassword: server.Password,
				Type:        service.Type,
			}

			go func() {
				if err := startTunnel(config); err != nil {
					log.Fatal(err)
				}
			}()
		}
	}

	select {}
}

func startTunnel(config *TunnelConfig) error {
	for {
		if err := createTunnel(config); err != nil {
			log.Printf("Tunnel error: %v, retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
			continue
		}
	}
}

func createTunnel(config *TunnelConfig) error {
	var authMethod ssh.AuthMethod

	if config.SSHPassword != "" {
		authMethod = ssh.Password(config.SSHPassword)
	} else {
		key, err := os.ReadFile(config.SSHKeyPath)
		if err != nil {
			return fmt.Errorf("unable to read private key: %v", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return fmt.Errorf("unable to parse private key: %v", err)
		}
		authMethod = ssh.PublicKeys(signer)
	}

	sshConfig := &ssh.ClientConfig{
		User: config.SSHUser,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	for {
		sshClient, err := ssh.Dial("tcp", config.SSHAddr, sshConfig)
		if err != nil {
			return fmt.Errorf("SSH Connection failed: %v", err)
		}
		defer sshClient.Close()

		if config.Type == "remote" {
			listener, err := net.Listen("tcp", config.LocalAddr)
			if err != nil {
				return fmt.Errorf("listen failed: %v", err)
			}
			defer listener.Close()

			log.Printf("Remote tunnel created: %s -> %s\n", config.LocalAddr, config.RemoteAddr)

			for {
				local, err := listener.Accept()
				if err != nil {
					log.Printf("Accept connection failed: %v", err)
					return err
				}
				remote, err := sshClient.Dial("tcp", config.RemoteAddr)
				if err != nil {
					log.Printf("Remote connection failed: %v", err)
					local.Close()
					continue
				}
				go handleConnection(local, remote)
			}
		} else if config.Type == "local" {
			listener, err := sshClient.Listen("tcp", config.RemoteAddr)
			if err != nil {
				return fmt.Errorf("remote listen failed: %v", err)
			}
			defer listener.Close()

			log.Printf("Local tunnel created: %s -> %s\n", config.LocalAddr, config.RemoteAddr)

			for {
				remote, err := listener.Accept()
				if err != nil {
					log.Printf("Accept connection failed: %v", err)
					return err
				}
				local, err := net.Dial("tcp", config.LocalAddr)
				if err != nil {
					log.Printf("Local connection failed: %v", err)
					remote.Close()
					continue
				}
				go handleConnection(local, remote)
			}
		}
	}
}

func handleConnection(conn1, conn2 net.Conn) {
	done := make(chan bool, 2)
	go func() {
		copyData(conn1, conn2)
		done <- true
	}()
	go func() {
		copyData(conn2, conn1)
		done <- true
	}()

	<-done
}

func copyData(dst net.Conn, src net.Conn) {
	defer dst.Close()
	defer src.Close()
	_, _ = io.Copy(dst, src)
}
