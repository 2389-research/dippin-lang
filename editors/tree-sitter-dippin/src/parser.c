#include "tree_sitter/parser.h"

#if defined(__GNUC__) || defined(__clang__)
#pragma GCC diagnostic ignored "-Wmissing-field-initializers"
#endif

#define LANGUAGE_VERSION 14
#define STATE_COUNT 148
#define LARGE_STATE_COUNT 2
#define SYMBOL_COUNT 103
#define ALIAS_COUNT 0
#define TOKEN_COUNT 53
#define EXTERNAL_TOKEN_COUNT 3
#define FIELD_COUNT 0
#define MAX_ALIAS_SEQUENCE_LENGTH 5
#define PRODUCTION_ID_COUNT 1

enum ts_symbol_identifiers {
  sym_identifier = 1,
  anon_sym_workflow = 2,
  anon_sym_goal = 3,
  anon_sym_start = 4,
  anon_sym_exit = 5,
  anon_sym_requires = 6,
  anon_sym_COLON = 7,
  anon_sym_defaults = 8,
  anon_sym_agent = 9,
  anon_sym_human = 10,
  anon_sym_tool = 11,
  anon_sym_subgraph = 12,
  anon_sym_conditional = 13,
  anon_sym_manager_loop = 14,
  anon_sym_parallel = 15,
  anon_sym_DASH_GT = 16,
  anon_sym_fan_in = 17,
  anon_sym_LT_DASH = 18,
  anon_sym_edges = 19,
  anon_sym_when = 20,
  anon_sym_on = 21,
  anon_sym_label = 22,
  anon_sym_weight = 23,
  anon_sym_restart = 24,
  anon_sym_override = 25,
  anon_sym_or = 26,
  anon_sym_and = 27,
  anon_sym_not = 28,
  anon_sym_EQ_EQ = 29,
  anon_sym_BANG_EQ = 30,
  anon_sym_EQ = 31,
  anon_sym_contains = 32,
  anon_sym_startswith = 33,
  anon_sym_endswith = 34,
  anon_sym_in = 35,
  anon_sym_DOT = 36,
  anon_sym_stylesheet = 37,
  anon_sym_STAR = 38,
  anon_sym_POUND = 39,
  sym_raw_inline = 40,
  sym_block_line = 41,
  anon_sym_COMMA = 42,
  anon_sym_DQUOTE = 43,
  aux_sym_string_token1 = 44,
  aux_sym_string_token2 = 45,
  anon_sym_SQUOTE = 46,
  aux_sym_string_token3 = 47,
  anon_sym_SQUOTE_SQUOTE = 48,
  sym_comment = 49,
  sym__indent = 50,
  sym__dedent = 51,
  sym__newline = 52,
  sym_source_file = 53,
  sym_workflow_decl = 54,
  sym_workflow_body = 55,
  sym_workflow_field = 56,
  sym_defaults_section = 57,
  sym_defaults_field = 58,
  sym_node_decl = 59,
  sym_agent_node = 60,
  sym_human_node = 61,
  sym_tool_node = 62,
  sym_subgraph_node = 63,
  sym_conditional_node = 64,
  sym_manager_loop_node = 65,
  sym_parallel_node = 66,
  sym_fan_in_node = 67,
  sym_node_attr_block = 68,
  sym_node_field = 69,
  sym_edges_section = 70,
  sym_edge_entry = 71,
  sym_edge_attr = 72,
  sym_condition = 73,
  sym_or_expr = 74,
  sym_and_expr = 75,
  sym_compare_expr = 76,
  sym_compare_op = 77,
  sym_operand = 78,
  sym_variable = 79,
  sym_stylesheet_section = 80,
  sym_stylesheet_rule = 81,
  sym_selector = 82,
  sym_field_name = 83,
  sym_field_value = 84,
  sym_multiline_block = 85,
  sym_block_content = 86,
  sym_identifier_list = 87,
  sym_string = 88,
  aux_sym_source_file_repeat1 = 89,
  aux_sym_workflow_body_repeat1 = 90,
  aux_sym_defaults_section_repeat1 = 91,
  aux_sym_agent_node_repeat1 = 92,
  aux_sym_edges_section_repeat1 = 93,
  aux_sym_edge_entry_repeat1 = 94,
  aux_sym_or_expr_repeat1 = 95,
  aux_sym_and_expr_repeat1 = 96,
  aux_sym_stylesheet_section_repeat1 = 97,
  aux_sym_stylesheet_rule_repeat1 = 98,
  aux_sym_block_content_repeat1 = 99,
  aux_sym_identifier_list_repeat1 = 100,
  aux_sym_string_repeat1 = 101,
  aux_sym_string_repeat2 = 102,
};

