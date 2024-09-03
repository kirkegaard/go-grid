const grid = document.querySelector("#grid");

const form = document.querySelector("#form");
form.addEventListener("change", send);

function send(event) {
  const { name } = event.target;
  fetch("http://localhost:6060/set", {
    method: "POST",
    body: { cell: name },
  })
    .then((res) => {
      console.log(res);
    })
    .catch((err) => {
      console.error(err);
    });
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
