package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"

	"github.com/quic-go/quic-go"
	"jarkom.cs.ui.ac.id/h01/project/utils"
)

type PIDSSubscriber struct {
	listener *quic.Listener
	address  string
}

func NewPIDSSubscriber(address string) (*PIDSSubscriber, error) {
	tlsConfig := generateTLSConfig()

	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen UDP: %v", err)
	}

	listener, err := quic.Listen(conn, tlsConfig, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create QUIC listener: %v", err)
	}

	return &PIDSSubscriber{
		listener: listener,
		address:  address,
	}, nil
}

func (s *PIDSSubscriber) Start() {
	fmt.Printf("PIDS Subscriber started on %s\n", s.address)
	fmt.Println("Waiting for connections...")

	for {
		conn, err := s.listener.Accept(context.Background())
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *PIDSSubscriber) handleConnection(conn quic.Connection) {
	fmt.Printf("New connection from: %s\n", conn.RemoteAddr())

	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			log.Printf("Failed to accept stream: %v", err)
			return
		}

		go s.handleStream(stream, conn)
	}
}

func (s *PIDSSubscriber) handleStream(stream quic.Stream, conn quic.Connection) {
	defer stream.Close()

	buffer := make([]byte, 1024)
	for {
		n, err := stream.Read(buffer)
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("Failed to read from stream: %v", err)
			}
			return
		}

		packet, err := utils.Decode(buffer[:n])
		if err != nil {
			log.Printf("Failed to decode packet: %v", err)
			continue
		}

		s.processPacket(packet, stream)
	}
}

func (s *PIDSSubscriber) processPacket(packet utils.LRTPIDSPacket, stream quic.Stream) {
	if packet.IsTrainArriving == 1 {
		fmt.Printf("Mohon perhatian, kereta tujuan %s akan tiba di Peron 1.\n", packet.Destination)
	}

	if packet.IsTrainDeparting == 1 {
		fmt.Printf("Mohon perhatian, kereta tujuan %s akan diberangkatkan dari Peron 1.\n", packet.Destination)
	}

	ackPacket := utils.LRTPIDSPacket{
		TransactionID:    packet.TransactionID,
		IsAck:            1,
		IsNewTrain:       0,
		IsUpdateTrain:    0,
		IsDeleteTrain:    0,
		IsTrainArriving:  0,
		IsTrainDeparting: 0,
		TrainNumber:      packet.TrainNumber,
		DestinationLength: 0,
		Destination:      "",
	}

	ackData, err := utils.Encode(ackPacket)
	if err != nil {
		log.Printf("Failed to encode ACK packet: %v", err)
		return
	}

	_, err = stream.Write(ackData)
	if err != nil {
		log.Printf("Failed to send ACK: %v", err)
	}
}

func (s *PIDSSubscriber) Close() error {
	return s.listener.Close()
}

func generateTLSConfig() *tls.Config {
	cert, _ := tls.LoadX509KeyPair("server.crt", "server.key")
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"lrt-jabodebek-2006142424"},
	}
}

func main() {
	subscriber, err := NewPIDSSubscriber(":8080")
	if err != nil {
		log.Fatalf("Failed to create subscriber: %v", err)
	}
	defer subscriber.Close()

	subscriber.Start()
}