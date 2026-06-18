#include "tree_sitter/parser.h"

#if defined(__GNUC__) || defined(__clang__)
#pragma GCC diagnostic ignored "-Wmissing-field-initializers"
#endif

#define LANGUAGE_VERSION 14
#define STATE_COUNT 158
#define LARGE_STATE_COUNT 2
#define SYMBOL_COUNT 109
#define ALIAS_COUNT 0
#define TOKEN_COUNT 57
#define EXTERNAL_TOKEN_COUNT 3
#define FIELD_COUNT 0
#define MAX_ALIAS_SEQUENCE_LENGTH 5
#define PRODUCTION_ID_COUNT 1

enum ts_symbol_identifiers {
  sym_identifier = 1,
  anon_sym_dip = 2,
  anon_sym_workflow = 3,
  anon_sym_goal = 4,
  anon_sym_start = 5,
  anon_sym_exit = 6,
  anon_sym_requires = 7,
  anon_sym_COLON = 8,
  anon_sym_defaults = 9,
  anon_sym_agent = 10,
  anon_sym_human = 11,
  anon_sym_tool = 12,
  anon_sym_subgraph = 13,
  anon_sym_conditional = 14,
  anon_sym_manager_loop = 15,
  anon_sym_parallel = 16,
  anon_sym_DASH_GT = 17,
  anon_sym_fan_in = 18,
  anon_sym_LT_DASH = 19,
  anon_sym_edges = 20,
  anon_sym_else = 21,
  anon_sym_when = 22,
  anon_sym_on = 23,
  anon_sym_loop = 24,
  anon_sym_label = 25,
  anon_sym_choice = 26,
  anon_sym_weight = 27,
  anon_sym_restart = 28,
  anon_sym_override = 29,
  anon_sym_or = 30,
  anon_sym_and = 31,
  anon_sym_not = 32,
  anon_sym_EQ_EQ = 33,
  anon_sym_BANG_EQ = 34,
  anon_sym_EQ = 35,
  anon_sym_contains = 36,
  anon_sym_startswith = 37,
  anon_sym_endswith = 38,
  anon_sym_in = 39,
  anon_sym_DOT = 40,
  anon_sym_stylesheet = 41,
  anon_sym_STAR = 42,
  anon_sym_POUND = 43,
  sym_raw_inline = 44,
  sym_block_line = 45,
  anon_sym_COMMA = 46,
  anon_sym_DQUOTE = 47,
  aux_sym_string_token1 = 48,
  aux_sym_string_token2 = 49,
  anon_sym_SQUOTE = 50,
  aux_sym_string_token3 = 51,
  anon_sym_SQUOTE_SQUOTE = 52,
  sym_comment = 53,
  sym__indent = 54,
  sym__dedent = 55,
  sym__newline = 56,
  sym_source_file = 57,
  sym_version_decl = 58,
  sym_workflow_decl = 59,
  sym_workflow_body = 60,
  sym_workflow_field = 61,
  sym_defaults_section = 62,
  sym_defaults_field = 63,
  sym_node_decl = 64,
  sym_agent_node = 65,
  sym_human_node = 66,
  sym_tool_node = 67,
  sym_subgraph_node = 68,
  sym_conditional_node = 69,
  sym_manager_loop_node = 70,
  sym_parallel_node = 71,
  sym_fan_in_node = 72,
  sym_node_attr_block = 73,
  sym_node_field = 74,
  sym_edges_section = 75,
  sym_edge_entry = 76,
  sym_else_default = 77,
  sym_edge_attr = 78,
  sym_condition = 79,
  sym_or_expr = 80,
  sym_and_expr = 81,
  sym_compare_expr = 82,
  sym_compare_op = 83,
  sym_operand = 84,
  sym_variable = 85,
  sym_stylesheet_section = 86,
  sym_stylesheet_rule = 87,
  sym_selector = 88,
  sym_field_name = 89,
  sym_field_value = 90,
  sym_multiline_block = 91,
  sym_block_content = 92,
  sym_identifier_list = 93,
  sym_string = 94,
  aux_sym_source_file_repeat1 = 95,
  aux_sym_workflow_body_repeat1 = 96,
  aux_sym_defaults_section_repeat1 = 97,
  aux_sym_agent_node_repeat1 = 98,
  aux_sym_edges_section_repeat1 = 99,
  aux_sym_edge_entry_repeat1 = 100,
  aux_sym_or_expr_repeat1 = 101,
  aux_sym_and_expr_repeat1 = 102,
  aux_sym_stylesheet_section_repeat1 = 103,
  aux_sym_stylesheet_rule_repeat1 = 104,
  aux_sym_block_content_repeat1 = 105,
  aux_sym_identifier_list_repeat1 = 106,
  aux_sym_string_repeat1 = 107,
  aux_sym_string_repeat2 = 108,
};

static const char * const ts_symbol_names[] = {
  [ts_builtin_sym_end] = "end",
  [sym_identifier] = "identifier",
  [anon_sym_dip] = "dip",
  [anon_sym_workflow] = "workflow",
  [anon_sym_goal] = "goal",
  [anon_sym_start] = "start",
  [anon_sym_exit] = "exit",
  [anon_sym_requires] = "requires",
  [anon_sym_COLON] = ":",
  [anon_sym_defaults] = "defaults",
  [anon_sym_agent] = "agent",
  [anon_sym_human] = "human",
  [anon_sym_tool] = "tool",
  [anon_sym_subgraph] = "subgraph",
  [anon_sym_conditional] = "conditional",
  [anon_sym_manager_loop] = "manager_loop",
  [anon_sym_parallel] = "parallel",
  [anon_sym_DASH_GT] = "->",
  [anon_sym_fan_in] = "fan_in",
  [anon_sym_LT_DASH] = "<-",
  [anon_sym_edges] = "edges",
  [anon_sym_else] = "else",
  [anon_sym_when] = "when",
  [anon_sym_on] = "on",
  [anon_sym_loop] = "loop",
  [anon_sym_label] = "label",
  [anon_sym_choice] = "choice",
  [anon_sym_weight] = "weight",
  [anon_sym_restart] = "restart",
  [anon_sym_override] = "override",
  [anon_sym_or] = "or",
  [anon_sym_and] = "and",
  [anon_sym_not] = "not",
  [anon_sym_EQ_EQ] = "==",
  [anon_sym_BANG_EQ] = "!=",
  [anon_sym_EQ] = "=",
  [anon_sym_contains] = "contains",
  [anon_sym_startswith] = "startswith",
  [anon_sym_endswith] = "endswith",
  [anon_sym_in] = "in",
  [anon_sym_DOT] = ".",
  [anon_sym_stylesheet] = "stylesheet",
  [anon_sym_STAR] = "*",
  [anon_sym_POUND] = "#",
  [sym_raw_inline] = "raw_inline",
  [sym_block_line] = "block_line",
  [anon_sym_COMMA] = ",",
  [anon_sym_DQUOTE] = "\"",
  [aux_sym_string_token1] = "string_token1",
  [aux_sym_string_token2] = "string_token2",
  [anon_sym_SQUOTE] = "'",
  [aux_sym_string_token3] = "string_token3",
  [anon_sym_SQUOTE_SQUOTE] = "''",
  [sym_comment] = "comment",
  [sym__indent] = "_indent",
  [sym__dedent] = "_dedent",
  [sym__newline] = "_newline",
  [sym_source_file] = "source_file",
  [sym_version_decl] = "version_decl",
  [sym_workflow_decl] = "workflow_decl",
  [sym_workflow_body] = "workflow_body",
  [sym_workflow_field] = "workflow_field",
  [sym_defaults_section] = "defaults_section",
  [sym_defaults_field] = "defaults_field",
  [sym_node_decl] = "node_decl",
  [sym_agent_node] = "agent_node",
  [sym_human_node] = "human_node",
  [sym_tool_node] = "tool_node",
  [sym_subgraph_node] = "subgraph_node",
  [sym_conditional_node] = "conditional_node",
  [sym_manager_loop_node] = "manager_loop_node",
  [sym_parallel_node] = "parallel_node",
  [sym_fan_in_node] = "fan_in_node",
  [sym_node_attr_block] = "node_attr_block",
  [sym_node_field] = "node_field",
  [sym_edges_section] = "edges_section",
  [sym_edge_entry] = "edge_entry",
  [sym_else_default] = "else_default",
  [sym_edge_attr] = "edge_attr",
  [sym_condition] = "condition",
  [sym_or_expr] = "or_expr",
  [sym_and_expr] = "and_expr",
  [sym_compare_expr] = "compare_expr",
  [sym_compare_op] = "compare_op",
  [sym_operand] = "operand",
  [sym_variable] = "variable",
  [sym_stylesheet_section] = "stylesheet_section",
  [sym_stylesheet_rule] = "stylesheet_rule",
  [sym_selector] = "selector",
  [sym_field_name] = "field_name",
  [sym_field_value] = "field_value",
  [sym_multiline_block] = "multiline_block",
  [sym_block_content] = "block_content",
  [sym_identifier_list] = "identifier_list",
  [sym_string] = "string",
  [aux_sym_source_file_repeat1] = "source_file_repeat1",
  [aux_sym_workflow_body_repeat1] = "workflow_body_repeat1",
  [aux_sym_defaults_section_repeat1] = "defaults_section_repeat1",
  [aux_sym_agent_node_repeat1] = "agent_node_repeat1",
  [aux_sym_edges_section_repeat1] = "edges_section_repeat1",
  [aux_sym_edge_entry_repeat1] = "edge_entry_repeat1",
  [aux_sym_or_expr_repeat1] = "or_expr_repeat1",
  [aux_sym_and_expr_repeat1] = "and_expr_repeat1",
  [aux_sym_stylesheet_section_repeat1] = "stylesheet_section_repeat1",
  [aux_sym_stylesheet_rule_repeat1] = "stylesheet_rule_repeat1",
  [aux_sym_block_content_repeat1] = "block_content_repeat1",
  [aux_sym_identifier_list_repeat1] = "identifier_list_repeat1",
  [aux_sym_string_repeat1] = "string_repeat1",
  [aux_sym_string_repeat2] = "string_repeat2",
};

static const TSSymbol ts_symbol_map[] = {
  [ts_builtin_sym_end] = ts_builtin_sym_end,
  [sym_identifier] = sym_identifier,
  [anon_sym_dip] = anon_sym_dip,
  [anon_sym_workflow] = anon_sym_workflow,
  [anon_sym_goal] = anon_sym_goal,
  [anon_sym_start] = anon_sym_start,
  [anon_sym_exit] = anon_sym_exit,
  [anon_sym_requires] = anon_sym_requires,
  [anon_sym_COLON] = anon_sym_COLON,
  [anon_sym_defaults] = anon_sym_defaults,
  [anon_sym_agent] = anon_sym_agent,
  [anon_sym_human] = anon_sym_human,
  [anon_sym_tool] = anon_sym_tool,
  [anon_sym_subgraph] = anon_sym_subgraph,
  [anon_sym_conditional] = anon_sym_conditional,
  [anon_sym_manager_loop] = anon_sym_manager_loop,
  [anon_sym_parallel] = anon_sym_parallel,
  [anon_sym_DASH_GT] = anon_sym_DASH_GT,
  [anon_sym_fan_in] = anon_sym_fan_in,
  [anon_sym_LT_DASH] = anon_sym_LT_DASH,
  [anon_sym_edges] = anon_sym_edges,
  [anon_sym_else] = anon_sym_else,
  [anon_sym_when] = anon_sym_when,
  [anon_sym_on] = anon_sym_on,
  [anon_sym_loop] = anon_sym_loop,
  [anon_sym_label] = anon_sym_label,
  [anon_sym_choice] = anon_sym_choice,
  [anon_sym_weight] = anon_sym_weight,
  [anon_sym_restart] = anon_sym_restart,
  [anon_sym_override] = anon_sym_override,
  [anon_sym_or] = anon_sym_or,
  [anon_sym_and] = anon_sym_and,
  [anon_sym_not] = anon_sym_not,
  [anon_sym_EQ_EQ] = anon_sym_EQ_EQ,
  [anon_sym_BANG_EQ] = anon_sym_BANG_EQ,
  [anon_sym_EQ] = anon_sym_EQ,
  [anon_sym_contains] = anon_sym_contains,
  [anon_sym_startswith] = anon_sym_startswith,
  [anon_sym_endswith] = anon_sym_endswith,
  [anon_sym_in] = anon_sym_in,
  [anon_sym_DOT] = anon_sym_DOT,
  [anon_sym_stylesheet] = anon_sym_stylesheet,
  [anon_sym_STAR] = anon_sym_STAR,
  [anon_sym_POUND] = anon_sym_POUND,
  [sym_raw_inline] = sym_raw_inline,
  [sym_block_line] = sym_block_line,
  [anon_sym_COMMA] = anon_sym_COMMA,
  [anon_sym_DQUOTE] = anon_sym_DQUOTE,
  [aux_sym_string_token1] = aux_sym_string_token1,
  [aux_sym_string_token2] = aux_sym_string_token2,
  [anon_sym_SQUOTE] = anon_sym_SQUOTE,
  [aux_sym_string_token3] = aux_sym_string_token3,
  [anon_sym_SQUOTE_SQUOTE] = anon_sym_SQUOTE_SQUOTE,
  [sym_comment] = sym_comment,
  [sym__indent] = sym__indent,
  [sym__dedent] = sym__dedent,
  [sym__newline] = sym__newline,
  [sym_source_file] = sym_source_file,
  [sym_version_decl] = sym_version_decl,
  [sym_workflow_decl] = sym_workflow_decl,
  [sym_workflow_body] = sym_workflow_body,
  [sym_workflow_field] = sym_workflow_field,
  [sym_defaults_section] = sym_defaults_section,
  [sym_defaults_field] = sym_defaults_field,
  [sym_node_decl] = sym_node_decl,
  [sym_agent_node] = sym_agent_node,
  [sym_human_node] = sym_human_node,
  [sym_tool_node] = sym_tool_node,
  [sym_subgraph_node] = sym_subgraph_node,
  [sym_conditional_node] = sym_conditional_node,
  [sym_manager_loop_node] = sym_manager_loop_node,
  [sym_parallel_node] = sym_parallel_node,
  [sym_fan_in_node] = sym_fan_in_node,
  [sym_node_attr_block] = sym_node_attr_block,
  [sym_node_field] = sym_node_field,
  [sym_edges_section] = sym_edges_section,
  [sym_edge_entry] = sym_edge_entry,
  [sym_else_default] = sym_else_default,
  [sym_edge_attr] = sym_edge_attr,
  [sym_condition] = sym_condition,
  [sym_or_expr] = sym_or_expr,
  [sym_and_expr] = sym_and_expr,
  [sym_compare_expr] = sym_compare_expr,
  [sym_compare_op] = sym_compare_op,
  [sym_operand] = sym_operand,
  [sym_variable] = sym_variable,
  [sym_stylesheet_section] = sym_stylesheet_section,
  [sym_stylesheet_rule] = sym_stylesheet_rule,
  [sym_selector] = sym_selector,
  [sym_field_name] = sym_field_name,
  [sym_field_value] = sym_field_value,
  [sym_multiline_block] = sym_multiline_block,
  [sym_block_content] = sym_block_content,
  [sym_identifier_list] = sym_identifier_list,
  [sym_string] = sym_string,
  [aux_sym_source_file_repeat1] = aux_sym_source_file_repeat1,
  [aux_sym_workflow_body_repeat1] = aux_sym_workflow_body_repeat1,
  [aux_sym_defaults_section_repeat1] = aux_sym_defaults_section_repeat1,
  [aux_sym_agent_node_repeat1] = aux_sym_agent_node_repeat1,
  [aux_sym_edges_section_repeat1] = aux_sym_edges_section_repeat1,
  [aux_sym_edge_entry_repeat1] = aux_sym_edge_entry_repeat1,
  [aux_sym_or_expr_repeat1] = aux_sym_or_expr_repeat1,
  [aux_sym_and_expr_repeat1] = aux_sym_and_expr_repeat1,
  [aux_sym_stylesheet_section_repeat1] = aux_sym_stylesheet_section_repeat1,
  [aux_sym_stylesheet_rule_repeat1] = aux_sym_stylesheet_rule_repeat1,
  [aux_sym_block_content_repeat1] = aux_sym_block_content_repeat1,
  [aux_sym_identifier_list_repeat1] = aux_sym_identifier_list_repeat1,
  [aux_sym_string_repeat1] = aux_sym_string_repeat1,
  [aux_sym_string_repeat2] = aux_sym_string_repeat2,
};

