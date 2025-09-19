package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type LRTPIDSPacket struct {
	TransactionID    uint16
	IsAck            uint8
	IsNewTrain       uint8
	IsUpdateTrain    uint8
	IsDeleteTrain    uint8
	IsTrainArriving  uint8
	IsTrainDeparting uint8
	TrainNumber      uint16
	DestinationLength uint8
	Destination      string
}

func Encode(packet LRTPIDSPacket) ([]byte, error) {
	var buffer bytes.Buffer

	if err := binary.Write(&buffer, binary.BigEndian, packet.TransactionID); err != nil {
		return nil, fmt.Errorf("error encoding TransactionID: %v", err)
	}

	if err := binary.Write(&buffer, binary.BigEndian, packet.IsAck); err != nil {
		return nil, fmt.Errorf("error encoding IsAck: %v", err)
	}

	if err := binary.Write(&buffer, binary.BigEndian, packet.IsNewTrain); err != nil {
		return nil, fmt.Errorf("error encoding IsNewTrain: %v", err)
	}

	if err := binary.Write(&buffer, binary.BigEndian, packet.IsUpdateTrain); err != nil {
		return nil, fmt.Errorf("error encoding IsUpdateTrain: %v", err)
	}

	if err := binary.Write(&buffer, binary.BigEndian, packet.IsDeleteTrain); err != nil {
		return nil, fmt.Errorf("error encoding IsDeleteTrain: %v", err)
	}

	if err := binary.Write(&buffer, binary.BigEndian, packet.IsTrainArriving); err != nil {
		return nil, fmt.Errorf("error encoding IsTrainArriving: %v", err)
	}

	if err := binary.Write(&buffer, binary.BigEndian, packet.IsTrainDeparting); err != nil {
		return nil, fmt.Errorf("error encoding IsTrainDeparting: %v", err)
	}

	if err := binary.Write(&buffer, binary.BigEndian, packet.TrainNumber); err != nil {
		return nil, fmt.Errorf("error encoding TrainNumber: %v", err)
	}

	packet.DestinationLength = uint8(len(packet.Destination))
	if err := binary.Write(&buffer, binary.BigEndian, packet.DestinationLength); err != nil {
		return nil, fmt.Errorf("error encoding DestinationLength: %v", err)
	}

	if _, err := buffer.WriteString(packet.Destination); err != nil {
		return nil, fmt.Errorf("error encoding Destination: %v", err)
	}

	return buffer.Bytes(), nil
}

func Decode(data []byte) (LRTPIDSPacket, error) {
	var packet LRTPIDSPacket
	buffer := bytes.NewReader(data)

	if err := binary.Read(buffer, binary.BigEndian, &packet.TransactionID); err != nil {
		return packet, fmt.Errorf("error decoding TransactionID: %v", err)
	}

	if err := binary.Read(buffer, binary.BigEndian, &packet.IsAck); err != nil {
		return packet, fmt.Errorf("error decoding IsAck: %v", err)
	}

	if err := binary.Read(buffer, binary.BigEndian, &packet.IsNewTrain); err != nil {
		return packet, fmt.Errorf("error decoding IsNewTrain: %v", err)
	}

	if err := binary.Read(buffer, binary.BigEndian, &packet.IsUpdateTrain); err != nil {
		return packet, fmt.Errorf("error decoding IsUpdateTrain: %v", err)
	}

	if err := binary.Read(buffer, binary.BigEndian, &packet.IsDeleteTrain); err != nil {
		return packet, fmt.Errorf("error decoding IsDeleteTrain: %v", err)
	}

	if err := binary.Read(buffer, binary.BigEndian, &packet.IsTrainArriving); err != nil {
		return packet, fmt.Errorf("error decoding IsTrainArriving: %v", err)
	}

	if err := binary.Read(buffer, binary.BigEndian, &packet.IsTrainDeparting); err != nil {
		return packet, fmt.Errorf("error decoding IsTrainDeparting: %v", err)
	}

	if err := binary.Read(buffer, binary.BigEndian, &packet.TrainNumber); err != nil {
		return packet, fmt.Errorf("error decoding TrainNumber: %v", err)
	}

	if err := binary.Read(buffer, binary.BigEndian, &packet.DestinationLength); err != nil {
		return packet, fmt.Errorf("error decoding DestinationLength: %v", err)
	}

	if packet.DestinationLength > 0 {
		destinationBytes := make([]byte, packet.DestinationLength)
		if _, err := buffer.Read(destinationBytes); err != nil {
			return packet, fmt.Errorf("error decoding Destination: %v", err)
		}
		packet.Destination = string(destinationBytes)
	} else {
		packet.Destination = ""
	}

	return packet, nil
}