package main

import (
	"server/cluster/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)



func main() {
	connect()
}

func connect() error{
	amqpUrl := "amqp://guest:guest@localhost:5672/"  //os.Getenv("AMQP_URL")

	conn, err := amqp.Dial(amqpUrl)
	if err != nil{
		logger.LogError(logger.MESSAGE_Q, "Exiting because of err %v", err)
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil{
		logger.LogError(logger.MESSAGE_Q, "Exiting because of err %v", err)
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
			"jobs", 
			false,
			false,
			false,
			false,
			nil)		

	return nil
}