static const TSSymbolMetadata ts_symbol_metadata[] = {
  [ts_builtin_sym_end] = {
    .visible = false,
    .named = true,
  },
  [sym_identifier] = {
    .visible = true,
    .named = true,
  },
  [anon_sym_dip] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_workflow] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_goal] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_start] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_exit] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_requires] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_COLON] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_defaults] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_agent] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_human] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_tool] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_subgraph] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_conditional] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_manager_loop] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_parallel] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_DASH_GT] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_fan_in] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_LT_DASH] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_edges] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_else] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_when] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_on] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_loop] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_label] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_choice] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_weight] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_restart] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_override] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_or] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_and] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_not] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_EQ_EQ] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_BANG_EQ] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_EQ] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_contains] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_startswith] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_endswith] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_in] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_DOT] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_stylesheet] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_STAR] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_POUND] = {
    .visible = true,
    .named = false,
  },
  [sym_raw_inline] = {
    .visible = true,
    .named = true,
  },
  [sym_block_line] = {
    .visible = true,
    .named = true,
  },
  [anon_sym_COMMA] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_DQUOTE] = {
    .visible = true,
    .named = false,
  },
  [aux_sym_string_token1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_string_token2] = {
    .visible = false,
    .named = false,
  },
  [anon_sym_SQUOTE] = {
    .visible = true,
    .named = false,
  },
  [aux_sym_string_token3] = {
    .visible = false,
    .named = false,
  },
  [anon_sym_SQUOTE_SQUOTE] = {
    .visible = true,
    .named = false,
  },
  [sym_comment] = {
    .visible = true,
    .named = true,
  },
  [sym__indent] = {
    .visible = false,
    .named = true,
  },
  [sym__dedent] = {
    .visible = false,
    .named = true,
  },
  [sym__newline] = {
    .visible = false,
    .named = true,
  },
  [sym_source_file] = {
    .visible = true,
    .named = true,
  },
  [sym_version_decl] = {
    .visible = true,
    .named = true,
  },
  [sym_workflow_decl] = {
    .visible = true,
    .named = true,
  },
  [sym_workflow_body] = {
    .visible = true,
    .named = true,
  },
  [sym_workflow_field] = {
    .visible = true,
    .named = true,
  },
  [sym_defaults_section] = {
    .visible = true,
    .named = true,
  },
  [sym_defaults_field] = {
    .visible = true,
    .named = true,
  },
  [sym_node_decl] = {
    .visible = true,
    .named = true,
  },
  [sym_agent_node] = {
    .visible = true,
    .named = true,
  },
  [sym_human_node] = {
    .visible = true,
    .named = true,
  },
  [sym_tool_node] = {
    .visible = true,
    .named = true,
  },
  [sym_subgraph_node] = {
    .visible = true,
    .named = true,
  },
  [sym_conditional_node] = {
    .visible = true,
    .named = true,
  },
  [sym_manager_loop_node] = {
    .visible = true,
    .named = true,
  },
  [sym_parallel_node] = {
    .visible = true,
    .named = true,
  },
  [sym_fan_in_node] = {
    .visible = true,
    .named = true,
  },
  [sym_node_attr_block] = {
    .visible = true,
    .named = true,
  },
  [sym_node_field] = {
    .visible = true,
    .named = true,
  },
  [sym_edges_section] = {
    .visible = true,
    .named = true,
  },
  [sym_edge_entry] = {
    .visible = true,
    .named = true,
  },
  [sym_else_default] = {
    .visible = true,
    .named = true,
  },
  [sym_edge_attr] = {
    .visible = true,
    .named = true,
  },
  [sym_condition] = {
    .visible = true,
    .named = true,
  },
  [sym_or_expr] = {
    .visible = true,
    .named = true,
  },
  [sym_and_expr] = {
    .visible = true,
    .named = true,
  },
  [sym_compare_expr] = {
    .visible = true,
    .named = true,
  },
  [sym_compare_op] = {
    .visible = true,
    .named = true,
  },
  [sym_operand] = {
    .visible = true,
    .named = true,
  },
  [sym_variable] = {
    .visible = true,
    .named = true,
  },
  [sym_stylesheet_section] = {
    .visible = true,
    .named = true,
  },
  [sym_stylesheet_rule] = {
    .visible = true,
    .named = true,
  },
  [sym_selector] = {
    .visible = true,
    .named = true,
  },
  [sym_field_name] = {
    .visible = true,
    .named = true,
  },
  [sym_field_value] = {
    .visible = true,
    .named = true,
  },
  [sym_multiline_block] = {
    .visible = true,
    .named = true,
  },
  [sym_block_content] = {
    .visible = true,
    .named = true,
  },
  [sym_identifier_list] = {
    .visible = true,
    .named = true,
  },
  [sym_string] = {
    .visible = true,
    .named = true,
  },
  [aux_sym_source_file_repeat1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_workflow_body_repeat1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_defaults_section_repeat1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_agent_node_repeat1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_edges_section_repeat1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_edge_entry_repeat1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_or_expr_repeat1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_and_expr_repeat1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_stylesheet_section_repeat1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_stylesheet_rule_repeat1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_block_content_repeat1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_identifier_list_repeat1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_string_repeat1] = {
    .visible = false,
    .named = false,
  },
  [aux_sym_string_repeat2] = {
    .visible = false,
    .named = false,
  },
};

static const TSSymbol ts_alias_sequences[PRODUCTION_ID_COUNT][MAX_ALIAS_SEQUENCE_LENGTH] = {
  [0] = {0},
};

static const uint16_t ts_non_terminal_alias_map[] = {
  0,
};

static const TSStateId ts_primary_state_ids[STATE_COUNT] = {
  [0] = 0,
  [1] = 1,
  [2] = 2,
  [3] = 3,
  [4] = 4,
  [5] = 5,
  [6] = 6,
  [7] = 7,
  [8] = 8,
  [9] = 9,
  [10] = 10,
  [11] = 11,
  [12] = 12,
  [13] = 13,
  [14] = 14,
  [15] = 15,
  [16] = 16,
  [17] = 17,
  [18] = 18,
  [19] = 19,
  [20] = 20,
  [21] = 21,
  [22] = 22,
  [23] = 23,
  [24] = 24,
  [25] = 25,
  [26] = 26,
  [27] = 27,
  [28] = 28,
  [29] = 29,
  [30] = 30,
  [31] = 31,
  [32] = 32,
  [33] = 33,
  [34] = 34,
  [35] = 35,
  [36] = 36,
  [37] = 37,
  [38] = 38,
  [39] = 39,
  [40] = 40,
  [41] = 41,
  [42] = 42,
  [43] = 43,
  [44] = 44,
  [45] = 45,
  [46] = 46,
  [47] = 47,
  [48] = 48,
  [49] = 49,
  [50] = 50,
  [51] = 51,
  [52] = 52,
  [53] = 53,
  [54] = 54,
  [55] = 55,
  [56] = 56,
  [57] = 57,
  [58] = 58,
  [59] = 59,
  [60] = 60,
  [61] = 61,
  [62] = 62,
  [63] = 63,
  [64] = 64,
  [65] = 65,
  [66] = 66,
  [67] = 67,
  [68] = 68,
  [69] = 69,
  [70] = 70,
  [71] = 71,
  [72] = 72,
  [73] = 73,
  [74] = 74,
  [75] = 75,
  [76] = 76,
  [77] = 77,
  [78] = 78,
  [79] = 79,
  [80] = 80,
  [81] = 81,
  [82] = 82,
  [83] = 83,
  [84] = 84,
  [85] = 85,
  [86] = 86,
  [87] = 87,
  [88] = 88,
  [89] = 89,
  [90] = 90,
  [91] = 91,
  [92] = 92,
  [93] = 93,
  [94] = 94,
  [95] = 95,
  [96] = 96,
  [97] = 97,
  [98] = 98,
  [99] = 99,
  [100] = 100,
  [101] = 101,
  [102] = 102,
  [103] = 103,
  [104] = 104,
  [105] = 105,
  [106] = 106,
  [107] = 107,
  [108] = 108,
  [109] = 109,
  [110] = 110,
  [111] = 111,
  [112] = 112,
  [113] = 113,
  [114] = 114,
  [115] = 115,
  [116] = 116,
  [117] = 117,
  [118] = 118,
  [119] = 119,
  [120] = 120,
  [121] = 121,
  [122] = 122,
  [123] = 123,
  [124] = 124,
  [125] = 125,
  [126] = 126,
  [127] = 127,
  [128] = 128,
  [129] = 129,
  [130] = 130,
  [131] = 131,
  [132] = 132,
  [133] = 133,
  [134] = 134,
  [135] = 135,
  [136] = 136,
  [137] = 137,
  [138] = 138,
  [139] = 139,
  [140] = 140,
  [141] = 141,
  [142] = 142,
  [143] = 143,
  [144] = 144,
  [145] = 145,
  [146] = 146,
  [147] = 147,
  [148] = 148,
  [149] = 149,
  [150] = 150,
  [151] = 151,
  [152] = 152,
  [153] = 153,
  [154] = 154,
  [155] = 155,
  [156] = 156,
  [157] = 157,
};

static bool ts_lex(TSLexer *lexer, TSStateId state) {
  START_LEXER();
  eof = lexer->eof(lexer);
  switch (state) {
    case 0:
      if (eof) ADVANCE(10);
      ADVANCE_MAP(
        '!', 6,
        '"', 26,
        '#', 19,
        '\'', 32,
        '*', 18,
        ',', 24,
        '-', 7,
        '.', 17,
        ':', 11,
        '<', 5,
        '=', 16,
        '\\', 8,
      );
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') SKIP(0);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(25);
      END_STATE();
    case 1:
      if (lookahead == '"') ADVANCE(26);
      if (lookahead == '#') ADVANCE(37);
      if (lookahead == '\'') ADVANCE(31);
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') ADVANCE(20);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n') ADVANCE(21);
      END_STATE();
    case 2:
      if (lookahead == '"') ADVANCE(26);
      if (lookahead == '#') ADVANCE(27);
      if (lookahead == '\\') ADVANCE(8);
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') ADVANCE(28);
      if (lookahead != 0) ADVANCE(29);
      END_STATE();
    case 3:
      if (lookahead == '#') ADVANCE(23);
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') ADVANCE(22);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n') ADVANCE(23);
      END_STATE();
    case 4:
      if (lookahead == '#') ADVANCE(33);
      if (lookahead == '\'') ADVANCE(32);
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') ADVANCE(34);
      if (lookahead != 0) ADVANCE(35);
      END_STATE();
    case 5:
      if (lookahead == '-') ADVANCE(13);
      END_STATE();
    case 6:
      if (lookahead == '=') ADVANCE(15);
      END_STATE();
    case 7:
      if (lookahead == '>') ADVANCE(12);
      END_STATE();
    case 8:
      if (lookahead != 0 &&
          lookahead != '\n') ADVANCE(30);
      END_STATE();
    case 9:
      if (eof) ADVANCE(10);
      ADVANCE_MAP(
        '!', 6,
        '"', 26,
        '#', 37,
        '\'', 31,
        ',', 24,
        '-', 7,
        '.', 17,
        ':', 11,
        '<', 5,
        '=', 16,
      );
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') SKIP(9);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(25);
      END_STATE();
    case 10:
      ACCEPT_TOKEN(ts_builtin_sym_end);
      END_STATE();
    case 11:
      ACCEPT_TOKEN(anon_sym_COLON);
      END_STATE();
    case 12:
      ACCEPT_TOKEN(anon_sym_DASH_GT);
      END_STATE();
    case 13:
      ACCEPT_TOKEN(anon_sym_LT_DASH);
      END_STATE();
    case 14:
      ACCEPT_TOKEN(anon_sym_EQ_EQ);
      END_STATE();
    case 15:
      ACCEPT_TOKEN(anon_sym_BANG_EQ);
      END_STATE();
    case 16:
      ACCEPT_TOKEN(anon_sym_EQ);
      if (lookahead == '=') ADVANCE(14);
      END_STATE();
    case 17:
      ACCEPT_TOKEN(anon_sym_DOT);
      END_STATE();
    case 18:
      ACCEPT_TOKEN(anon_sym_STAR);
      END_STATE();
    case 19:
      ACCEPT_TOKEN(anon_sym_POUND);
      if (lookahead != 0 &&
          lookahead != '\n') ADVANCE(37);
      END_STATE();
    case 20:
      ACCEPT_TOKEN(sym_raw_inline);
      if (lookahead == '"') ADVANCE(26);
      if (lookahead == '#') ADVANCE(37);
      if (lookahead == '\'') ADVANCE(31);
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') ADVANCE(20);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n') ADVANCE(21);
      END_STATE();
    case 21:
      ACCEPT_TOKEN(sym_raw_inline);
      if (lookahead != 0 &&
          lookahead != '\n') ADVANCE(21);
      END_STATE();
    case 22:
      ACCEPT_TOKEN(sym_block_line);
      if (lookahead == '#') ADVANCE(23);
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') ADVANCE(22);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n') ADVANCE(23);
      END_STATE();
    case 23:
      ACCEPT_TOKEN(sym_block_line);
      if (lookahead != 0 &&
          lookahead != '\n') ADVANCE(23);
      END_STATE();
    case 24:
      ACCEPT_TOKEN(anon_sym_COMMA);
      END_STATE();
    case 25:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '-' ||
          ('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(25);
      END_STATE();
    case 26:
      ACCEPT_TOKEN(anon_sym_DQUOTE);
      END_STATE();
    case 27:
      ACCEPT_TOKEN(aux_sym_string_token1);
      if (lookahead == '\n') ADVANCE(29);
      if (lookahead == '"' ||
          lookahead == '\\') ADVANCE(37);
      if (lookahead != 0) ADVANCE(27);
      END_STATE();
    case 28:
      ACCEPT_TOKEN(aux_sym_string_token1);
      if (lookahead == '#') ADVANCE(27);
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') ADVANCE(28);
      if (lookahead != 0 &&
          lookahead != '"' &&
          lookahead != '#' &&
          lookahead != '\\') ADVANCE(29);
      END_STATE();
    case 29:
      ACCEPT_TOKEN(aux_sym_string_token1);
      if (lookahead != 0 &&
          lookahead != '"' &&
          lookahead != '\\') ADVANCE(29);
      END_STATE();
    case 30:
      ACCEPT_TOKEN(aux_sym_string_token2);
      END_STATE();
    case 31:
      ACCEPT_TOKEN(anon_sym_SQUOTE);
      END_STATE();
    case 32:
      ACCEPT_TOKEN(anon_sym_SQUOTE);
      if (lookahead == '\'') ADVANCE(36);
      END_STATE();
    case 33:
      ACCEPT_TOKEN(aux_sym_string_token3);
      if (lookahead == '\n') ADVANCE(35);
      if (lookahead == '\'') ADVANCE(37);
      if (lookahead != 0) ADVANCE(33);
      END_STATE();
    case 34:
      ACCEPT_TOKEN(aux_sym_string_token3);
      if (lookahead == '#') ADVANCE(33);
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') ADVANCE(34);
      if (lookahead != 0 &&
          lookahead != '\'') ADVANCE(35);
      END_STATE();
    case 35:
      ACCEPT_TOKEN(aux_sym_string_token3);
      if (lookahead != 0 &&
          lookahead != '\'') ADVANCE(35);
      END_STATE();
    case 36:
      ACCEPT_TOKEN(anon_sym_SQUOTE_SQUOTE);
      END_STATE();
    case 37:
      ACCEPT_TOKEN(sym_comment);
      if (lookahead != 0 &&
          lookahead != '\n') ADVANCE(37);
      END_STATE();
    default:
      return false;
  }
}

