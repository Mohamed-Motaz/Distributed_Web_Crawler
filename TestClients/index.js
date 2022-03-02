const WebSocket = require("ws");
let conns = [];
let ctr = 0;
function connect(id) {
  let ws = new WebSocket("ws://127.0.0.1:5555/");
  conns.push(ws);
  ws.onmessage = (m) => {
    console.log("Received message " + JSON.stringify(JSON.parse(m.data)));
    ctr++;
    ws.close();
  };
  ws.onopen = function (e) {
    console.log("[open] Connection established");
    console.log("Sending to server");
    ws.send(
      JSON.stringify({
        jobId: "JOB" + id,
        urlToCrawl: "https://www.facebook.com/",
        depthToCrawl: 1,
      })
    );
  };
}
for (let i = 1; i <= 10; i++) {
  connect(i);
  //   (async () => await Promise.resolve(setTimeout(() => {}, 50)))();
}

for (let i = 1; i < 100; i++) {
  (async () =>
    await Promise.resolve(
      setTimeout(() => {
        console.log(ctr);
      }, 10 * 1000)
    ))();
}

console.log(ctr);