static const char * const ts_symbol_names[] = {
  [ts_builtin_sym_end] = "end",
  [sym_identifier] = "identifier",
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
  [anon_sym_when] = "when",
  [anon_sym_on] = "on",
  [anon_sym_label] = "label",
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
  [anon_sym_when] = anon_sym_when,
  [anon_sym_on] = anon_sym_on,
  [anon_sym_label] = anon_sym_label,
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
  [anon_sym_when] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_on] = {
    .visible = true,
    .named = false,
  },
  [anon_sym_label] = {
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
      if (lookahead == 'o') ADVANCE(20);
      END_STATE();
    case 3:
      if (lookahead == 'e') ADVANCE(21);
      END_STATE();
    case 4:
      if (lookahead == 'd') ADVANCE(22);
      if (lookahead == 'n') ADVANCE(23);
      if (lookahead == 'x') ADVANCE(24);
      END_STATE();
    case 5:
      if (lookahead == 'a') ADVANCE(25);
      END_STATE();
    case 6:
      if (lookahead == 'o') ADVANCE(26);
      END_STATE();
    case 7:
      if (lookahead == 'u') ADVANCE(27);
      END_STATE();
    case 8:
      if (lookahead == 'n') ADVANCE(28);
      END_STATE();
    case 9:
      if (lookahead == 'a') ADVANCE(29);
      END_STATE();
    case 10:
      if (lookahead == 'a') ADVANCE(30);
      END_STATE();
    case 11:
      if (lookahead == 'o') ADVANCE(31);
      END_STATE();
    case 12:
      if (lookahead == 'n') ADVANCE(32);
      if (lookahead == 'r') ADVANCE(33);
      if (lookahead == 'v') ADVANCE(34);
      END_STATE();
    case 13:
      if (lookahead == 'a') ADVANCE(35);
      END_STATE();
    case 14:
      if (lookahead == 'e') ADVANCE(36);
      END_STATE();
    case 15:
      if (lookahead == 't') ADVANCE(37);
      if (lookahead == 'u') ADVANCE(38);
      END_STATE();
    case 16:
      if (lookahead == 'o') ADVANCE(39);
      END_STATE();
    case 17:
      if (lookahead == 'e') ADVANCE(40);
      if (lookahead == 'h') ADVANCE(41);
      if (lookahead == 'o') ADVANCE(42);
      END_STATE();
    case 18:
      if (lookahead == 'e') ADVANCE(43);
      END_STATE();
    case 19:
      if (lookahead == 'd') ADVANCE(44);
      END_STATE();
    case 20:
      if (lookahead == 'n') ADVANCE(45);
      END_STATE();
    case 21:
      if (lookahead == 'f') ADVANCE(46);
      END_STATE();
    case 22:
      if (lookahead == 'g') ADVANCE(47);
      END_STATE();
    case 23:
      if (lookahead == 'd') ADVANCE(48);
      END_STATE();
    case 24:
      if (lookahead == 'i') ADVANCE(49);
      END_STATE();
    case 25:
      if (lookahead == 'n') ADVANCE(50);
      END_STATE();
    case 26:
      if (lookahead == 'a') ADVANCE(51);
      END_STATE();
    case 27:
      if (lookahead == 'm') ADVANCE(52);
      END_STATE();
    case 28:
      ACCEPT_TOKEN(anon_sym_in);
      END_STATE();
    case 29:
      if (lookahead == 'b') ADVANCE(53);
      END_STATE();
    case 30:
      if (lookahead == 'n') ADVANCE(54);
      END_STATE();
    case 31:
      if (lookahead == 't') ADVANCE(55);
      END_STATE();
    case 32:
      ACCEPT_TOKEN(anon_sym_on);
      END_STATE();
    case 33:
      ACCEPT_TOKEN(anon_sym_or);
      END_STATE();
    case 34:
      if (lookahead == 'e') ADVANCE(56);
      END_STATE();
    case 35:
      if (lookahead == 'r') ADVANCE(57);
      END_STATE();
    case 36:
      if (lookahead == 'q') ADVANCE(58);
      if (lookahead == 's') ADVANCE(59);
      END_STATE();
    case 37:
      if (lookahead == 'a') ADVANCE(60);
      if (lookahead == 'y') ADVANCE(61);
      END_STATE();
    case 38:
      if (lookahead == 'b') ADVANCE(62);
      END_STATE();
    case 39:
      if (lookahead == 'o') ADVANCE(63);
      END_STATE();
    case 40:
      if (lookahead == 'i') ADVANCE(64);
      END_STATE();
    case 41:
      if (lookahead == 'e') ADVANCE(65);
      END_STATE();
    case 42:
      if (lookahead == 'r') ADVANCE(66);
      END_STATE();
    case 43:
      if (lookahead == 'n') ADVANCE(67);
      END_STATE();
    case 44:
      ACCEPT_TOKEN(anon_sym_and);
      END_STATE();
    case 45:
      if (lookahead == 'd') ADVANCE(68);
      if (lookahead == 't') ADVANCE(69);
      END_STATE();
    case 46:
      if (lookahead == 'a') ADVANCE(70);
      END_STATE();
    case 47:
      if (lookahead == 'e') ADVANCE(71);
      END_STATE();
    case 48:
      if (lookahead == 's') ADVANCE(72);
      END_STATE();
    case 49:
      if (lookahead == 't') ADVANCE(73);
      END_STATE();
    case 50:
      if (lookahead == '_') ADVANCE(74);
      END_STATE();
    case 51:
      if (lookahead == 'l') ADVANCE(75);
      END_STATE();
    case 52:
      if (lookahead == 'a') ADVANCE(76);
      END_STATE();
    case 53:
      if (lookahead == 'e') ADVANCE(77);
      END_STATE();
    case 54:
      if (lookahead == 'a') ADVANCE(78);
      END_STATE();
    case 55:
      ACCEPT_TOKEN(anon_sym_not);
      END_STATE();
    case 56:
      if (lookahead == 'r') ADVANCE(79);
      END_STATE();
    case 57:
      if (lookahead == 'a') ADVANCE(80);
      END_STATE();
    case 58:
      if (lookahead == 'u') ADVANCE(81);
      END_STATE();
    case 59:
      if (lookahead == 't') ADVANCE(82);
      END_STATE();
    case 60:
      if (lookahead == 'r') ADVANCE(83);
      END_STATE();
    case 61:
      if (lookahead == 'l') ADVANCE(84);
      END_STATE();
    case 62:
      if (lookahead == 'g') ADVANCE(85);
      END_STATE();
    case 63:
      if (lookahead == 'l') ADVANCE(86);
      END_STATE();
    case 64:
      if (lookahead == 'g') ADVANCE(87);
      END_STATE();
    case 65:
      if (lookahead == 'n') ADVANCE(88);
      END_STATE();
    case 66:
      if (lookahead == 'k') ADVANCE(89);
      END_STATE();
    case 67:
      if (lookahead == 't') ADVANCE(90);
      END_STATE();
    case 68:
      if (lookahead == 'i') ADVANCE(91);
      END_STATE();
    case 69:
      if (lookahead == 'a') ADVANCE(92);
      END_STATE();
    case 70:
      if (lookahead == 'u') ADVANCE(93);
      END_STATE();
    case 71:
      if (lookahead == 's') ADVANCE(94);
      END_STATE();
    case 72:
      if (lookahead == 'w') ADVANCE(95);
      END_STATE();
    case 73:
      ACCEPT_TOKEN(anon_sym_exit);
      END_STATE();
    case 74:
      if (lookahead == 'i') ADVANCE(96);
      END_STATE();
    case 75:
      ACCEPT_TOKEN(anon_sym_goal);
      END_STATE();
    case 76:
      if (lookahead == 'n') ADVANCE(97);
      END_STATE();
    case 77:
      if (lookahead == 'l') ADVANCE(98);
      END_STATE();
    case 78:
      if (lookahead == 'g') ADVANCE(99);
      END_STATE();
    case 79:
      if (lookahead == 'r') ADVANCE(100);
      END_STATE();
    case 80:
      if (lookahead == 'l') ADVANCE(101);
      END_STATE();
    case 81:
      if (lookahead == 'i') ADVANCE(102);
      END_STATE();
    case 82:
      if (lookahead == 'a') ADVANCE(103);
      END_STATE();
    case 83:
      if (lookahead == 't') ADVANCE(104);
      END_STATE();
    case 84:
      if (lookahead == 'e') ADVANCE(105);
      END_STATE();
    case 85:
      if (lookahead == 'r') ADVANCE(106);
      END_STATE();
    case 86:
      ACCEPT_TOKEN(anon_sym_tool);
      END_STATE();
    case 87:
      if (lookahead == 'h') ADVANCE(107);
      END_STATE();
    case 88:
      ACCEPT_TOKEN(anon_sym_when);
      END_STATE();
    case 89:
      if (lookahead == 'f') ADVANCE(108);
      END_STATE();
    case 90:
      ACCEPT_TOKEN(anon_sym_agent);
      END_STATE();
    case 91:
      if (lookahead == 't') ADVANCE(109);
      END_STATE();
    case 92:
      if (lookahead == 'i') ADVANCE(110);
      END_STATE();
    case 93:
      if (lookahead == 'l') ADVANCE(111);
      END_STATE();
    case 94:
      ACCEPT_TOKEN(anon_sym_edges);
      END_STATE();
    case 95:
      if (lookahead == 'i') ADVANCE(112);
      END_STATE();
    case 96:
      if (lookahead == 'n') ADVANCE(113);
      END_STATE();
    case 97:
      ACCEPT_TOKEN(anon_sym_human);
      END_STATE();
    case 98:
      ACCEPT_TOKEN(anon_sym_label);
      END_STATE();
    case 99:
      if (lookahead == 'e') ADVANCE(114);
      END_STATE();
    case 100:
      if (lookahead == 'i') ADVANCE(115);
      END_STATE();
    case 101:
      if (lookahead == 'l') ADVANCE(116);
      END_STATE();
    case 102:
      if (lookahead == 'r') ADVANCE(117);
      END_STATE();
    case 103:
      if (lookahead == 'r') ADVANCE(118);
      END_STATE();
    case 104:
      ACCEPT_TOKEN(anon_sym_start);
      if (lookahead == 's') ADVANCE(119);
      END_STATE();
    case 105:
      if (lookahead == 's') ADVANCE(120);
      END_STATE();
    case 106:
      if (lookahead == 'a') ADVANCE(121);
      END_STATE();
    case 107:
      if (lookahead == 't') ADVANCE(122);
      END_STATE();
    case 108:
      if (lookahead == 'l') ADVANCE(123);
      END_STATE();
    case 109:
      if (lookahead == 'i') ADVANCE(124);
      END_STATE();
    case 110:
      if (lookahead == 'n') ADVANCE(125);
      END_STATE();
    case 111:
      if (lookahead == 't') ADVANCE(126);
      END_STATE();
    case 112:
      if (lookahead == 't') ADVANCE(127);
      END_STATE();
    case 113:
      ACCEPT_TOKEN(anon_sym_fan_in);
      END_STATE();
    case 114:
      if (lookahead == 'r') ADVANCE(128);
      END_STATE();
    case 115:
      if (lookahead == 'd') ADVANCE(129);
      END_STATE();
    case 116:
      if (lookahead == 'e') ADVANCE(130);
      END_STATE();
    case 117:
      if (lookahead == 'e') ADVANCE(131);
      END_STATE();
    case 118:
      if (lookahead == 't') ADVANCE(132);
      END_STATE();
    case 119:
      if (lookahead == 'w') ADVANCE(133);
      END_STATE();
    case 120:
      if (lookahead == 'h') ADVANCE(134);
      END_STATE();
    case 121:
      if (lookahead == 'p') ADVANCE(135);
      END_STATE();
    case 122:
      ACCEPT_TOKEN(anon_sym_weight);
      END_STATE();
    case 123:
      if (lookahead == 'o') ADVANCE(136);
      END_STATE();
    case 124:
      if (lookahead == 'o') ADVANCE(137);
      END_STATE();
    case 125:
      if (lookahead == 's') ADVANCE(138);
      END_STATE();
    case 126:
      if (lookahead == 's') ADVANCE(139);
      END_STATE();
    case 127:
      if (lookahead == 'h') ADVANCE(140);
      END_STATE();
    case 128:
      if (lookahead == '_') ADVANCE(141);
      END_STATE();
    case 129:
      if (lookahead == 'e') ADVANCE(142);
      END_STATE();
    case 130:
      if (lookahead == 'l') ADVANCE(143);
      END_STATE();
    case 131:
      if (lookahead == 's') ADVANCE(144);
      END_STATE();
    case 132:
      ACCEPT_TOKEN(anon_sym_restart);
      END_STATE();
    case 133:
      if (lookahead == 'i') ADVANCE(145);
      END_STATE();
    case 134:
      if (lookahead == 'e') ADVANCE(146);
      END_STATE();
    case 135:
      if (lookahead == 'h') ADVANCE(147);
      END_STATE();
    case 136:
      if (lookahead == 'w') ADVANCE(148);
      END_STATE();
    case 137:
      if (lookahead == 'n') ADVANCE(149);
      END_STATE();
    case 138:
      ACCEPT_TOKEN(anon_sym_contains);
      END_STATE();
    case 139:
      ACCEPT_TOKEN(anon_sym_defaults);
      END_STATE();
    case 140:
      ACCEPT_TOKEN(anon_sym_endswith);
      END_STATE();
    case 141:
      if (lookahead == 'l') ADVANCE(150);
      END_STATE();
    case 142:
      ACCEPT_TOKEN(anon_sym_override);
      END_STATE();
    case 143:
      ACCEPT_TOKEN(anon_sym_parallel);
      END_STATE();
    case 144:
      ACCEPT_TOKEN(anon_sym_requires);
      END_STATE();
    case 145:
      if (lookahead == 't') ADVANCE(151);
      END_STATE();
    case 146:
      if (lookahead == 'e') ADVANCE(152);
      END_STATE();
    case 147:
      ACCEPT_TOKEN(anon_sym_subgraph);
      END_STATE();
    case 148:
      ACCEPT_TOKEN(anon_sym_workflow);
      END_STATE();
    case 149:
      if (lookahead == 'a') ADVANCE(153);
      END_STATE();
    case 150:
      if (lookahead == 'o') ADVANCE(154);
      END_STATE();
    case 151:
      if (lookahead == 'h') ADVANCE(155);
      END_STATE();
    case 152:
      if (lookahead == 't') ADVANCE(156);
      END_STATE();
    case 153:
      if (lookahead == 'l') ADVANCE(157);
      END_STATE();
    case 154:
      if (lookahead == 'o') ADVANCE(158);
      END_STATE();
    case 155:
      ACCEPT_TOKEN(anon_sym_startswith);
      END_STATE();
    case 156:
      ACCEPT_TOKEN(anon_sym_stylesheet);
      END_STATE();
    case 157:
      ACCEPT_TOKEN(anon_sym_conditional);
      END_STATE();
    case 158:
      if (lookahead == 'p') ADVANCE(159);
      END_STATE();
    case 159:
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
  [5] = {.lex_state = 9, .external_lex_state = 3},
  [6] = {.lex_state = 9, .external_lex_state = 2},
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
  [39] = {.lex_state = 9},
  [40] = {.lex_state = 9, .external_lex_state = 3},
  [41] = {.lex_state = 9, .external_lex_state = 3},
  [42] = {.lex_state = 9, .external_lex_state = 3},
  [43] = {.lex_state = 9, .external_lex_state = 3},
  [44] = {.lex_state = 0, .external_lex_state = 3},
  [45] = {.lex_state = 0, .external_lex_state = 3},
  [46] = {.lex_state = 9},
  [47] = {.lex_state = 0, .external_lex_state = 2},
  [48] = {.lex_state = 9},
  [49] = {.lex_state = 1, .external_lex_state = 4},
  [50] = {.lex_state = 1, .external_lex_state = 4},
  [51] = {.lex_state = 1, .external_lex_state = 4},
  [52] = {.lex_state = 9},
  [53] = {.lex_state = 1, .external_lex_state = 4},
  [54] = {.lex_state = 1, .external_lex_state = 4},
  [55] = {.lex_state = 9},
  [56] = {.lex_state = 0, .external_lex_state = 3},
  [57] = {.lex_state = 9, .external_lex_state = 3},
  [58] = {.lex_state = 9, .external_lex_state = 3},
  [59] = {.lex_state = 9, .external_lex_state = 3},
  [60] = {.lex_state = 9, .external_lex_state = 3},
  [61] = {.lex_state = 9, .external_lex_state = 3},
  [62] = {.lex_state = 9, .external_lex_state = 3},
  [63] = {.lex_state = 9},
  [64] = {.lex_state = 9, .external_lex_state = 3},
  [65] = {.lex_state = 9, .external_lex_state = 3},
  [66] = {.lex_state = 9, .external_lex_state = 3},
  [67] = {.lex_state = 9, .external_lex_state = 3},
  [68] = {.lex_state = 9, .external_lex_state = 2},
  [69] = {.lex_state = 9, .external_lex_state = 2},
  [70] = {.lex_state = 9, .external_lex_state = 2},
  [71] = {.lex_state = 9, .external_lex_state = 2},
  [72] = {.lex_state = 9, .external_lex_state = 2},
  [73] = {.lex_state = 9, .external_lex_state = 3},
  [74] = {.lex_state = 9, .external_lex_state = 3},
  [75] = {.lex_state = 9, .external_lex_state = 3},
  [76] = {.lex_state = 9, .external_lex_state = 2},
  [77] = {.lex_state = 9, .external_lex_state = 2},
  [78] = {.lex_state = 9, .external_lex_state = 2},
  [79] = {.lex_state = 9, .external_lex_state = 3},
  [80] = {.lex_state = 2},
  [81] = {.lex_state = 2},
  [82] = {.lex_state = 3, .external_lex_state = 3},
  [83] = {.lex_state = 4},
  [84] = {.lex_state = 4},
  [85] = {.lex_state = 9, .external_lex_state = 2},
  [86] = {.lex_state = 3, .external_lex_state = 2},
  [87] = {.lex_state = 9, .external_lex_state = 5},
  [88] = {.lex_state = 9, .external_lex_state = 2},
  [89] = {.lex_state = 9, .external_lex_state = 2},
  [90] = {.lex_state = 9, .external_lex_state = 5},
  [91] = {.lex_state = 2},
  [92] = {.lex_state = 4},
  [93] = {.lex_state = 3, .external_lex_state = 3},
  [94] = {.lex_state = 9, .external_lex_state = 5},
  [95] = {.lex_state = 9, .external_lex_state = 2},
  [96] = {.lex_state = 9, .external_lex_state = 5},
  [97] = {.lex_state = 9, .external_lex_state = 3},
  [98] = {.lex_state = 9, .external_lex_state = 3},
  [99] = {.lex_state = 9},
  [100] = {.lex_state = 9, .external_lex_state = 5},
  [101] = {.lex_state = 9, .external_lex_state = 3},
  [102] = {.lex_state = 9, .external_lex_state = 5},
  [103] = {.lex_state = 9},
  [104] = {.lex_state = 9},
  [105] = {.lex_state = 9},
  [106] = {.lex_state = 9, .external_lex_state = 4},
  [107] = {.lex_state = 9},
  [108] = {.lex_state = 9},
  [109] = {.lex_state = 9},
  [110] = {.lex_state = 9},
  [111] = {.lex_state = 9},
  [112] = {.lex_state = 9},
  [113] = {.lex_state = 9},
  [114] = {.lex_state = 9},
  [115] = {.lex_state = 9},
  [116] = {.lex_state = 9, .external_lex_state = 4},
  [117] = {.lex_state = 9},
  [118] = {.lex_state = 9},
  [119] = {.lex_state = 9, .external_lex_state = 4},
  [120] = {.lex_state = 9},
  [121] = {.lex_state = 9, .external_lex_state = 6},
  [122] = {.lex_state = 9},
  [123] = {.lex_state = 9},
  [124] = {.lex_state = 9},
  [125] = {.lex_state = 9, .external_lex_state = 4},
  [126] = {.lex_state = 9, .external_lex_state = 4},
  [127] = {.lex_state = 9},
  [128] = {.lex_state = 9, .external_lex_state = 6},
  [129] = {.lex_state = 9},
  [130] = {.lex_state = 9},
  [131] = {.lex_state = 9},
  [132] = {.lex_state = 9},
  [133] = {.lex_state = 9},
  [134] = {.lex_state = 9, .external_lex_state = 4},
  [135] = {.lex_state = 9},
  [136] = {.lex_state = 9, .external_lex_state = 4},
  [137] = {.lex_state = 9, .external_lex_state = 4},
  [138] = {.lex_state = 9, .external_lex_state = 4},
  [139] = {.lex_state = 9},
  [140] = {.lex_state = 9, .external_lex_state = 4},
  [141] = {.lex_state = 9},
  [142] = {.lex_state = 9},
  [143] = {.lex_state = 9, .external_lex_state = 4},
  [144] = {.lex_state = 9, .external_lex_state = 4},
  [145] = {.lex_state = 9},
  [146] = {.lex_state = 9, .external_lex_state = 4},
  [147] = {.lex_state = 9},
};

