const grid = document.querySelector("#grid");

const form = document.querySelector("#form");
form.addEventListener("change", send);

async function send(event) {
  const { name } = event.target;
  try {
    const res = await fetch("http://localhost:6060/set", {
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

function createGrid() {
  for (let i = 0; i < 25 * 25; i++) {
    const input = document.createElement("input");
    input.type = "checkbox";
    input.name = i;
    input.value = 1;
    grid.appendChild(input);
  }
}

createGrid();
