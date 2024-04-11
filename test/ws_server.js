// ws://127.0.0.1:18099

const ws = require("ws");
const wss = new ws.WebSocketServer({ port: 18099 });

wss.on("connection", function connection(ws) {
  console.log(`new connection`);
  ws.isAlive = true;
  ws.on("error", console.error);
  ws.on("pong", () => {
    console.log("on pong");
  });
});

const interval = setInterval(() => {
  wss.clients.forEach(function each(ws) {
    ws.send(JSON.stringify({ test: new Date().getTime() }));
  });
}, 1000 * 10);

wss.on("close", function close() {
  clearInterval(interval);
});
