const WebSocket = require("ws");
let conns = [];
function connect(id) {
  let ws = new WebSocket("ws://127.0.0.1:8080/");
  //conns.push(ws);
  ws.onmessage = (m) =>
    console.log("Received message " + JSON.stringify(JSON.parse(m.data)));
  ws.onopen = function (e) {
    console.log("[open] Connection established");
    console.log("Sending to server");
    ws.send(
      JSON.stringify({
        jobId: "JOB" + id,
        urlToCrawl: "https://www.instagram.com/",
        depthToCrawl: 1,
      })
    );
    ws.close();
  };
}
for (let i = 1; i <= 500; i++) {
  connect(i);
  (async () => await Promise.resolve(setTimeout(() => {}, 50)))();
}

(async () => await Promise.resolve(setTimeout(() => {}, 1000 * 1000)))();
