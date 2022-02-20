const WebSocket = require("ws");

async function connect(id) {
  try {
    let ws = new WebSocket("ws://127.0.0.1:2222/jobs");
    ws.onmessage = (m) => console.log("Received message " + m);
    ws.onopen = function (e) {
      console.log("[open] Connection established");
      console.log("Sending to server");
      ws.send(
        JSON.stringify({
          clientId: "Client" + id,
          jobId: "JOB" + id,
          urlToCrawl: "https://www.google.com/",
          depthToCrawl: 1,
        })
      );
      ws.close();
    };
  } catch (err) {
    console.log(err);
  }
}

async function connectN(N) {
  for (let i = 1; i <= N; i++) {
    await connect(i);
  }
}

(async () => await connectN(10))();