static bool ts_lex_keywords(TSLexer *lexer, TSStateId state) {
  START_LEXER();
  eof = lexer->eof(lexer);
  switch (state) {
    case 0:
      ADVANCE_MAP(
        'a', 1,
        'c', 2,
        'd', 3,
        'e', 4,
        'f', 5,
        'g', 6,
        'h', 7,
        'i', 8,
        'l', 9,
        'm', 10,
        'n', 11,
        'o', 12,
        'p', 13,
        'r', 14,
        's', 15,
        't', 16,
        'w', 17,
      );
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') SKIP(0);
      END_STATE();
    case 1:
      if (lookahead == 'g') ADVANCE(18);
      if (lookahead == 'n') ADVANCE(19);
      END_STATE();
    case 2:
      if (lookahead == 'h') ADVANCE(20);
      if (lookahead == 'o') ADVANCE(21);
      END_STATE();
    case 3:
      if (lookahead == 'e') ADVANCE(22);
      if (lookahead == 'i') ADVANCE(23);
      END_STATE();
    case 4:
      if (lookahead == 'd') ADVANCE(24);
      if (lookahead == 'l') ADVANCE(25);
      if (lookahead == 'n') ADVANCE(26);
      if (lookahead == 'x') ADVANCE(27);
      END_STATE();
    case 5:
      if (lookahead == 'a') ADVANCE(28);
      END_STATE();
    case 6:
      if (lookahead == 'o') ADVANCE(29);
      END_STATE();
    case 7:
      if (lookahead == 'u') ADVANCE(30);
      END_STATE();
    case 8:
      if (lookahead == 'n') ADVANCE(31);
      END_STATE();
    case 9:
      if (lookahead == 'a') ADVANCE(32);
      if (lookahead == 'o') ADVANCE(33);
      END_STATE();
    case 10:
      if (lookahead == 'a') ADVANCE(34);
      END_STATE();
    case 11:
      if (lookahead == 'o') ADVANCE(35);
      END_STATE();
    case 12:
      if (lookahead == 'n') ADVANCE(36);
      if (lookahead == 'r') ADVANCE(37);
      if (lookahead == 'v') ADVANCE(38);
      END_STATE();
    case 13:
      if (lookahead == 'a') ADVANCE(39);
      END_STATE();
    case 14:
      if (lookahead == 'e') ADVANCE(40);
      END_STATE();
    case 15:
      if (lookahead == 't') ADVANCE(41);
      if (lookahead == 'u') ADVANCE(42);
      END_STATE();
    case 16:
      if (lookahead == 'o') ADVANCE(43);
      END_STATE();
    case 17:
      if (lookahead == 'e') ADVANCE(44);
      if (lookahead == 'h') ADVANCE(45);
      if (lookahead == 'o') ADVANCE(46);
      END_STATE();
    case 18:
      if (lookahead == 'e') ADVANCE(47);
      END_STATE();
    case 19:
      if (lookahead == 'd') ADVANCE(48);
      END_STATE();
    case 20:
      if (lookahead == 'o') ADVANCE(49);
      END_STATE();
    case 21:
      if (lookahead == 'n') ADVANCE(50);
      END_STATE();
    case 22:
      if (lookahead == 'f') ADVANCE(51);
      END_STATE();
    case 23:
      if (lookahead == 'p') ADVANCE(52);
      END_STATE();
    case 24:
      if (lookahead == 'g') ADVANCE(53);
      END_STATE();
    case 25:
      if (lookahead == 's') ADVANCE(54);
      END_STATE();
    case 26:
      if (lookahead == 'd') ADVANCE(55);
      END_STATE();
    case 27:
      if (lookahead == 'i') ADVANCE(56);
      END_STATE();
    case 28:
      if (lookahead == 'n') ADVANCE(57);
      END_STATE();
    case 29:
      if (lookahead == 'a') ADVANCE(58);
      END_STATE();
    case 30:
      if (lookahead == 'm') ADVANCE(59);
      END_STATE();
    case 31:
      ACCEPT_TOKEN(anon_sym_in);
      END_STATE();
    case 32:
      if (lookahead == 'b') ADVANCE(60);
      END_STATE();
    case 33:
      if (lookahead == 'o') ADVANCE(61);
      END_STATE();
    case 34:
      if (lookahead == 'n') ADVANCE(62);
      END_STATE();
    case 35:
      if (lookahead == 't') ADVANCE(63);
      END_STATE();
    case 36:
      ACCEPT_TOKEN(anon_sym_on);
      END_STATE();
    case 37:
      ACCEPT_TOKEN(anon_sym_or);
      END_STATE();
    case 38:
      if (lookahead == 'e') ADVANCE(64);
      END_STATE();
    case 39:
      if (lookahead == 'r') ADVANCE(65);
      END_STATE();
    case 40:
      if (lookahead == 'q') ADVANCE(66);
      if (lookahead == 's') ADVANCE(67);
      END_STATE();
    case 41:
      if (lookahead == 'a') ADVANCE(68);
      if (lookahead == 'y') ADVANCE(69);
      END_STATE();
    case 42:
      if (lookahead == 'b') ADVANCE(70);
      END_STATE();
    case 43:
      if (lookahead == 'o') ADVANCE(71);
      END_STATE();
    case 44:
      if (lookahead == 'i') ADVANCE(72);
      END_STATE();
    case 45:
      if (lookahead == 'e') ADVANCE(73);
      END_STATE();
    case 46:
      if (lookahead == 'r') ADVANCE(74);
      END_STATE();
    case 47:
      if (lookahead == 'n') ADVANCE(75);
      END_STATE();
    case 48:
      ACCEPT_TOKEN(anon_sym_and);
      END_STATE();
    case 49:
      if (lookahead == 'i') ADVANCE(76);
      END_STATE();
    case 50:
      if (lookahead == 'd') ADVANCE(77);
      if (lookahead == 't') ADVANCE(78);
      END_STATE();
    case 51:
      if (lookahead == 'a') ADVANCE(79);
      END_STATE();
    case 52:
      ACCEPT_TOKEN(anon_sym_dip);
      END_STATE();
    case 53:
      if (lookahead == 'e') ADVANCE(80);
      END_STATE();
    case 54:
      if (lookahead == 'e') ADVANCE(81);
      END_STATE();
    case 55:
      if (lookahead == 's') ADVANCE(82);
      END_STATE();
    case 56:
      if (lookahead == 't') ADVANCE(83);
      END_STATE();
    case 57:
      if (lookahead == '_') ADVANCE(84);
      END_STATE();
    case 58:
      if (lookahead == 'l') ADVANCE(85);
      END_STATE();
    case 59:
      if (lookahead == 'a') ADVANCE(86);
      END_STATE();
    case 60:
      if (lookahead == 'e') ADVANCE(87);
      END_STATE();
    case 61:
      if (lookahead == 'p') ADVANCE(88);
      END_STATE();
    case 62:
      if (lookahead == 'a') ADVANCE(89);
      END_STATE();
    case 63:
      ACCEPT_TOKEN(anon_sym_not);
      END_STATE();
    case 64:
      if (lookahead == 'r') ADVANCE(90);
      END_STATE();
    case 65:
      if (lookahead == 'a') ADVANCE(91);
      END_STATE();
    case 66:
      if (lookahead == 'u') ADVANCE(92);
      END_STATE();
    case 67:
      if (lookahead == 't') ADVANCE(93);
      END_STATE();
    case 68:
      if (lookahead == 'r') ADVANCE(94);
      END_STATE();
    case 69:
      if (lookahead == 'l') ADVANCE(95);
      END_STATE();
    case 70:
      if (lookahead == 'g') ADVANCE(96);
      END_STATE();
    case 71:
      if (lookahead == 'l') ADVANCE(97);
      END_STATE();
    case 72:
      if (lookahead == 'g') ADVANCE(98);
      END_STATE();
    case 73:
      if (lookahead == 'n') ADVANCE(99);
      END_STATE();
    case 74:
      if (lookahead == 'k') ADVANCE(100);
      END_STATE();
    case 75:
      if (lookahead == 't') ADVANCE(101);
      END_STATE();
    case 76:
      if (lookahead == 'c') ADVANCE(102);
      END_STATE();
    case 77:
      if (lookahead == 'i') ADVANCE(103);
      END_STATE();
    case 78:
      if (lookahead == 'a') ADVANCE(104);
      END_STATE();
    case 79:
      if (lookahead == 'u') ADVANCE(105);
      END_STATE();
    case 80:
      if (lookahead == 's') ADVANCE(106);
      END_STATE();
    case 81:
      ACCEPT_TOKEN(anon_sym_else);
      END_STATE();
    case 82:
      if (lookahead == 'w') ADVANCE(107);
      END_STATE();
    case 83:
      ACCEPT_TOKEN(anon_sym_exit);
      END_STATE();
    case 84:
      if (lookahead == 'i') ADVANCE(108);
      END_STATE();
    case 85:
      ACCEPT_TOKEN(anon_sym_goal);
      END_STATE();
    case 86:
      if (lookahead == 'n') ADVANCE(109);
      END_STATE();
    case 87:
      if (lookahead == 'l') ADVANCE(110);
      END_STATE();
    case 88:
      ACCEPT_TOKEN(anon_sym_loop);
      END_STATE();
    case 89:
      if (lookahead == 'g') ADVANCE(111);
      END_STATE();
    case 90:
      if (lookahead == 'r') ADVANCE(112);
      END_STATE();
    case 91:
      if (lookahead == 'l') ADVANCE(113);
      END_STATE();
    case 92:
      if (lookahead == 'i') ADVANCE(114);
      END_STATE();
    case 93:
      if (lookahead == 'a') ADVANCE(115);
      END_STATE();
    case 94:
      if (lookahead == 't') ADVANCE(116);
      END_STATE();
    case 95:
      if (lookahead == 'e') ADVANCE(117);
      END_STATE();
    case 96:
      if (lookahead == 'r') ADVANCE(118);
      END_STATE();
    case 97:
      ACCEPT_TOKEN(anon_sym_tool);
      END_STATE();
    case 98:
      if (lookahead == 'h') ADVANCE(119);
      END_STATE();
    case 99:
      ACCEPT_TOKEN(anon_sym_when);
      END_STATE();
    case 100:
      if (lookahead == 'f') ADVANCE(120);
      END_STATE();
    case 101:
      ACCEPT_TOKEN(anon_sym_agent);
      END_STATE();
    case 102:
      if (lookahead == 'e') ADVANCE(121);
      END_STATE();
    case 103:
      if (lookahead == 't') ADVANCE(122);
      END_STATE();
    case 104:
      if (lookahead == 'i') ADVANCE(123);
      END_STATE();
    case 105:
      if (lookahead == 'l') ADVANCE(124);
      END_STATE();
    case 106:
      ACCEPT_TOKEN(anon_sym_edges);
      END_STATE();
    case 107:
      if (lookahead == 'i') ADVANCE(125);
      END_STATE();
    case 108:
      if (lookahead == 'n') ADVANCE(126);
      END_STATE();
    case 109:
      ACCEPT_TOKEN(anon_sym_human);
      END_STATE();
    case 110:
      ACCEPT_TOKEN(anon_sym_label);
      END_STATE();
    case 111:
      if (lookahead == 'e') ADVANCE(127);
      END_STATE();
    case 112:
      if (lookahead == 'i') ADVANCE(128);
      END_STATE();
    case 113:
      if (lookahead == 'l') ADVANCE(129);
      END_STATE();
    case 114:
      if (lookahead == 'r') ADVANCE(130);
      END_STATE();
    case 115:
      if (lookahead == 'r') ADVANCE(131);
      END_STATE();
    case 116:
      ACCEPT_TOKEN(anon_sym_start);
      if (lookahead == 's') ADVANCE(132);
      END_STATE();
    case 117:
      if (lookahead == 's') ADVANCE(133);
      END_STATE();
    case 118:
      if (lookahead == 'a') ADVANCE(134);
      END_STATE();
    case 119:
      if (lookahead == 't') ADVANCE(135);
      END_STATE();
    case 120:
      if (lookahead == 'l') ADVANCE(136);
      END_STATE();
    case 121:
      ACCEPT_TOKEN(anon_sym_choice);
      END_STATE();
    case 122:
      if (lookahead == 'i') ADVANCE(137);
      END_STATE();
    case 123:
      if (lookahead == 'n') ADVANCE(138);
      END_STATE();
    case 124:
      if (lookahead == 't') ADVANCE(139);
      END_STATE();
    case 125:
      if (lookahead == 't') ADVANCE(140);
      END_STATE();
    case 126:
      ACCEPT_TOKEN(anon_sym_fan_in);
      END_STATE();
    case 127:
      if (lookahead == 'r') ADVANCE(141);
      END_STATE();
    case 128:
      if (lookahead == 'd') ADVANCE(142);
      END_STATE();
    case 129:
      if (lookahead == 'e') ADVANCE(143);
      END_STATE();
    case 130:
      if (lookahead == 'e') ADVANCE(144);
      END_STATE();
    case 131:
      if (lookahead == 't') ADVANCE(145);
      END_STATE();
    case 132:
      if (lookahead == 'w') ADVANCE(146);
      END_STATE();
    case 133:
      if (lookahead == 'h') ADVANCE(147);
      END_STATE();
    case 134:
      if (lookahead == 'p') ADVANCE(148);
      END_STATE();
    case 135:
      ACCEPT_TOKEN(anon_sym_weight);
      END_STATE();
    case 136:
      if (lookahead == 'o') ADVANCE(149);
      END_STATE();
    case 137:
      if (lookahead == 'o') ADVANCE(150);
      END_STATE();
    case 138:
      if (lookahead == 's') ADVANCE(151);
      END_STATE();
    case 139:
      if (lookahead == 's') ADVANCE(152);
      END_STATE();
    case 140:
      if (lookahead == 'h') ADVANCE(153);
      END_STATE();
    case 141:
      if (lookahead == '_') ADVANCE(154);
      END_STATE();
    case 142:
      if (lookahead == 'e') ADVANCE(155);
      END_STATE();
    case 143:
      if (lookahead == 'l') ADVANCE(156);
      END_STATE();
    case 144:
      if (lookahead == 's') ADVANCE(157);
      END_STATE();
    case 145:
      ACCEPT_TOKEN(anon_sym_restart);
      END_STATE();
    case 146:
      if (lookahead == 'i') ADVANCE(158);
      END_STATE();
    case 147:
      if (lookahead == 'e') ADVANCE(159);
      END_STATE();
    case 148:
      if (lookahead == 'h') ADVANCE(160);
      END_STATE();
    case 149:
      if (lookahead == 'w') ADVANCE(161);
      END_STATE();
    case 150:
      if (lookahead == 'n') ADVANCE(162);
      END_STATE();
    case 151:
      ACCEPT_TOKEN(anon_sym_contains);
      END_STATE();
    case 152:
      ACCEPT_TOKEN(anon_sym_defaults);
      END_STATE();
    case 153:
      ACCEPT_TOKEN(anon_sym_endswith);
      END_STATE();
    case 154:
      if (lookahead == 'l') ADVANCE(163);
      END_STATE();
    case 155:
      ACCEPT_TOKEN(anon_sym_override);
      END_STATE();
    case 156:
      ACCEPT_TOKEN(anon_sym_parallel);
      END_STATE();
    case 157:
      ACCEPT_TOKEN(anon_sym_requires);
      END_STATE();
    case 158:
      if (lookahead == 't') ADVANCE(164);
      END_STATE();
    case 159:
      if (lookahead == 'e') ADVANCE(165);
      END_STATE();
    case 160:
      ACCEPT_TOKEN(anon_sym_subgraph);
      END_STATE();
    case 161:
      ACCEPT_TOKEN(anon_sym_workflow);
      END_STATE();
    case 162:
      if (lookahead == 'a') ADVANCE(166);
      END_STATE();
    case 163:
      if (lookahead == 'o') ADVANCE(167);
      END_STATE();
    case 164:
      if (lookahead == 'h') ADVANCE(168);
      END_STATE();
    case 165:
      if (lookahead == 't') ADVANCE(169);
      END_STATE();
    case 166:
      if (lookahead == 'l') ADVANCE(170);
      END_STATE();
    case 167:
      if (lookahead == 'o') ADVANCE(171);
      END_STATE();
    case 168:
      ACCEPT_TOKEN(anon_sym_startswith);
      END_STATE();
    case 169:
      ACCEPT_TOKEN(anon_sym_stylesheet);
      END_STATE();
    case 170:
      ACCEPT_TOKEN(anon_sym_conditional);
      END_STATE();
    case 171:
      if (lookahead == 'p') ADVANCE(172);
      END_STATE();
    case 172:
      ACCEPT_TOKEN(anon_sym_manager_loop);
      END_STATE();
    default:
      return false;
  }
}

