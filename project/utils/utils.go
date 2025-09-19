package utils

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"io"
	"log"
	"math/big"
)

// Tipe data komponen boleh diubah, namun variabelnya jangan diubah
type LRTPIDSPacketFixed struct {
	TransactionId     uint16
	IsAck             bool
	IsNewTrain        bool
	IsUpdateTrain     bool
	IsDeleteTrain     bool
	IsTrainArriving   bool
	IsTrainDeparting  bool
	TrainNumber       uint16
	DestinationLength uint8
}

type LRTPIDSPacket struct {
	LRTPIDSPacketFixed
	Destination string
}

func Encoder(packet LRTPIDSPacket) []byte {
	buffer := new(bytes.Buffer)

	// Write TransactionId (16 bit, Big Endian)
	binary.Write(buffer, binary.BigEndian, packet.TransactionId)

	// Pack the 6 flag bits into a single byte
	var flags uint8 = 0
	if packet.IsAck {
		flags |= (1 << 7) // bit 7
	}
	if packet.IsNewTrain {
		flags |= (1 << 6) // bit 6
	}
	if packet.IsUpdateTrain {
		flags |= (1 << 5) // bit 5
	}
	if packet.IsDeleteTrain {
		flags |= (1 << 4) // bit 4
	}
	if packet.IsTrainArriving {
		flags |= (1 << 3) // bit 3
	}
	if packet.IsTrainDeparting {
		flags |= (1 << 2) // bit 2
	}
	// bits 1 and 0 are unused/reserved
	buffer.WriteByte(flags)

	// Write TrainNumber (16 bit, Little Endian as specified)
	binary.Write(buffer, binary.LittleEndian, packet.TrainNumber)

	// Write DestinationLength (8 bit)
	buffer.WriteByte(packet.DestinationLength)

	// Write Destination string
	buffer.WriteString(packet.Destination)

	// Add padding to make total length of Destination + padding a multiple of 4
	destLen := len(packet.Destination)
	paddingNeeded := (4 - (destLen % 4)) % 4
	for i := 0; i < paddingNeeded; i++ {
		buffer.WriteByte(0xFF)
	}

	return buffer.Bytes()
}

func Decoder(rawMessage []byte) LRTPIDSPacket {
	if len(rawMessage) < 6 {
		return LRTPIDSPacket{}
	}

	buffer := bytes.NewReader(rawMessage)
	packet := LRTPIDSPacket{}

	// Read TransactionId (16 bit, Big Endian)
	binary.Read(buffer, binary.BigEndian, &packet.TransactionId)

	// Read flags byte
	var flags uint8
	binary.Read(buffer, binary.BigEndian, &flags)

	// Extract flag bits
	packet.IsAck = (flags & (1 << 7)) != 0
	packet.IsNewTrain = (flags & (1 << 6)) != 0
	packet.IsUpdateTrain = (flags & (1 << 5)) != 0
	packet.IsDeleteTrain = (flags & (1 << 4)) != 0
	packet.IsTrainArriving = (flags & (1 << 3)) != 0
	packet.IsTrainDeparting = (flags & (1 << 2)) != 0

	// Read TrainNumber (16 bit, Little Endian)
	binary.Read(buffer, binary.LittleEndian, &packet.TrainNumber)

	// Read DestinationLength (8 bit)
	binary.Read(buffer, binary.BigEndian, &packet.DestinationLength)

	// Read Destination string
	if packet.DestinationLength > 0 && len(rawMessage) >= 6+int(packet.DestinationLength) {
		destBytes := make([]byte, packet.DestinationLength)
		buffer.Read(destBytes)
		packet.Destination = string(destBytes)
	}

	return packet
}

type qlogWriter struct {
	*bufio.Writer
	io.Closer
}

func GenerateTLSSelfSignedCertificates() []tls.Certificate {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalln(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		log.Fatalln(err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		log.Fatalln(err)
	}
	return []tls.Certificate{tlsCert}
}

