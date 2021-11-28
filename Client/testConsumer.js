const ampq = require("amqplib")

const testQueueName = "jobs"

async function connect(){
    try{
        const connection = await ampq.connect("amqp://localhost:5672")  //create connection
        const channel = await connection.createChannel() //create a channel
        await channel.assertQueue(testQueueName) //assert queue exists, and create it if it doesnt
        
        channel.consume(testQueueName, message => {
            console.log(`This message arrived ${message.content.toString()}`)
            channel.ack(message)
        })
        
        console.log("Waiting for messages...")
        
    }catch (err){
        console.log(err)
    }
}

connect()