static const uint16_t ts_parse_table[LARGE_STATE_COUNT][SYMBOL_COUNT] = {
  [0] = {
    [ts_builtin_sym_end] = ACTIONS(1),
    [sym_identifier] = ACTIONS(1),
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
    [anon_sym_when] = ACTIONS(1),
    [anon_sym_on] = ACTIONS(1),
    [anon_sym_label] = ACTIONS(1),
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
    [sym_source_file] = STATE(131),
    [sym_workflow_decl] = STATE(132),
    [aux_sym_source_file_repeat1] = STATE(89),
    [anon_sym_workflow] = ACTIONS(5),
    [sym_comment] = ACTIONS(7),
    [sym__newline] = ACTIONS(9),
  },
};

static const uint16_t ts_small_parse_table[] = {
  [0] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(13), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(11), 30,
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
      anon_sym_when,
      anon_sym_on,
      anon_sym_label,
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
  [42] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(17), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(15), 30,
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
      anon_sym_when,
      anon_sym_on,
      anon_sym_label,
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
  [84] = 17,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(22), 1,
      anon_sym_defaults,
    ACTIONS(25), 1,
      anon_sym_agent,
    ACTIONS(28), 1,
      anon_sym_human,
    ACTIONS(31), 1,
      anon_sym_tool,
    ACTIONS(34), 1,
      anon_sym_subgraph,
    ACTIONS(37), 1,
      anon_sym_conditional,
    ACTIONS(40), 1,
      anon_sym_manager_loop,
    ACTIONS(43), 1,
      anon_sym_parallel,
    ACTIONS(46), 1,
      anon_sym_fan_in,
    ACTIONS(49), 1,
      anon_sym_edges,
    ACTIONS(52), 1,
      anon_sym_stylesheet,
    ACTIONS(55), 1,
      sym__dedent,
    ACTIONS(57), 1,
      sym__newline,
    ACTIONS(19), 4,
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
    STATE(14), 8,
      sym_agent_node,
      sym_human_node,
      sym_tool_node,
      sym_subgraph_node,
      sym_conditional_node,
      sym_manager_loop_node,
      sym_parallel_node,
      sym_fan_in_node,
  [151] = 17,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(62), 1,
      anon_sym_defaults,
    ACTIONS(64), 1,
      anon_sym_agent,
    ACTIONS(66), 1,
      anon_sym_human,
    ACTIONS(68), 1,
      anon_sym_tool,
    ACTIONS(70), 1,
      anon_sym_subgraph,
    ACTIONS(72), 1,
      anon_sym_conditional,
    ACTIONS(74), 1,
      anon_sym_manager_loop,
    ACTIONS(76), 1,
      anon_sym_parallel,
    ACTIONS(78), 1,
      anon_sym_fan_in,
    ACTIONS(80), 1,
      anon_sym_edges,
    ACTIONS(82), 1,
      anon_sym_stylesheet,
    ACTIONS(84), 1,
      sym__dedent,
    ACTIONS(86), 1,
      sym__newline,
    ACTIONS(60), 4,
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
    STATE(14), 8,
      sym_agent_node,
      sym_human_node,
      sym_tool_node,
      sym_subgraph_node,
      sym_conditional_node,
      sym_manager_loop_node,
      sym_parallel_node,
      sym_fan_in_node,
  [218] = 17,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(62), 1,
      anon_sym_defaults,
    ACTIONS(64), 1,
      anon_sym_agent,
    ACTIONS(66), 1,
      anon_sym_human,
    ACTIONS(68), 1,
      anon_sym_tool,
    ACTIONS(70), 1,
      anon_sym_subgraph,
    ACTIONS(72), 1,
      anon_sym_conditional,
    ACTIONS(74), 1,
      anon_sym_manager_loop,
    ACTIONS(76), 1,
      anon_sym_parallel,
    ACTIONS(78), 1,
      anon_sym_fan_in,
    ACTIONS(80), 1,
      anon_sym_edges,
    ACTIONS(82), 1,
      anon_sym_stylesheet,
    ACTIONS(88), 1,
      sym__newline,
    STATE(128), 1,
      sym_workflow_body,
    ACTIONS(60), 4,
      anon_sym_goal,
      anon_sym_start,
      anon_sym_exit,
      anon_sym_requires,
    STATE(5), 6,
      sym_workflow_field,
      sym_defaults_section,
      sym_node_decl,
      sym_edges_section,
      sym_stylesheet_section,
      aux_sym_workflow_body_repeat1,
    STATE(14), 8,
      sym_agent_node,
      sym_human_node,
      sym_tool_node,
      sym_subgraph_node,
      sym_conditional_node,
      sym_manager_loop_node,
      sym_parallel_node,
      sym_fan_in_node,
  [285] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(92), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(90), 22,
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
      anon_sym_when,
      anon_sym_on,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_stylesheet,
      sym_identifier,
  [317] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(96), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(94), 22,
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
      anon_sym_when,
      anon_sym_on,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_stylesheet,
      sym_identifier,
  [349] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(102), 1,
      anon_sym_DOT,
    ACTIONS(100), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(98), 15,
      anon_sym_when,
      anon_sym_on,
      anon_sym_label,
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
  [379] = 7,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(106), 1,
      anon_sym_not,
    STATE(63), 1,
      sym_compare_op,
    ACTIONS(108), 2,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(112), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(110), 5,
      anon_sym_EQ,
      anon_sym_contains,
      anon_sym_startswith,
      anon_sym_endswith,
      anon_sym_in,
    ACTIONS(104), 9,
      anon_sym_when,
      anon_sym_on,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      anon_sym_and,
      sym_identifier,
  [415] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(100), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(98), 15,
      anon_sym_when,
      anon_sym_on,
      anon_sym_label,
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
  [442] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(116), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(114), 15,
      anon_sym_when,
      anon_sym_on,
      anon_sym_label,
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
  [469] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(118), 17,
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
  [492] = 2,
    ACTIONS(7), 1,
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
  [515] = 2,
    ACTIONS(7), 1,
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
  [538] = 2,
    ACTIONS(7), 1,
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
  [561] = 2,
    ACTIONS(7), 1,
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
  [584] = 2,
    ACTIONS(7), 1,
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
  [607] = 2,
    ACTIONS(7), 1,
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
  [630] = 2,
    ACTIONS(7), 1,
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
  [653] = 2,
    ACTIONS(7), 1,
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
  [676] = 2,
    ACTIONS(7), 1,
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
  [699] = 2,
    ACTIONS(7), 1,
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
  [722] = 2,
    ACTIONS(7), 1,
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
  [745] = 2,
    ACTIONS(7), 1,
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
  [768] = 2,
    ACTIONS(7), 1,
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
  [791] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(148), 1,
      anon_sym_and,
    STATE(29), 1,
      aux_sym_and_expr_repeat1,
    ACTIONS(150), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(146), 8,
      anon_sym_when,
      anon_sym_on,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      sym_identifier,
  [815] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(148), 1,
      anon_sym_and,
    STATE(27), 1,
      aux_sym_and_expr_repeat1,
    ACTIONS(154), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(152), 8,
      anon_sym_when,
      anon_sym_on,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      sym_identifier,
  [839] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(158), 1,
      anon_sym_and,
    STATE(29), 1,
      aux_sym_and_expr_repeat1,
    ACTIONS(161), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(156), 8,
      anon_sym_when,
      anon_sym_on,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      sym_identifier,
  [863] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(165), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(163), 9,
      anon_sym_when,
      anon_sym_on,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      anon_sym_and,
      sym_identifier,
  [882] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(169), 1,
      anon_sym_or,
    STATE(34), 1,
      aux_sym_or_expr_repeat1,
    ACTIONS(171), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(167), 7,
      anon_sym_when,
      anon_sym_on,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [905] = 7,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(173), 1,
      sym_identifier,
    ACTIONS(175), 1,
      anon_sym_when,
    ACTIONS(177), 1,
      anon_sym_on,
    ACTIONS(181), 2,
      sym__dedent,
      sym__newline,
    STATE(33), 2,
      sym_edge_attr,
      aux_sym_edge_entry_repeat1,
    ACTIONS(179), 4,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
  [932] = 7,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(183), 1,
      sym_identifier,
    ACTIONS(185), 1,
      anon_sym_when,
    ACTIONS(188), 1,
      anon_sym_on,
    ACTIONS(194), 2,
      sym__dedent,
      sym__newline,
    STATE(33), 2,
      sym_edge_attr,
      aux_sym_edge_entry_repeat1,
    ACTIONS(191), 4,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
  [959] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(169), 1,
      anon_sym_or,
    STATE(36), 1,
      aux_sym_or_expr_repeat1,
    ACTIONS(198), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(196), 7,
      anon_sym_when,
      anon_sym_on,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [982] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(202), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(200), 9,
      anon_sym_when,
      anon_sym_on,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      anon_sym_and,
      sym_identifier,
  [1001] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(206), 1,
      anon_sym_or,
    STATE(36), 1,
      aux_sym_or_expr_repeat1,
    ACTIONS(209), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(204), 7,
      anon_sym_when,
      anon_sym_on,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [1024] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(161), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(156), 9,
      anon_sym_when,
      anon_sym_on,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      anon_sym_and,
      sym_identifier,
  [1043] = 7,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(175), 1,
      anon_sym_when,
    ACTIONS(177), 1,
      anon_sym_on,
    ACTIONS(211), 1,
      sym_identifier,
    ACTIONS(213), 2,
      sym__dedent,
      sym__newline,
    STATE(32), 2,
      sym_edge_attr,
      aux_sym_edge_entry_repeat1,
    ACTIONS(179), 4,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
  [1070] = 10,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(215), 1,
      sym_identifier,
    ACTIONS(217), 1,
      anon_sym_DQUOTE,
    ACTIONS(219), 1,
      anon_sym_SQUOTE,
    STATE(10), 1,
      sym_operand,
    STATE(28), 1,
      sym_compare_expr,
    STATE(31), 1,
      sym_and_expr,
    STATE(41), 1,
      sym_or_expr,
    STATE(43), 1,
      sym_condition,
    STATE(11), 2,
      sym_variable,
      sym_string,
  [1102] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(209), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(204), 8,
      anon_sym_when,
      anon_sym_on,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      sym_identifier,
  [1120] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(223), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(221), 7,
      anon_sym_when,
      anon_sym_on,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [1137] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(227), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(225), 7,
      anon_sym_when,
      anon_sym_on,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [1154] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(231), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(229), 7,
      anon_sym_when,
      anon_sym_on,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [1171] = 8,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(235), 1,
      anon_sym_DOT,
    ACTIONS(237), 1,
      anon_sym_POUND,
    ACTIONS(239), 1,
      sym__dedent,
    ACTIONS(241), 1,
      sym__newline,
    STATE(119), 1,
      sym_selector,
    ACTIONS(233), 2,
      anon_sym_STAR,
      sym_identifier,
    STATE(45), 2,
      sym_stylesheet_rule,
      aux_sym_stylesheet_section_repeat1,
  [1198] = 8,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(246), 1,
      anon_sym_DOT,
    ACTIONS(249), 1,
      anon_sym_POUND,
    ACTIONS(252), 1,
      sym__dedent,
    ACTIONS(254), 1,
      sym__newline,
    STATE(119), 1,
      sym_selector,
    ACTIONS(243), 2,
      anon_sym_STAR,
      sym_identifier,
    STATE(45), 2,
      sym_stylesheet_rule,
      aux_sym_stylesheet_section_repeat1,
  [1225] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(110), 1,
      anon_sym_EQ,
    STATE(55), 1,
      sym_compare_op,
    ACTIONS(108), 6,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
      anon_sym_contains,
      anon_sym_startswith,
      anon_sym_endswith,
      anon_sym_in,
  [1243] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(235), 1,
      anon_sym_DOT,
    ACTIONS(237), 1,
      anon_sym_POUND,
    ACTIONS(257), 1,
      sym__newline,
    STATE(119), 1,
      sym_selector,
    ACTIONS(233), 2,
      anon_sym_STAR,
      sym_identifier,
    STATE(44), 2,
      sym_stylesheet_rule,
      aux_sym_stylesheet_section_repeat1,
  [1267] = 8,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(215), 1,
      sym_identifier,
    ACTIONS(217), 1,
      anon_sym_DQUOTE,
    ACTIONS(219), 1,
      anon_sym_SQUOTE,
    STATE(10), 1,
      sym_operand,
    STATE(28), 1,
      sym_compare_expr,
    STATE(40), 1,
      sym_and_expr,
    STATE(11), 2,
      sym_variable,
      sym_string,
  [1293] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(259), 1,
      sym_raw_inline,
    ACTIONS(261), 1,
      anon_sym_DQUOTE,
    ACTIONS(263), 1,
      anon_sym_SQUOTE,
    ACTIONS(265), 1,
      sym__indent,
    STATE(98), 1,
      sym_field_value,
    STATE(8), 2,
      sym_multiline_block,
      sym_string,
  [1316] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(259), 1,
      sym_raw_inline,
    ACTIONS(261), 1,
      anon_sym_DQUOTE,
    ACTIONS(263), 1,
      anon_sym_SQUOTE,
    ACTIONS(265), 1,
      sym__indent,
    STATE(101), 1,
      sym_field_value,
    STATE(8), 2,
      sym_multiline_block,
      sym_string,
  [1339] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(259), 1,
      sym_raw_inline,
    ACTIONS(261), 1,
      anon_sym_DQUOTE,
    ACTIONS(263), 1,
      anon_sym_SQUOTE,
    ACTIONS(265), 1,
      sym__indent,
    STATE(18), 1,
      sym_field_value,
    STATE(8), 2,
      sym_multiline_block,
      sym_string,
  [1362] = 7,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(215), 1,
      sym_identifier,
    ACTIONS(217), 1,
      anon_sym_DQUOTE,
    ACTIONS(219), 1,
      anon_sym_SQUOTE,
    STATE(10), 1,
      sym_operand,
    STATE(37), 1,
      sym_compare_expr,
    STATE(11), 2,
      sym_variable,
      sym_string,
  [1385] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(259), 1,
      sym_raw_inline,
    ACTIONS(261), 1,
      anon_sym_DQUOTE,
    ACTIONS(263), 1,
      anon_sym_SQUOTE,
    ACTIONS(265), 1,
      sym__indent,
    STATE(42), 1,
      sym_field_value,
    STATE(8), 2,
      sym_multiline_block,
      sym_string,
  [1408] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(259), 1,
      sym_raw_inline,
    ACTIONS(261), 1,
      anon_sym_DQUOTE,
    ACTIONS(263), 1,
      anon_sym_SQUOTE,
    ACTIONS(265), 1,
      sym__indent,
    STATE(97), 1,
      sym_field_value,
    STATE(8), 2,
      sym_multiline_block,
      sym_string,
  [1431] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(215), 1,
      sym_identifier,
    ACTIONS(217), 1,
      anon_sym_DQUOTE,
    ACTIONS(219), 1,
      anon_sym_SQUOTE,
    STATE(35), 1,
      sym_operand,
    STATE(11), 2,
      sym_variable,
      sym_string,
  [1451] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(269), 1,
      anon_sym_POUND,
    ACTIONS(267), 5,
      sym__dedent,
      sym__newline,
      anon_sym_DOT,
      anon_sym_STAR,
      sym_identifier,
  [1465] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(271), 1,
      sym_identifier,
    ACTIONS(274), 1,
      sym__dedent,
    ACTIONS(276), 1,
      sym__newline,
    STATE(133), 1,
      sym_field_name,
    STATE(57), 2,
      sym_defaults_field,
      aux_sym_defaults_section_repeat1,
  [1485] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(279), 1,
      sym_identifier,
    ACTIONS(281), 1,
      sym__dedent,
    ACTIONS(283), 1,
      sym__newline,
    STATE(130), 1,
      sym_field_name,
    STATE(64), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1505] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(279), 1,
      sym_identifier,
    ACTIONS(283), 1,
      sym__newline,
    ACTIONS(285), 1,
      sym__dedent,
    STATE(130), 1,
      sym_field_name,
    STATE(64), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1525] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(279), 1,
      sym_identifier,
    ACTIONS(283), 1,
      sym__newline,
    ACTIONS(287), 1,
      sym__dedent,
    STATE(130), 1,
      sym_field_name,
    STATE(64), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1545] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(279), 1,
      sym_identifier,
    ACTIONS(283), 1,
      sym__newline,
    ACTIONS(289), 1,
      sym__dedent,
    STATE(130), 1,
      sym_field_name,
    STATE(64), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1565] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(279), 1,
      sym_identifier,
    ACTIONS(283), 1,
      sym__newline,
    ACTIONS(291), 1,
      sym__dedent,
    STATE(130), 1,
      sym_field_name,
    STATE(64), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1585] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(215), 1,
      sym_identifier,
    ACTIONS(217), 1,
      anon_sym_DQUOTE,
    ACTIONS(219), 1,
      anon_sym_SQUOTE,
    STATE(30), 1,
      sym_operand,
    STATE(11), 2,
      sym_variable,
      sym_string,
  [1605] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(293), 1,
      sym_identifier,
    ACTIONS(296), 1,
      sym__dedent,
    ACTIONS(298), 1,
      sym__newline,
    STATE(130), 1,
      sym_field_name,
    STATE(64), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1625] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(279), 1,
      sym_identifier,
    ACTIONS(283), 1,
      sym__newline,
    ACTIONS(301), 1,
      sym__dedent,
    STATE(130), 1,
      sym_field_name,
    STATE(64), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1645] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(279), 1,
      sym_identifier,
    ACTIONS(283), 1,
      sym__newline,
    ACTIONS(303), 1,
      sym__dedent,
    STATE(130), 1,
      sym_field_name,
    STATE(64), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1665] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(279), 1,
      sym_identifier,
    ACTIONS(305), 1,
      sym__dedent,
    ACTIONS(307), 1,
      sym__newline,
    STATE(133), 1,
      sym_field_name,
    STATE(57), 2,
      sym_defaults_field,
      aux_sym_defaults_section_repeat1,
  [1685] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(279), 1,
      sym_identifier,
    ACTIONS(309), 1,
      sym__newline,
    STATE(130), 1,
      sym_field_name,
    STATE(65), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1702] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(279), 1,
      sym_identifier,
    ACTIONS(311), 1,
      sym__newline,
    STATE(130), 1,
      sym_field_name,
    STATE(60), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1719] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(279), 1,
      sym_identifier,
    ACTIONS(313), 1,
      sym__newline,
    STATE(130), 1,
      sym_field_name,
    STATE(59), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1736] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(279), 1,
      sym_identifier,
    ACTIONS(315), 1,
      sym__newline,
    STATE(130), 1,
      sym_field_name,
    STATE(61), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1753] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(279), 1,
      sym_identifier,
    ACTIONS(317), 1,
      sym__newline,
    STATE(133), 1,
      sym_field_name,
    STATE(67), 2,
      sym_defaults_field,
      aux_sym_defaults_section_repeat1,
  [1770] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(319), 1,
      sym_identifier,
    ACTIONS(322), 1,
      sym__dedent,
    ACTIONS(324), 1,
      sym__newline,
    STATE(73), 2,
      sym_edge_entry,
      aux_sym_edges_section_repeat1,
  [1787] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(327), 1,
      sym_identifier,
    ACTIONS(330), 1,
      sym__dedent,
    ACTIONS(332), 1,
      sym__newline,
    STATE(74), 1,
      aux_sym_stylesheet_rule_repeat1,
    STATE(114), 1,
      sym_field_name,
  [1806] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(279), 1,
      sym_identifier,
    ACTIONS(335), 1,
      sym__dedent,
    ACTIONS(337), 1,
      sym__newline,
    STATE(74), 1,
      aux_sym_stylesheet_rule_repeat1,
    STATE(114), 1,
      sym_field_name,
  [1825] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(279), 1,
      sym_identifier,
    ACTIONS(339), 1,
      sym__newline,
    STATE(130), 1,
      sym_field_name,
    STATE(58), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1842] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(279), 1,
      sym_identifier,
    ACTIONS(341), 1,
      sym__newline,
    STATE(130), 1,
      sym_field_name,
    STATE(62), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1859] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(279), 1,
      sym_identifier,
    ACTIONS(343), 1,
      sym__newline,
    STATE(130), 1,
      sym_field_name,
    STATE(66), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1876] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(345), 1,
      sym_identifier,
    ACTIONS(347), 1,
      sym__dedent,
    ACTIONS(349), 1,
      sym__newline,
    STATE(73), 2,
      sym_edge_entry,
      aux_sym_edges_section_repeat1,
  [1893] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(351), 1,
      anon_sym_DQUOTE,
    STATE(80), 1,
      aux_sym_string_repeat1,
    ACTIONS(353), 2,
      aux_sym_string_token1,
      aux_sym_string_token2,
  [1907] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(356), 1,
      anon_sym_DQUOTE,
    STATE(91), 1,
      aux_sym_string_repeat1,
    ACTIONS(358), 2,
      aux_sym_string_token1,
      aux_sym_string_token2,
  [1921] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(363), 1,
      sym__dedent,
    STATE(82), 1,
      aux_sym_block_content_repeat1,
    ACTIONS(360), 2,
      sym__newline,
      sym_block_line,
  [1935] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(365), 1,
      anon_sym_SQUOTE,
    STATE(83), 1,
      aux_sym_string_repeat2,
    ACTIONS(367), 2,
      aux_sym_string_token3,
      anon_sym_SQUOTE_SQUOTE,
  [1949] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(356), 1,
      anon_sym_SQUOTE,
    STATE(92), 1,
      aux_sym_string_repeat2,
    ACTIONS(370), 2,
      aux_sym_string_token3,
      anon_sym_SQUOTE_SQUOTE,
  [1963] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(345), 1,
      sym_identifier,
    ACTIONS(372), 1,
      sym__newline,
    STATE(79), 2,
      sym_edge_entry,
      aux_sym_edges_section_repeat1,
  [1977] = 4,
    ACTIONS(3), 1,
      sym_comment,
    STATE(93), 1,
      aux_sym_block_content_repeat1,
    STATE(121), 1,
      sym_block_content,
    ACTIONS(374), 2,
      sym__newline,
      sym_block_line,
  [1991] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(376), 1,
      anon_sym_COMMA,
    STATE(90), 1,
      aux_sym_identifier_list_repeat1,
    ACTIONS(378), 2,
      sym__indent,
      sym__newline,
  [2005] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(279), 1,
      sym_identifier,
    ACTIONS(380), 1,
      sym__newline,
    STATE(75), 1,
      aux_sym_stylesheet_rule_repeat1,
    STATE(114), 1,
      sym_field_name,
  [2021] = 5,
    ACTIONS(5), 1,
      anon_sym_workflow,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(382), 1,
      sym__newline,
    STATE(95), 1,
      aux_sym_source_file_repeat1,
    STATE(139), 1,
      sym_workflow_decl,
  [2037] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(384), 1,
      anon_sym_COMMA,
    STATE(90), 1,
      aux_sym_identifier_list_repeat1,
    ACTIONS(387), 2,
      sym__indent,
      sym__newline,
  [2051] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(389), 1,
      anon_sym_DQUOTE,
    STATE(80), 1,
      aux_sym_string_repeat1,
    ACTIONS(391), 2,
      aux_sym_string_token1,
      aux_sym_string_token2,
  [2065] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(389), 1,
      anon_sym_SQUOTE,
    STATE(83), 1,
      aux_sym_string_repeat2,
    ACTIONS(393), 2,
      aux_sym_string_token3,
      anon_sym_SQUOTE_SQUOTE,
  [2079] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(397), 1,
      sym__dedent,
    STATE(82), 1,
      aux_sym_block_content_repeat1,
    ACTIONS(395), 2,
      sym__newline,
      sym_block_line,
  [2093] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(376), 1,
      anon_sym_COMMA,
    STATE(87), 1,
      aux_sym_identifier_list_repeat1,
    ACTIONS(399), 2,
      sym__indent,
      sym__newline,
  [2107] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(401), 1,
      anon_sym_workflow,
    ACTIONS(403), 1,
      sym__newline,
    STATE(95), 1,
      aux_sym_source_file_repeat1,
  [2120] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(406), 1,
      sym__indent,
    ACTIONS(408), 1,
      sym__newline,
    STATE(24), 1,
      sym_node_attr_block,
  [2133] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(410), 3,
      sym__dedent,
      sym__newline,
      sym_identifier,
  [2142] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(412), 3,
      sym__dedent,
      sym__newline,
      sym_identifier,
  [2151] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(414), 3,
      sym_identifier,
      anon_sym_DQUOTE,
      anon_sym_SQUOTE,
  [2160] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(387), 3,
      sym__indent,
      sym__newline,
      anon_sym_COMMA,
  [2169] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(416), 3,
      sym__dedent,
      sym__newline,
      sym_identifier,
  [2178] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(406), 1,
      sym__indent,
    ACTIONS(418), 1,
      sym__newline,
    STATE(23), 1,
      sym_node_attr_block,
  [2191] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(420), 1,
      sym_identifier,
    STATE(102), 1,
      sym_identifier_list,
  [2201] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(420), 1,
      sym_identifier,
    STATE(96), 1,
      sym_identifier_list,
  [2211] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(422), 1,
      sym_identifier,
  [2218] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(424), 1,
      sym__indent,
  [2225] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(426), 1,
      anon_sym_DASH_GT,
  [2232] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(428), 1,
      anon_sym_LT_DASH,
  [2239] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(430), 1,
      sym_identifier,
  [2246] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(432), 1,
      sym_identifier,
  [2253] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(434), 1,
      sym_identifier,
  [2260] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(436), 1,
      anon_sym_COLON,
  [2267] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(438), 1,
      sym_identifier,
  [2274] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(440), 1,
      anon_sym_COLON,
  [2281] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(442), 1,
      sym_identifier,
  [2288] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(444), 1,
      sym__indent,
  [2295] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(446), 1,
      sym_identifier,
  [2302] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(448), 1,
      anon_sym_DASH_GT,
  [2309] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(450), 1,
      sym__indent,
  [2316] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(452), 1,
      sym_identifier,
  [2323] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(454), 1,
      sym__dedent,
  [2330] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(456), 1,
      ts_builtin_sym_end,
  [2337] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(458), 1,
      sym_identifier,
  [2344] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(460), 1,
      sym_identifier,
  [2351] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(462), 1,
      sym__indent,
  [2358] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(464), 1,
      sym__indent,
  [2365] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(466), 1,
      anon_sym_COLON,
  [2372] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(468), 1,
      sym__dedent,
  [2379] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(470), 1,
      sym_identifier,
  [2386] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(472), 1,
      anon_sym_COLON,
  [2393] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(474), 1,
      ts_builtin_sym_end,
  [2400] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(476), 1,
      ts_builtin_sym_end,
  [2407] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(478), 1,
      anon_sym_COLON,
  [2414] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(480), 1,
      sym__indent,
  [2421] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(482), 1,
      anon_sym_COLON,
  [2428] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(484), 1,
      sym__indent,
  [2435] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(486), 1,
      sym__indent,
  [2442] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(488), 1,
      sym__indent,
  [2449] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(490), 1,
      ts_builtin_sym_end,
  [2456] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(492), 1,
      sym__indent,
  [2463] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(494), 1,
      sym_identifier,
  [2470] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(496), 1,
      sym_identifier,
  [2477] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(498), 1,
      sym__indent,
  [2484] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(500), 1,
      sym__indent,
  [2491] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(502), 1,
      sym_identifier,
  [2498] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(504), 1,
      sym__indent,
  [2505] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(506), 1,
      anon_sym_COLON,
};

