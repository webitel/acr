/**
 * Created by I. Navrotskyj on 01.09.17.
 */

package rpc

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"github.com/tidwall/gjson"
	"github.com/webitel/acr/src/pkg/config"
	"github.com/webitel/acr/src/pkg/logger"
	"sync"
	"time"
)

type ApiArgsT struct {
	json.RawMessage
	CallUuid string `json:"callId"`
	SetVar   string `json:"setVar"`
}

type ApiT struct {
	Api  string          `json:"exec-api"`
	Args json.RawMessage `json:"exec-args"`
}

const exchangeEventName = "ACR.Event"
const exchangeEventFormat = "ACR-Hostname,Event-Name,Event-Subclass,Domain"

const exchangeCommandsName = "ACR.Commands"
const exchangeCommandsFormat = "acr.commands.inbound"

type RPC struct {
	sync.Mutex
	queueCommands amqp.Queue
	queueEvents   amqp.Queue
	callbacks     map[string]chan ApiT
	channel       *amqp.Channel
	connection    *amqp.Connection
}

func (rpc *RPC) GetCommandsQueueName() string {
	return rpc.queueCommands.Name
}

func (rpc *RPC) AddCommands(uuid string) ApiT {
	rpc.Lock()
	rpc.callbacks[uuid] = make(chan ApiT, 1)
	rpc.Unlock()
	return <-rpc.callbacks[uuid]
}

func (rpc *RPC) RemoveCommands(uuid string, data ApiT) {
	if c, ok := rpc.callbacks[uuid]; ok {
		rpc.Lock()
		delete(rpc.callbacks, uuid)
		rpc.Unlock()
		c <- data
	}
}

func (rpc *RPC) initExchange() error {
	err := rpc.channel.ExchangeDeclare(
		exchangeCommandsName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		logger.Error("Create exchange %s error: %s", exchangeCommandsName, err.Error())
		return err
	}

	rpc.queueCommands, err = rpc.channel.QueueDeclare(
		"",
		true,
		true,
		true,
		false,
		nil,
	)

	if err != nil {
		logger.Error("Create queue exchange %s error: %s", exchangeCommandsName, err.Error())
		return err
	}

	err = rpc.channel.QueueBind(
		rpc.queueCommands.Name,
		exchangeCommandsFormat,
		exchangeCommandsName,
		false,
		nil,
	)

	if err != nil {
		logger.Error("Bind queue %s to exchange %s error: %s", rpc.queueCommands.Name, exchangeCommandsName, err.Error())
		return err
	}

	msgs, err := rpc.channel.Consume(
		rpc.queueCommands.Name,
		"",
		false,
		true,
		false,
		false,
		nil,
	)

	if err != nil {
		logger.Error("Create consume queue %s exchange %s error: %s", rpc.queueCommands.Name, exchangeCommandsName, err.Error())
		return err
	}

	logger.Info("Success init exchange commands")
	go func() {
		msg := ApiT{}
		//var ok bool
		var err error
		//var id string
		for d := range msgs {
			logger.Debug("Received a message: %s", d.Body)

			if err = json.Unmarshal(d.Body, &msg); err != nil {
				logger.Error("Read response amqp error: %s", err.Error())
				continue
			}

			rpc.RemoveCommands(gjson.GetBytes(msg.Args, "callId").String(), msg)

			d.Ack(false)
		}

		rpc.reconnect()
	}()

	return nil
}

func (rpc *RPC) Fire(body []byte, rk string) error {
	logger.Debug("RPC: send to engine %d bytes %s", len(body), body)
	return rpc.channel.Publish(
		"engine",
		rk,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			ContentType:  "text/plain",
			Body:         body,
		},
	)
}

func (rpc *RPC) connect() error {
	connectionString := config.Conf.Get("broker:connectionString")
	if connectionString == "" {
		panic("Bad broker connectionString config.")
	}
	conn, err := amqp.Dial(connectionString)
	if err != nil {
		logger.Error("Connect to %s error: %s", connectionString, err.Error())
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		logger.Error("Create channel to %s error: %s", connectionString, err.Error())
		return err
	}

	rpc.connection = conn
	rpc.channel = ch
	rpc.initExchange()
	return nil
}

func (rpc *RPC) reconnect() {
	var err error
	for {
		logger.Debug("Try reconnect to amqp")
		if err = rpc.connect(); err == nil {
			return
		}
		time.Sleep(time.Second)
	}
}

func New() *RPC {

	r := &RPC{
		callbacks: make(map[string]chan ApiT),
	}

	if err := r.connect(); err != nil {
		r.reconnect()
	}

	return r
}
