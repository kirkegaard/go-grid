const API = "";
const ws = new WebSocket(`${API}/ws`);

ws.onopen = () => {
  console.log("Connected to the server");
};

ws.onmessage = (event) => {
  if (event.data.indexOf("set:") === 0) {
    const [_, cell, checked] = event.data.split(":");
    const input = document.querySelector(`input[name="${cell}"]`);
    input.checked = checked === "1" ? true : false;
  } else {
    // Setup grid
    const bits = reverseXOR(hexToBytes(event.data));
    setupGrid(bits);
  }
};

const form = document.querySelector("#form");
form.addEventListener("change", (event) => {
  event.preventDefault();
  const { name } = event.target;
  try {
    ws.send(`set:${name}`);
  } catch (err) {
    console.error(err);
  }
});

function setupGrid(bits) {
  const grid = document.querySelector("#grid");
  const fragment = document.createDocumentFragment();

  for (let i = 0; i < bits.length; i++) {
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
  const gridSize = bytes.length * 8;
  console.log(gridSize);
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
