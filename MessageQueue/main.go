package MessageQueue

import (
	logger "Distributed_Web_Crawler/Logger"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MQ struct{
	conn *amqp.Connection 
	ch *amqp.Channel
	jobsQ *amqp.Queue
	jobsFinishedQ *amqp.Queue
}


// func main() {
// 	amqpAddr := "amqp://guest:guest@localhost:5672/"  //os.Getenv("AMQP_URL")
// 	mQ := New(amqpAddr)
// 	mQ.Publish(JOBS_QUEUE, "hiiiiiii there")
// 	mQ.Close()
// }

func New(amqpAddr string) *MQ{
	mq := &MQ{
		conn: nil,
		ch: nil,
		jobsQ: nil,
		jobsFinishedQ: nil,
	}
	//try to reconnect and sleep 2 seconds on failure
	var err error = fmt.Errorf("error")
	ctr := 0
	for ctr < 3 && err != nil{
		err = mq.connect(amqpAddr)
		ctr++
		time.Sleep(time.Second * 2)
	}
	
	return mq
}

func(mq *MQ) Close() {
	mq.ch.Close()
	mq.conn.Close()
}

func (mq *MQ) Publish(qName string, body []byte) error{

	//if jobsFinishedQ is nill, declare it
	if mq.jobsFinishedQ == nil{
		q, err := mq.ch.QueueDeclare(
			qName, 
			true,  //durable
			false,  //autoDelete
			false,  //exclusive -> send errors when another consumer tries to connect to it
			false, //noWait
			nil,
		)		
		if err != nil{
			return err
		}
		mq.jobsFinishedQ = &q
	}
	

	err := mq.ch.Publish(
		"", //exchange,   empty is default
		mq.jobsFinishedQ.Name, //routing key
		false, //mandatory 
		false, //immediate 
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType: "text/plain",
			Body: body,
		},
	)
	if err != nil{
		return err
	}

	logger.LogInfo(logger.MESSAGE_Q, "Successfully published to queue %v a message with this data:\n%v", qName, string(body))
	return nil
}

func (mq *MQ) Consume(qName string) (<-chan amqp.Delivery, error) {

	//if jobsQ is nill, declare it
	if mq.jobsQ == nil{
		q, err := mq.ch.QueueDeclare(
			qName, 
			true,  //durable
			false,  //autoDelete
			false,  //exclusive -> send errors when another consumer tries to connect to it
			false, //noWait
			nil,
		)		
		if err != nil{
			return nil, err
		}
		mq.jobsQ = &q
	}
	err := mq.ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil{
		return nil, err
	}
	msgs, err := mq.ch.Consume(
		mq.jobsQ.Name,
		"", //consumer --> unique identifier that rabbit can just generate
		false, //auto ack
		false, //exclusive --> errors if other consumers consume
		false, //nolocal, dont receive back if i send on channel
		false, //no wait
		nil,
	)
	if err != nil{
		return nil, err
	}

	logger.LogInfo(logger.MESSAGE_Q, "Successfully subscribed to queue %v", qName)
	return msgs, nil
}


func(mq *MQ) connect(amqpAddr string) error{

	conn, err := amqp.Dial(amqpAddr)
	if err != nil{
		logger.FailOnError(logger.MESSAGE_Q, "Exiting because of err while establishing connection with message queue %v", err)
	}
	mq.conn = conn


	ch, err := conn.Channel()
	if err != nil{
		logger.FailOnError(logger.MESSAGE_Q, "Exiting because of err while opening a channel %v", err)
	}
	mq.ch = ch

	logger.LogInfo(logger.MESSAGE_Q, "Connection to message queue has been successfully established")
	return nil
}