static const TSLexMode ts_lex_modes[STATE_COUNT] = {
  [0] = {.lex_state = 0, .external_lex_state = 1},
  [1] = {.lex_state = 9, .external_lex_state = 2},
  [2] = {.lex_state = 9, .external_lex_state = 3},
  [3] = {.lex_state = 9, .external_lex_state = 3},
  [4] = {.lex_state = 9, .external_lex_state = 3},
  [5] = {.lex_state = 9, .external_lex_state = 2},
  [6] = {.lex_state = 9, .external_lex_state = 3},
  [7] = {.lex_state = 9, .external_lex_state = 3},
  [8] = {.lex_state = 9, .external_lex_state = 3},
  [9] = {.lex_state = 9, .external_lex_state = 3},
  [10] = {.lex_state = 9, .external_lex_state = 3},
  [11] = {.lex_state = 9, .external_lex_state = 3},
  [12] = {.lex_state = 9, .external_lex_state = 3},
  [13] = {.lex_state = 9, .external_lex_state = 3},
  [14] = {.lex_state = 9, .external_lex_state = 3},
  [15] = {.lex_state = 9, .external_lex_state = 3},
  [16] = {.lex_state = 9, .external_lex_state = 3},
  [17] = {.lex_state = 9, .external_lex_state = 3},
  [18] = {.lex_state = 9, .external_lex_state = 3},
  [19] = {.lex_state = 9, .external_lex_state = 3},
  [20] = {.lex_state = 9, .external_lex_state = 3},
  [21] = {.lex_state = 9, .external_lex_state = 3},
  [22] = {.lex_state = 9, .external_lex_state = 3},
  [23] = {.lex_state = 9, .external_lex_state = 3},
  [24] = {.lex_state = 9, .external_lex_state = 3},
  [25] = {.lex_state = 9, .external_lex_state = 3},
  [26] = {.lex_state = 9, .external_lex_state = 3},
  [27] = {.lex_state = 9, .external_lex_state = 3},
  [28] = {.lex_state = 9, .external_lex_state = 3},
  [29] = {.lex_state = 9, .external_lex_state = 3},
  [30] = {.lex_state = 9, .external_lex_state = 3},
  [31] = {.lex_state = 9, .external_lex_state = 3},
  [32] = {.lex_state = 9, .external_lex_state = 3},
  [33] = {.lex_state = 9, .external_lex_state = 3},
  [34] = {.lex_state = 9, .external_lex_state = 3},
  [35] = {.lex_state = 9, .external_lex_state = 3},
  [36] = {.lex_state = 9, .external_lex_state = 3},
  [37] = {.lex_state = 9, .external_lex_state = 3},
  [38] = {.lex_state = 9, .external_lex_state = 3},
  [39] = {.lex_state = 9, .external_lex_state = 3},
  [40] = {.lex_state = 9, .external_lex_state = 3},
  [41] = {.lex_state = 9, .external_lex_state = 3},
  [42] = {.lex_state = 9, .external_lex_state = 3},
  [43] = {.lex_state = 9, .external_lex_state = 3},
  [44] = {.lex_state = 9},
  [45] = {.lex_state = 0, .external_lex_state = 3},
  [46] = {.lex_state = 0, .external_lex_state = 3},
  [47] = {.lex_state = 9},
  [48] = {.lex_state = 0, .external_lex_state = 2},
  [49] = {.lex_state = 9},
  [50] = {.lex_state = 1, .external_lex_state = 4},
  [51] = {.lex_state = 9, .external_lex_state = 3},
  [52] = {.lex_state = 9, .external_lex_state = 3},
  [53] = {.lex_state = 1, .external_lex_state = 4},
  [54] = {.lex_state = 1, .external_lex_state = 4},
  [55] = {.lex_state = 1, .external_lex_state = 4},
  [56] = {.lex_state = 1, .external_lex_state = 4},
  [57] = {.lex_state = 9},
  [58] = {.lex_state = 9, .external_lex_state = 2},
  [59] = {.lex_state = 9, .external_lex_state = 3},
  [60] = {.lex_state = 9, .external_lex_state = 3},
  [61] = {.lex_state = 9, .external_lex_state = 3},
  [62] = {.lex_state = 9, .external_lex_state = 3},
  [63] = {.lex_state = 9, .external_lex_state = 3},
  [64] = {.lex_state = 9, .external_lex_state = 3},
  [65] = {.lex_state = 9, .external_lex_state = 3},
  [66] = {.lex_state = 9, .external_lex_state = 3},
  [67] = {.lex_state = 9, .external_lex_state = 3},
  [68] = {.lex_state = 0, .external_lex_state = 3},
  [69] = {.lex_state = 9, .external_lex_state = 3},
  [70] = {.lex_state = 9},
  [71] = {.lex_state = 9, .external_lex_state = 2},
  [72] = {.lex_state = 9},
  [73] = {.lex_state = 9, .external_lex_state = 2},
  [74] = {.lex_state = 9, .external_lex_state = 2},
  [75] = {.lex_state = 9, .external_lex_state = 2},
  [76] = {.lex_state = 9, .external_lex_state = 2},
  [77] = {.lex_state = 9, .external_lex_state = 2},
  [78] = {.lex_state = 9, .external_lex_state = 3},
  [79] = {.lex_state = 9, .external_lex_state = 2},
  [80] = {.lex_state = 9, .external_lex_state = 3},
  [81] = {.lex_state = 9, .external_lex_state = 2},
  [82] = {.lex_state = 9, .external_lex_state = 2},
  [83] = {.lex_state = 9, .external_lex_state = 3},
  [84] = {.lex_state = 3, .external_lex_state = 3},
  [85] = {.lex_state = 2},
  [86] = {.lex_state = 4},
  [87] = {.lex_state = 9, .external_lex_state = 2},
  [88] = {.lex_state = 9, .external_lex_state = 5},
  [89] = {.lex_state = 9, .external_lex_state = 5},
  [90] = {.lex_state = 9, .external_lex_state = 5},
  [91] = {.lex_state = 2},
  [92] = {.lex_state = 3, .external_lex_state = 3},
  [93] = {.lex_state = 2},
  [94] = {.lex_state = 4},
  [95] = {.lex_state = 4},
  [96] = {.lex_state = 3, .external_lex_state = 2},
  [97] = {.lex_state = 9, .external_lex_state = 2},
  [98] = {.lex_state = 9, .external_lex_state = 3},
  [99] = {.lex_state = 9, .external_lex_state = 3},
  [100] = {.lex_state = 9, .external_lex_state = 5},
  [101] = {.lex_state = 9},
  [102] = {.lex_state = 9, .external_lex_state = 3},
  [103] = {.lex_state = 9, .external_lex_state = 5},
  [104] = {.lex_state = 9, .external_lex_state = 5},
  [105] = {.lex_state = 9},
  [106] = {.lex_state = 9},
  [107] = {.lex_state = 9},
  [108] = {.lex_state = 9},
  [109] = {.lex_state = 9},
  [110] = {.lex_state = 9},
  [111] = {.lex_state = 9, .external_lex_state = 4},
  [112] = {.lex_state = 9},
  [113] = {.lex_state = 9},
  [114] = {.lex_state = 9},
  [115] = {.lex_state = 9},
  [116] = {.lex_state = 9},
  [117] = {.lex_state = 9},
  [118] = {.lex_state = 9},
  [119] = {.lex_state = 9},
  [120] = {.lex_state = 9},
  [121] = {.lex_state = 9, .external_lex_state = 4},
  [122] = {.lex_state = 9},
  [123] = {.lex_state = 9},
  [124] = {.lex_state = 9},
  [125] = {.lex_state = 9, .external_lex_state = 4},
  [126] = {.lex_state = 9, .external_lex_state = 6},
  [127] = {.lex_state = 9},
  [128] = {.lex_state = 9},
  [129] = {.lex_state = 9, .external_lex_state = 6},
  [130] = {.lex_state = 9},
  [131] = {.lex_state = 9, .external_lex_state = 4},
  [132] = {.lex_state = 9, .external_lex_state = 4},
  [133] = {.lex_state = 9},
  [134] = {.lex_state = 9},
  [135] = {.lex_state = 9, .external_lex_state = 4},
  [136] = {.lex_state = 9, .external_lex_state = 2},
  [137] = {.lex_state = 9},
  [138] = {.lex_state = 9},
  [139] = {.lex_state = 9},
  [140] = {.lex_state = 9, .external_lex_state = 4},
  [141] = {.lex_state = 9, .external_lex_state = 4},
  [142] = {.lex_state = 9},
  [143] = {.lex_state = 9},
  [144] = {.lex_state = 9, .external_lex_state = 4},
  [145] = {.lex_state = 9},
  [146] = {.lex_state = 9, .external_lex_state = 4},
  [147] = {.lex_state = 9},
  [148] = {.lex_state = 9},
  [149] = {.lex_state = 9, .external_lex_state = 4},
  [150] = {.lex_state = 9, .external_lex_state = 4},
  [151] = {.lex_state = 9},
  [152] = {.lex_state = 9},
  [153] = {.lex_state = 9},
  [154] = {.lex_state = 9},
  [155] = {.lex_state = 9},
  [156] = {.lex_state = 9},
  [157] = {.lex_state = 9, .external_lex_state = 4},
};

static const uint16_t ts_parse_table[LARGE_STATE_COUNT][SYMBOL_COUNT] = {
  [0] = {
    [ts_builtin_sym_end] = ACTIONS(1),
    [sym_identifier] = ACTIONS(1),
    [anon_sym_dip] = ACTIONS(1),
    [anon_sym_workflow] = ACTIONS(1),
    [anon_sym_goal] = ACTIONS(1),
    [anon_sym_start] = ACTIONS(1),
    [anon_sym_exit] = ACTIONS(1),
    [anon_sym_requires] = ACTIONS(1),
    [anon_sym_COLON] = ACTIONS(1),
    [anon_sym_defaults] = ACTIONS(1),
    [anon_sym_agent] = ACTIONS(1),
    [anon_sym_human] = ACTIONS(1),
    [anon_sym_tool] = ACTIONS(1),
    [anon_sym_subgraph] = ACTIONS(1),
    [anon_sym_conditional] = ACTIONS(1),
    [anon_sym_manager_loop] = ACTIONS(1),
    [anon_sym_parallel] = ACTIONS(1),
    [anon_sym_DASH_GT] = ACTIONS(1),
    [anon_sym_fan_in] = ACTIONS(1),
    [anon_sym_LT_DASH] = ACTIONS(1),
    [anon_sym_edges] = ACTIONS(1),
    [anon_sym_else] = ACTIONS(1),
    [anon_sym_when] = ACTIONS(1),
    [anon_sym_on] = ACTIONS(1),
    [anon_sym_loop] = ACTIONS(1),
    [anon_sym_label] = ACTIONS(1),
    [anon_sym_choice] = ACTIONS(1),
    [anon_sym_weight] = ACTIONS(1),
    [anon_sym_restart] = ACTIONS(1),
    [anon_sym_override] = ACTIONS(1),
    [anon_sym_or] = ACTIONS(1),
    [anon_sym_and] = ACTIONS(1),
    [anon_sym_not] = ACTIONS(1),
    [anon_sym_EQ_EQ] = ACTIONS(1),
    [anon_sym_BANG_EQ] = ACTIONS(1),
    [anon_sym_EQ] = ACTIONS(1),
    [anon_sym_contains] = ACTIONS(1),
    [anon_sym_startswith] = ACTIONS(1),
    [anon_sym_endswith] = ACTIONS(1),
    [anon_sym_in] = ACTIONS(1),
    [anon_sym_DOT] = ACTIONS(1),
    [anon_sym_stylesheet] = ACTIONS(1),
    [anon_sym_STAR] = ACTIONS(1),
    [anon_sym_POUND] = ACTIONS(1),
    [anon_sym_COMMA] = ACTIONS(1),
    [anon_sym_DQUOTE] = ACTIONS(1),
    [aux_sym_string_token2] = ACTIONS(1),
    [anon_sym_SQUOTE] = ACTIONS(1),
    [anon_sym_SQUOTE_SQUOTE] = ACTIONS(1),
    [sym_comment] = ACTIONS(3),
    [sym__indent] = ACTIONS(1),
    [sym__dedent] = ACTIONS(1),
    [sym__newline] = ACTIONS(1),
  },
  [1] = {
    [sym_source_file] = STATE(128),
    [sym_version_decl] = STATE(107),
    [sym_workflow_decl] = STATE(155),
    [aux_sym_source_file_repeat1] = STATE(58),
    [anon_sym_dip] = ACTIONS(5),
    [anon_sym_workflow] = ACTIONS(7),
    [sym_comment] = ACTIONS(9),
    [sym__newline] = ACTIONS(11),
  },
};

