package main

import (
	"crypto/rand"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

var (
	crt          = flag.String("crt", "", "Usage: -crt=<path/to/crt_file>")
	key          = flag.String("key", "", "Usage: -key=<path/to/key_file>")
	local_addr   = flag.String("local_addr", "0.0.0.0:443", "Usage: -local_addr=<local_addr:local_port>")
	forward_port = flag.Int("forward_port", 9000, "Usage: -forward_port=<forward_port>")
	remote_addr  = flag.String("remote_addr", "", "Usage: -remote_addr=<remote_addr:remote_port>")
)

func main() {
	var err error

	flag.Parse()

	if *forward_port < 20 || *forward_port > 65535 || *local_addr == "" || *remote_addr == "" {
		flag.PrintDefaults()
		log.Fatal()
	}

	go func() {
		local, err := net.Listen("tcp", "127.0.0.1:"+fmt.Sprintf("%d", *forward_port))
		if local == nil {
			log.Fatalf("cannot listen: %v", err)
		}
		for {
			conn, err := local.Accept()
			if conn == nil {
				log.Fatalf("accept failed: %v", err)
			}
			go forward_remote(conn, *remote_addr)
		}
	}()

	var local net.Listener
	if *crt != "" && *key != "" {
		cert, err := tls.LoadX509KeyPair(*crt, *key)
		if err != nil {
			log.Fatalf("Cannot load KeyPair: %v", err)
		}

		local, err = tls.Listen("tcp", *local_addr, &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
			Rand:               rand.Reader,
		})
	} else {
		local, err = net.Listen("tcp", *local_addr)
	}
	if err != nil {
		log.Fatalf("cannot listen: %v", err)
	}

	for {
		conn, err := local.Accept()
		if conn == nil {
			log.Fatalf("accept failed: %v", err)
		}
		go forward_local(conn, "127.0.0.1:"+fmt.Sprintf("%d", *forward_port))
	}

}

func forward_local(conn net.Conn, remoteAddr string) {
	defer conn.Close()

	remote, err := net.Dial("tcp", remoteAddr)
	if remote == nil {
		fmt.Fprintf(os.Stderr, "remote dial failed: %v\n", err)
		return
	}

	forward(conn, remote)
}

func forward_remote(conn net.Conn, remoteAddr string) {
	defer conn.Close()

	remote, err := tls.Dial("tcp", remoteAddr, &tls.Config{InsecureSkipVerify: true})
	if remote == nil {
		log.Printf("remote dial failed: %v\n", err)
		return
	}

	forward(conn, remote)
}

func forward(local, remote net.Conn) {
	done := make(chan struct{}, 3)
	go func() {
		_, e := io.Copy(local, remote)
		if e != nil {
			log.Printf("remote dial: %v\n", e)
		}
		done <- struct{}{}
	}()
	go func() {
		_, e := io.Copy(remote, local)
		if e != nil {
			log.Printf("remote dial: %v\n", e)
		}
		done <- struct{}{}
	}()

	<-done

	fmt.Fprintf(os.Stderr, "remote dial end: %s %s\n", remote.RemoteAddr(), local.RemoteAddr())
}
