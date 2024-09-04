// const ws = new WebSocket("ws://localhost:6060/ws");
//
// ws.onmessage = (event) => {
//   const data = JSON.parse(event.data);
//   console.log(data);
// };

const API = "";

const grid = document.querySelector("#grid");

const form = document.querySelector("#form");
form.addEventListener("change", send);

async function send(event) {
  const { name } = event.target;
  try {
    const res = await fetch(`${API}/api/set`, {
      method: "POST",
      headers: {
        "Content-Type": "application/x-www-form-urlencoded",
      },
      body: `cell=${name}`,
    });
    const data = await res.text();
    console.log(data);
  } catch (err) {
    console.error(err);
  }
}

async function get() {
  try {
    const res = await fetch(`${API}/api/get`);
    const data = await res.text();
    return reverseXOR(hexToBytes(data));
  } catch (err) {
    console.error(err);
  }
}

async function createGrid() {
  const fragment = document.createDocumentFragment();

  const bits = await get();

  for (let i = 0; i < 25 * 25; i++) {
    const input = document.createElement("input");
    input.type = "checkbox";
    input.name = i;
    input.value = 1;
    input.checked = bits[i] === 1 ? true : false;
    fragment.appendChild(input);
  }

  grid.innerHTML = "";
  grid.appendChild(fragment);
}

// Function to convert hex string to byte array
function hexToBytes(hex) {
  const bytes = [];
  for (let i = 0; i < hex.length; i += 2) {
    bytes.push(parseInt(hex.substr(i, 2), 16));
  }
  return bytes;
}

// Reverse the XOR operation
function reverseXOR(bytes) {
  const gridSize = 625;
  const bits = [];

  for (let i = 0; i < gridSize; i++) {
    // Determine which byte the bit is in and which bit position within that byte
    const byteIndex = Math.floor(i / 8);
    const bitPosition = i % 8;

    // XOR operation to reverse the bit
    const bit = (bytes[byteIndex] >> bitPosition) & 1;
    bits.push(bit);
  }

  return bits;
}

setInterval(createGrid, 2500);
createGrid();
