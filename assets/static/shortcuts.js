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

  var ALIGN_VALUES = {
    center: true,
    justify: true,
    left: true,
    right: true,
  };

  var CODE_CLASS_RE =
    /^(?:lang|language)-[A-Za-z0-9_-]+(?:\s+(?:lang|language)-[A-Za-z0-9_-]+)*$/;

  var HTML_ENTITY_MAP = {
    amp: "&",
    apos: "'",
    gt: ">",
    lt: "<",
    nbsp: "\u00a0",
    quot: '"',
  };

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

  function appendNode(target, node) {
    if (target && node) {
      target.appendChild(node);
    }
  }

  function decodeEntities(text) {
    return String(text || "").replace(
      /&(?:#(\d+)|#x([0-9a-fA-F]+)|([A-Za-z][A-Za-z0-9]+));/g,
      function (match, dec, hex, named) {
        var value;
        if (dec) {
          value = parseInt(dec, 10);
        } else if (hex) {
          value = parseInt(hex, 16);
        } else if (named) {
          if (!Object.prototype.hasOwnProperty.call(HTML_ENTITY_MAP, named)) {
            return match;
          }
          return HTML_ENTITY_MAP[named];
        }
        if (!isFinite(value) || value < 0 || value > 1114111) {
          return match;
        }
        try {
          return String.fromCodePoint(value);
        } catch (error) {
          return match;
        }
      },
    );
  }

  function appendDecodedText(target, text) {
    if (text) {
      appendNode(target, document.createTextNode(decodeEntities(text)));
    }
  }

  function appendLiteralText(target, text) {
    if (text) {
      appendNode(target, document.createTextNode(String(text)));
    }
  }

  function createMarkdownOptions(renderer) {
    var options = {
      breaks: true,
      gfm: true,
      pedantic: false,
      smartypants: false,
      tables: true,
    };
    var config =
      renderer &&
      renderer.options &&
      renderer.options.renderingConfig;
    if (config && config.singleLineBreaks === false) {
      options.breaks = false;
    }
    return options;
  }

  function getMarkedEngine(renderer) {
    if (
      renderer &&
      renderer.constructor &&
      renderer.constructor.marked
    ) {
      return renderer.constructor.marked;
    }
    if (global.SimpleMDE && global.SimpleMDE.marked) {
      return global.SimpleMDE.marked;
    }
    return null;
  }

  function getMarkdownTokens(markdownText, renderer) {
    var marked = getMarkedEngine(renderer);
    var options = createMarkdownOptions(renderer);
    var lexer = marked && (marked.lexer || (marked.Lexer && marked.Lexer.lex));
    if (typeof lexer !== "function") {
      return null;
    }
    var tokens = lexer(String(markdownText || ""), options);
    if (tokens) {
      tokens.marked = marked;
      tokens.options = options;
    }
    return tokens;
  }

  function getInlineRules(marked, options) {
    var inlineLexer = marked && marked.InlineLexer;
    if (!inlineLexer || !inlineLexer.rules) {
      return null;
    }
    if (options.gfm) {
      return options.breaks ? inlineLexer.rules.breaks : inlineLexer.rules.gfm;
    }
    if (options.pedantic) {
      return inlineLexer.rules.pedantic;
    }
    return inlineLexer.rules.normal;
  }

  function createInlineContext(links, options, marked) {
    return {
      inLink: false,
      links: links || {},
      marked: marked,
      options: options || {},
      rules: getInlineRules(marked, options || {}),
    };
  }

  function cloneInlineContext(context, inLink) {
    return {
      inLink: !!inLink,
      links: context.links,
      marked: context.marked,
      options: context.options,
      rules: context.rules,
    };
  }

  function createSafeLink(href, title, content) {
    var safeHref = sanitizeUrl(href);
    if (!safeHref) {
      return content;
    }
    var link = document.createElement("a");
    link.setAttribute("href", safeHref);
    if (title) {
      link.setAttribute("title", title);
    }
    appendNode(link, content);
    return link;
  }

  function createSafeImage(src, title, altText) {
    var safeSrc = sanitizeUrl(src, { allowDataImage: true });
    if (!safeSrc) {
      return document.createTextNode(altText || "");
    }
    var image = document.createElement("img");
    image.setAttribute("src", safeSrc);
    if (altText) {
      image.setAttribute("alt", altText);
    }
    if (title) {
      image.setAttribute("title", title);
    }
    return image;
  }

  function createInlineElement(tagName, content) {
    var element = document.createElement(tagName);
    appendNode(element, content);
    return element;
  }

  function createBlockElement(tagName, content) {
    var element = document.createElement(tagName);
    appendNode(element, content);
    return element;
  }

  function createLiteralBlock(text) {
    var value = String(text || "");
    if (value.indexOf("\n") >= 0) {
      var pre = document.createElement("pre");
      pre.appendChild(document.createTextNode(value));
      return pre;
    }
    return createBlockElement("p", document.createTextNode(value));
  }

  function inlineContextForTokens(tokens) {
    return createInlineContext(
      (tokens && tokens.links) || {},
      (tokens && tokens.options) || {},
      tokens && tokens.marked,
    );
  }

  function collectTextTokens(firstToken, tokens, state) {
    var text = String((firstToken && firstToken.text) || "");
    while (
      state.index < tokens.length &&
      tokens[state.index] &&
      tokens[state.index].type === "text"
    ) {
      text += "\n" + String(tokens[state.index].text || "");
      state.index += 1;
    }
    return text;
  }

  function renderInline(text, context) {
    var source = String(text || "");
    var rules = context && context.rules;
    var fragment = emptyFragment();
    if (!rules) {
      appendDecodedText(fragment, source);
      return fragment;
    }

    while (source) {
      var match;
      var label;
      var definition;
      if ((match = rules.escape.exec(source))) {
        source = source.substring(match[0].length);
        appendDecodedText(fragment, match[1]);
        continue;
      }
      if ((match = rules.autolink.exec(source))) {
        source = source.substring(match[0].length);
        if (match[2] === "@") {
          label =
            match[1].charAt(6) === ":"
              ? match[1].substring(7)
              : match[1];
          appendNode(
            fragment,
            createSafeLink(
              "mailto:" + label,
              null,
              document.createTextNode(label),
            ),
          );
        } else {
          label = match[1];
          appendNode(
            fragment,
            createSafeLink(label, null, document.createTextNode(label)),
          );
        }
        continue;
      }
      if (!context.inLink && (match = rules.url.exec(source))) {
        source = source.substring(match[0].length);
        label = match[0];
        appendNode(
          fragment,
          createSafeLink(label, null, document.createTextNode(label)),
        );
        continue;
      }
      if ((match = rules.tag.exec(source))) {
        source = source.substring(match[0].length);
        appendLiteralText(fragment, match[0]);
        continue;
      }
      if ((match = rules.link.exec(source))) {
        source = source.substring(match[0].length);
        if (match[0].charAt(0) === "!") {
          appendNode(
            fragment,
            createSafeImage(match[2], match[3], decodeEntities(match[1] || "")),
          );
        } else {
          appendNode(
            fragment,
            createSafeLink(
              match[2],
              match[3],
              renderInline(match[1], cloneInlineContext(context, true)),
            ),
          );
        }
        continue;
      }
      match = rules.reflink.exec(source) || rules.nolink.exec(source);
      if (match) {
        source = source.substring(match[0].length);
        label = (match[2] || match[1] || "").replace(/\s+/g, " ");
        definition = context.links[String(label).toLowerCase()];
        if (!definition || !definition.href) {
          appendLiteralText(fragment, match[0].charAt(0));
          source = match[0].substring(1) + source;
          continue;
        }
        if (match[0].charAt(0) === "!") {
          appendNode(
            fragment,
            createSafeImage(
              definition.href,
              definition.title,
              decodeEntities(match[1] || ""),
            ),
          );
        } else {
          appendNode(
            fragment,
            createSafeLink(
              definition.href,
              definition.title,
              renderInline(match[1], cloneInlineContext(context, true)),
            ),
          );
        }
        continue;
      }
      if ((match = rules.strong.exec(source))) {
        source = source.substring(match[0].length);
        appendNode(
          fragment,
          createInlineElement(
            "strong",
            renderInline(
              match[2] || match[1],
              cloneInlineContext(context, context.inLink),
            ),
          ),
        );
        continue;
      }
      if ((match = rules.em.exec(source))) {
        source = source.substring(match[0].length);
        appendNode(
          fragment,
          createInlineElement(
            "em",
            renderInline(
              match[2] || match[1],
              cloneInlineContext(context, context.inLink),
            ),
          ),
        );
        continue;
      }
      if ((match = rules.code.exec(source))) {
        source = source.substring(match[0].length);
        var inlineCode = document.createElement("code");
        inlineCode.textContent = match[2] || "";
        appendNode(fragment, inlineCode);
        continue;
      }
      if ((match = rules.br.exec(source))) {
        source = source.substring(match[0].length);
        appendNode(fragment, document.createElement("br"));
        continue;
      }
      if ((match = rules.del.exec(source))) {
        source = source.substring(match[0].length);
        appendNode(
          fragment,
          createInlineElement(
            "del",
            renderInline(
              match[1],
              cloneInlineContext(context, context.inLink),
            ),
          ),
        );
        continue;
      }
      if ((match = rules.text.exec(source))) {
        source = source.substring(match[0].length);
        appendDecodedText(fragment, match[0]);
        continue;
      }
      appendLiteralText(fragment, source.charAt(0));
      source = source.substring(1);
    }

    return fragment;
  }

  function renderTextParagraph(text, tokens) {
    return createBlockElement("p", renderInline(text, inlineContextForTokens(tokens)));
  }

  function renderHeading(token, tokens) {
    var depth = parseInt(token.depth, 10);
    if (isNaN(depth) || depth < 1 || depth > 6) {
      depth = 1;
    }
    return createBlockElement(
      "h" + String(depth),
      renderInline(token.text, inlineContextForTokens(tokens)),
    );
  }

  function renderCodeBlock(token) {
    var pre = document.createElement("pre");
    var code = document.createElement("code");
    if (token.lang) {
      var className = "lang-" + String(token.lang).trim();
      if (CODE_CLASS_RE.test(className)) {
        code.className = className;
      }
    }
    code.textContent = token.text || "";
    pre.appendChild(code);
    return pre;
  }

  function renderTable(token, tokens) {
    var table = document.createElement("table");
    var thead = document.createElement("thead");
    var tbody = document.createElement("tbody");
    var row = document.createElement("tr");
    var inlineContext = inlineContextForTokens(tokens);
    var i;
    var j;
    for (i = 0; i < token.header.length; i++) {
      var th = document.createElement("th");
      appendNode(th, renderInline(token.header[i], cloneInlineContext(inlineContext, false)));
      applyTextAlign(th, token.align && token.align[i]);
      row.appendChild(th);
    }
    thead.appendChild(row);
    table.appendChild(thead);
    for (i = 0; i < token.cells.length; i++) {
      row = document.createElement("tr");
      for (j = 0; j < token.cells[i].length; j++) {
        var td = document.createElement("td");
        appendNode(td, renderInline(token.cells[i][j], cloneInlineContext(inlineContext, false)));
        applyTextAlign(td, token.align && token.align[j]);
        row.appendChild(td);
      }
      tbody.appendChild(row);
    }
    table.appendChild(tbody);
    return table;
  }

  function renderListItem(loose, tokens, state) {
    var item = document.createElement("li");
    while (state.index < tokens.length) {
      var token = tokens[state.index++];
      if (!token || token.type === "list_item_end") {
        break;
      }
      if (token.type === "space") {
        continue;
      }
      if (!loose && token.type === "text") {
        appendNode(
          item,
          renderInline(
            collectTextTokens(token, tokens, state),
            inlineContextForTokens(tokens),
          ),
        );
        continue;
      }
      appendNode(item, renderBlockToken(token, tokens, state));
    }
    return item;
  }

  function renderList(token, tokens, state) {
    var list = document.createElement(token.ordered ? "ol" : "ul");
    while (state.index < tokens.length) {
      var itemToken = tokens[state.index++];
      if (!itemToken) {
        break;
      }
      if (itemToken.type === "list_end") {
        break;
      }
      if (
        itemToken.type !== "list_item_start" &&
        itemToken.type !== "loose_item_start"
      ) {
        continue;
      }
      list.appendChild(
        renderListItem(itemToken.type === "loose_item_start", tokens, state),
      );
    }
    return list;
  }

  function renderBlockToken(token, tokens, state) {
    switch (token.type) {
      case "space":
        return emptyFragment();
      case "hr":
        return document.createElement("hr");
      case "heading":
        return renderHeading(token, tokens);
      case "code":
        return renderCodeBlock(token);
      case "table":
        return renderTable(token, tokens);
      case "blockquote_start":
        return createBlockElement(
          "blockquote",
          renderBlocks(tokens, state, "blockquote_end"),
        );
      case "list_start":
        return renderList(token, tokens, state);
      case "html":
        return createLiteralBlock(token.text);
      case "paragraph":
        return renderTextParagraph(token.text, tokens);
      case "text":
        return renderTextParagraph(collectTextTokens(token, tokens, state), tokens);
      default:
        return createLiteralBlock(token.text || "");
    }
  }

  function renderBlocks(tokens, state, stopType) {
    var fragment = emptyFragment();
    while (state.index < tokens.length) {
      var token = tokens[state.index++];
      if (!token) {
        break;
      }
      if (stopType && token.type === stopType) {
        break;
      }
      appendNode(fragment, renderBlockToken(token, tokens, state));
    }
    return fragment;
  }

  function buildMarkdownFragment(markdownText, renderer) {
    var tokens = getMarkdownTokens(markdownText, renderer);
    if (!tokens) {
      var fallback = document.createElement("pre");
      fallback.textContent = String(markdownText || "");
      var fragment = emptyFragment();
      fragment.appendChild(fallback);
      return fragment;
    }
    return renderBlocks(tokens, { index: 0 });
  }

  global.NotebookSafeMarkdown = {
    buildFragment: function (markdownText, renderer) {
      return buildMarkdownFragment(markdownText, renderer);
    },
    renderInto: function (container, markdownText, renderer) {
      if (!container) {
        return;
      }
      container.replaceChildren(buildMarkdownFragment(markdownText, renderer));
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
