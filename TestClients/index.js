const WebSocket = require("ws");

function connect(id) {
  let ws = new WebSocket("ws://127.0.0.1:8080/");
  ws.onmessage = (m) =>
    console.log("Received message " + JSON.stringify(JSON.parse(m.data)));
  ws.onopen = function (e) {
    console.log("[open] Connection established");
    console.log("Sending to server");
    ws.send(
      JSON.stringify({
        jobId: "JOB" + id,
        urlToCrawl: "https://www.twitter.com/",
        depthToCrawl: 1,
      })
    );
  };
}

function connectN(N) {
  for (let i = 1; i <= N; i++) {
    connect(i);
  }
}

(async () => connectN(10))();

(async () => await Promise.resolve(setTimeout(() => {}, 1000 * 1000)))();
