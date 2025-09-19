package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/quic-go/quic-go"
	"jarkom.cs.ui.ac.id/h01/project/utils"
)

const (
	serverIP          = "3.81.118.89"
	serverPort        = "4510"
	serverType        = "udp4"
	bufferSize        = 2048
	appLayerProto     = "lrt-jabodebek-2306214510"
	sslKeyLogFileName = "ssl-key.log"
)

func main() {
	sslKeyLogFile, err := os.Create(sslKeyLogFileName)
	if err != nil {
		log.Fatalln(err)
	}
	defer sslKeyLogFile.Close()

	fmt.Printf("QUIC Publisher (Node Kendali Stasiun) - PIDS System\n")

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{appLayerProto},
		KeyLogWriter:       sslKeyLogFile,
	}

	connection, err := quic.DialAddr(context.Background(), net.JoinHostPort(serverIP, serverPort), tlsConfig, &quic.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	defer connection.CloseWithError(0x0, "No Error")

	fmt.Printf("[quic] Connected to %s\n", connection.RemoteAddr())

	// Create Packet A: Train 42 to Harjamukti arriving
	destination := "Harjamukti"
	packetA := utils.LRTPIDSPacket{
		LRTPIDSPacketFixed: utils.LRTPIDSPacketFixed{
			TransactionId:     1,
			IsAck:             false,
			IsNewTrain:        false,
			IsUpdateTrain:     false,
			IsDeleteTrain:     false,
			IsTrainArriving:   true,
			IsTrainDeparting:  false,
			TrainNumber:       42,
			DestinationLength: uint8(len(destination)),
		},
		Destination: destination,
	}

	// Create Packet B: Train 42 to Harjamukti departing
	packetB := utils.LRTPIDSPacket{
		LRTPIDSPacketFixed: utils.LRTPIDSPacketFixed{
			TransactionId:     2,
			IsAck:             false,
			IsNewTrain:        false,
			IsUpdateTrain:     false,
			IsDeleteTrain:     false,
			IsTrainArriving:   false,
			IsTrainDeparting:  true,
			TrainNumber:       42,
			DestinationLength: uint8(len(destination)),
		},
		Destination: destination,
	}

	// Send Packet A and receive ACK
	sendPacketAndReceiveACK(connection, packetA, "A")

	// Send Packet B and receive ACK
	sendPacketAndReceiveACK(connection, packetB, "B")

	fmt.Printf("[quic] Closing connection\n")
}

func sendPacketAndReceiveACK(connection quic.Connection, packet utils.LRTPIDSPacket, packetName string) {
	// Encode and send packet
	encodedData := utils.Encoder(packet)

	stream, err := connection.OpenStreamSync(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("[quic] Sending Packet %s (Train %d to %s)\n", packetName, packet.TrainNumber, packet.Destination)
	_, err = stream.Write(encodedData)
	if err != nil {
		log.Fatalln(err)
	}

	// Receive ACK
	receiveBuffer := make([]byte, bufferSize)
	receiveLength, err := stream.Read(receiveBuffer)
	if err != nil {
		log.Fatalln(err)
	}

	// Decode ACK
	ackPacket := utils.Decoder(receiveBuffer[:receiveLength])
	fmt.Printf("[quic] Received ACK for Packet %s (Transaction ID: %d)\n", packetName, ackPacket.TransactionId)

	stream.Close()
}