static const uint32_t ts_small_parse_table_map[] = {
  [SMALL_STATE(2)] = 0,
  [SMALL_STATE(3)] = 42,
  [SMALL_STATE(4)] = 84,
  [SMALL_STATE(5)] = 151,
  [SMALL_STATE(6)] = 218,
  [SMALL_STATE(7)] = 285,
  [SMALL_STATE(8)] = 317,
  [SMALL_STATE(9)] = 349,
  [SMALL_STATE(10)] = 379,
  [SMALL_STATE(11)] = 415,
  [SMALL_STATE(12)] = 442,
  [SMALL_STATE(13)] = 469,
  [SMALL_STATE(14)] = 492,
  [SMALL_STATE(15)] = 515,
  [SMALL_STATE(16)] = 538,
  [SMALL_STATE(17)] = 561,
  [SMALL_STATE(18)] = 584,
  [SMALL_STATE(19)] = 607,
  [SMALL_STATE(20)] = 630,
  [SMALL_STATE(21)] = 653,
  [SMALL_STATE(22)] = 676,
  [SMALL_STATE(23)] = 699,
  [SMALL_STATE(24)] = 722,
  [SMALL_STATE(25)] = 745,
  [SMALL_STATE(26)] = 768,
  [SMALL_STATE(27)] = 791,
  [SMALL_STATE(28)] = 815,
  [SMALL_STATE(29)] = 839,
  [SMALL_STATE(30)] = 863,
  [SMALL_STATE(31)] = 882,
  [SMALL_STATE(32)] = 905,
  [SMALL_STATE(33)] = 932,
  [SMALL_STATE(34)] = 959,
  [SMALL_STATE(35)] = 982,
  [SMALL_STATE(36)] = 1001,
  [SMALL_STATE(37)] = 1024,
  [SMALL_STATE(38)] = 1043,
  [SMALL_STATE(39)] = 1070,
  [SMALL_STATE(40)] = 1102,
  [SMALL_STATE(41)] = 1120,
  [SMALL_STATE(42)] = 1137,
  [SMALL_STATE(43)] = 1154,
  [SMALL_STATE(44)] = 1171,
  [SMALL_STATE(45)] = 1198,
  [SMALL_STATE(46)] = 1225,
  [SMALL_STATE(47)] = 1243,
  [SMALL_STATE(48)] = 1267,
  [SMALL_STATE(49)] = 1293,
  [SMALL_STATE(50)] = 1316,
  [SMALL_STATE(51)] = 1339,
  [SMALL_STATE(52)] = 1362,
  [SMALL_STATE(53)] = 1385,
  [SMALL_STATE(54)] = 1408,
  [SMALL_STATE(55)] = 1431,
  [SMALL_STATE(56)] = 1451,
  [SMALL_STATE(57)] = 1465,
  [SMALL_STATE(58)] = 1485,
  [SMALL_STATE(59)] = 1505,
  [SMALL_STATE(60)] = 1525,
  [SMALL_STATE(61)] = 1545,
  [SMALL_STATE(62)] = 1565,
  [SMALL_STATE(63)] = 1585,
  [SMALL_STATE(64)] = 1605,
  [SMALL_STATE(65)] = 1625,
  [SMALL_STATE(66)] = 1645,
  [SMALL_STATE(67)] = 1665,
  [SMALL_STATE(68)] = 1685,
  [SMALL_STATE(69)] = 1702,
  [SMALL_STATE(70)] = 1719,
  [SMALL_STATE(71)] = 1736,
  [SMALL_STATE(72)] = 1753,
  [SMALL_STATE(73)] = 1770,
  [SMALL_STATE(74)] = 1787,
  [SMALL_STATE(75)] = 1806,
  [SMALL_STATE(76)] = 1825,
  [SMALL_STATE(77)] = 1842,
  [SMALL_STATE(78)] = 1859,
  [SMALL_STATE(79)] = 1876,
  [SMALL_STATE(80)] = 1893,
  [SMALL_STATE(81)] = 1907,
  [SMALL_STATE(82)] = 1921,
  [SMALL_STATE(83)] = 1935,
  [SMALL_STATE(84)] = 1949,
  [SMALL_STATE(85)] = 1963,
  [SMALL_STATE(86)] = 1977,
  [SMALL_STATE(87)] = 1991,
  [SMALL_STATE(88)] = 2005,
  [SMALL_STATE(89)] = 2021,
  [SMALL_STATE(90)] = 2037,
  [SMALL_STATE(91)] = 2051,
  [SMALL_STATE(92)] = 2065,
  [SMALL_STATE(93)] = 2079,
  [SMALL_STATE(94)] = 2093,
  [SMALL_STATE(95)] = 2107,
  [SMALL_STATE(96)] = 2120,
  [SMALL_STATE(97)] = 2133,
  [SMALL_STATE(98)] = 2142,
  [SMALL_STATE(99)] = 2151,
  [SMALL_STATE(100)] = 2160,
  [SMALL_STATE(101)] = 2169,
  [SMALL_STATE(102)] = 2178,
  [SMALL_STATE(103)] = 2191,
  [SMALL_STATE(104)] = 2201,
  [SMALL_STATE(105)] = 2211,
  [SMALL_STATE(106)] = 2218,
  [SMALL_STATE(107)] = 2225,
  [SMALL_STATE(108)] = 2232,
  [SMALL_STATE(109)] = 2239,
  [SMALL_STATE(110)] = 2246,
  [SMALL_STATE(111)] = 2253,
  [SMALL_STATE(112)] = 2260,
  [SMALL_STATE(113)] = 2267,
  [SMALL_STATE(114)] = 2274,
  [SMALL_STATE(115)] = 2281,
  [SMALL_STATE(116)] = 2288,
  [SMALL_STATE(117)] = 2295,
  [SMALL_STATE(118)] = 2302,
  [SMALL_STATE(119)] = 2309,
  [SMALL_STATE(120)] = 2316,
  [SMALL_STATE(121)] = 2323,
  [SMALL_STATE(122)] = 2330,
  [SMALL_STATE(123)] = 2337,
  [SMALL_STATE(124)] = 2344,
  [SMALL_STATE(125)] = 2351,
  [SMALL_STATE(126)] = 2358,
  [SMALL_STATE(127)] = 2365,
  [SMALL_STATE(128)] = 2372,
  [SMALL_STATE(129)] = 2379,
  [SMALL_STATE(130)] = 2386,
  [SMALL_STATE(131)] = 2393,
  [SMALL_STATE(132)] = 2400,
  [SMALL_STATE(133)] = 2407,
  [SMALL_STATE(134)] = 2414,
  [SMALL_STATE(135)] = 2421,
  [SMALL_STATE(136)] = 2428,
  [SMALL_STATE(137)] = 2435,
  [SMALL_STATE(138)] = 2442,
  [SMALL_STATE(139)] = 2449,
  [SMALL_STATE(140)] = 2456,
  [SMALL_STATE(141)] = 2463,
  [SMALL_STATE(142)] = 2470,
  [SMALL_STATE(143)] = 2477,
  [SMALL_STATE(144)] = 2484,
  [SMALL_STATE(145)] = 2491,
  [SMALL_STATE(146)] = 2498,
  [SMALL_STATE(147)] = 2505,
};

