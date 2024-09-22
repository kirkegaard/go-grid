const WEBSOCKET = "ws://localhost:6060/ws";
const API = "";

let clientId = "";

const players = [];

const getSocket = () => {
  const socket = new WebSocket(`${WEBSOCKET}`);

  socket.onopen = () => {
    console.log("Connection established");
  };

  socket.onmessage = ({ data }) => {
    const [type, ...payload] = data.split(":");
    switch (type) {
      case "c":
        clientId = payload[0];
        break;

      case "r": {
        const [id] = payload;
        players.push({ id, x: 0, y: 0 });

        addPlayer(id);
        updateCount();

        break;
      }

      case "s": {
        const [cell, checked] = payload;
        const input = document.querySelector(`input[name="${cell}"]`);
        input.checked = checked === "1" ? true : false;
        break;
      }

      case "d": {
        const [id] = payload;
        const index = players.findIndex((player) => player.id === id);
        players.splice(index, 1);

        removePlayer(id);
        updateCount();

        break;
      }

      case "p": {
        const [id, x, y] = payload;
        const p = players.find((player) => player.id === id);

        if (!p) {
          players.push({ id, x: 0, y: 0 });
          addPlayer(id);
          updateCount();
          return;
        }

        p.x = x;
        p.y = y;

        movePlayer(id, x, y);

        break;
      }

      default: {
        const binaryString = atob(data);
        const len = binaryString.length;
        const bytes = new Uint8Array(len);

        for (let i = 0; i < len; i++) {
          bytes[i] = binaryString.charCodeAt(i);
        }

        buildGrid(bytesToBitArray(bytes));
        break;
      }
    }
  };

  socket.onclose = () => {
    console.log("Connection closed");
    setTimeout(() => {
      getSocket();
    }, 1000);
  };

  return socket;
};

const ws = getSocket();
const grid = document.querySelector("#grid");

document.querySelector("#toggle").onclick = (event) => {
  event.preventDefault();
  grid.classList.toggle("fixed");
};

function handleMouseMove(event) {
  if (!clientId) return;
  const { x, y } = getCursorPosition(event);
  ws.send(`p:${clientId}:${x}:${y}`);
}
document.addEventListener("mousemove", throttle(handleMouseMove, 100));

const p = document.querySelector("#players");
function addPlayer(id) {
  if (
    players.includes(id) ||
    document.querySelector(`.player[data-id="${id}"]`) !== null
  ) {
    return;
  }

  const player = document.createElement("div");
  player.dataset.id = id;
  player.classList.add("player");
  player.innerHTML = id;

  p.appendChild(player);
}

function removePlayer(id) {
  const player = document.querySelector(`.player[data-id="${id}"]`);
  if (!player) return;
  p.removeChild(player);
}

function movePlayer(id, x, y) {
  const player = document.querySelector(`.player[data-id="${id}"]`);
  if (!player) return;
  player.style.left = `${x}px`;
  player.style.top = `${y}px`;
}

function updateCount() {
  document.querySelector("#count").innerHTML = players.length;
}

function buildGrid(bits) {
  const fragment = document.createDocumentFragment();

  for (let i = 0; i < bits.length; i++) {
    const input = document.createElement("input");
    input.type = "checkbox";
    input.name = i;
    input.checked = bits[i] === 1 ? true : false;
    input.onclick = (event) => {
      event.preventDefault();
      ws.send(`s:${i}`);
      input.disabled = true;
      setTimeout(() => (input.disabled = false), 250);
    };
    fragment.appendChild(input);
  }

  grid.innerHTML = "";
  grid.appendChild(fragment);
}

function bytesToBitArray(bytes) {
  let bitArray = [];
  bytes.forEach((byte) => {
    for (let i = 0; i < 8; i++) {
      bitArray.push((byte >> (7 - i)) & 1);
    }
  });
  return bitArray;
}

// Get cursor position relative to body
function getCursorPosition(event) {
  const x = event.clientX;
  const y = event.clientY;
  return { x, y };
}

let throttleTimer;
function throttle(fn, delay) {
  return function (...args) {
    if (!throttleTimer) {
      fn.apply(this, args);
      throttleTimer = setTimeout(() => {
        throttleTimer = null;
      }, delay);
    }
  };
}
