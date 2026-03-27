// Lightweight syntax highlighter for Dippin site code blocks.
// Auto-detects Dippin, shell, terminal, and diagnostic blocks.
(function () {
  "use strict";

  function span(cls, text) {
    return '<span class="hl-' + cls + '">' + text + "</span>";
  }

  function esc(s) {
    return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  }

  // Highlight a Dippin code block (already HTML-escaped).
  function highlightDippin(h) {
    // Comments first (greedy to EOL).
    h = h.replace(/(^|\n)(\s*#[^\n]*)/g, function (_, pre, cmt) {
      return pre + span("cmt", cmt);
    });
    // Strings.
    h = h.replace(/("(?:[^"\\]|\\.)*")/g, function (_, s) {
      return span("str", s);
    });
    // ${ctx.*} variables.
    h = h.replace(/(\$\{[^}]+\})/g, function (_, v) {
      return span("shvar", v);
    });
    // Node declarations: agent Foo, tool Bar, workflow Name, etc.
    h = h.replace(
      /\b(workflow|agent|human|tool|subgraph)\s+([A-Z]\w*)/g,
      function (_, kw, name) {
        return span("kw", kw) + " " + span("node", name);
      }
    );
    // Parallel/fan_in with targets.
    h = h.replace(
      /\b(parallel)\s+(\w+)\s*(-&gt;)/g,
      function (_, kw, name, arrow) {
        return span("kw", kw) + " " + span("node", name) + " " + span("op", arrow);
      }
    );
    h = h.replace(
      /\b(fan_in)\s+(\w+)\s*(&lt;-)/g,
      function (_, kw, name, arrow) {
        return span("kw", kw) + " " + span("node", name) + " " + span("op", arrow);
      }
    );
    // Remaining keywords.
    h = h.replace(
      /\b(edges|defaults|stylesheet)\b/g,
      function (_, kw) { return span("kw", kw); }
    );
    // Condition keywords.
    h = h.replace(
      /\b(when|and|or|not|contains|startswith|endswith)\b/g,
      function (_, kw) { return span("cond", kw); }
    );
    // Booleans.
    h = h.replace(/\b(true|false)\b/g, function (_, b) {
      return span("bool", b);
    });
    // Arrows (already escaped).
    h = h.replace(/(-&gt;|&lt;-)/g, function (_, op) {
      return span("op", op);
    });
    // Operators.
    h = h.replace(/(==|!=)/g, function (_, op) {
      return span("op", op);
    });
    // Field names (word followed by colon at start of line, indented).
    h = h.replace(/(^|\n)(\s+)(\w[\w_]*)(:)/g, function (_, nl, ws, name, colon) {
      return nl + ws + span("field", name) + colon;
    });
    return h;
  }

  // Highlight a shell script block.
  function highlightShell(h) {
    // Shebangs.
    h = h.replace(/^(#![^\n]+)/gm, function (_, s) {
      return span("cmt", s);
    });
    // Comments (but not inside strings — good enough heuristic).
    h = h.replace(/(^|\s)(#[^\n]*)/gm, function (_, pre, cmt) {
      return pre + span("cmt", cmt);
    });
    // Strings.
    h = h.replace(/('[^']*')/g, function (_, s) {
      return span("str", s);
    });
    h = h.replace(/("(?:[^"\\]|\\.)*")/g, function (_, s) {
      return span("str", s);
    });
    // Variables.
    h = h.replace(/(\$\w+|\$\{[^}]+\})/g, function (_, v) {
      return span("shvar", v);
    });
    // Shell keywords.
    h = h.replace(
      /\b(if|then|else|elif|fi|for|while|do|done|case|esac|set|exit|printf|echo|cat|grep|mkdir|cd|export)\b/g,
      function (_, kw) { return span("shkw", kw); }
    );
    return h;
  }

  // Highlight terminal / CLI output.
  function highlightTerminal(h) {
    // Diagnostic severity tags.
    h = h.replace(/\b(error)(\[DIP\d+\])/g, function (_, sev, code) {
      return span("fail", sev + code);
    });
    h = h.replace(/\b(warning)(\[DIP\d+\])/g, function (_, sev, code) {
      return span("warn", sev + code);
    });
    h = h.replace(/\b(hint)(\[DIP\d+\])/g, function (_, sev, code) {
      return span("dim", sev + code);
    });
    // $ prompt lines.
    h = h.replace(/^(\$) (\S+)/gm, function (_, p, cmd) {
      return span("prompt", p) + " " + span("cmd", cmd);
    });
    // PASS / FAIL.
    h = h.replace(/\b(PASS)\b/g, function (_, w) { return span("pass", w); });
    h = h.replace(/\b(FAIL)\b/g, function (_, w) { return span("fail", w); });
    // Flags.
    h = h.replace(/(\s)(--?\w[\w-]*)/g, function (_, ws, flag) {
      return ws + span("flag", flag);
    });
    // File paths with extensions.
    h = h.replace(/\b([\w./-]+\.(?:dip|dot|json|csv|out|html|js))\b/g, function (_, f) {
      return span("file", f);
    });
    // help lines.
    h = h.replace(/(= help:[^\n]*)/g, function (_, help) {
      return span("dim", help);
    });
    // Source locations.
    h = h.replace(/(--&gt;\s*[\w./-]+:\d+:\d+)/g, function (_, loc) {
      return span("dim", loc);
    });
    // Strings.
    h = h.replace(/("(?:[^"\\]|\\.)*")/g, function (_, s) {
      return span("str", s);
    });
    return h;
  }

  // Highlight JSON output.
  function highlightJSON(h) {
    // Keys.
    h = h.replace(/("[\w_]+")\s*:/g, function (_, key) {
      return span("field", key) + ":";
    });
    // String values.
    h = h.replace(/:\s*("(?:[^"\\]|\\.)*")/g, function (m, val) {
      return ": " + span("str", val);
    });
    // Booleans and numbers.
    h = h.replace(/:\s*(true|false)\b/g, function (_, b) {
      return ": " + span("bool", b);
    });
    h = h.replace(/:\s*(\d+)/g, function (_, n) {
      return ": " + span("num", n);
    });
    return h;
  }

  function isDippin(t) {
    return /\b(workflow|agent |human |tool |edges\b|defaults\b)/.test(t);
  }
  function isShell(t) {
    return /^(\s*#!\/bin\/|set -e)/.test(t.trim());
  }
  function isTerminal(t) {
    return /^\$\s/.test(t.trim());
  }
  function isDiagnostic(t) {
    return /\b(error|warning|hint)\[DIP\d+\]/.test(t);
  }
  function isJSON(t) {
    return /^\s*[\[{]/.test(t.trim());
  }

  // Auto-highlight a string and return HTML. Exported for playground use.
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
    dippin: function (raw) { return highlightDippin(esc(raw)); },
    json: function (raw) { return highlightJSON(esc(raw)); },
    auto: autoHighlight
  };

  document.addEventListener("DOMContentLoaded", function () {
    // Target all pre elements on the page, not just .doc-body.
    var pres = document.querySelectorAll("pre");
    pres.forEach(function (pre) {
      // Skip blocks with existing manual markup.
      if (pre.querySelector("span")) return;
      // Skip blocks inside compare-code (already hand-highlighted).
      if (pre.closest(".compare-code")) return;

      var raw = pre.textContent;
      var h = autoHighlight(raw);
      if (h) pre.innerHTML = h;
    });
  });
})();