static const TSParseActionEntry ts_parse_actions[] = {
  [0] = {.entry = {.count = 0, .reusable = false}},
  [1] = {.entry = {.count = 1, .reusable = false}}, RECOVER(),
  [3] = {.entry = {.count = 1, .reusable = false}}, SHIFT_EXTRA(),
  [5] = {.entry = {.count = 1, .reusable = true}}, SHIFT(124),
  [7] = {.entry = {.count = 1, .reusable = true}}, SHIFT_EXTRA(),
  [9] = {.entry = {.count = 1, .reusable = true}}, SHIFT(89),
  [11] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_string, 2, 0, 0),
  [13] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_string, 2, 0, 0),
  [15] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_string, 3, 0, 0),
  [17] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_string, 3, 0, 0),
  [19] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(135),
  [22] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(140),
  [25] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(141),
  [28] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(109),
  [31] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(110),
  [34] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(113),
  [37] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(115),
  [40] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(105),
  [43] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(120),
  [46] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(123),
  [49] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(126),
  [52] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(127),
  [55] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0),
  [57] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(4),
  [60] = {.entry = {.count = 1, .reusable = true}}, SHIFT(135),
  [62] = {.entry = {.count = 1, .reusable = true}}, SHIFT(140),
  [64] = {.entry = {.count = 1, .reusable = true}}, SHIFT(141),
  [66] = {.entry = {.count = 1, .reusable = true}}, SHIFT(109),
  [68] = {.entry = {.count = 1, .reusable = true}}, SHIFT(110),
  [70] = {.entry = {.count = 1, .reusable = true}}, SHIFT(113),
  [72] = {.entry = {.count = 1, .reusable = true}}, SHIFT(115),
  [74] = {.entry = {.count = 1, .reusable = true}}, SHIFT(105),
  [76] = {.entry = {.count = 1, .reusable = true}}, SHIFT(120),
  [78] = {.entry = {.count = 1, .reusable = true}}, SHIFT(123),
  [80] = {.entry = {.count = 1, .reusable = true}}, SHIFT(126),
  [82] = {.entry = {.count = 1, .reusable = true}}, SHIFT(127),
  [84] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_workflow_body, 1, 0, 0),
  [86] = {.entry = {.count = 1, .reusable = true}}, SHIFT(4),
  [88] = {.entry = {.count = 1, .reusable = true}}, SHIFT(5),
  [90] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_multiline_block, 3, 0, 0),
  [92] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_multiline_block, 3, 0, 0),
  [94] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_field_value, 1, 0, 0),
  [96] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_field_value, 1, 0, 0),
  [98] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_operand, 1, 0, 0),
  [100] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_operand, 1, 0, 0),
  [102] = {.entry = {.count = 1, .reusable = true}}, SHIFT(129),
  [104] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_compare_expr, 1, 0, 0),
  [106] = {.entry = {.count = 1, .reusable = false}}, SHIFT(46),
  [108] = {.entry = {.count = 1, .reusable = true}}, SHIFT(99),
  [110] = {.entry = {.count = 1, .reusable = false}}, SHIFT(99),
  [112] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_expr, 1, 0, 0),
  [114] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_variable, 3, 0, 0),
  [116] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_variable, 3, 0, 0),
  [118] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_conditional_node, 5, 0, 0),
  [120] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_node_decl, 1, 0, 0),
  [122] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_defaults_section, 4, 0, 0),
  [124] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edges_section, 4, 0, 0),
  [126] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_subgraph_node, 5, 0, 0),
  [128] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_workflow_field, 3, 0, 0),
  [130] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_agent_node, 5, 0, 0),
  [132] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_human_node, 5, 0, 0),
  [134] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_tool_node, 5, 0, 0),
  [136] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_manager_loop_node, 5, 0, 0),
  [138] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_parallel_node, 5, 0, 0),
  [140] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_fan_in_node, 5, 0, 0),
  [142] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_node_attr_block, 3, 0, 0),
  [144] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_stylesheet_section, 5, 0, 0),
  [146] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_and_expr, 2, 0, 0),
  [148] = {.entry = {.count = 1, .reusable = false}}, SHIFT(52),
  [150] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_and_expr, 2, 0, 0),
  [152] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_and_expr, 1, 0, 0),
  [154] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_and_expr, 1, 0, 0),
  [156] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_and_expr_repeat1, 2, 0, 0),
  [158] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_and_expr_repeat1, 2, 0, 0), SHIFT_REPEAT(52),
  [161] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_and_expr_repeat1, 2, 0, 0),
  [163] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_compare_expr, 3, 0, 0),
  [165] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_expr, 3, 0, 0),
  [167] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_or_expr, 1, 0, 0),
  [169] = {.entry = {.count = 1, .reusable = false}}, SHIFT(48),
  [171] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_or_expr, 1, 0, 0),
  [173] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_entry, 4, 0, 0),
  [175] = {.entry = {.count = 1, .reusable = false}}, SHIFT(39),
  [177] = {.entry = {.count = 1, .reusable = false}}, SHIFT(111),
  [179] = {.entry = {.count = 1, .reusable = false}}, SHIFT(112),
  [181] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_entry, 4, 0, 0),
  [183] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0),
  [185] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0), SHIFT_REPEAT(39),
  [188] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0), SHIFT_REPEAT(111),
  [191] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0), SHIFT_REPEAT(112),
  [194] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0),
  [196] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_or_expr, 2, 0, 0),
  [198] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_or_expr, 2, 0, 0),
  [200] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_compare_expr, 4, 0, 0),
  [202] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_expr, 4, 0, 0),
  [204] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_or_expr_repeat1, 2, 0, 0),
  [206] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_or_expr_repeat1, 2, 0, 0), SHIFT_REPEAT(48),
  [209] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_or_expr_repeat1, 2, 0, 0),
  [211] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_entry, 3, 0, 0),
  [213] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_entry, 3, 0, 0),
  [215] = {.entry = {.count = 1, .reusable = true}}, SHIFT(9),
  [217] = {.entry = {.count = 1, .reusable = true}}, SHIFT(81),
  [219] = {.entry = {.count = 1, .reusable = true}}, SHIFT(84),
  [221] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_condition, 1, 0, 0),
  [223] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_condition, 1, 0, 0),
  [225] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_attr, 3, 0, 0),
  [227] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_attr, 3, 0, 0),
  [229] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_attr, 2, 0, 0),
  [231] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_attr, 2, 0, 0),
  [233] = {.entry = {.count = 1, .reusable = true}}, SHIFT(146),
  [235] = {.entry = {.count = 1, .reusable = true}}, SHIFT(145),
  [237] = {.entry = {.count = 1, .reusable = false}}, SHIFT(145),
  [239] = {.entry = {.count = 1, .reusable = true}}, SHIFT(26),
  [241] = {.entry = {.count = 1, .reusable = true}}, SHIFT(45),
  [243] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(146),
  [246] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(145),
  [249] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(145),
  [252] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0),
  [254] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(45),
  [257] = {.entry = {.count = 1, .reusable = true}}, SHIFT(44),
  [259] = {.entry = {.count = 1, .reusable = false}}, SHIFT(8),
  [261] = {.entry = {.count = 1, .reusable = false}}, SHIFT(81),
  [263] = {.entry = {.count = 1, .reusable = false}}, SHIFT(84),
  [265] = {.entry = {.count = 1, .reusable = true}}, SHIFT(86),
  [267] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_stylesheet_rule, 4, 0, 0),
  [269] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_stylesheet_rule, 4, 0, 0),
  [271] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_defaults_section_repeat1, 2, 0, 0), SHIFT_REPEAT(147),
  [274] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_defaults_section_repeat1, 2, 0, 0),
  [276] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_defaults_section_repeat1, 2, 0, 0), SHIFT_REPEAT(57),
  [279] = {.entry = {.count = 1, .reusable = true}}, SHIFT(147),
  [281] = {.entry = {.count = 1, .reusable = true}}, SHIFT(19),
  [283] = {.entry = {.count = 1, .reusable = true}}, SHIFT(64),
  [285] = {.entry = {.count = 1, .reusable = true}}, SHIFT(20),
  [287] = {.entry = {.count = 1, .reusable = true}}, SHIFT(21),
  [289] = {.entry = {.count = 1, .reusable = true}}, SHIFT(25),
  [291] = {.entry = {.count = 1, .reusable = true}}, SHIFT(17),
  [293] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_agent_node_repeat1, 2, 0, 0), SHIFT_REPEAT(147),
  [296] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_agent_node_repeat1, 2, 0, 0),
  [298] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_agent_node_repeat1, 2, 0, 0), SHIFT_REPEAT(64),
  [301] = {.entry = {.count = 1, .reusable = true}}, SHIFT(22),
  [303] = {.entry = {.count = 1, .reusable = true}}, SHIFT(13),
  [305] = {.entry = {.count = 1, .reusable = true}}, SHIFT(15),
  [307] = {.entry = {.count = 1, .reusable = true}}, SHIFT(57),
  [309] = {.entry = {.count = 1, .reusable = true}}, SHIFT(65),
  [311] = {.entry = {.count = 1, .reusable = true}}, SHIFT(60),
  [313] = {.entry = {.count = 1, .reusable = true}}, SHIFT(59),
  [315] = {.entry = {.count = 1, .reusable = true}}, SHIFT(61),
  [317] = {.entry = {.count = 1, .reusable = true}}, SHIFT(67),
  [319] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_edges_section_repeat1, 2, 0, 0), SHIFT_REPEAT(118),
  [322] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_edges_section_repeat1, 2, 0, 0),
  [324] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_edges_section_repeat1, 2, 0, 0), SHIFT_REPEAT(73),
  [327] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 2, 0, 0), SHIFT_REPEAT(147),
  [330] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 2, 0, 0),
  [332] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 2, 0, 0), SHIFT_REPEAT(74),
  [335] = {.entry = {.count = 1, .reusable = true}}, SHIFT(56),
  [337] = {.entry = {.count = 1, .reusable = true}}, SHIFT(74),
  [339] = {.entry = {.count = 1, .reusable = true}}, SHIFT(58),
  [341] = {.entry = {.count = 1, .reusable = true}}, SHIFT(62),
  [343] = {.entry = {.count = 1, .reusable = true}}, SHIFT(66),
  [345] = {.entry = {.count = 1, .reusable = true}}, SHIFT(118),
  [347] = {.entry = {.count = 1, .reusable = true}}, SHIFT(16),
  [349] = {.entry = {.count = 1, .reusable = true}}, SHIFT(73),
  [351] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_string_repeat1, 2, 0, 0),
  [353] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_string_repeat1, 2, 0, 0), SHIFT_REPEAT(80),
  [356] = {.entry = {.count = 1, .reusable = false}}, SHIFT(2),
  [358] = {.entry = {.count = 1, .reusable = false}}, SHIFT(91),
  [360] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_block_content_repeat1, 2, 0, 0), SHIFT_REPEAT(82),
  [363] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_block_content_repeat1, 2, 0, 0),
  [365] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_string_repeat2, 2, 0, 0),
  [367] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_string_repeat2, 2, 0, 0), SHIFT_REPEAT(83),
  [370] = {.entry = {.count = 1, .reusable = false}}, SHIFT(92),
  [372] = {.entry = {.count = 1, .reusable = true}}, SHIFT(79),
  [374] = {.entry = {.count = 1, .reusable = true}}, SHIFT(93),
  [376] = {.entry = {.count = 1, .reusable = true}}, SHIFT(117),
  [378] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_identifier_list, 2, 0, 0),
  [380] = {.entry = {.count = 1, .reusable = true}}, SHIFT(75),
  [382] = {.entry = {.count = 1, .reusable = true}}, SHIFT(95),
  [384] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_identifier_list_repeat1, 2, 0, 0), SHIFT_REPEAT(117),
  [387] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_identifier_list_repeat1, 2, 0, 0),
  [389] = {.entry = {.count = 1, .reusable = false}}, SHIFT(3),
  [391] = {.entry = {.count = 1, .reusable = false}}, SHIFT(80),
  [393] = {.entry = {.count = 1, .reusable = false}}, SHIFT(83),
  [395] = {.entry = {.count = 1, .reusable = true}}, SHIFT(82),
  [397] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_block_content, 1, 0, 0),
  [399] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_identifier_list, 1, 0, 0),
  [401] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0),
  [403] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(95),
  [406] = {.entry = {.count = 1, .reusable = true}}, SHIFT(71),
  [408] = {.entry = {.count = 1, .reusable = true}}, SHIFT(24),
  [410] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_defaults_field, 3, 0, 0),
  [412] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_node_field, 3, 0, 0),
  [414] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_op, 1, 0, 0),
  [416] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 3, 0, 0),
  [418] = {.entry = {.count = 1, .reusable = true}}, SHIFT(23),
  [420] = {.entry = {.count = 1, .reusable = true}}, SHIFT(94),
  [422] = {.entry = {.count = 1, .reusable = true}}, SHIFT(106),
  [424] = {.entry = {.count = 1, .reusable = true}}, SHIFT(68),
  [426] = {.entry = {.count = 1, .reusable = true}}, SHIFT(103),
  [428] = {.entry = {.count = 1, .reusable = true}}, SHIFT(104),
  [430] = {.entry = {.count = 1, .reusable = true}}, SHIFT(137),
  [432] = {.entry = {.count = 1, .reusable = true}}, SHIFT(143),
  [434] = {.entry = {.count = 1, .reusable = true}}, SHIFT(43),
  [436] = {.entry = {.count = 1, .reusable = true}}, SHIFT(53),
  [438] = {.entry = {.count = 1, .reusable = true}}, SHIFT(144),
  [440] = {.entry = {.count = 1, .reusable = true}}, SHIFT(50),
  [442] = {.entry = {.count = 1, .reusable = true}}, SHIFT(125),
  [444] = {.entry = {.count = 1, .reusable = true}}, SHIFT(47),
  [446] = {.entry = {.count = 1, .reusable = true}}, SHIFT(100),
  [448] = {.entry = {.count = 1, .reusable = true}}, SHIFT(142),
  [450] = {.entry = {.count = 1, .reusable = true}}, SHIFT(88),
  [452] = {.entry = {.count = 1, .reusable = true}}, SHIFT(107),
  [454] = {.entry = {.count = 1, .reusable = true}}, SHIFT(7),
  [456] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_workflow_decl, 5, 0, 0),
  [458] = {.entry = {.count = 1, .reusable = true}}, SHIFT(108),
  [460] = {.entry = {.count = 1, .reusable = true}}, SHIFT(134),
  [462] = {.entry = {.count = 1, .reusable = true}}, SHIFT(78),
  [464] = {.entry = {.count = 1, .reusable = true}}, SHIFT(85),
  [466] = {.entry = {.count = 1, .reusable = true}}, SHIFT(116),
  [468] = {.entry = {.count = 1, .reusable = true}}, SHIFT(122),
  [470] = {.entry = {.count = 1, .reusable = true}}, SHIFT(12),
  [472] = {.entry = {.count = 1, .reusable = true}}, SHIFT(49),
  [474] = {.entry = {.count = 1, .reusable = true}},  ACCEPT_INPUT(),
  [476] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 1, 0, 0),
  [478] = {.entry = {.count = 1, .reusable = true}}, SHIFT(54),
  [480] = {.entry = {.count = 1, .reusable = true}}, SHIFT(6),
  [482] = {.entry = {.count = 1, .reusable = true}}, SHIFT(51),
  [484] = {.entry = {.count = 1, .reusable = true}}, SHIFT(76),
  [486] = {.entry = {.count = 1, .reusable = true}}, SHIFT(70),
  [488] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_selector, 2, 0, 0),
  [490] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 2, 0, 0),
  [492] = {.entry = {.count = 1, .reusable = true}}, SHIFT(72),
  [494] = {.entry = {.count = 1, .reusable = true}}, SHIFT(136),
  [496] = {.entry = {.count = 1, .reusable = true}}, SHIFT(38),
  [498] = {.entry = {.count = 1, .reusable = true}}, SHIFT(69),
  [500] = {.entry = {.count = 1, .reusable = true}}, SHIFT(77),
  [502] = {.entry = {.count = 1, .reusable = true}}, SHIFT(138),
  [504] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_selector, 1, 0, 0),
  [506] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_field_name, 1, 0, 0),
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