static const uint16_t ts_small_parse_table[] = {
  [0] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(15), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(13), 33,
      anon_sym_goal,
      anon_sym_start,
      anon_sym_exit,
      anon_sym_requires,
      anon_sym_defaults,
      anon_sym_agent,
      anon_sym_human,
      anon_sym_tool,
      anon_sym_subgraph,
      anon_sym_conditional,
      anon_sym_manager_loop,
      anon_sym_parallel,
      anon_sym_fan_in,
      anon_sym_edges,
      anon_sym_else,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      anon_sym_and,
      anon_sym_not,
      anon_sym_EQ,
      anon_sym_contains,
      anon_sym_startswith,
      anon_sym_endswith,
      anon_sym_in,
      anon_sym_stylesheet,
      sym_identifier,
  [45] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(19), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(17), 33,
      anon_sym_goal,
      anon_sym_start,
      anon_sym_exit,
      anon_sym_requires,
      anon_sym_defaults,
      anon_sym_agent,
      anon_sym_human,
      anon_sym_tool,
      anon_sym_subgraph,
      anon_sym_conditional,
      anon_sym_manager_loop,
      anon_sym_parallel,
      anon_sym_fan_in,
      anon_sym_edges,
      anon_sym_else,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      anon_sym_and,
      anon_sym_not,
      anon_sym_EQ,
      anon_sym_contains,
      anon_sym_startswith,
      anon_sym_endswith,
      anon_sym_in,
      anon_sym_stylesheet,
      sym_identifier,
  [90] = 17,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(24), 1,
      anon_sym_defaults,
    ACTIONS(27), 1,
      anon_sym_agent,
    ACTIONS(30), 1,
      anon_sym_human,
    ACTIONS(33), 1,
      anon_sym_tool,
    ACTIONS(36), 1,
      anon_sym_subgraph,
    ACTIONS(39), 1,
      anon_sym_conditional,
    ACTIONS(42), 1,
      anon_sym_manager_loop,
    ACTIONS(45), 1,
      anon_sym_parallel,
    ACTIONS(48), 1,
      anon_sym_fan_in,
    ACTIONS(51), 1,
      anon_sym_edges,
    ACTIONS(54), 1,
      anon_sym_stylesheet,
    ACTIONS(57), 1,
      sym__dedent,
    ACTIONS(59), 1,
      sym__newline,
    ACTIONS(21), 4,
      anon_sym_goal,
      anon_sym_start,
      anon_sym_exit,
      anon_sym_requires,
    STATE(4), 6,
      sym_workflow_field,
      sym_defaults_section,
      sym_node_decl,
      sym_edges_section,
      sym_stylesheet_section,
      aux_sym_workflow_body_repeat1,
    STATE(24), 8,
      sym_agent_node,
      sym_human_node,
      sym_tool_node,
      sym_subgraph_node,
      sym_conditional_node,
      sym_manager_loop_node,
      sym_parallel_node,
      sym_fan_in_node,
  [157] = 17,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(64), 1,
      anon_sym_defaults,
    ACTIONS(66), 1,
      anon_sym_agent,
    ACTIONS(68), 1,
      anon_sym_human,
    ACTIONS(70), 1,
      anon_sym_tool,
    ACTIONS(72), 1,
      anon_sym_subgraph,
    ACTIONS(74), 1,
      anon_sym_conditional,
    ACTIONS(76), 1,
      anon_sym_manager_loop,
    ACTIONS(78), 1,
      anon_sym_parallel,
    ACTIONS(80), 1,
      anon_sym_fan_in,
    ACTIONS(82), 1,
      anon_sym_edges,
    ACTIONS(84), 1,
      anon_sym_stylesheet,
    ACTIONS(86), 1,
      sym__newline,
    STATE(126), 1,
      sym_workflow_body,
    ACTIONS(62), 4,
      anon_sym_goal,
      anon_sym_start,
      anon_sym_exit,
      anon_sym_requires,
    STATE(6), 6,
      sym_workflow_field,
      sym_defaults_section,
      sym_node_decl,
      sym_edges_section,
      sym_stylesheet_section,
      aux_sym_workflow_body_repeat1,
    STATE(24), 8,
      sym_agent_node,
      sym_human_node,
      sym_tool_node,
      sym_subgraph_node,
      sym_conditional_node,
      sym_manager_loop_node,
      sym_parallel_node,
      sym_fan_in_node,
  [224] = 17,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(64), 1,
      anon_sym_defaults,
    ACTIONS(66), 1,
      anon_sym_agent,
    ACTIONS(68), 1,
      anon_sym_human,
    ACTIONS(70), 1,
      anon_sym_tool,
    ACTIONS(72), 1,
      anon_sym_subgraph,
    ACTIONS(74), 1,
      anon_sym_conditional,
    ACTIONS(76), 1,
      anon_sym_manager_loop,
    ACTIONS(78), 1,
      anon_sym_parallel,
    ACTIONS(80), 1,
      anon_sym_fan_in,
    ACTIONS(82), 1,
      anon_sym_edges,
    ACTIONS(84), 1,
      anon_sym_stylesheet,
    ACTIONS(88), 1,
      sym__dedent,
    ACTIONS(90), 1,
      sym__newline,
    ACTIONS(62), 4,
      anon_sym_goal,
      anon_sym_start,
      anon_sym_exit,
      anon_sym_requires,
    STATE(4), 6,
      sym_workflow_field,
      sym_defaults_section,
      sym_node_decl,
      sym_edges_section,
      sym_stylesheet_section,
      aux_sym_workflow_body_repeat1,
    STATE(24), 8,
      sym_agent_node,
      sym_human_node,
      sym_tool_node,
      sym_subgraph_node,
      sym_conditional_node,
      sym_manager_loop_node,
      sym_parallel_node,
      sym_fan_in_node,
  [291] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(94), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(92), 25,
      anon_sym_goal,
      anon_sym_start,
      anon_sym_exit,
      anon_sym_requires,
      anon_sym_defaults,
      anon_sym_agent,
      anon_sym_human,
      anon_sym_tool,
      anon_sym_subgraph,
      anon_sym_conditional,
      anon_sym_manager_loop,
      anon_sym_parallel,
      anon_sym_fan_in,
      anon_sym_edges,
      anon_sym_else,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_stylesheet,
      sym_identifier,
  [326] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(98), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(96), 25,
      anon_sym_goal,
      anon_sym_start,
      anon_sym_exit,
      anon_sym_requires,
      anon_sym_defaults,
      anon_sym_agent,
      anon_sym_human,
      anon_sym_tool,
      anon_sym_subgraph,
      anon_sym_conditional,
      anon_sym_manager_loop,
      anon_sym_parallel,
      anon_sym_fan_in,
      anon_sym_edges,
      anon_sym_else,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_stylesheet,
      sym_identifier,
  [361] = 7,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(102), 1,
      anon_sym_not,
    STATE(70), 1,
      sym_compare_op,
    ACTIONS(104), 2,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(108), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(106), 5,
      anon_sym_EQ,
      anon_sym_contains,
      anon_sym_startswith,
      anon_sym_endswith,
      anon_sym_in,
    ACTIONS(100), 12,
      anon_sym_else,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      anon_sym_and,
      sym_identifier,
  [400] = 4,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(114), 1,
      anon_sym_DOT,
    ACTIONS(112), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(110), 18,
      anon_sym_else,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      anon_sym_and,
      anon_sym_not,
      anon_sym_EQ,
      anon_sym_contains,
      anon_sym_startswith,
      anon_sym_endswith,
      anon_sym_in,
      sym_identifier,
  [433] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(112), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(110), 18,
      anon_sym_else,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      anon_sym_and,
      anon_sym_not,
      anon_sym_EQ,
      anon_sym_contains,
      anon_sym_startswith,
      anon_sym_endswith,
      anon_sym_in,
      sym_identifier,
  [463] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(118), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(116), 18,
      anon_sym_else,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      anon_sym_and,
      anon_sym_not,
      anon_sym_EQ,
      anon_sym_contains,
      anon_sym_startswith,
      anon_sym_endswith,
      anon_sym_in,
      sym_identifier,
  [493] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(120), 17,
      sym__dedent,
      sym__newline,
      anon_sym_goal,
      anon_sym_start,
      anon_sym_exit,
      anon_sym_requires,
      anon_sym_defaults,
      anon_sym_agent,
      anon_sym_human,
      anon_sym_tool,
      anon_sym_subgraph,
      anon_sym_conditional,
      anon_sym_manager_loop,
      anon_sym_parallel,
      anon_sym_fan_in,
      anon_sym_edges,
      anon_sym_stylesheet,
  [516] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(122), 17,
      sym__dedent,
      sym__newline,
      anon_sym_goal,
      anon_sym_start,
      anon_sym_exit,
      anon_sym_requires,
      anon_sym_defaults,
      anon_sym_agent,
      anon_sym_human,
      anon_sym_tool,
      anon_sym_subgraph,
      anon_sym_conditional,
      anon_sym_manager_loop,
      anon_sym_parallel,
      anon_sym_fan_in,
      anon_sym_edges,
      anon_sym_stylesheet,
  [539] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(124), 17,
      sym__dedent,
      sym__newline,
      anon_sym_goal,
      anon_sym_start,
      anon_sym_exit,
      anon_sym_requires,
      anon_sym_defaults,
      anon_sym_agent,
      anon_sym_human,
      anon_sym_tool,
      anon_sym_subgraph,
      anon_sym_conditional,
      anon_sym_manager_loop,
      anon_sym_parallel,
      anon_sym_fan_in,
      anon_sym_edges,
      anon_sym_stylesheet,
  [562] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(126), 17,
      sym__dedent,
      sym__newline,
      anon_sym_goal,
      anon_sym_start,
      anon_sym_exit,
      anon_sym_requires,
      anon_sym_defaults,
      anon_sym_agent,
      anon_sym_human,
      anon_sym_tool,
      anon_sym_subgraph,
      anon_sym_conditional,
      anon_sym_manager_loop,
      anon_sym_parallel,
      anon_sym_fan_in,
      anon_sym_edges,
      anon_sym_stylesheet,
  [585] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(128), 17,
      sym__dedent,
      sym__newline,
      anon_sym_goal,
      anon_sym_start,
      anon_sym_exit,
      anon_sym_requires,
      anon_sym_defaults,
      anon_sym_agent,
      anon_sym_human,
      anon_sym_tool,
      anon_sym_subgraph,
      anon_sym_conditional,
      anon_sym_manager_loop,
      anon_sym_parallel,
      anon_sym_fan_in,
      anon_sym_edges,
      anon_sym_stylesheet,
  [608] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(130), 17,
      sym__dedent,
      sym__newline,
      anon_sym_goal,
      anon_sym_start,
      anon_sym_exit,
      anon_sym_requires,
      anon_sym_defaults,
      anon_sym_agent,
      anon_sym_human,
      anon_sym_tool,
      anon_sym_subgraph,
      anon_sym_conditional,
      anon_sym_manager_loop,
      anon_sym_parallel,
      anon_sym_fan_in,
      anon_sym_edges,
      anon_sym_stylesheet,
  [631] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(132), 17,
      sym__dedent,
      sym__newline,
      anon_sym_goal,
      anon_sym_start,
      anon_sym_exit,
      anon_sym_requires,
      anon_sym_defaults,
      anon_sym_agent,
      anon_sym_human,
      anon_sym_tool,
      anon_sym_subgraph,
      anon_sym_conditional,
      anon_sym_manager_loop,
      anon_sym_parallel,
      anon_sym_fan_in,
      anon_sym_edges,
      anon_sym_stylesheet,
  [654] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(134), 17,
      sym__dedent,
      sym__newline,
      anon_sym_goal,
      anon_sym_start,
      anon_sym_exit,
      anon_sym_requires,
      anon_sym_defaults,
      anon_sym_agent,
      anon_sym_human,
      anon_sym_tool,
      anon_sym_subgraph,
      anon_sym_conditional,
      anon_sym_manager_loop,
      anon_sym_parallel,
      anon_sym_fan_in,
      anon_sym_edges,
      anon_sym_stylesheet,
  [677] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(136), 17,
      sym__dedent,
      sym__newline,
      anon_sym_goal,
      anon_sym_start,
      anon_sym_exit,
      anon_sym_requires,
      anon_sym_defaults,
      anon_sym_agent,
      anon_sym_human,
      anon_sym_tool,
      anon_sym_subgraph,
      anon_sym_conditional,
      anon_sym_manager_loop,
      anon_sym_parallel,
      anon_sym_fan_in,
      anon_sym_edges,
      anon_sym_stylesheet,
  [700] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(138), 17,
      sym__dedent,
      sym__newline,
      anon_sym_goal,
      anon_sym_start,
      anon_sym_exit,
      anon_sym_requires,
      anon_sym_defaults,
      anon_sym_agent,
      anon_sym_human,
      anon_sym_tool,
      anon_sym_subgraph,
      anon_sym_conditional,
      anon_sym_manager_loop,
      anon_sym_parallel,
      anon_sym_fan_in,
      anon_sym_edges,
      anon_sym_stylesheet,
  [723] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(140), 17,
      sym__dedent,
      sym__newline,
      anon_sym_goal,
      anon_sym_start,
      anon_sym_exit,
      anon_sym_requires,
      anon_sym_defaults,
      anon_sym_agent,
      anon_sym_human,
      anon_sym_tool,
      anon_sym_subgraph,
      anon_sym_conditional,
      anon_sym_manager_loop,
      anon_sym_parallel,
      anon_sym_fan_in,
      anon_sym_edges,
      anon_sym_stylesheet,
  [746] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(142), 17,
      sym__dedent,
      sym__newline,
      anon_sym_goal,
      anon_sym_start,
      anon_sym_exit,
      anon_sym_requires,
      anon_sym_defaults,
      anon_sym_agent,
      anon_sym_human,
      anon_sym_tool,
      anon_sym_subgraph,
      anon_sym_conditional,
      anon_sym_manager_loop,
      anon_sym_parallel,
      anon_sym_fan_in,
      anon_sym_edges,
      anon_sym_stylesheet,
  [769] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(144), 17,
      sym__dedent,
      sym__newline,
      anon_sym_goal,
      anon_sym_start,
      anon_sym_exit,
      anon_sym_requires,
      anon_sym_defaults,
      anon_sym_agent,
      anon_sym_human,
      anon_sym_tool,
      anon_sym_subgraph,
      anon_sym_conditional,
      anon_sym_manager_loop,
      anon_sym_parallel,
      anon_sym_fan_in,
      anon_sym_edges,
      anon_sym_stylesheet,
  [792] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(146), 17,
      sym__dedent,
      sym__newline,
      anon_sym_goal,
      anon_sym_start,
      anon_sym_exit,
      anon_sym_requires,
      anon_sym_defaults,
      anon_sym_agent,
      anon_sym_human,
      anon_sym_tool,
      anon_sym_subgraph,
      anon_sym_conditional,
      anon_sym_manager_loop,
      anon_sym_parallel,
      anon_sym_fan_in,
      anon_sym_edges,
      anon_sym_stylesheet,
  [815] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(150), 1,
      anon_sym_and,
    STATE(28), 1,
      aux_sym_and_expr_repeat1,
    ACTIONS(152), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(148), 11,
      anon_sym_else,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      sym_identifier,
  [842] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(150), 1,
      anon_sym_and,
    STATE(29), 1,
      aux_sym_and_expr_repeat1,
    ACTIONS(156), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(154), 11,
      anon_sym_else,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      sym_identifier,
  [869] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(160), 1,
      anon_sym_and,
    STATE(29), 1,
      aux_sym_and_expr_repeat1,
    ACTIONS(163), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(158), 11,
      anon_sym_else,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      sym_identifier,
  [896] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(167), 1,
      anon_sym_or,
    STATE(30), 1,
      aux_sym_or_expr_repeat1,
    ACTIONS(170), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(165), 10,
      anon_sym_else,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [922] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(174), 1,
      anon_sym_or,
    STATE(32), 1,
      aux_sym_or_expr_repeat1,
    ACTIONS(176), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(172), 10,
      anon_sym_else,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [948] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(174), 1,
      anon_sym_or,
    STATE(30), 1,
      aux_sym_or_expr_repeat1,
    ACTIONS(180), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(178), 10,
      anon_sym_else,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [974] = 8,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(184), 1,
      anon_sym_when,
    ACTIONS(186), 1,
      anon_sym_on,
    ACTIONS(188), 1,
      anon_sym_loop,
    ACTIONS(182), 2,
      anon_sym_else,
      sym_identifier,
    ACTIONS(192), 2,
      sym__dedent,
      sym__newline,
    STATE(36), 2,
      sym_edge_attr,
      aux_sym_edge_entry_repeat1,
    ACTIONS(190), 5,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
  [1006] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(163), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(158), 12,
      anon_sym_else,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      anon_sym_and,
      sym_identifier,
  [1028] = 8,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(196), 1,
      anon_sym_when,
    ACTIONS(199), 1,
      anon_sym_on,
    ACTIONS(202), 1,
      anon_sym_loop,
    ACTIONS(194), 2,
      anon_sym_else,
      sym_identifier,
    ACTIONS(208), 2,
      sym__dedent,
      sym__newline,
    STATE(35), 2,
      sym_edge_attr,
      aux_sym_edge_entry_repeat1,
    ACTIONS(205), 5,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
  [1060] = 8,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(184), 1,
      anon_sym_when,
    ACTIONS(186), 1,
      anon_sym_on,
    ACTIONS(188), 1,
      anon_sym_loop,
    ACTIONS(210), 2,
      anon_sym_else,
      sym_identifier,
    ACTIONS(212), 2,
      sym__dedent,
      sym__newline,
    STATE(35), 2,
      sym_edge_attr,
      aux_sym_edge_entry_repeat1,
    ACTIONS(190), 5,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
  [1092] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(216), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(214), 12,
      anon_sym_else,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      anon_sym_and,
      sym_identifier,
  [1114] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(220), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(218), 12,
      anon_sym_else,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      anon_sym_and,
      sym_identifier,
  [1136] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(170), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(165), 11,
      anon_sym_else,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      sym_identifier,
  [1157] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(224), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(222), 10,
      anon_sym_else,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [1177] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(228), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(226), 10,
      anon_sym_else,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [1197] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(232), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(230), 10,
      anon_sym_else,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [1217] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(236), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(234), 10,
      anon_sym_else,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [1237] = 10,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(238), 1,
      sym_identifier,
    ACTIONS(240), 1,
      anon_sym_DQUOTE,
    ACTIONS(242), 1,
      anon_sym_SQUOTE,
    STATE(9), 1,
      sym_operand,
    STATE(27), 1,
      sym_compare_expr,
    STATE(31), 1,
      sym_and_expr,
    STATE(42), 1,
      sym_condition,
    STATE(43), 1,
      sym_or_expr,
    STATE(11), 2,
      sym_variable,
      sym_string,
  [1269] = 8,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(246), 1,
      anon_sym_DOT,
    ACTIONS(248), 1,
      anon_sym_POUND,
    ACTIONS(250), 1,
      sym__dedent,
    ACTIONS(252), 1,
      sym__newline,
    STATE(149), 1,
      sym_selector,
    ACTIONS(244), 2,
      anon_sym_STAR,
      sym_identifier,
    STATE(46), 2,
      sym_stylesheet_rule,
      aux_sym_stylesheet_section_repeat1,
  [1296] = 8,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(257), 1,
      anon_sym_DOT,
    ACTIONS(260), 1,
      anon_sym_POUND,
    ACTIONS(263), 1,
      sym__dedent,
    ACTIONS(265), 1,
      sym__newline,
    STATE(149), 1,
      sym_selector,
    ACTIONS(254), 2,
      anon_sym_STAR,
      sym_identifier,
    STATE(46), 2,
      sym_stylesheet_rule,
      aux_sym_stylesheet_section_repeat1,
  [1323] = 8,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(238), 1,
      sym_identifier,
    ACTIONS(240), 1,
      anon_sym_DQUOTE,
    ACTIONS(242), 1,
      anon_sym_SQUOTE,
    STATE(9), 1,
      sym_operand,
    STATE(27), 1,
      sym_compare_expr,
    STATE(39), 1,
      sym_and_expr,
    STATE(11), 2,
      sym_variable,
      sym_string,
  [1349] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(246), 1,
      anon_sym_DOT,
    ACTIONS(248), 1,
      anon_sym_POUND,
    ACTIONS(268), 1,
      sym__newline,
    STATE(149), 1,
      sym_selector,
    ACTIONS(244), 2,
      anon_sym_STAR,
      sym_identifier,
    STATE(45), 2,
      sym_stylesheet_rule,
      aux_sym_stylesheet_section_repeat1,
  [1373] = 4,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(106), 1,
      anon_sym_EQ,
    STATE(72), 1,
      sym_compare_op,
    ACTIONS(104), 6,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
      anon_sym_contains,
      anon_sym_startswith,
      anon_sym_endswith,
      anon_sym_in,
  [1391] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(270), 1,
      sym_raw_inline,
    ACTIONS(272), 1,
      anon_sym_DQUOTE,
    ACTIONS(274), 1,
      anon_sym_SQUOTE,
    ACTIONS(276), 1,
      sym__indent,
    STATE(98), 1,
      sym_field_value,
    STATE(7), 2,
      sym_multiline_block,
      sym_string,
  [1414] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(278), 1,
      sym_identifier,
    ACTIONS(280), 1,
      anon_sym_else,
    ACTIONS(282), 1,
      sym__dedent,
    ACTIONS(284), 1,
      sym__newline,
    STATE(52), 3,
      sym_edge_entry,
      sym_else_default,
      aux_sym_edges_section_repeat1,
  [1435] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(286), 1,
      sym_identifier,
    ACTIONS(289), 1,
      anon_sym_else,
    ACTIONS(292), 1,
      sym__dedent,
    ACTIONS(294), 1,
      sym__newline,
    STATE(52), 3,
      sym_edge_entry,
      sym_else_default,
      aux_sym_edges_section_repeat1,
  [1456] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(270), 1,
      sym_raw_inline,
    ACTIONS(272), 1,
      anon_sym_DQUOTE,
    ACTIONS(274), 1,
      anon_sym_SQUOTE,
    ACTIONS(276), 1,
      sym__indent,
    STATE(99), 1,
      sym_field_value,
    STATE(7), 2,
      sym_multiline_block,
      sym_string,
  [1479] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(270), 1,
      sym_raw_inline,
    ACTIONS(272), 1,
      anon_sym_DQUOTE,
    ACTIONS(274), 1,
      anon_sym_SQUOTE,
    ACTIONS(276), 1,
      sym__indent,
    STATE(102), 1,
      sym_field_value,
    STATE(7), 2,
      sym_multiline_block,
      sym_string,
  [1502] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(270), 1,
      sym_raw_inline,
    ACTIONS(272), 1,
      anon_sym_DQUOTE,
    ACTIONS(274), 1,
      anon_sym_SQUOTE,
    ACTIONS(276), 1,
      sym__indent,
    STATE(23), 1,
      sym_field_value,
    STATE(7), 2,
      sym_multiline_block,
      sym_string,
  [1525] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(270), 1,
      sym_raw_inline,
    ACTIONS(272), 1,
      anon_sym_DQUOTE,
    ACTIONS(274), 1,
      anon_sym_SQUOTE,
    ACTIONS(276), 1,
      sym__indent,
    STATE(40), 1,
      sym_field_value,
    STATE(7), 2,
      sym_multiline_block,
      sym_string,
  [1548] = 7,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(238), 1,
      sym_identifier,
    ACTIONS(240), 1,
      anon_sym_DQUOTE,
    ACTIONS(242), 1,
      anon_sym_SQUOTE,
    STATE(9), 1,
      sym_operand,
    STATE(34), 1,
      sym_compare_expr,
    STATE(11), 2,
      sym_variable,
      sym_string,
  [1571] = 7,
    ACTIONS(5), 1,
      anon_sym_dip,
    ACTIONS(7), 1,
      anon_sym_workflow,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(297), 1,
      sym__newline,
    STATE(97), 1,
      aux_sym_source_file_repeat1,
    STATE(108), 1,
      sym_version_decl,
    STATE(112), 1,
      sym_workflow_decl,
  [1593] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(299), 1,
      sym_identifier,
    ACTIONS(301), 1,
      sym__dedent,
    ACTIONS(303), 1,
      sym__newline,
    STATE(137), 1,
      sym_field_name,
    STATE(60), 2,
      sym_defaults_field,
      aux_sym_defaults_section_repeat1,
  [1613] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(305), 1,
      sym_identifier,
    ACTIONS(308), 1,
      sym__dedent,
    ACTIONS(310), 1,
      sym__newline,
    STATE(137), 1,
      sym_field_name,
    STATE(60), 2,
      sym_defaults_field,
      aux_sym_defaults_section_repeat1,
  [1633] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(299), 1,
      sym_identifier,
    ACTIONS(313), 1,
      sym__dedent,
    ACTIONS(315), 1,
      sym__newline,
    STATE(154), 1,
      sym_field_name,
    STATE(69), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1653] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(299), 1,
      sym_identifier,
    ACTIONS(315), 1,
      sym__newline,
    ACTIONS(317), 1,
      sym__dedent,
    STATE(154), 1,
      sym_field_name,
    STATE(69), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1673] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(299), 1,
      sym_identifier,
    ACTIONS(315), 1,
      sym__newline,
    ACTIONS(319), 1,
      sym__dedent,
    STATE(154), 1,
      sym_field_name,
    STATE(69), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1693] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(299), 1,
      sym_identifier,
    ACTIONS(315), 1,
      sym__newline,
    ACTIONS(321), 1,
      sym__dedent,
    STATE(154), 1,
      sym_field_name,
    STATE(69), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1713] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(299), 1,
      sym_identifier,
    ACTIONS(315), 1,
      sym__newline,
    ACTIONS(323), 1,
      sym__dedent,
    STATE(154), 1,
      sym_field_name,
    STATE(69), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1733] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(299), 1,
      sym_identifier,
    ACTIONS(315), 1,
      sym__newline,
    ACTIONS(325), 1,
      sym__dedent,
    STATE(154), 1,
      sym_field_name,
    STATE(69), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1753] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(299), 1,
      sym_identifier,
    ACTIONS(315), 1,
      sym__newline,
    ACTIONS(327), 1,
      sym__dedent,
    STATE(154), 1,
      sym_field_name,
    STATE(69), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1773] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(331), 1,
      anon_sym_POUND,
    ACTIONS(329), 5,
      sym__dedent,
      sym__newline,
      anon_sym_DOT,
      anon_sym_STAR,
      sym_identifier,
  [1787] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(333), 1,
      sym_identifier,
    ACTIONS(336), 1,
      sym__dedent,
    ACTIONS(338), 1,
      sym__newline,
    STATE(154), 1,
      sym_field_name,
    STATE(69), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1807] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(238), 1,
      sym_identifier,
    ACTIONS(240), 1,
      anon_sym_DQUOTE,
    ACTIONS(242), 1,
      anon_sym_SQUOTE,
    STATE(37), 1,
      sym_operand,
    STATE(11), 2,
      sym_variable,
      sym_string,
  [1827] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(278), 1,
      sym_identifier,
    ACTIONS(280), 1,
      anon_sym_else,
    ACTIONS(341), 1,
      sym__newline,
    STATE(51), 3,
      sym_edge_entry,
      sym_else_default,
      aux_sym_edges_section_repeat1,
  [1845] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(238), 1,
      sym_identifier,
    ACTIONS(240), 1,
      anon_sym_DQUOTE,
    ACTIONS(242), 1,
      anon_sym_SQUOTE,
    STATE(38), 1,
      sym_operand,
    STATE(11), 2,
      sym_variable,
      sym_string,
  [1865] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(299), 1,
      sym_identifier,
    ACTIONS(343), 1,
      sym__newline,
    STATE(154), 1,
      sym_field_name,
    STATE(63), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1882] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(299), 1,
      sym_identifier,
    ACTIONS(345), 1,
      sym__newline,
    STATE(154), 1,
      sym_field_name,
    STATE(65), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1899] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(299), 1,
      sym_identifier,
    ACTIONS(347), 1,
      sym__newline,
    STATE(154), 1,
      sym_field_name,
    STATE(66), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1916] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(299), 1,
      sym_identifier,
    ACTIONS(349), 1,
      sym__newline,
    STATE(154), 1,
      sym_field_name,
    STATE(67), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1933] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(299), 1,
      sym_identifier,
    ACTIONS(351), 1,
      sym__newline,
    STATE(154), 1,
      sym_field_name,
    STATE(64), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1950] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(299), 1,
      sym_identifier,
    ACTIONS(353), 1,
      sym__dedent,
    ACTIONS(355), 1,
      sym__newline,
    STATE(80), 1,
      aux_sym_stylesheet_rule_repeat1,
    STATE(124), 1,
      sym_field_name,
  [1969] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(299), 1,
      sym_identifier,
    ACTIONS(357), 1,
      sym__newline,
    STATE(137), 1,
      sym_field_name,
    STATE(59), 2,
      sym_defaults_field,
      aux_sym_defaults_section_repeat1,
  [1986] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(359), 1,
      sym_identifier,
    ACTIONS(362), 1,
      sym__dedent,
    ACTIONS(364), 1,
      sym__newline,
    STATE(80), 1,
      aux_sym_stylesheet_rule_repeat1,
    STATE(124), 1,
      sym_field_name,
  [2005] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(299), 1,
      sym_identifier,
    ACTIONS(367), 1,
      sym__newline,
    STATE(154), 1,
      sym_field_name,
    STATE(61), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [2022] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(299), 1,
      sym_identifier,
    ACTIONS(369), 1,
      sym__newline,
    STATE(154), 1,
      sym_field_name,
    STATE(62), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [2039] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(371), 2,
      anon_sym_else,
      sym_identifier,
    ACTIONS(373), 2,
      sym__dedent,
      sym__newline,
  [2051] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(378), 1,
      sym__dedent,
    STATE(84), 1,
      aux_sym_block_content_repeat1,
    ACTIONS(375), 2,
      sym__newline,
      sym_block_line,
  [2065] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(380), 1,
      anon_sym_DQUOTE,
    STATE(85), 1,
      aux_sym_string_repeat1,
    ACTIONS(382), 2,
      aux_sym_string_token1,
      aux_sym_string_token2,
  [2079] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(385), 1,
      anon_sym_SQUOTE,
    STATE(86), 1,
      aux_sym_string_repeat2,
    ACTIONS(387), 2,
      aux_sym_string_token3,
      anon_sym_SQUOTE_SQUOTE,
  [2093] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(299), 1,
      sym_identifier,
    ACTIONS(390), 1,
      sym__newline,
    STATE(78), 1,
      aux_sym_stylesheet_rule_repeat1,
    STATE(124), 1,
      sym_field_name,
  [2109] = 4,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(392), 1,
      anon_sym_COMMA,
    STATE(89), 1,
      aux_sym_identifier_list_repeat1,
    ACTIONS(394), 2,
      sym__indent,
      sym__newline,
  [2123] = 4,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(396), 1,
      anon_sym_COMMA,
    STATE(89), 1,
      aux_sym_identifier_list_repeat1,
    ACTIONS(399), 2,
      sym__indent,
      sym__newline,
  [2137] = 4,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(392), 1,
      anon_sym_COMMA,
    STATE(88), 1,
      aux_sym_identifier_list_repeat1,
    ACTIONS(401), 2,
      sym__indent,
      sym__newline,
  [2151] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(403), 1,
      anon_sym_DQUOTE,
    STATE(93), 1,
      aux_sym_string_repeat1,
    ACTIONS(405), 2,
      aux_sym_string_token1,
      aux_sym_string_token2,
  [2165] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(409), 1,
      sym__dedent,
    STATE(84), 1,
      aux_sym_block_content_repeat1,
    ACTIONS(407), 2,
      sym__newline,
      sym_block_line,
  [2179] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(411), 1,
      anon_sym_DQUOTE,
    STATE(85), 1,
      aux_sym_string_repeat1,
    ACTIONS(413), 2,
      aux_sym_string_token1,
      aux_sym_string_token2,
  [2193] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(403), 1,
      anon_sym_SQUOTE,
    STATE(95), 1,
      aux_sym_string_repeat2,
    ACTIONS(415), 2,
      aux_sym_string_token3,
      anon_sym_SQUOTE_SQUOTE,
  [2207] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(411), 1,
      anon_sym_SQUOTE,
    STATE(86), 1,
      aux_sym_string_repeat2,
    ACTIONS(417), 2,
      aux_sym_string_token3,
      anon_sym_SQUOTE_SQUOTE,
  [2221] = 4,
    ACTIONS(3), 1,
      sym_comment,
    STATE(92), 1,
      aux_sym_block_content_repeat1,
    STATE(129), 1,
      sym_block_content,
    ACTIONS(419), 2,
      sym__newline,
      sym_block_line,
  [2235] = 4,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(423), 1,
      sym__newline,
    STATE(97), 1,
      aux_sym_source_file_repeat1,
    ACTIONS(421), 2,
      anon_sym_dip,
      anon_sym_workflow,
  [2249] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(426), 3,
      sym__dedent,
      sym__newline,
      sym_identifier,
  [2258] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(428), 3,
      sym__dedent,
      sym__newline,
      sym_identifier,
  [2267] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(399), 3,
      sym__indent,
      sym__newline,
      anon_sym_COMMA,
  [2276] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(430), 3,
      sym_identifier,
      anon_sym_DQUOTE,
      anon_sym_SQUOTE,
  [2285] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(432), 3,
      sym__dedent,
      sym__newline,
      sym_identifier,
  [2294] = 4,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(434), 1,
      sym__indent,
    ACTIONS(436), 1,
      sym__newline,
    STATE(13), 1,
      sym_node_attr_block,
  [2307] = 4,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(434), 1,
      sym__indent,
    ACTIONS(438), 1,
      sym__newline,
    STATE(18), 1,
      sym_node_attr_block,
  [2320] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(440), 1,
      sym_identifier,
    STATE(104), 1,
      sym_identifier_list,
  [2330] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(440), 1,
      sym_identifier,
    STATE(103), 1,
      sym_identifier_list,
  [2340] = 3,
    ACTIONS(7), 1,
      anon_sym_workflow,
    ACTIONS(9), 1,
      sym_comment,
    STATE(112), 1,
      sym_workflow_decl,
  [2350] = 3,
    ACTIONS(7), 1,
      anon_sym_workflow,
    ACTIONS(9), 1,
      sym_comment,
    STATE(138), 1,
      sym_workflow_decl,
  [2360] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(442), 1,
      anon_sym_LT_DASH,
  [2367] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(444), 1,
      sym_identifier,
  [2374] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(446), 1,
      sym__indent,
  [2381] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(448), 1,
      ts_builtin_sym_end,
  [2388] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(450), 1,
      sym_identifier,
  [2395] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(452), 1,
      sym_identifier,
  [2402] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(454), 1,
      anon_sym_workflow,
  [2409] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(456), 1,
      anon_sym_DASH_GT,
  [2416] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(458), 1,
      sym_identifier,
  [2423] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(460), 1,
      anon_sym_COLON,
  [2430] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(462), 1,
      sym_identifier,
  [2437] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(464), 1,
      sym_identifier,
  [2444] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(466), 1,
      sym__indent,
  [2451] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(468), 1,
      anon_sym_COLON,
  [2458] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(470), 1,
      anon_sym_COLON,
  [2465] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(472), 1,
      anon_sym_COLON,
  [2472] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(474), 1,
      sym__indent,
  [2479] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(476), 1,
      sym__dedent,
  [2486] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(478), 1,
      sym_identifier,
  [2493] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(480), 1,
      ts_builtin_sym_end,
  [2500] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(482), 1,
      sym__dedent,
  [2507] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(484), 1,
      anon_sym_DASH_GT,
  [2514] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(486), 1,
      sym__indent,
  [2521] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(488), 1,
      sym__indent,
  [2528] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(490), 1,
      sym_identifier,
  [2535] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(492), 1,
      sym_identifier,
  [2542] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(494), 1,
      sym__indent,
  [2549] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(496), 1,
      sym__newline,
  [2556] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(498), 1,
      anon_sym_COLON,
  [2563] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(500), 1,
      ts_builtin_sym_end,
  [2570] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(502), 1,
      sym_identifier,
  [2577] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(504), 1,
      sym__indent,
  [2584] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(506), 1,
      sym__indent,
  [2591] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(508), 1,
      anon_sym_COLON,
  [2598] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(510), 1,
      anon_sym_DASH_GT,
  [2605] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(512), 1,
      sym__indent,
  [2612] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(514), 1,
      sym_identifier,
  [2619] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(516), 1,
      sym__indent,
  [2626] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(518), 1,
      sym_identifier,
  [2633] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(520), 1,
      sym_identifier,
  [2640] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(522), 1,
      sym__indent,
  [2647] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(524), 1,
      sym__indent,
  [2654] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(526), 1,
      ts_builtin_sym_end,
  [2661] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(528), 1,
      sym_identifier,
  [2668] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(530), 1,
      sym_identifier,
  [2675] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(532), 1,
      anon_sym_COLON,
  [2682] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(534), 1,
      ts_builtin_sym_end,
  [2689] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(536), 1,
      sym_identifier,
  [2696] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(538), 1,
      sym__indent,
};

