// Syntax highlighter for the Dippin site.
// Auto-detects Dippin, shell, terminal, JSON, and diagnostic blocks.
(function () {
  "use strict";

  // Placeholder system: generated <span> tags are replaced with \x00N\x00
  // tokens so later regexes don't match inside them.
  var tokens;
  function reset() { tokens = []; }
  function push(html) {
    var i = tokens.length;
    tokens.push(html);
    return "\x00" + i + "\x00";
  }
  function pop(s) {
    return s.replace(/\x00(\d+)\x00/g, function (_, i) { return tokens[+i]; });
  }
  function s(cls, text) {
    return push('<span class="hl-' + cls + '">' + text + "</span>");
  }

  function esc(t) {
    return t.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  }

  // ── Dippin ──────────────────────────────────────────────
  function highlightDippin(h) {
    reset();
    // Comments.
    h = h.replace(/(^|\n)(\s*#[^\n]*)/g, function (_, p, c) { return p + s("cmt", c); });
    // Strings.
    h = h.replace(/("(?:[^"\\]|\\.)*")/g, function (_, v) { return s("str", v); });
    // ${ctx.*} variables.
    h = h.replace(/(\$\{[^}]+\})/g, function (_, v) { return s("shvar", v); });
    // Node declarations.
    h = h.replace(/\b(workflow|agent|human|tool|subgraph|conditional)\s+([A-Z]\w*)/g,
      function (_, kw, n) { return s("kw", kw) + " " + s("node", n); });
    // Parallel/fan_in.
    h = h.replace(/\b(parallel)\s+(\w+)\s*(-&gt;)/g,
      function (_, kw, n, a) { return s("kw", kw) + " " + s("node", n) + " " + s("op", a); });
    h = h.replace(/\b(fan_in)\s+(\w+)\s*(&lt;-)/g,
      function (_, kw, n, a) { return s("kw", kw) + " " + s("node", n) + " " + s("op", a); });
    // Section keywords.
    h = h.replace(/\b(edges|defaults|vars|stylesheet)\b/g, function (_, k) { return s("kw", k); });
    // Condition keywords.
    h = h.replace(/\b(when|and|or|not|contains|startswith|endswith)\b/g, function (_, k) { return s("cond", k); });
    // Booleans.
    h = h.replace(/\b(true|false)\b/g, function (_, b) { return s("bool", b); });
    // Arrows.
    h = h.replace(/(-&gt;|&lt;-)/g, function (_, o) { return s("op", o); });
    // Equality operators.
    h = h.replace(/(==|!=)/g, function (_, o) { return s("op", o); });
    // Field names.
    h = h.replace(/(^|\n)(\s+)(\w[\w_]*)(:)/g, function (_, nl, ws, n, c) { return nl + ws + s("field", n) + c; });
    return pop(h);
  }

  // ── Shell ───────────────────────────────────────────────
  function highlightShell(h) {
    reset();
    h = h.replace(/^(#![^\n]+)/gm, function (_, v) { return s("cmt", v); });
    h = h.replace(/(^|\s)(#[^\n]*)/gm, function (_, p, c) { return p + s("cmt", c); });
    h = h.replace(/('[^']*')/g, function (_, v) { return s("str", v); });
    h = h.replace(/("(?:[^"\\]|\\.)*")/g, function (_, v) { return s("str", v); });
    h = h.replace(/(\$\w+|\$\{[^}]+\})/g, function (_, v) { return s("shvar", v); });
    h = h.replace(/\b(if|then|else|elif|fi|for|while|do|done|case|esac|set|exit|printf|echo|cat|grep|mkdir|cd|export)\b/g,
      function (_, k) { return s("shkw", k); });
    return pop(h);
  }

  // ── Terminal ────────────────────────────────────────────
  function highlightTerminal(h) {
    reset();
    // Diagnostic severity.
    h = h.replace(/\b(error)(\[DIP\d+\])/g, function (_, a, b) { return s("fail", a + b); });
    h = h.replace(/\b(warning)(\[DIP\d+\])/g, function (_, a, b) { return s("warn", a + b); });
    h = h.replace(/\b(hint)(\[DIP\d+\])/g, function (_, a, b) { return s("dim", a + b); });
    // Prompt lines.
    h = h.replace(/^(\$) (\S+)/gm, function (_, p, c) { return s("prompt", p) + " " + s("cmd", c); });
    // PASS / FAIL.
    h = h.replace(/\b(PASS)\b/g, function (_, w) { return s("pass", w); });
    h = h.replace(/\b(FAIL)\b/g, function (_, w) { return s("fail", w); });
    // Strings (in diagnostic output).
    h = h.replace(/("(?:[^"\\]|\\.)*")/g, function (_, v) { return s("str", v); });
    // Flags.
    h = h.replace(/(\s)(--?\w[\w-]*)/g, function (_, ws, f) { return ws + s("flag", f); });
    // File paths.
    h = h.replace(/\b([\w./-]+\.(?:dip|dot|json|csv|out|html|js))\b/g, function (_, f) { return s("file", f); });
    // Help and source locations.
    h = h.replace(/(= help:[^\n]*)/g, function (_, v) { return s("dim", v); });
    h = h.replace(/(--&gt;\s*[\w./-]+:\d+:\d+)/g, function (_, v) { return s("dim", v); });
    // JSON body embedded in terminal output.
    h = h.replace(/([\n])([ \t]*\{[\s\S]*\})\s*$/g, function (_, nl, json) {
      return nl + highlightJSONInner(json);
    });
    return pop(h);
  }

  // ── JSON ────────────────────────────────────────────────
  function highlightJSON(h) {
    reset();
    return pop(highlightJSONInner(h));
  }
  function highlightJSONInner(h) {
    // Keys.
    h = h.replace(/("[\w_]+")\s*:/g, function (_, k) { return s("field", k) + ":"; });
    // String values (after colon or in arrays).
    h = h.replace(/(:\s*)("(?:[^"\\]|\\.)*")/g, function (_, pre, v) { return pre + s("str", v); });
    h = h.replace(/(\[)("(?:[^"\\]|\\.)*")/g, function (_, br, v) { return br + s("str", v); });
    h = h.replace(/(,\s*)("(?:[^"\\]|\\.)*")/g, function (_, pre, v) { return pre + s("str", v); });
    // Booleans.
    h = h.replace(/\b(true|false)\b/g, function (_, b) { return s("bool", b); });
    // Numbers.
    h = h.replace(/(:\s*)(\d+)/g, function (_, pre, n) { return pre + s("num", n); });
    // null.
    h = h.replace(/\b(null)\b/g, function (_, n) { return s("dim", n); });
    return h;
  }

  // ── Detection ───────────────────────────────────────────
  function isDippin(t) { return /\b(workflow|agent |human |tool |conditional |edges\b|defaults\b|vars\b)/.test(t); }
  function isShell(t) { return /^(\s*#!\/bin\/|set -e)/.test(t.trim()); }
  function isTerminal(t) { return /^\$\s/.test(t.trim()); }
  function isDiagnostic(t) { return /\b(error|warning|hint)\[DIP\d+\]/.test(t); }
  function isJSON(t) { return /^\s*[\[{]/.test(t.trim()); }

  function autoHighlight(raw) {
    var h = esc(raw);
    if (isDiagnostic(raw)) return highlightTerminal(h);
    if (isDippin(raw)) return highlightDippin(h);
    if (isTerminal(raw)) return highlightTerminal(h);
    if (isShell(raw)) return highlightShell(h);
    if (isJSON(raw)) return highlightJSON(h);
    return null;
  }

  // Expose for playground.
  window.dippinHighlight = {
    dippin: function (raw) { reset(); var r = highlightDippin(esc(raw)); return r; },
    json: function (raw) { reset(); return pop(highlightJSONInner(esc(raw))); },
    auto: autoHighlight
  };

  document.addEventListener("DOMContentLoaded", function () {
    var pres = document.querySelectorAll("pre");
    pres.forEach(function (pre) {
      // Skip blocks that already have manual span highlighting.
      if (pre.querySelector("span")) return;
      var raw = pre.textContent;
      var h = autoHighlight(raw);
      if (h) pre.innerHTML = h;
    });
  });
})();
