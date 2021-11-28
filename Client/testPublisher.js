const ampq = require("amqplib");

const testJob = {
  website: "https://google.com",
  numToCrawl: 10,
};
const testQueueName = "jobs";

async function connect() {
  try {
    const connection = await ampq.connect("amqp://localhost:5672"); //create connection
    const channel = await connection.createChannel(); //create a channel
    await channel.assertQueue(testQueueName); //assert queue exists, and create it if it doesnt
    channel.sendToQueue(testQueueName, Buffer.from(JSON.stringify(testJob)));
    console.log(`Job sent successfully ${testJob}`);
  } catch (err) {
    console.log(err);
  }
}

connect();