static const uint32_t ts_small_parse_table_map[] = {
  [SMALL_STATE(2)] = 0,
  [SMALL_STATE(3)] = 45,
  [SMALL_STATE(4)] = 90,
  [SMALL_STATE(5)] = 157,
  [SMALL_STATE(6)] = 224,
  [SMALL_STATE(7)] = 291,
  [SMALL_STATE(8)] = 326,
  [SMALL_STATE(9)] = 361,
  [SMALL_STATE(10)] = 400,
  [SMALL_STATE(11)] = 433,
  [SMALL_STATE(12)] = 463,
  [SMALL_STATE(13)] = 493,
  [SMALL_STATE(14)] = 516,
  [SMALL_STATE(15)] = 539,
  [SMALL_STATE(16)] = 562,
  [SMALL_STATE(17)] = 585,
  [SMALL_STATE(18)] = 608,
  [SMALL_STATE(19)] = 631,
  [SMALL_STATE(20)] = 654,
  [SMALL_STATE(21)] = 677,
  [SMALL_STATE(22)] = 700,
  [SMALL_STATE(23)] = 723,
  [SMALL_STATE(24)] = 746,
  [SMALL_STATE(25)] = 769,
  [SMALL_STATE(26)] = 792,
  [SMALL_STATE(27)] = 815,
  [SMALL_STATE(28)] = 842,
  [SMALL_STATE(29)] = 869,
  [SMALL_STATE(30)] = 896,
  [SMALL_STATE(31)] = 922,
  [SMALL_STATE(32)] = 948,
  [SMALL_STATE(33)] = 974,
  [SMALL_STATE(34)] = 1006,
  [SMALL_STATE(35)] = 1028,
  [SMALL_STATE(36)] = 1060,
  [SMALL_STATE(37)] = 1092,
  [SMALL_STATE(38)] = 1114,
  [SMALL_STATE(39)] = 1136,
  [SMALL_STATE(40)] = 1157,
  [SMALL_STATE(41)] = 1177,
  [SMALL_STATE(42)] = 1197,
  [SMALL_STATE(43)] = 1217,
  [SMALL_STATE(44)] = 1237,
  [SMALL_STATE(45)] = 1269,
  [SMALL_STATE(46)] = 1296,
  [SMALL_STATE(47)] = 1323,
  [SMALL_STATE(48)] = 1349,
  [SMALL_STATE(49)] = 1373,
  [SMALL_STATE(50)] = 1391,
  [SMALL_STATE(51)] = 1414,
  [SMALL_STATE(52)] = 1435,
  [SMALL_STATE(53)] = 1456,
  [SMALL_STATE(54)] = 1479,
  [SMALL_STATE(55)] = 1502,
  [SMALL_STATE(56)] = 1525,
  [SMALL_STATE(57)] = 1548,
  [SMALL_STATE(58)] = 1571,
  [SMALL_STATE(59)] = 1593,
  [SMALL_STATE(60)] = 1613,
  [SMALL_STATE(61)] = 1633,
  [SMALL_STATE(62)] = 1653,
  [SMALL_STATE(63)] = 1673,
  [SMALL_STATE(64)] = 1693,
  [SMALL_STATE(65)] = 1713,
  [SMALL_STATE(66)] = 1733,
  [SMALL_STATE(67)] = 1753,
  [SMALL_STATE(68)] = 1773,
  [SMALL_STATE(69)] = 1787,
  [SMALL_STATE(70)] = 1807,
  [SMALL_STATE(71)] = 1827,
  [SMALL_STATE(72)] = 1845,
  [SMALL_STATE(73)] = 1865,
  [SMALL_STATE(74)] = 1882,
  [SMALL_STATE(75)] = 1899,
  [SMALL_STATE(76)] = 1916,
  [SMALL_STATE(77)] = 1933,
  [SMALL_STATE(78)] = 1950,
  [SMALL_STATE(79)] = 1969,
  [SMALL_STATE(80)] = 1986,
  [SMALL_STATE(81)] = 2005,
  [SMALL_STATE(82)] = 2022,
  [SMALL_STATE(83)] = 2039,
  [SMALL_STATE(84)] = 2051,
  [SMALL_STATE(85)] = 2065,
  [SMALL_STATE(86)] = 2079,
  [SMALL_STATE(87)] = 2093,
  [SMALL_STATE(88)] = 2109,
  [SMALL_STATE(89)] = 2123,
  [SMALL_STATE(90)] = 2137,
  [SMALL_STATE(91)] = 2151,
  [SMALL_STATE(92)] = 2165,
  [SMALL_STATE(93)] = 2179,
  [SMALL_STATE(94)] = 2193,
  [SMALL_STATE(95)] = 2207,
  [SMALL_STATE(96)] = 2221,
  [SMALL_STATE(97)] = 2235,
  [SMALL_STATE(98)] = 2249,
  [SMALL_STATE(99)] = 2258,
  [SMALL_STATE(100)] = 2267,
  [SMALL_STATE(101)] = 2276,
  [SMALL_STATE(102)] = 2285,
  [SMALL_STATE(103)] = 2294,
  [SMALL_STATE(104)] = 2307,
  [SMALL_STATE(105)] = 2320,
  [SMALL_STATE(106)] = 2330,
  [SMALL_STATE(107)] = 2340,
  [SMALL_STATE(108)] = 2350,
  [SMALL_STATE(109)] = 2360,
  [SMALL_STATE(110)] = 2367,
  [SMALL_STATE(111)] = 2374,
  [SMALL_STATE(112)] = 2381,
  [SMALL_STATE(113)] = 2388,
  [SMALL_STATE(114)] = 2395,
  [SMALL_STATE(115)] = 2402,
  [SMALL_STATE(116)] = 2409,
  [SMALL_STATE(117)] = 2416,
  [SMALL_STATE(118)] = 2423,
  [SMALL_STATE(119)] = 2430,
  [SMALL_STATE(120)] = 2437,
  [SMALL_STATE(121)] = 2444,
  [SMALL_STATE(122)] = 2451,
  [SMALL_STATE(123)] = 2458,
  [SMALL_STATE(124)] = 2465,
  [SMALL_STATE(125)] = 2472,
  [SMALL_STATE(126)] = 2479,
  [SMALL_STATE(127)] = 2486,
  [SMALL_STATE(128)] = 2493,
  [SMALL_STATE(129)] = 2500,
  [SMALL_STATE(130)] = 2507,
  [SMALL_STATE(131)] = 2514,
  [SMALL_STATE(132)] = 2521,
  [SMALL_STATE(133)] = 2528,
  [SMALL_STATE(134)] = 2535,
  [SMALL_STATE(135)] = 2542,
  [SMALL_STATE(136)] = 2549,
  [SMALL_STATE(137)] = 2556,
  [SMALL_STATE(138)] = 2563,
  [SMALL_STATE(139)] = 2570,
  [SMALL_STATE(140)] = 2577,
  [SMALL_STATE(141)] = 2584,
  [SMALL_STATE(142)] = 2591,
  [SMALL_STATE(143)] = 2598,
  [SMALL_STATE(144)] = 2605,
  [SMALL_STATE(145)] = 2612,
  [SMALL_STATE(146)] = 2619,
  [SMALL_STATE(147)] = 2626,
  [SMALL_STATE(148)] = 2633,
  [SMALL_STATE(149)] = 2640,
  [SMALL_STATE(150)] = 2647,
  [SMALL_STATE(151)] = 2654,
  [SMALL_STATE(152)] = 2661,
  [SMALL_STATE(153)] = 2668,
  [SMALL_STATE(154)] = 2675,
  [SMALL_STATE(155)] = 2682,
  [SMALL_STATE(156)] = 2689,
  [SMALL_STATE(157)] = 2696,
};

