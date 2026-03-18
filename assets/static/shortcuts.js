(function (global, document) {
  "use strict";

  var isMac = /Mac|iPhone|iPad|iPod/.test(global.navigator.platform);
  var config = global.NOTEBOOK_KB || {};
  var defaults = {
    about: "alt+a",
    new_page: "ctrl+n",
    settings: "alt+s",
    toggle_pipe: "alt+o",
    new_command: "alt+c",
    new_markdown: "alt+m",
    new_script: "alt+j",
    search_notebook: "mod+shift+f",
    save: "mod+s",
    cancel: "escape",
    execute: "ctrl+enter",
    edit_markdown: "alt+e",
  };

  function normalizeKey(key) {
    if (!key) return "";
    var normalized = String(key).toLowerCase();
    switch (normalized) {
      case " ":
      case "spacebar":
        return "space";
      case "esc":
        return "escape";
      case "arrowup":
        return "up";
      case "arrowdown":
        return "down";
      case "arrowleft":
        return "left";
      case "arrowright":
        return "right";
      default:
        return normalized;
    }
  }

  function expandToken(token) {
    token = normalizeKey(token);
    if (token === "mod") {
      return isMac ? "meta" : "ctrl";
    }
    return token;
  }

  function parseCombo(combo) {
    var parts = String(combo || "")
      .split("+")
      .map(function (part) {
        return expandToken(part.trim());
      })
      .filter(Boolean);
    return {
      ctrl: parts.indexOf("ctrl") >= 0,
      meta: parts.indexOf("meta") >= 0,
      alt: parts.indexOf("alt") >= 0,
      shift: parts.indexOf("shift") >= 0,
      key:
        parts.filter(function (part) {
          return ["ctrl", "meta", "alt", "shift"].indexOf(part) < 0;
        })[0] || "",
    };
  }

  function matches(event, combo) {
    var spec = parseCombo(combo);
    if (!spec.key) return false;
    if (!!event.ctrlKey !== spec.ctrl) return false;
    if (!!event.metaKey !== spec.meta) return false;
    if (!!event.altKey !== spec.alt) return false;
    if (!!event.shiftKey !== spec.shift) return false;
    return normalizeKey(event.key) === spec.key;
  }

  function isEditableTarget(target) {
    if (!target || target === document.body) return false;
    if (target.isContentEditable) return true;
    var tag = (target.tagName || "").toLowerCase();
    return tag === "input" || tag === "textarea" || tag === "select";
  }

  function formatCombo(combo) {
    if (!combo) return "";
    var tokens = String(combo)
      .split("+")
      .map(function (token) {
        token = token.trim().toLowerCase();
        if (token === "mod") return isMac ? "⌘" : "Ctrl";
        if (token === "ctrl") return isMac ? "⌃" : "Ctrl";
        if (token === "meta") return isMac ? "⌘" : "Meta";
        if (token === "alt") return isMac ? "⌥" : "Alt";
        if (token === "shift") return "⇧";
        if (token === "escape") return "Esc";
        if (token === "enter") return "↩";
        if (token === "space") return "Space";
        return token.length === 1
          ? token.toUpperCase()
          : token.charAt(0).toUpperCase() + token.slice(1);
      });
    return tokens.join(isMac ? "" : "+");
  }

  function combo(action) {
    return config[action] || defaults[action] || "";
  }

  function bind(bindings) {
    document.addEventListener(
      "keydown",
      function (event) {
        if (!bindings || !bindings.length) return;
        for (var i = 0; i < bindings.length; i++) {
          var binding = bindings[i];
          if (!binding) continue;

          var comboString = "";
          if (binding.action) {
            comboString = combo(binding.action);
          } else if (binding.combo) {
            comboString = String(binding.combo);
          }
          if (!comboString) continue;
          if (!matches(event, comboString)) continue;
          if (isEditableTarget(event.target) && !binding.allowInInputs) continue;
          if (binding.when && !binding.when(event)) continue;

          event.preventDefault();
          event.stopPropagation();
          binding.handler(event);
          return;
        }
      },
      true,
    );
  }

  function annotate(element, actionOrCombo) {
    if (!element) return;
    var comboString = defaults[actionOrCombo] ? combo(actionOrCombo) : actionOrCombo;
    if (!comboString) return;
    var display = formatCombo(comboString);
    var text = (element.textContent || "").replace(/\s+/g, " ").trim();
    var tooltip = text ? text + " (" + display + ")" : display;
    element.setAttribute("data-tooltip", tooltip);
    element.setAttribute("aria-keyshortcuts", display);
    if (element.classList) {
      element.classList.add("tooltip");
      element.classList.add("tooltip-bottom");
    }
  }

  function record(callback) {
    function handler(event) {
      event.preventDefault();
      event.stopPropagation();
      var parts = [];
      if (event.ctrlKey) parts.push("ctrl");
      if (event.altKey) parts.push("alt");
      if (event.shiftKey) parts.push("shift");
      if (event.metaKey) parts.push("meta");
      var key = normalizeKey(event.key);
      if (["control", "alt", "shift", "meta"].indexOf(key) >= 0) return;
      parts.push(key);
      document.removeEventListener("keydown", handler, true);
      if (callback) callback(parts.join("+"));
    }

    document.addEventListener("keydown", handler, true);
  }

  var ALLOWED_MARKDOWN_TAGS = {
    a: true,
    b: true,
    blockquote: true,
    br: true,
    code: true,
    del: true,
    details: true,
    div: true,
    em: true,
    h1: true,
    h2: true,
    h3: true,
    h4: true,
    h5: true,
    h6: true,
    hr: true,
    i: true,
    img: true,
    input: true,
    ins: true,
    kbd: true,
    li: true,
    ol: true,
    p: true,
    pre: true,
    samp: true,
    span: true,
    strike: true,
    strong: true,
    sub: true,
    summary: true,
    sup: true,
    table: true,
    tbody: true,
    td: true,
    th: true,
    thead: true,
    tr: true,
    u: true,
    ul: true,
  };

  var BLOCKED_MARKDOWN_TAGS = {
    audio: true,
    base: true,
    button: true,
    canvas: true,
    embed: true,
    form: true,
    iframe: true,
    link: true,
    math: true,
    meta: true,
    object: true,
    option: true,
    script: true,
    select: true,
    source: true,
    style: true,
    svg: true,
    textarea: true,
    video: true,
  };

  var ALIGN_VALUES = {
    center: true,
    justify: true,
    left: true,
    right: true,
  };

  var CODE_CLASS_RE =
    /^(?:lang|language)-[A-Za-z0-9_-]+(?:\s+(?:lang|language)-[A-Za-z0-9_-]+)*$/;

  function emptyFragment() {
    return document.createDocumentFragment();
  }

  function normalizeUrlCandidate(url) {
    return String(url || "").replace(/[\u0000-\u001F\u007F-\u009F\s]+/g, "");
  }

  function sanitizeUrl(url, options) {
    var value = String(url || "").trim();
    if (!value) {
      return "";
    }
    if (value.charAt(0) === "#") {
      return value;
    }

    var normalized = normalizeUrlCandidate(value).toLowerCase();
    if (
      normalized.indexOf("javascript:") === 0 ||
      normalized.indexOf("vbscript:") === 0
    ) {
      return "";
    }
    if (normalized.indexOf("data:") === 0) {
      if (
        options &&
        options.allowDataImage &&
        /^data:image\/(?:bmp|gif|jpe?g|png|webp);/i.test(value)
      ) {
        return value;
      }
      return "";
    }
    if (value.indexOf("//") === 0) {
      return value;
    }
    if (/^[A-Za-z][A-Za-z0-9+.-]*:/.test(normalized)) {
      var scheme = normalized.split(":", 1)[0];
      if (
        scheme === "ftp" ||
        scheme === "http" ||
        scheme === "https" ||
        scheme === "mailto"
      ) {
        return value;
      }
      return "";
    }
    return value;
  }

  function applyTextAlign(target, value) {
    var normalized = String(value || "").trim().toLowerCase();
    if (ALIGN_VALUES[normalized]) {
      target.style.textAlign = normalized;
    }
  }

  function copyTextAlign(source, target) {
    applyTextAlign(target, source.getAttribute("align"));

    var style = source.getAttribute("style");
    if (!style) {
      return;
    }

    style.split(";").forEach(function (declaration) {
      var parts = declaration.split(":");
      if (parts.length < 2) {
        return;
      }
      if (parts[0].trim().toLowerCase() === "text-align") {
        applyTextAlign(target, parts.slice(1).join(":"));
      }
    });
  }

  function sanitizeChildren(node) {
    var fragment = document.createDocumentFragment();
    for (var child = node.firstChild; child; child = child.nextSibling) {
      fragment.appendChild(sanitizeNode(child));
    }
    return fragment;
  }

  function sanitizeImage(source) {
    var src = sanitizeUrl(source.getAttribute("src"), { allowDataImage: true });
    if (!src) {
      return emptyFragment();
    }

    var img = document.createElement("img");
    img.setAttribute("src", src);

    if (source.hasAttribute("alt")) {
      img.setAttribute("alt", source.getAttribute("alt"));
    }
    if (source.hasAttribute("title")) {
      img.setAttribute("title", source.getAttribute("title"));
    }

    return img;
  }

  function sanitizeInput(source) {
    var type = String(source.getAttribute("type") || "").toLowerCase();
    if (type !== "checkbox") {
      return emptyFragment();
    }

    var input = document.createElement("input");
    input.setAttribute("type", "checkbox");
    input.setAttribute("disabled", "disabled");
    input.disabled = true;

    if (source.hasAttribute("checked")) {
      input.setAttribute("checked", "checked");
      input.checked = true;
    }

    return input;
  }

  function sanitizeElement(source) {
    var tag = source.tagName.toLowerCase();
    if (BLOCKED_MARKDOWN_TAGS[tag]) {
      return emptyFragment();
    }
    if (tag === "img") {
      return sanitizeImage(source);
    }
    if (tag === "input") {
      return sanitizeInput(source);
    }
    if (!ALLOWED_MARKDOWN_TAGS[tag]) {
      return sanitizeChildren(source);
    }

    var element = document.createElement(tag);

    if (tag === "a") {
      var href = sanitizeUrl(source.getAttribute("href"));
      if (href) {
        element.setAttribute("href", href);
      }
      if (source.hasAttribute("title")) {
        element.setAttribute("title", source.getAttribute("title"));
      }
    } else if ((tag === "code" || tag === "pre") && source.className) {
      var className = String(source.className || "").trim();
      if (CODE_CLASS_RE.test(className)) {
        element.className = className;
      }
    } else if (tag === "ol" && source.hasAttribute("start")) {
      var start = parseInt(source.getAttribute("start"), 10);
      if (!isNaN(start) && start > 0) {
        element.setAttribute("start", String(start));
      }
    } else if (tag === "details" && source.hasAttribute("open")) {
      element.setAttribute("open", "open");
    }

    if (tag === "td" || tag === "th") {
      copyTextAlign(source, element);
    }

    if (tag !== "br" && tag !== "hr") {
      element.appendChild(sanitizeChildren(source));
    }

    return element;
  }

  function sanitizeNode(node) {
    if (node.nodeType === Node.TEXT_NODE) {
      return document.createTextNode(node.data || "");
    }
    if (node.nodeType !== Node.ELEMENT_NODE) {
      return emptyFragment();
    }
    return sanitizeElement(node);
  }

  function buildSafeFragment(html) {
    var parser = new DOMParser();
    var parsed = parser.parseFromString(
      "<!doctype html><body>" + String(html || "") + "</body>",
      "text/html",
    );
    return sanitizeChildren(parsed.body);
  }

  function renderMarkdown(markdownText, renderer) {
    if (renderer && typeof renderer.markdown === "function") {
      return renderer.markdown(markdownText || "");
    }
    return String(markdownText || "");
  }

  function fragmentToHTML(fragment) {
    var container = document.createElement("div");
    container.appendChild(fragment);
    return container.innerHTML;
  }

  global.NotebookSafeMarkdown = {
    buildFragment: function (markdownText, renderer) {
      return buildSafeFragment(renderMarkdown(markdownText, renderer));
    },
    renderInto: function (container, markdownText, renderer) {
      if (!container) {
        return;
      }
      container.replaceChildren(
        buildSafeFragment(renderMarkdown(markdownText, renderer)),
      );
    },
    renderToHTML: function (markdownText, renderer) {
      return fragmentToHTML(
        buildSafeFragment(renderMarkdown(markdownText, renderer)),
      );
    },
  };

  global.NotebookShortcuts = {
    bind: bind,
    annotate: annotate,
    combo: combo,
    formatCombo: formatCombo,
    record: record,
    isEditableTarget: isEditableTarget,
    defaults: defaults,
  };
})(window, document);
