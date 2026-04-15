// Dark-mode toggle
function toggleDark() {
  var next = document.documentElement.dataset.theme === 'dark' ? 'light' : 'dark';
  document.documentElement.dataset.theme = next;
  localStorage.setItem('dark', next);
}

// Client-side search
(function () {
  var idx = null;
  var timer = null;
  var input = document.getElementById('search-input');
  var box = document.getElementById('search-results');
  if (!input || !box) return;

  // Fetch index once on load so searches work fully offline after.
  fetch('/_search_index.json')
    .then(function (r) { return r.json(); })
    .then(function (data) { idx = data; })
    .catch(function () {});

  input.addEventListener('input', function () {
    clearTimeout(timer);
    var q = input.value.trim();
    if (!q) { hide(); return; }
    timer = setTimeout(function () { search(q); }, 200);
  });

  input.addEventListener('keydown', function (e) {
    // Let Enter submit the form (server-side search fallback).
    if (e.key === 'Escape') { hide(); input.blur(); }
  });

  document.addEventListener('click', function (e) {
    if (!box.contains(e.target) && e.target !== input) hide();
  });

  function search(q) {
    if (!idx) { hide(); return; }
    var lower = q.toLowerCase();
    var results = [];
    for (var i = 0; i < idx.length; i++) {
      var e = idx[i];
      var score = 0;
      if (e.title.toLowerCase().indexOf(lower) >= 0) score += 2;
      if (e.snippet.toLowerCase().indexOf(lower) >= 0) score += 1;
      if (score > 0) results.push({ e: e, score: score });
    }
    results.sort(function (a, b) { return b.score - a.score; });
    render(results.slice(0, 12), q);
  }

  function render(results, q) {
    if (results.length === 0) {
      box.innerHTML = '<div class="no-match">No results</div>';
    } else {
      var html = '';
      for (var i = 0; i < results.length; i++) {
        var e = results[i].e;
        html += '<a href="/' + esc(e.path) + '"><strong>' + esc(e.title) + '</strong>'
          + (e.snippet ? '<span class="snip">' + esc(e.snippet.slice(0, 80)) + '</span>' : '')
          + '</a>';
      }
      box.innerHTML = html;
    }
    box.hidden = false;
  }

  function hide() {
    box.hidden = true;
    box.innerHTML = '';
  }

  function esc(s) {
    return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
  }
})();
