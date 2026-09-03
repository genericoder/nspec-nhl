(function () {
  var script = document.currentScript;
  var name = script.dataset.name;
  var editor = document.getElementById("editor");
  var preview = document.getElementById("preview");
  var status = document.getElementById("status");
  var saveBtn = document.getElementById("save");
  var debounceTimer;

  function renderPreview() {
    fetch("/api/preview", { method: "POST", body: editor.value })
      .then(function (r) { return r.text(); })
      .then(function (html) { preview.innerHTML = html; })
      .catch(function () { /* preview is best-effort */ });
  }

  function save() {
    status.textContent = "Saving…";
    fetch("/api/save/" + encodeURIComponent(name), { method: "POST", body: editor.value })
      .then(function (r) {
        if (!r.ok) {
          return r.text().then(function (t) { throw new Error(t || r.statusText); });
        }
        return r.json();
      })
      .then(function () {
        status.textContent = "Saved";
        setTimeout(function () { status.textContent = ""; }, 1500);
      })
      .catch(function (err) {
        status.textContent = "Error: " + err.message;
      });
  }

  editor.addEventListener("input", function () {
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(renderPreview, 300);
  });

  saveBtn.addEventListener("click", save);

  document.addEventListener("keydown", function (e) {
    if ((e.metaKey || e.ctrlKey) && e.key === "s") {
      e.preventDefault();
      save();
    }
  });

  renderPreview();
})();
