package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/quic-go/quic-go"
	"jarkom.cs.ui.ac.id/h01/project/utils"
)

const (
	serverIP          = ""
	serverPort        = "4510"
	serverType        = "udp4"
	bufferSize        = 2048
	appLayerProto     = "lrt-jabodebek-2306214510"
	sslKeyLogFileName = "ssl-key.log"
)

func Handler(packet utils.LRTPIDSPacket) string {
	if packet.IsTrainArriving {
		return fmt.Sprintf("Mohon perhatian, kereta tujuan %s akan tiba di Peron 1.", packet.Destination)
	} else if packet.IsTrainDeparting {
		return fmt.Sprintf("Mohon perhatian, kereta tujuan %s akan diberangkatkan dari Peron 1.", packet.Destination)
	}
	return ""
}

func main() {
	localUdpAddress, err := net.ResolveUDPAddr(serverType, net.JoinHostPort(serverIP, serverPort))
	if err != nil {
		log.Fatalln(err)
	}
	socket, err := net.ListenUDP(serverType, localUdpAddress)
	if err != nil {
		log.Fatalln(err)
	}
	defer socket.Close()

	fmt.Printf("QUIC Subscriber (Node Layar PIDS) - PIDS System\n")
	fmt.Printf("[%s] Preparing UDP listening socket on %s\n", serverType, socket.LocalAddr())

	tlsConfig := &tls.Config{
		Certificates: utils.GenerateTLSSelfSignedCertificates(),
		NextProtos:   []string{appLayerProto},
	}
	listener, err := quic.Listen(socket, tlsConfig, &quic.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	defer listener.Close()

	fmt.Printf("[quic] Listening QUIC connections on %s\n", listener.Addr())

	for {
		connection, err := listener.Accept(context.Background())
		if err != nil {
			log.Fatalln(err)
		}

		go handleConnection(connection)
	}
}

func handleConnection(connection quic.Connection) {
	fmt.Printf("[quic] Receiving connection from %s\n", connection.RemoteAddr())

	for {
		stream, err := connection.AcceptStream(context.Background())
		if err != nil {
			log.Printf("Error accepting stream: %v\n", err)
			return
		}
		go handleStream(connection.RemoteAddr(), stream)
	}
}

func handleStream(clientAddress net.Addr, stream quic.Stream) {
	fmt.Printf("[quic] [Client: %s] Receive stream open request with ID %d\n", clientAddress, stream.StreamID())

	_, err := io.Copy(pidsProcessor{stream}, stream)
	if err != nil {
		fmt.Println(err)
	}
}

type pidsProcessor struct{ io.Writer }

func (pp pidsProcessor) Write(receivedMessageRaw []byte) (int, error) {
	// Decode the received packet
	packet := utils.Decoder(receivedMessageRaw)

	// Handle the packet and generate message
	message := Handler(packet)
	if message != "" {
		fmt.Printf("%s\n", message)
	}

	// Create ACK packet
	ackPacket := packet
	ackPacket.IsAck = true
	ackPacket.TransactionId = packet.TransactionId ^ 0xABCD // XOR with 0xABCD as specified

	// Encode and send ACK
	ackData := utils.Encoder(ackPacket)
	writeLength, err := pp.Writer.Write(ackData)

	fmt.Printf("[quic] Sent ACK for Transaction ID: %d\n", packet.TransactionId)

	return writeLength, err
}