static const TSParseActionEntry ts_parse_actions[] = {
  [0] = {.entry = {.count = 0, .reusable = false}},
  [1] = {.entry = {.count = 1, .reusable = false}}, RECOVER(),
  [3] = {.entry = {.count = 1, .reusable = false}}, SHIFT_EXTRA(),
  [5] = {.entry = {.count = 1, .reusable = true}}, SHIFT(134),
  [7] = {.entry = {.count = 1, .reusable = true}}, SHIFT(127),
  [9] = {.entry = {.count = 1, .reusable = true}}, SHIFT_EXTRA(),
  [11] = {.entry = {.count = 1, .reusable = true}}, SHIFT(58),
  [13] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_string, 3, 0, 0),
  [15] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_string, 3, 0, 0),
  [17] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_string, 2, 0, 0),
  [19] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_string, 2, 0, 0),
  [21] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(142),
  [24] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(144),
  [27] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(145),
  [30] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(147),
  [33] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(153),
  [36] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(156),
  [39] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(110),
  [42] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(113),
  [45] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(114),
  [48] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(119),
  [51] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(121),
  [54] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(123),
  [57] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0),
  [59] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(4),
  [62] = {.entry = {.count = 1, .reusable = true}}, SHIFT(142),
  [64] = {.entry = {.count = 1, .reusable = true}}, SHIFT(144),
  [66] = {.entry = {.count = 1, .reusable = true}}, SHIFT(145),
  [68] = {.entry = {.count = 1, .reusable = true}}, SHIFT(147),
  [70] = {.entry = {.count = 1, .reusable = true}}, SHIFT(153),
  [72] = {.entry = {.count = 1, .reusable = true}}, SHIFT(156),
  [74] = {.entry = {.count = 1, .reusable = true}}, SHIFT(110),
  [76] = {.entry = {.count = 1, .reusable = true}}, SHIFT(113),
  [78] = {.entry = {.count = 1, .reusable = true}}, SHIFT(114),
  [80] = {.entry = {.count = 1, .reusable = true}}, SHIFT(119),
  [82] = {.entry = {.count = 1, .reusable = true}}, SHIFT(121),
  [84] = {.entry = {.count = 1, .reusable = true}}, SHIFT(123),
  [86] = {.entry = {.count = 1, .reusable = true}}, SHIFT(6),
  [88] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_workflow_body, 1, 0, 0),
  [90] = {.entry = {.count = 1, .reusable = true}}, SHIFT(4),
  [92] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_field_value, 1, 0, 0),
  [94] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_field_value, 1, 0, 0),
  [96] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_multiline_block, 3, 0, 0),
  [98] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_multiline_block, 3, 0, 0),
  [100] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_compare_expr, 1, 0, 0),
  [102] = {.entry = {.count = 1, .reusable = false}}, SHIFT(49),
  [104] = {.entry = {.count = 1, .reusable = true}}, SHIFT(101),
  [106] = {.entry = {.count = 1, .reusable = false}}, SHIFT(101),
  [108] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_expr, 1, 0, 0),
  [110] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_operand, 1, 0, 0),
  [112] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_operand, 1, 0, 0),
  [114] = {.entry = {.count = 1, .reusable = true}}, SHIFT(139),
  [116] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_variable, 3, 0, 0),
  [118] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_variable, 3, 0, 0),
  [120] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_fan_in_node, 5, 0, 0),
  [122] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_tool_node, 5, 0, 0),
  [124] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_subgraph_node, 5, 0, 0),
  [126] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_conditional_node, 5, 0, 0),
  [128] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_manager_loop_node, 5, 0, 0),
  [130] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_parallel_node, 5, 0, 0),
  [132] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_defaults_section, 4, 0, 0),
  [134] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_stylesheet_section, 5, 0, 0),
  [136] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_node_attr_block, 3, 0, 0),
  [138] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edges_section, 4, 0, 0),
  [140] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_workflow_field, 3, 0, 0),
  [142] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_node_decl, 1, 0, 0),
  [144] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_agent_node, 5, 0, 0),
  [146] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_human_node, 5, 0, 0),
  [148] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_and_expr, 1, 0, 0),
  [150] = {.entry = {.count = 1, .reusable = false}}, SHIFT(57),
  [152] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_and_expr, 1, 0, 0),
  [154] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_and_expr, 2, 0, 0),
  [156] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_and_expr, 2, 0, 0),
  [158] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_and_expr_repeat1, 2, 0, 0),
  [160] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_and_expr_repeat1, 2, 0, 0), SHIFT_REPEAT(57),
  [163] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_and_expr_repeat1, 2, 0, 0),
  [165] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_or_expr_repeat1, 2, 0, 0),
  [167] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_or_expr_repeat1, 2, 0, 0), SHIFT_REPEAT(47),
  [170] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_or_expr_repeat1, 2, 0, 0),
  [172] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_or_expr, 1, 0, 0),
  [174] = {.entry = {.count = 1, .reusable = false}}, SHIFT(47),
  [176] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_or_expr, 1, 0, 0),
  [178] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_or_expr, 2, 0, 0),
  [180] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_or_expr, 2, 0, 0),
  [182] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_entry, 3, 0, 0),
  [184] = {.entry = {.count = 1, .reusable = false}}, SHIFT(44),
  [186] = {.entry = {.count = 1, .reusable = false}}, SHIFT(120),
  [188] = {.entry = {.count = 1, .reusable = false}}, SHIFT(41),
  [190] = {.entry = {.count = 1, .reusable = false}}, SHIFT(122),
  [192] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_entry, 3, 0, 0),
  [194] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0),
  [196] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0), SHIFT_REPEAT(44),
  [199] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0), SHIFT_REPEAT(120),
  [202] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0), SHIFT_REPEAT(41),
  [205] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0), SHIFT_REPEAT(122),
  [208] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0),
  [210] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_entry, 4, 0, 0),
  [212] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_entry, 4, 0, 0),
  [214] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_compare_expr, 3, 0, 0),
  [216] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_expr, 3, 0, 0),
  [218] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_compare_expr, 4, 0, 0),
  [220] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_expr, 4, 0, 0),
  [222] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_attr, 3, 0, 0),
  [224] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_attr, 3, 0, 0),
  [226] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_attr, 1, 0, 0),
  [228] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_attr, 1, 0, 0),
  [230] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_attr, 2, 0, 0),
  [232] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_attr, 2, 0, 0),
  [234] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_condition, 1, 0, 0),
  [236] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_condition, 1, 0, 0),
  [238] = {.entry = {.count = 1, .reusable = true}}, SHIFT(10),
  [240] = {.entry = {.count = 1, .reusable = true}}, SHIFT(91),
  [242] = {.entry = {.count = 1, .reusable = true}}, SHIFT(94),
  [244] = {.entry = {.count = 1, .reusable = true}}, SHIFT(125),
  [246] = {.entry = {.count = 1, .reusable = true}}, SHIFT(117),
  [248] = {.entry = {.count = 1, .reusable = false}}, SHIFT(117),
  [250] = {.entry = {.count = 1, .reusable = true}}, SHIFT(20),
  [252] = {.entry = {.count = 1, .reusable = true}}, SHIFT(46),
  [254] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(125),
  [257] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(117),
  [260] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(117),
  [263] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0),
  [265] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(46),
  [268] = {.entry = {.count = 1, .reusable = true}}, SHIFT(45),
  [270] = {.entry = {.count = 1, .reusable = false}}, SHIFT(7),
  [272] = {.entry = {.count = 1, .reusable = false}}, SHIFT(91),
  [274] = {.entry = {.count = 1, .reusable = false}}, SHIFT(94),
  [276] = {.entry = {.count = 1, .reusable = true}}, SHIFT(96),
  [278] = {.entry = {.count = 1, .reusable = false}}, SHIFT(130),
  [280] = {.entry = {.count = 1, .reusable = false}}, SHIFT(116),
  [282] = {.entry = {.count = 1, .reusable = true}}, SHIFT(22),
  [284] = {.entry = {.count = 1, .reusable = true}}, SHIFT(52),
  [286] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_edges_section_repeat1, 2, 0, 0), SHIFT_REPEAT(130),
  [289] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_edges_section_repeat1, 2, 0, 0), SHIFT_REPEAT(116),
  [292] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_edges_section_repeat1, 2, 0, 0),
  [294] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_edges_section_repeat1, 2, 0, 0), SHIFT_REPEAT(52),
  [297] = {.entry = {.count = 1, .reusable = true}}, SHIFT(97),
  [299] = {.entry = {.count = 1, .reusable = true}}, SHIFT(118),
  [301] = {.entry = {.count = 1, .reusable = true}}, SHIFT(19),
  [303] = {.entry = {.count = 1, .reusable = true}}, SHIFT(60),
  [305] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_defaults_section_repeat1, 2, 0, 0), SHIFT_REPEAT(118),
  [308] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_defaults_section_repeat1, 2, 0, 0),
  [310] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_defaults_section_repeat1, 2, 0, 0), SHIFT_REPEAT(60),
  [313] = {.entry = {.count = 1, .reusable = true}}, SHIFT(25),
  [315] = {.entry = {.count = 1, .reusable = true}}, SHIFT(69),
  [317] = {.entry = {.count = 1, .reusable = true}}, SHIFT(26),
  [319] = {.entry = {.count = 1, .reusable = true}}, SHIFT(14),
  [321] = {.entry = {.count = 1, .reusable = true}}, SHIFT(15),
  [323] = {.entry = {.count = 1, .reusable = true}}, SHIFT(16),
  [325] = {.entry = {.count = 1, .reusable = true}}, SHIFT(17),
  [327] = {.entry = {.count = 1, .reusable = true}}, SHIFT(21),
  [329] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_stylesheet_rule, 4, 0, 0),
  [331] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_stylesheet_rule, 4, 0, 0),
  [333] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_agent_node_repeat1, 2, 0, 0), SHIFT_REPEAT(118),
  [336] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_agent_node_repeat1, 2, 0, 0),
  [338] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_agent_node_repeat1, 2, 0, 0), SHIFT_REPEAT(69),
  [341] = {.entry = {.count = 1, .reusable = true}}, SHIFT(51),
  [343] = {.entry = {.count = 1, .reusable = true}}, SHIFT(63),
  [345] = {.entry = {.count = 1, .reusable = true}}, SHIFT(65),
  [347] = {.entry = {.count = 1, .reusable = true}}, SHIFT(66),
  [349] = {.entry = {.count = 1, .reusable = true}}, SHIFT(67),
  [351] = {.entry = {.count = 1, .reusable = true}}, SHIFT(64),
  [353] = {.entry = {.count = 1, .reusable = true}}, SHIFT(68),
  [355] = {.entry = {.count = 1, .reusable = true}}, SHIFT(80),
  [357] = {.entry = {.count = 1, .reusable = true}}, SHIFT(59),
  [359] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 2, 0, 0), SHIFT_REPEAT(118),
  [362] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 2, 0, 0),
  [364] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 2, 0, 0), SHIFT_REPEAT(80),
  [367] = {.entry = {.count = 1, .reusable = true}}, SHIFT(61),
  [369] = {.entry = {.count = 1, .reusable = true}}, SHIFT(62),
  [371] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_else_default, 3, 0, 0),
  [373] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_else_default, 3, 0, 0),
  [375] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_block_content_repeat1, 2, 0, 0), SHIFT_REPEAT(84),
  [378] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_block_content_repeat1, 2, 0, 0),
  [380] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_string_repeat1, 2, 0, 0),
  [382] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_string_repeat1, 2, 0, 0), SHIFT_REPEAT(85),
  [385] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_string_repeat2, 2, 0, 0),
  [387] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_string_repeat2, 2, 0, 0), SHIFT_REPEAT(86),
  [390] = {.entry = {.count = 1, .reusable = true}}, SHIFT(78),
  [392] = {.entry = {.count = 1, .reusable = true}}, SHIFT(148),
  [394] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_identifier_list, 2, 0, 0),
  [396] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_identifier_list_repeat1, 2, 0, 0), SHIFT_REPEAT(148),
  [399] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_identifier_list_repeat1, 2, 0, 0),
  [401] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_identifier_list, 1, 0, 0),
  [403] = {.entry = {.count = 1, .reusable = false}}, SHIFT(3),
  [405] = {.entry = {.count = 1, .reusable = false}}, SHIFT(93),
  [407] = {.entry = {.count = 1, .reusable = true}}, SHIFT(84),
  [409] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_block_content, 1, 0, 0),
  [411] = {.entry = {.count = 1, .reusable = false}}, SHIFT(2),
  [413] = {.entry = {.count = 1, .reusable = false}}, SHIFT(85),
  [415] = {.entry = {.count = 1, .reusable = false}}, SHIFT(95),
  [417] = {.entry = {.count = 1, .reusable = false}}, SHIFT(86),
  [419] = {.entry = {.count = 1, .reusable = true}}, SHIFT(92),
  [421] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0),
  [423] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(97),
  [426] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_defaults_field, 3, 0, 0),
  [428] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_node_field, 3, 0, 0),
  [430] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_op, 1, 0, 0),
  [432] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 3, 0, 0),
  [434] = {.entry = {.count = 1, .reusable = true}}, SHIFT(76),
  [436] = {.entry = {.count = 1, .reusable = true}}, SHIFT(13),
  [438] = {.entry = {.count = 1, .reusable = true}}, SHIFT(18),
  [440] = {.entry = {.count = 1, .reusable = true}}, SHIFT(90),
  [442] = {.entry = {.count = 1, .reusable = true}}, SHIFT(106),
  [444] = {.entry = {.count = 1, .reusable = true}}, SHIFT(140),
  [446] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_selector, 2, 0, 0),
  [448] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 2, 0, 0),
  [450] = {.entry = {.count = 1, .reusable = true}}, SHIFT(141),
  [452] = {.entry = {.count = 1, .reusable = true}}, SHIFT(143),
  [454] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_version_decl, 3, 0, 0),
  [456] = {.entry = {.count = 1, .reusable = true}}, SHIFT(152),
  [458] = {.entry = {.count = 1, .reusable = true}}, SHIFT(111),
  [460] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_field_name, 1, 0, 0),
  [462] = {.entry = {.count = 1, .reusable = true}}, SHIFT(109),
  [464] = {.entry = {.count = 1, .reusable = true}}, SHIFT(42),
  [466] = {.entry = {.count = 1, .reusable = true}}, SHIFT(71),
  [468] = {.entry = {.count = 1, .reusable = true}}, SHIFT(56),
  [470] = {.entry = {.count = 1, .reusable = true}}, SHIFT(150),
  [472] = {.entry = {.count = 1, .reusable = true}}, SHIFT(54),
  [474] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_selector, 1, 0, 0),
  [476] = {.entry = {.count = 1, .reusable = true}}, SHIFT(151),
  [478] = {.entry = {.count = 1, .reusable = true}}, SHIFT(146),
  [480] = {.entry = {.count = 1, .reusable = true}},  ACCEPT_INPUT(),
  [482] = {.entry = {.count = 1, .reusable = true}}, SHIFT(8),
  [484] = {.entry = {.count = 1, .reusable = true}}, SHIFT(133),
  [486] = {.entry = {.count = 1, .reusable = true}}, SHIFT(81),
  [488] = {.entry = {.count = 1, .reusable = true}}, SHIFT(82),
  [490] = {.entry = {.count = 1, .reusable = true}}, SHIFT(33),
  [492] = {.entry = {.count = 1, .reusable = true}}, SHIFT(136),
  [494] = {.entry = {.count = 1, .reusable = true}}, SHIFT(77),
  [496] = {.entry = {.count = 1, .reusable = true}}, SHIFT(115),
  [498] = {.entry = {.count = 1, .reusable = true}}, SHIFT(50),
  [500] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 3, 0, 0),
  [502] = {.entry = {.count = 1, .reusable = true}}, SHIFT(12),
  [504] = {.entry = {.count = 1, .reusable = true}}, SHIFT(74),
  [506] = {.entry = {.count = 1, .reusable = true}}, SHIFT(75),
  [508] = {.entry = {.count = 1, .reusable = true}}, SHIFT(55),
  [510] = {.entry = {.count = 1, .reusable = true}}, SHIFT(105),
  [512] = {.entry = {.count = 1, .reusable = true}}, SHIFT(79),
  [514] = {.entry = {.count = 1, .reusable = true}}, SHIFT(131),
  [516] = {.entry = {.count = 1, .reusable = true}}, SHIFT(5),
  [518] = {.entry = {.count = 1, .reusable = true}}, SHIFT(132),
  [520] = {.entry = {.count = 1, .reusable = true}}, SHIFT(100),
  [522] = {.entry = {.count = 1, .reusable = true}}, SHIFT(87),
  [524] = {.entry = {.count = 1, .reusable = true}}, SHIFT(48),
  [526] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_workflow_decl, 5, 0, 0),
  [528] = {.entry = {.count = 1, .reusable = true}}, SHIFT(83),
  [530] = {.entry = {.count = 1, .reusable = true}}, SHIFT(157),
  [532] = {.entry = {.count = 1, .reusable = true}}, SHIFT(53),
  [534] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 1, 0, 0),
  [536] = {.entry = {.count = 1, .reusable = true}}, SHIFT(135),
  [538] = {.entry = {.count = 1, .reusable = true}}, SHIFT(73),
};

