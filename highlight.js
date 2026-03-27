// Lightweight syntax highlighter for Dippin site code blocks.
// Runs on DOMContentLoaded, highlights .doc-body pre elements.
// Detects: Dippin (.dip), shell, and terminal blocks.
(function () {
  "use strict";

  var DIP_KW =
    /\b(workflow|agent|human|tool|subgraph|parallel|fan_in|edges|defaults|stylesheet)\b/g;
  var DIP_COND =
    /\b(when|and|or|not|contains|startswith|endswith|in)\b/g;
  var DIP_BOOL = /\b(true|false)\b/g;
  var DIP_OP = /(-&gt;|&lt;-|->|<-|==|!=)/g;
  var DIP_STR = /("(?:[^"\\]|\\.)*")/g;
  var DIP_CMT = /(#[^\n]*)/g;
  var DIP_FIELD = /^(\s*\w[\w_]*)(:)/gm;
  var DIP_NODE_DECL =
    /\b(agent|human|tool|subgraph|parallel|fan_in|workflow)\s+([A-Z]\w*)/g;
  var DIP_VAR = /(\$\{[^}]+\})/g;
  var DIP_NUM = /\b(\d+[smh]?\b)/g;

  var SH_KW =
    /\b(if|then|else|elif|fi|for|while|do|done|case|esac|set|exit|printf|echo|cat|grep|mkdir|cd|export)\b/g;
  var SH_VAR = /(\$\w+|\$\{[^}]+\})/g;
  var SH_CMT = /(#[^\n]*)/g;
  var SH_STR = /('(?:[^'\\]|\\.)*'|"(?:[^"\\]|\\.)*")/g;
  var SH_SHEBANG = /^(#![^\n]+)/gm;

  var TERM_PROMPT = /^(\$\s)/gm;
  var TERM_CMD =
    /(?:^\$ )(dippin|go|git|brew|just|curl|pytest|npm|npx|tree-sitter)\b/gm;

  function isDippin(text) {
    return /\b(workflow|agent|human|tool|edges|defaults)\b/.test(text);
  }

  function isShell(text) {
    return (
      /^#!\/bin\/(sh|bash)/.test(text.trim()) ||
      /^\s*(set -|#!/)/.test(text.trim())
    );
  }

  function isTerminal(text) {
    return /^\$\s/.test(text.trim());
  }

  function isDiagnostic(text) {
    return /\b(error|warning|hint)\[DIP\d+\]/.test(text);
  }

  function esc(s) {
    return s
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;");
  }

  // Protect already-highlighted spans from re-processing.
  var PH = [];
  function protect(html) {
    return html.replace(/<span class="hl-[^"]*">[^<]*<\/span>/g, function (m) {
      var idx = PH.length;
      PH.push(m);
      return "\x00PH" + idx + "\x00";
    });
  }
  function restore(html) {
    return html.replace(/\x00PH(\d+)\x00/g, function (_, i) {
      return PH[+i];
    });
  }

  function highlightDippin(text) {
    PH = [];
    var h = esc(text);
    // Order matters: strings and comments first, then keywords.
    h = h.replace(DIP_CMT, '<span class="hl-cmt">$1</span>');
    h = protect(h);
    h = h.replace(DIP_STR, '<span class="hl-str">$1</span>');
    h = protect(h);
    h = h.replace(DIP_VAR, '<span class="hl-shvar">$1</span>');
    h = protect(h);
    h = h.replace(
      DIP_NODE_DECL,
      '<span class="hl-kw">$1</span> <span class="hl-node">$2</span>'
    );
    h = protect(h);
    h = h.replace(DIP_KW, '<span class="hl-kw">$1</span>');
    h = protect(h);
    h = h.replace(DIP_COND, '<span class="hl-cond">$1</span>');
    h = protect(h);
    h = h.replace(DIP_BOOL, '<span class="hl-bool">$1</span>');
    h = protect(h);
    h = h.replace(DIP_OP, '<span class="hl-op">$1</span>');
    h = protect(h);
    h = h.replace(DIP_FIELD, '<span class="hl-field">$1</span>$2');
    h = protect(h);
    h = restore(h);
    return h;
  }

  function highlightShell(text) {
    PH = [];
    var h = esc(text);
    h = h.replace(SH_SHEBANG, '<span class="hl-cmt">$1</span>');
    h = protect(h);
    h = h.replace(SH_CMT, '<span class="hl-cmt">$1</span>');
    h = protect(h);
    h = h.replace(SH_STR, '<span class="hl-str">$1</span>');
    h = protect(h);
    h = h.replace(SH_VAR, '<span class="hl-shvar">$1</span>');
    h = protect(h);
    h = h.replace(SH_KW, '<span class="hl-shkw">$1</span>');
    h = protect(h);
    h = restore(h);
    return h;
  }

  function highlightTerminal(text) {
    PH = [];
    var h = esc(text);
    // Color diagnostic lines.
    h = h.replace(
      /\b(error)(\[DIP\d+\])/g,
      '<span class="hl-fail">$1$2</span>'
    );
    h = h.replace(
      /\b(warning)(\[DIP\d+\])/g,
      '<span class="hl-warn">$1$2</span>'
    );
    h = h.replace(
      /\b(hint)(\[DIP\d+\])/g,
      '<span class="hl-dim">$1$2</span>'
    );
    h = protect(h);
    h = h.replace(
      /^(\$) ([\w-]+)/gm,
      '<span class="hl-prompt">$1</span> <span class="hl-cmd">$2</span>'
    );
    h = protect(h);
    h = h.replace(
      /\b(PASS)\b/g,
      '<span class="hl-pass">$1</span>'
    );
    h = h.replace(
      /\b(FAIL)\b/g,
      '<span class="hl-fail">$1</span>'
    );
    h = protect(h);
    // Flags.
    h = h.replace(
      /\s(--?\w[\w-]*)/g,
      ' <span class="hl-flag">$1</span>'
    );
    h = protect(h);
    // Filenames with extensions.
    h = h.replace(
      /\b([\w./-]+\.(dip|dot|json|csv|dip))\b/g,
      '<span class="hl-file">$1</span>'
    );
    h = protect(h);
    h = restore(h);
    return h;
  }

  function highlightDiagnostic(text) {
    PH = [];
    var h = esc(text);
    h = h.replace(
      /\b(error)(\[DIP\d+\])/g,
      '<span class="hl-fail">$1$2</span>'
    );
    h = h.replace(
      /\b(warning)(\[DIP\d+\])/g,
      '<span class="hl-warn">$1$2</span>'
    );
    h = h.replace(
      /\b(hint)(\[DIP\d+\])/g,
      '<span class="hl-dim">$1$2</span>'
    );
    h = protect(h);
    h = h.replace(DIP_STR, '<span class="hl-str">$1</span>');
    h = protect(h);
    h = h.replace(
      /(--&gt;\s*[\w./]+:\d+:\d+)/g,
      '<span class="hl-dim">$1</span>'
    );
    h = protect(h);
    h = h.replace(
      /(= help:.*)/g,
      '<span class="hl-dim">$1</span>'
    );
    h = protect(h);
    h = restore(h);
    return h;
  }

  document.addEventListener("DOMContentLoaded", function () {
    var pres = document.querySelectorAll(".doc-body pre");
    pres.forEach(function (pre) {
      // Skip already-highlighted blocks (manual <span> tags).
      if (pre.querySelector("span")) return;

      var raw = pre.textContent;
      var html;

      if (isDiagnostic(raw)) {
        html = highlightDiagnostic(raw);
      } else if (isDippin(raw)) {
        html = highlightDippin(raw);
      } else if (isTerminal(raw)) {
        html = highlightTerminal(raw);
      } else if (isShell(raw)) {
        html = highlightShell(raw);
      } else {
        return; // Leave unknown blocks alone.
      }

      pre.innerHTML = html;
    });
  });
})();
