(function () {
  var script = document.currentScript;
  var name = script.dataset.name;
  var editor = document.getElementById("editor");
  var preview = document.getElementById("preview");
  var status = document.getElementById("status");
  var saveBtn = document.getElementById("save");
  var previewTimer;
  var autosaveTimer;
  var lastSaved = editor.value;
  var saveSeq = 0;

  var PREVIEW_DELAY_MS = 300;
  var AUTOSAVE_DELAY_MS = 800;

  function renderPreview() {
    fetch("/api/preview", { method: "POST", body: editor.value })
      .then(function (r) { return r.text(); })
      .then(function (html) { preview.innerHTML = html; })
      .catch(function () { /* preview is best-effort */ });
  }

  // save persists the editor's current content: to the .md file and, by
  // re-parsing, back to the .yaml file. It also fires whenever typing
  // pauses (see AUTOSAVE_DELAY_MS below), which is what keeps a running
  // `mdgen view` in sync without requiring an explicit Save click.
  function save() {
    var content = editor.value;
    if (content === lastSaved) {
      return;
    }
    var seq = ++saveSeq;
    status.textContent = "Saving…";
    fetch("/api/save/" + encodeURIComponent(name), { method: "POST", body: content })
      .then(function (r) {
        if (!r.ok) {
          return r.text().then(function (t) { throw new Error(t || r.statusText); });
        }
        return r.json();
      })
      .then(function () {
        lastSaved = content;
        if (seq !== saveSeq) return; // a newer save has since started
        status.textContent = "Saved";
        setTimeout(function () {
          if (seq === saveSeq) status.textContent = "";
        }, 1500);
      })
      .catch(function (err) {
        if (seq !== saveSeq) return;
        status.textContent = "Error: " + err.message;
      });
  }

  editor.addEventListener("input", function () {
    clearTimeout(previewTimer);
    previewTimer = setTimeout(renderPreview, PREVIEW_DELAY_MS);

    clearTimeout(autosaveTimer);
    autosaveTimer = setTimeout(save, AUTOSAVE_DELAY_MS);
  });

  saveBtn.addEventListener("click", function () {
    clearTimeout(autosaveTimer);
    save();
  });

  document.addEventListener("keydown", function (e) {
    if ((e.metaKey || e.ctrlKey) && e.key === "s") {
      e.preventDefault();
      clearTimeout(autosaveTimer);
      save();
    }
  });

  renderPreview();
})();
