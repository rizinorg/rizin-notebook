/* autocomplete.js — Rizin Notebook command autocomplete
 *
 * Provides inline autocomplete for rizin command input fields.
 * Fetches command data lazily from the server API and caches it.
 *
 * API:
 *   NotebookAutocomplete.attach(inputElement, options)
 *     options.apiUrl     — URL to fetch commands (default: inferred from page root)
 *     options.maxResults — maximum suggestions shown (default: 12)
 *     options.minChars   — minimum chars before suggesting (default: 2)
 */
(function (global, document) {
  "use strict";

  /* ── Command cache ──────────────────────────────────── */

  var _cache = null;
  var _fetching = null;

  /**
   * Fetch commands from the server API. Returns a sorted array of
   * { name: string, cmd: RizinCommand } objects. Cached after first call.
   */
  function fetchCommands(apiUrl) {
    if (_cache) return Promise.resolve(_cache);
    if (_fetching) return _fetching;
    _fetching = fetch(apiUrl)
      .then(function (r) {
        return r.json();
      })
      .then(function (data) {
        var list = [];
        for (var key in data) {
          if (data.hasOwnProperty(key)) {
            list.push({ name: key, cmd: data[key] });
          }
        }
        list.sort(function (a, b) {
          return a.name < b.name ? -1 : a.name > b.name ? 1 : 0;
        });
        _cache = list;
        _fetching = null;
        return list;
      })
      .catch(function () {
        _fetching = null;
        return [];
      });
    return _fetching;
  }

  /* ── Attach autocomplete to an input ────────────────── */

  function attach(input, options) {
    var opts = options || {};
    var apiUrl = opts.apiUrl || (global.NB_ROOT || "/") + "api/v1/commands";
    var maxResults = opts.maxResults || 12;
    var minChars = opts.minChars || 2;

    var commands = null;
    var selIdx = -1;
    var items = [];
    var isOpen = false;

    // Wrap input so dropdown can position relatively
    var wrap = input.parentNode;
    if (!wrap.classList.contains("nb-ac-wrap")) {
      var newWrap = document.createElement("div");
      newWrap.className = "nb-ac-wrap";
      newWrap.style.width = "100%";
      input.parentNode.insertBefore(newWrap, input);
      newWrap.appendChild(input);
      wrap = newWrap;
    }

    // Create dropdown
    var dropdown = document.createElement("div");
    dropdown.className = "nb-ac-dropdown";
    wrap.appendChild(dropdown);

    // Kick off lazy fetch
    fetchCommands(apiUrl).then(function (list) {
      commands = list;
    });

    function show(matches) {
      items = matches;
      selIdx = -1;
      if (!matches.length) {
        hide();
        return;
      }
      var fragment = document.createDocumentFragment();
      for (var i = 0; i < matches.length; i++) {
        var m = matches[i];
        var summ = m.cmd.summary || m.cmd.description || "";
        if (summ.length > 60) summ = summ.substring(0, 57) + "…";
        var item = document.createElement("div");
        item.className = "nb-ac-item";
        item.setAttribute("data-idx", String(i));

        var cmd = document.createElement("span");
        cmd.className = "nb-ac-cmd";
        cmd.textContent = m.name;
        item.appendChild(cmd);

        var args = document.createElement("span");
        args.className = "nb-ac-args";
        args.textContent = m.cmd.args_str || "";
        item.appendChild(args);

        if (summ) {
          var summary = document.createElement("span");
          summary.className = "nb-ac-summ";
          summary.textContent = summ;
          item.appendChild(summary);
        }

        fragment.appendChild(item);
      }
      dropdown.replaceChildren(fragment);
      dropdown.style.display = "block";
      isOpen = true;
      notifyResize();
    }

    function hide() {
      if (!isOpen) return;
      dropdown.style.display = "none";
      dropdown.replaceChildren();
      items = [];
      selIdx = -1;
      isOpen = false;
      notifyResize();
    }

    function selectItem(idx) {
      if (idx < 0 || idx >= items.length) return;
      var val = input.value;
      var firstSpace = val.indexOf(" ");
      // Replace the first token (command) with selected command
      if (firstSpace >= 0) {
        input.value = items[idx].name + val.substring(firstSpace);
      } else {
        input.value = items[idx].name;
      }
      hide();
      input.focus();
      input.setSelectionRange(input.value.length, input.value.length);
    }

    function highlight(idx) {
      var ch = dropdown.children;
      for (var i = 0; i < ch.length; i++) {
        ch[i].classList.toggle("nb-ac-sel", i === idx);
      }
      selIdx = idx;
      if (idx >= 0 && ch[idx]) {
        ch[idx].scrollIntoView({ block: "nearest" });
      }
    }

    function filter(text) {
      if (!commands || !text) {
        hide();
        return;
      }
      // Use only the first token being typed
      var first = text.split(/\s+/)[0].toLowerCase();
      if (first.length < minChars) {
        hide();
        return;
      }
      // If input already exactly matches a command AND there are args after it, don't show
      if (text.indexOf(" ") >= 0) {
        for (var c = 0; c < commands.length; c++) {
          if (commands[c].name === first) {
            hide();
            return;
          }
        }
      }
      // Prefix matches first, then substring matches
      var prefix = [],
        substr = [];
      for (var i = 0; i < commands.length; i++) {
        var nm = commands[i].name.toLowerCase();
        if (nm.indexOf(first) === 0) {
          prefix.push(commands[i]);
        } else if (nm.indexOf(first) > 0) {
          substr.push(commands[i]);
        }
      }
      var results = prefix.concat(substr);
      if (results.length > maxResults) results = results.slice(0, maxResults);
      // If only exact match remains, hide autocomplete
      if (
        results.length === 1 &&
        results[0].name === first &&
        text.indexOf(" ") < 0
      ) {
        hide();
        return;
      }
      show(results);
    }

    function notifyResize() {
      if (global.parent !== global) {
        var h = Math.max(
          document.body.scrollHeight,
          document.documentElement.scrollHeight,
        );
        global.parent.postMessage(
          { type: "iframe-resize", height: h + 10 },
          "*",
        );
      }
    }

    /* ── Events ────────────────────────────────────────── */

    input.addEventListener("input", function () {
      filter(input.value.trim());
    });

    input.addEventListener("focus", function () {
      var v = input.value.trim();
      if (v.length >= minChars) filter(v);
    });

    input.addEventListener("keydown", function (e) {
      if (!isOpen) return;

      if (e.key === "ArrowDown") {
        e.preventDefault();
        var next = selIdx + 1;
        if (next >= items.length) next = 0;
        highlight(next);
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        var prev = selIdx - 1;
        if (prev < 0) prev = items.length - 1;
        highlight(prev);
      } else if (e.key === "Tab") {
        e.preventDefault();
        if (selIdx >= 0) {
          selectItem(selIdx);
        } else if (items.length > 0) {
          selectItem(0);
        }
      } else if (e.key === "Enter" && !e.ctrlKey) {
        if (selIdx >= 0) {
          e.preventDefault();
          e.stopPropagation();
          selectItem(selIdx);
        }
        // If nothing highlighted, let Enter submit the form normally
      } else if (e.key === "Escape") {
        e.preventDefault();
        hide();
      }
    });

    // Click on item
    dropdown.addEventListener("mousedown", function (e) {
      e.preventDefault(); // don't steal focus
      var target = e.target;
      while (target && !target.classList.contains("nb-ac-item")) {
        target = target.parentNode;
        if (target === dropdown) {
          target = null;
          break;
        }
      }
      if (target) {
        var idx = parseInt(target.getAttribute("data-idx"), 10);
        selectItem(idx);
      }
    });

    // Hide on blur (with delay for click handling)
    input.addEventListener("blur", function () {
      setTimeout(function () {
        hide();
      }, 200);
    });

    return { hide: hide, filter: filter };
  }

  /* ── Export ──────────────────────────────────────────── */

  global.NotebookAutocomplete = {
    attach: attach,
    fetchCommands: fetchCommands,
  };
})(window, document);
