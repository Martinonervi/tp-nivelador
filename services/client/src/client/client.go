package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/lottery"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/protocol"
)

const CONNECTION_ATTEMPTS_MAX = 3
const CONNECTION_ATTEMPS_DELAY_MS = 200

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   string
	InputFile  string
	OutputFile string
	BatchSize  int
}

type Client struct {
	config ClientConfig
	proto  *protocol.Protocol
}

func NewClient(config ClientConfig) (*Client, error) {
	protocol, err := connectToServer(config.ServerHost, config.ServerPort)
	if err != nil {
		logger.Warn("connect-to-server", logger.Fail)
		return nil, err
	}

	client := &Client{config: config, proto: protocol}
	return client, nil
}

func connectToServer(host, port string) (*protocol.Protocol, error) {
	const action = "connect-to-server"
	var err error
	var conn net.Conn
	var p *protocol.Protocol

	logger.Info(action, logger.InProgress)
	for i := range CONNECTION_ATTEMPTS_MAX {
		conn, err = net.Dial("tcp", host+":"+port)
		if err != nil {
			logger.Warn(action, logger.Fail, "attempt", i)
			time.Sleep(CONNECTION_ATTEMPS_DELAY_MS * time.Millisecond)
			continue
		}
		logger.Info(action, logger.Success)
		break
	}
	if err != nil {
		return nil, err
	}
	p, err = protocol.NewProtocol(conn)

	return p, err
}

func (client *Client) Run() error {
	defer client.proto.Close()
	const mainAction = "test-echo-server"

	inputFile, err := os.Open(client.config.InputFile)
	if err != nil {
		logger.Error("file error", logger.Fail, err)
		return err
	}
	defer inputFile.Close()

	outputFile, err := os.Create(client.config.OutputFile)
	if err != nil {
		logger.Error("file error", logger.Fail, err)
		return err
	}
	defer outputFile.Close()
	dataWriter := bufio.NewWriter(outputFile)

	scanner := bufio.NewScanner(inputFile)
	agencyId, err := strconv.Atoi(client.config.AgencyId)
	if err != nil {
		return err
	}

	listOfBets := []lottery.Bet{}
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ",")
		if len(fields) < 5 {
			logger.Error("bad line", logger.Fail, line)
			continue
		}
		document, err := strconv.Atoi(fields[2])
		if err != nil {
			return err
		}
		betNumber, err := strconv.Atoi(fields[4])
		if err != nil {
			return err
		}

		bet := lottery.Bet{
			AgencyId:  agencyId,
			FirstName: fields[0],
			LastName:  fields[1],
			Document:  document,
			Birthdate: fields[3],
			Number:    betNumber,
		}
		listOfBets = append(listOfBets, bet)

		if len(listOfBets) == client.config.BatchSize {
			if err := client.proto.SendBets(listOfBets); err != nil {
				return err
			}
			listOfBets = []lottery.Bet{}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	if len(listOfBets) > 0 {
		if err := client.proto.SendBets(listOfBets); err != nil {
			return err
		}
	}

	err = client.proto.SendNoMoreBets()
	if err != nil {
		return err
	}

	for {
		ListOfBets, moreBets, err := client.proto.RecvWinners()
		if err != nil {
			return err
		}
		if err := writeBetsToFile(dataWriter, ListOfBets); err != nil {
			return err
		}
		if !moreBets {
			break
		}
	}

	if err := dataWriter.Flush(); err != nil {
		logger.Error("flush-output", logger.Fail)
		return err
	}
	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId)
	return nil
}

func writeBetsToFile(writer *bufio.Writer, bet []lottery.Bet) error {
	for _, bet := range bet {
		line := fmt.Sprintf("%s,%s,%d,%s,%d\n", bet.FirstName, bet.LastName, bet.Document, bet.Birthdate, bet.Number)
		_, err := writer.WriteString(line)
		if err != nil {
			return err
		}
	}
	return nil

}
