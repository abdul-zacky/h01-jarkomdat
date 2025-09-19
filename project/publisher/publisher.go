package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"github.com/quic-go/quic-go"
	"jarkom.cs.ui.ac.id/h01/project/utils"
)

type PIDSPublisher struct {
	connection quic.Connection
	address    string
}

func NewPIDSPublisher(address string) (*PIDSPublisher, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"lrt-jabodebek-2006142424"},
		ServerName:         "localhost",
	}

	conn, err := quic.DialAddr(context.Background(), address, tlsConfig, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %v", err)
	}

	return &PIDSPublisher{
		connection: conn,
		address:    address,
	}, nil
}

func (p *PIDSPublisher) SendPacket(packet utils.LRTPIDSPacket) error {
	stream, err := p.connection.OpenStreamSync(context.Background())
	if err != nil {
		return fmt.Errorf("failed to open stream: %v", err)
	}
	defer stream.Close()

	data, err := utils.Encode(packet)
	if err != nil {
		return fmt.Errorf("failed to encode packet: %v", err)
	}

	_, err = stream.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data: %v", err)
	}

	buffer := make([]byte, 1024)
	n, err := stream.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to read ACK: %v", err)
	}

	ackPacket, err := utils.Decode(buffer[:n])
	if err != nil {
		return fmt.Errorf("failed to decode ACK: %v", err)
	}

	if ackPacket.IsAck == 1 && ackPacket.TransactionID == packet.TransactionID {
		fmt.Printf("ACK received for Transaction ID: %d\n", packet.TransactionID)
		return nil
	}

	return fmt.Errorf("invalid ACK received")
}

func (p *PIDSPublisher) Close() error {
	return p.connection.CloseWithError(0, "publisher closed")
}

func main() {
	publisher, err := NewPIDSPublisher("localhost:8080")
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}
	defer publisher.Close()

	fmt.Println("PIDS Publisher connected to server")

	packetA := utils.LRTPIDSPacket{
		TransactionID:    42,
		IsAck:            0,
		IsNewTrain:       0,
		IsUpdateTrain:    0,
		IsDeleteTrain:    0,
		IsTrainArriving:  1,
		IsTrainDeparting: 0,
		TrainNumber:      42,
		DestinationLength: 0,
		Destination:      "Harjamukti",
	}

	fmt.Println("Sending Packet A (Train Arriving)...")
	err = publisher.SendPacket(packetA)
	if err != nil {
		log.Printf("Failed to send Packet A: %v", err)
	}

	time.Sleep(2 * time.Second)

	packetB := utils.LRTPIDSPacket{
		TransactionID:    42,
		IsAck:            0,
		IsNewTrain:       0,
		IsUpdateTrain:    0,
		IsDeleteTrain:    0,
		IsTrainArriving:  0,
		IsTrainDeparting: 1,
		TrainNumber:      42,
		DestinationLength: 0,
		Destination:      "Harjamukti",
	}

	fmt.Println("Sending Packet B (Train Departing)...")
	err = publisher.SendPacket(packetB)
	if err != nil {
		log.Printf("Failed to send Packet B: %v", err)
	}

	fmt.Println("All packets sent successfully!")
}