// Tree-sitter grammar for the Dippin workflow language.
// Dippin is indentation-sensitive (like Python); INDENT/DEDENT tokens
// are produced by the external scanner in src/scanner.c.
//
// The scanner emits exactly one of INDENT, DEDENT, or NEWLINE per line
// boundary. Rules that contain indented blocks use _indent/_dedent
// directly. Line-ending _newline tokens are consumed at the repeat
// level, not inside individual field rules, so the scanner can emit
// DEDENT when valid.

module.exports = grammar({
  name: "dippin",

  externals: ($) => [$._indent, $._dedent, $._newline],

  extras: ($) => [/[ \t\r]/, $.comment],

  word: ($) => $.identifier,

  rules: {
    source_file: ($) => seq(repeat($._newline), $.workflow_decl),

    workflow_decl: ($) =>
      seq("workflow", $.identifier, $._indent, $.workflow_body, $._dedent),

    workflow_body: ($) =>
      repeat1(
        choice(
          $.workflow_field,
          $.defaults_section,
          $.node_decl,
          $.edges_section,
          $.stylesheet_section,
          $._newline
        )
      ),

    workflow_field: ($) =>
      seq(choice("goal", "start", "exit", "requires"), ":", $.field_value),

    // ── Defaults ──────────────────────────────────────────────
    defaults_section: ($) =>
      seq("defaults", $._indent, repeat1(choice($.defaults_field, $._newline)), $._dedent),

    defaults_field: ($) => seq($.field_name, ":", $.field_value),

    // ── Nodes ─────────────────────────────────────────────────
    node_decl: ($) =>
      choice(
        $.agent_node,
        $.human_node,
        $.tool_node,
        $.subgraph_node,
        $.conditional_node,
        $.parallel_node,
        $.fan_in_node,
        $.manager_loop_node
      ),

    agent_node: ($) =>
      seq("agent", $.identifier, $._indent, repeat1(choice($.node_field, $._newline)), $._dedent),

    human_node: ($) =>
      seq("human", $.identifier, $._indent, repeat1(choice($.node_field, $._newline)), $._dedent),

    tool_node: ($) =>
      seq("tool", $.identifier, $._indent, repeat1(choice($.node_field, $._newline)), $._dedent),

    subgraph_node: ($) =>
      seq("subgraph", $.identifier, $._indent, repeat1(choice($.node_field, $._newline)), $._dedent),

    conditional_node: ($) =>
      seq("conditional", $.identifier, $._indent, repeat1(choice($.node_field, $._newline)), $._dedent),

    manager_loop_node: ($) =>
      seq("manager_loop", $.identifier, $._indent, repeat1(choice($.node_field, $._newline)), $._dedent),

    parallel_node: ($) =>
      seq(
        "parallel",
        $.identifier,
        "->",
        $.identifier_list,
        choice($._newline, $.node_attr_block)
      ),

    fan_in_node: ($) =>
      seq(
        "fan_in",
        $.identifier,
        "<-",
        $.identifier_list,
        choice($._newline, $.node_attr_block)
      ),

    // An optional indented block of node fields (e.g. a params: block) beneath
    // an inline parallel/fan_in declaration. When the scanner sees a more-indented
    // line after the inline declaration it emits INDENT directly (no NEWLINE),
    // mirroring how agent_node opens its body.
    node_attr_block: ($) =>
      seq($._indent, repeat1(choice($.node_field, $._newline)), $._dedent),

    node_field: ($) => seq($.field_name, ":", $.field_value),

    // ── Edges ─────────────────────────────────────────────────
    edges_section: ($) =>
      seq("edges", $._indent, repeat1(choice($.edge_entry, $._newline)), $._dedent),

    edge_entry: ($) =>
      seq($.identifier, "->", $.identifier, repeat($.edge_attr)),

    edge_attr: ($) =>
      choice(
        seq("when", $.condition),
        seq("on", $.identifier),
        "loop",
        seq("label", ":", $.field_value),
        seq("weight", ":", $.field_value),
        seq("restart", ":", $.field_value),
        seq("override", ":", $.field_value)
      ),

    condition: ($) => $.or_expr,

    or_expr: ($) => prec.left(1, seq($.and_expr, repeat(seq("or", $.and_expr)))),

    and_expr: ($) => prec.left(2, seq($.compare_expr, repeat(seq("and", $.compare_expr)))),

    compare_expr: ($) =>
      prec.left(
        3,
        seq(
          $.operand,
          optional(seq(optional("not"), $.compare_op, $.operand))
        )
      ),

    compare_op: ($) =>
      choice("==", "!=", "=", "contains", "startswith", "endswith", "in"),

    operand: ($) => choice($.variable, $.string, $.identifier),

    variable: ($) => seq($.identifier, ".", $.identifier),

    // ── Stylesheet ────────────────────────────────────────────
    stylesheet_section: ($) =>
      seq("stylesheet", ":", $._indent, repeat1(choice($.stylesheet_rule, $._newline)), $._dedent),

    stylesheet_rule: ($) =>
      seq(
        $.selector,
        $._indent,
        repeat1(choice(seq($.field_name, ":", $.field_value), $._newline)),
        $._dedent
      ),

    selector: ($) =>
      choice(
        "*",
        seq(".", $.identifier),
        seq("#", $.identifier),
        $.identifier
      ),

    // ── Shared ────────────────────────────────────────────────
    field_name: ($) => $.identifier,

    field_value: ($) =>
      choice($.string, $.multiline_block, $.raw_inline),

    raw_inline: ($) => token(prec(-1, /[^\n]+/)),

    multiline_block: ($) =>
      seq($._indent, $.block_content, $._dedent),

    block_content: ($) => repeat1(choice($.block_line, $._newline)),

    block_line: ($) => /[^\n]+/,

    identifier_list: ($) => seq($.identifier, repeat(seq(",", $.identifier))),

    identifier: ($) => /[a-zA-Z0-9][a-zA-Z0-9_\-]*/,

    string: ($) =>
      choice(
        // Double-quoted: \" and \\ escapes.
        seq('"', repeat(choice(/[^"\\]+/, /\\./)), '"'),
        // Single-quoted (YAML-style): literal content; the only escape is '' → '.
        seq("'", repeat(choice(/[^']+/, "''")), "'")
      ),

    comment: ($) => token(seq("#", /.*/)),
  },
});
