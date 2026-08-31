package client

import (
	"bufio"
	"net"
	"os"
	"time"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/protocol"
)

const CONNECTION_ATTEMPTS_MAX = 3
const CONNECTION_ATTEMPS_DELAY_MS = 200

const ECHO_CLIENT_MESSAGE_AMOUNT = 3
const ECHO_CLIENT_MESSAGE_DELAY_MS = 1000

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   string
	InputFile  string
	OutputFile string
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
	for scanner.Scan() {
		line := scanner.Text()

		err := client.proto.Send(line)
		if err != nil {
			return err
		}

		response, err := client.proto.Recv()
		if err != nil {
			return err
		}

		_, err = dataWriter.WriteString(response + "\n")
		if err != nil {
			logger.Error("write-response", logger.Fail)
			return err
		}
	}
	if err := dataWriter.Flush(); err != nil {
		logger.Error("flush-output", logger.Fail)
		return err
	}
	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId)
	err = client.proto.Close()
	if err != nil {
		return err
	}
	return nil
}