enum ts_external_scanner_symbol_identifiers {
  ts_external_token__indent = 0,
  ts_external_token__dedent = 1,
  ts_external_token__newline = 2,
};

static const TSSymbol ts_external_scanner_symbol_map[EXTERNAL_TOKEN_COUNT] = {
  [ts_external_token__indent] = sym__indent,
  [ts_external_token__dedent] = sym__dedent,
  [ts_external_token__newline] = sym__newline,
};

static const bool ts_external_scanner_states[7][EXTERNAL_TOKEN_COUNT] = {
  [1] = {
    [ts_external_token__indent] = true,
    [ts_external_token__dedent] = true,
    [ts_external_token__newline] = true,
  },
  [2] = {
    [ts_external_token__newline] = true,
  },
  [3] = {
    [ts_external_token__dedent] = true,
    [ts_external_token__newline] = true,
  },
  [4] = {
    [ts_external_token__indent] = true,
  },
  [5] = {
    [ts_external_token__indent] = true,
    [ts_external_token__newline] = true,
  },
  [6] = {
    [ts_external_token__dedent] = true,
  },
};

#ifdef __cplusplus
extern "C" {
#endif
void *tree_sitter_dippin_external_scanner_create(void);
void tree_sitter_dippin_external_scanner_destroy(void *);
bool tree_sitter_dippin_external_scanner_scan(void *, TSLexer *, const bool *);
unsigned tree_sitter_dippin_external_scanner_serialize(void *, char *);
void tree_sitter_dippin_external_scanner_deserialize(void *, const char *, unsigned);

#ifdef TREE_SITTER_HIDE_SYMBOLS
#define TS_PUBLIC
#elif defined(_WIN32)
#define TS_PUBLIC __declspec(dllexport)
#else
#define TS_PUBLIC __attribute__((visibility("default")))
#endif

TS_PUBLIC const TSLanguage *tree_sitter_dippin(void) {
  static const TSLanguage language = {
    .version = LANGUAGE_VERSION,
    .symbol_count = SYMBOL_COUNT,
    .alias_count = ALIAS_COUNT,
    .token_count = TOKEN_COUNT,
    .external_token_count = EXTERNAL_TOKEN_COUNT,
    .state_count = STATE_COUNT,
    .large_state_count = LARGE_STATE_COUNT,
    .production_id_count = PRODUCTION_ID_COUNT,
    .field_count = FIELD_COUNT,
    .max_alias_sequence_length = MAX_ALIAS_SEQUENCE_LENGTH,
    .parse_table = &ts_parse_table[0][0],
    .small_parse_table = ts_small_parse_table,
    .small_parse_table_map = ts_small_parse_table_map,
    .parse_actions = ts_parse_actions,
    .symbol_names = ts_symbol_names,
    .symbol_metadata = ts_symbol_metadata,
    .public_symbol_map = ts_symbol_map,
    .alias_map = ts_non_terminal_alias_map,
    .alias_sequences = &ts_alias_sequences[0][0],
    .lex_modes = ts_lex_modes,
    .lex_fn = ts_lex,
    .keyword_lex_fn = ts_lex_keywords,
    .keyword_capture_token = sym_identifier,
    .external_scanner = {
      &ts_external_scanner_states[0][0],
      ts_external_scanner_symbol_map,
      tree_sitter_dippin_external_scanner_create,
      tree_sitter_dippin_external_scanner_destroy,
      tree_sitter_dippin_external_scanner_scan,
      tree_sitter_dippin_external_scanner_serialize,
      tree_sitter_dippin_external_scanner_deserialize,
    },
    .primary_state_ids = ts_primary_state_ids,
  };
  return &language;
}
#ifdef __cplusplus
}
#endif
