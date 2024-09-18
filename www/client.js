const WEBSOCKET = "/ws";
const API = "";

const getSocket = () => {
  const socket = new WebSocket(`${WEBSOCKET}`);

  socket.onopen = () => {
    console.log("Connection established");
  };

  socket.onmessage = ({ data }) => {
    if (data.indexOf("set:") === 0) {
      const [_, cell, checked] = data.split(":");
      const input = document.querySelector(`input[name="${cell}"]`);
      input.checked = checked === "1" ? true : false;
    } else {
      const binaryString = atob(data);
      const len = binaryString.length;
      const bytes = new Uint8Array(len);

      for (let i = 0; i < len; i++) {
        bytes[i] = binaryString.charCodeAt(i);
      }

      buildGrid(bytesToBitArray(bytes));
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

function getCount() {
  fetch(`${API}/ws/count`)
    .then((response) => response.text())
    .then((data) => {
      document.querySelector("#count").textContent = data;
    });
}
getCount();
setInterval(getCount, 10000);

function buildGrid(bits) {
  const fragment = document.createDocumentFragment();

  for (let i = 0; i < bits.length; i++) {
    const input = document.createElement("input");
    input.type = "checkbox";
    input.name = i;
    input.checked = bits[i] === 1 ? true : false;
    input.onclick = (event) => {
      event.preventDefault();
      ws.send(`set:${i}`);
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
