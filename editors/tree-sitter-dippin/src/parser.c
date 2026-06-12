#include "tree_sitter/parser.h"

#if defined(__GNUC__) || defined(__clang__)
#pragma GCC diagnostic ignored "-Wmissing-field-initializers"
#endif

#define LANGUAGE_VERSION 14
#define STATE_COUNT 149
#define LARGE_STATE_COUNT 2
#define SYMBOL_COUNT 104
#define ALIAS_COUNT 0
#define TOKEN_COUNT 54
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
  anon_sym_loop = 22,
  anon_sym_label = 23,
  anon_sym_weight = 24,
  anon_sym_restart = 25,
  anon_sym_override = 26,
  anon_sym_or = 27,
  anon_sym_and = 28,
  anon_sym_not = 29,
  anon_sym_EQ_EQ = 30,
  anon_sym_BANG_EQ = 31,
  anon_sym_EQ = 32,
  anon_sym_contains = 33,
  anon_sym_startswith = 34,
  anon_sym_endswith = 35,
  anon_sym_in = 36,
  anon_sym_DOT = 37,
  anon_sym_stylesheet = 38,
  anon_sym_STAR = 39,
  anon_sym_POUND = 40,
  sym_raw_inline = 41,
  sym_block_line = 42,
  anon_sym_COMMA = 43,
  anon_sym_DQUOTE = 44,
  aux_sym_string_token1 = 45,
  aux_sym_string_token2 = 46,
  anon_sym_SQUOTE = 47,
  aux_sym_string_token3 = 48,
  anon_sym_SQUOTE_SQUOTE = 49,
  sym_comment = 50,
  sym__indent = 51,
  sym__dedent = 52,
  sym__newline = 53,
  sym_source_file = 54,
  sym_workflow_decl = 55,
  sym_workflow_body = 56,
  sym_workflow_field = 57,
  sym_defaults_section = 58,
  sym_defaults_field = 59,
  sym_node_decl = 60,
  sym_agent_node = 61,
  sym_human_node = 62,
  sym_tool_node = 63,
  sym_subgraph_node = 64,
  sym_conditional_node = 65,
  sym_manager_loop_node = 66,
  sym_parallel_node = 67,
  sym_fan_in_node = 68,
  sym_node_attr_block = 69,
  sym_node_field = 70,
  sym_edges_section = 71,
  sym_edge_entry = 72,
  sym_edge_attr = 73,
  sym_condition = 74,
  sym_or_expr = 75,
  sym_and_expr = 76,
  sym_compare_expr = 77,
  sym_compare_op = 78,
  sym_operand = 79,
  sym_variable = 80,
  sym_stylesheet_section = 81,
  sym_stylesheet_rule = 82,
  sym_selector = 83,
  sym_field_name = 84,
  sym_field_value = 85,
  sym_multiline_block = 86,
  sym_block_content = 87,
  sym_identifier_list = 88,
  sym_string = 89,
  aux_sym_source_file_repeat1 = 90,
  aux_sym_workflow_body_repeat1 = 91,
  aux_sym_defaults_section_repeat1 = 92,
  aux_sym_agent_node_repeat1 = 93,
  aux_sym_edges_section_repeat1 = 94,
  aux_sym_edge_entry_repeat1 = 95,
  aux_sym_or_expr_repeat1 = 96,
  aux_sym_and_expr_repeat1 = 97,
  aux_sym_stylesheet_section_repeat1 = 98,
  aux_sym_stylesheet_rule_repeat1 = 99,
  aux_sym_block_content_repeat1 = 100,
  aux_sym_identifier_list_repeat1 = 101,
  aux_sym_string_repeat1 = 102,
  aux_sym_string_repeat2 = 103,
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
  [anon_sym_loop] = "loop",
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
  [anon_sym_loop] = anon_sym_loop,
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
  [anon_sym_loop] = {
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
  [148] = 148,
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
      if (lookahead == 'o') ADVANCE(30);
      END_STATE();
    case 10:
      if (lookahead == 'a') ADVANCE(31);
      END_STATE();
    case 11:
      if (lookahead == 'o') ADVANCE(32);
      END_STATE();
    case 12:
      if (lookahead == 'n') ADVANCE(33);
      if (lookahead == 'r') ADVANCE(34);
      if (lookahead == 'v') ADVANCE(35);
      END_STATE();
    case 13:
      if (lookahead == 'a') ADVANCE(36);
      END_STATE();
    case 14:
      if (lookahead == 'e') ADVANCE(37);
      END_STATE();
    case 15:
      if (lookahead == 't') ADVANCE(38);
      if (lookahead == 'u') ADVANCE(39);
      END_STATE();
    case 16:
      if (lookahead == 'o') ADVANCE(40);
      END_STATE();
    case 17:
      if (lookahead == 'e') ADVANCE(41);
      if (lookahead == 'h') ADVANCE(42);
      if (lookahead == 'o') ADVANCE(43);
      END_STATE();
    case 18:
      if (lookahead == 'e') ADVANCE(44);
      END_STATE();
    case 19:
      if (lookahead == 'd') ADVANCE(45);
      END_STATE();
    case 20:
      if (lookahead == 'n') ADVANCE(46);
      END_STATE();
    case 21:
      if (lookahead == 'f') ADVANCE(47);
      END_STATE();
    case 22:
      if (lookahead == 'g') ADVANCE(48);
      END_STATE();
    case 23:
      if (lookahead == 'd') ADVANCE(49);
      END_STATE();
    case 24:
      if (lookahead == 'i') ADVANCE(50);
      END_STATE();
    case 25:
      if (lookahead == 'n') ADVANCE(51);
      END_STATE();
    case 26:
      if (lookahead == 'a') ADVANCE(52);
      END_STATE();
    case 27:
      if (lookahead == 'm') ADVANCE(53);
      END_STATE();
    case 28:
      ACCEPT_TOKEN(anon_sym_in);
      END_STATE();
    case 29:
      if (lookahead == 'b') ADVANCE(54);
      END_STATE();
    case 30:
      if (lookahead == 'o') ADVANCE(55);
      END_STATE();
    case 31:
      if (lookahead == 'n') ADVANCE(56);
      END_STATE();
    case 32:
      if (lookahead == 't') ADVANCE(57);
      END_STATE();
    case 33:
      ACCEPT_TOKEN(anon_sym_on);
      END_STATE();
    case 34:
      ACCEPT_TOKEN(anon_sym_or);
      END_STATE();
    case 35:
      if (lookahead == 'e') ADVANCE(58);
      END_STATE();
    case 36:
      if (lookahead == 'r') ADVANCE(59);
      END_STATE();
    case 37:
      if (lookahead == 'q') ADVANCE(60);
      if (lookahead == 's') ADVANCE(61);
      END_STATE();
    case 38:
      if (lookahead == 'a') ADVANCE(62);
      if (lookahead == 'y') ADVANCE(63);
      END_STATE();
    case 39:
      if (lookahead == 'b') ADVANCE(64);
      END_STATE();
    case 40:
      if (lookahead == 'o') ADVANCE(65);
      END_STATE();
    case 41:
      if (lookahead == 'i') ADVANCE(66);
      END_STATE();
    case 42:
      if (lookahead == 'e') ADVANCE(67);
      END_STATE();
    case 43:
      if (lookahead == 'r') ADVANCE(68);
      END_STATE();
    case 44:
      if (lookahead == 'n') ADVANCE(69);
      END_STATE();
    case 45:
      ACCEPT_TOKEN(anon_sym_and);
      END_STATE();
    case 46:
      if (lookahead == 'd') ADVANCE(70);
      if (lookahead == 't') ADVANCE(71);
      END_STATE();
    case 47:
      if (lookahead == 'a') ADVANCE(72);
      END_STATE();
    case 48:
      if (lookahead == 'e') ADVANCE(73);
      END_STATE();
    case 49:
      if (lookahead == 's') ADVANCE(74);
      END_STATE();
    case 50:
      if (lookahead == 't') ADVANCE(75);
      END_STATE();
    case 51:
      if (lookahead == '_') ADVANCE(76);
      END_STATE();
    case 52:
      if (lookahead == 'l') ADVANCE(77);
      END_STATE();
    case 53:
      if (lookahead == 'a') ADVANCE(78);
      END_STATE();
    case 54:
      if (lookahead == 'e') ADVANCE(79);
      END_STATE();
    case 55:
      if (lookahead == 'p') ADVANCE(80);
      END_STATE();
    case 56:
      if (lookahead == 'a') ADVANCE(81);
      END_STATE();
    case 57:
      ACCEPT_TOKEN(anon_sym_not);
      END_STATE();
    case 58:
      if (lookahead == 'r') ADVANCE(82);
      END_STATE();
    case 59:
      if (lookahead == 'a') ADVANCE(83);
      END_STATE();
    case 60:
      if (lookahead == 'u') ADVANCE(84);
      END_STATE();
    case 61:
      if (lookahead == 't') ADVANCE(85);
      END_STATE();
    case 62:
      if (lookahead == 'r') ADVANCE(86);
      END_STATE();
    case 63:
      if (lookahead == 'l') ADVANCE(87);
      END_STATE();
    case 64:
      if (lookahead == 'g') ADVANCE(88);
      END_STATE();
    case 65:
      if (lookahead == 'l') ADVANCE(89);
      END_STATE();
    case 66:
      if (lookahead == 'g') ADVANCE(90);
      END_STATE();
    case 67:
      if (lookahead == 'n') ADVANCE(91);
      END_STATE();
    case 68:
      if (lookahead == 'k') ADVANCE(92);
      END_STATE();
    case 69:
      if (lookahead == 't') ADVANCE(93);
      END_STATE();
    case 70:
      if (lookahead == 'i') ADVANCE(94);
      END_STATE();
    case 71:
      if (lookahead == 'a') ADVANCE(95);
      END_STATE();
    case 72:
      if (lookahead == 'u') ADVANCE(96);
      END_STATE();
    case 73:
      if (lookahead == 's') ADVANCE(97);
      END_STATE();
    case 74:
      if (lookahead == 'w') ADVANCE(98);
      END_STATE();
    case 75:
      ACCEPT_TOKEN(anon_sym_exit);
      END_STATE();
    case 76:
      if (lookahead == 'i') ADVANCE(99);
      END_STATE();
    case 77:
      ACCEPT_TOKEN(anon_sym_goal);
      END_STATE();
    case 78:
      if (lookahead == 'n') ADVANCE(100);
      END_STATE();
    case 79:
      if (lookahead == 'l') ADVANCE(101);
      END_STATE();
    case 80:
      ACCEPT_TOKEN(anon_sym_loop);
      END_STATE();
    case 81:
      if (lookahead == 'g') ADVANCE(102);
      END_STATE();
    case 82:
      if (lookahead == 'r') ADVANCE(103);
      END_STATE();
    case 83:
      if (lookahead == 'l') ADVANCE(104);
      END_STATE();
    case 84:
      if (lookahead == 'i') ADVANCE(105);
      END_STATE();
    case 85:
      if (lookahead == 'a') ADVANCE(106);
      END_STATE();
    case 86:
      if (lookahead == 't') ADVANCE(107);
      END_STATE();
    case 87:
      if (lookahead == 'e') ADVANCE(108);
      END_STATE();
    case 88:
      if (lookahead == 'r') ADVANCE(109);
      END_STATE();
    case 89:
      ACCEPT_TOKEN(anon_sym_tool);
      END_STATE();
    case 90:
      if (lookahead == 'h') ADVANCE(110);
      END_STATE();
    case 91:
      ACCEPT_TOKEN(anon_sym_when);
      END_STATE();
    case 92:
      if (lookahead == 'f') ADVANCE(111);
      END_STATE();
    case 93:
      ACCEPT_TOKEN(anon_sym_agent);
      END_STATE();
    case 94:
      if (lookahead == 't') ADVANCE(112);
      END_STATE();
    case 95:
      if (lookahead == 'i') ADVANCE(113);
      END_STATE();
    case 96:
      if (lookahead == 'l') ADVANCE(114);
      END_STATE();
    case 97:
      ACCEPT_TOKEN(anon_sym_edges);
      END_STATE();
    case 98:
      if (lookahead == 'i') ADVANCE(115);
      END_STATE();
    case 99:
      if (lookahead == 'n') ADVANCE(116);
      END_STATE();
    case 100:
      ACCEPT_TOKEN(anon_sym_human);
      END_STATE();
    case 101:
      ACCEPT_TOKEN(anon_sym_label);
      END_STATE();
    case 102:
      if (lookahead == 'e') ADVANCE(117);
      END_STATE();
    case 103:
      if (lookahead == 'i') ADVANCE(118);
      END_STATE();
    case 104:
      if (lookahead == 'l') ADVANCE(119);
      END_STATE();
    case 105:
      if (lookahead == 'r') ADVANCE(120);
      END_STATE();
    case 106:
      if (lookahead == 'r') ADVANCE(121);
      END_STATE();
    case 107:
      ACCEPT_TOKEN(anon_sym_start);
      if (lookahead == 's') ADVANCE(122);
      END_STATE();
    case 108:
      if (lookahead == 's') ADVANCE(123);
      END_STATE();
    case 109:
      if (lookahead == 'a') ADVANCE(124);
      END_STATE();
    case 110:
      if (lookahead == 't') ADVANCE(125);
      END_STATE();
    case 111:
      if (lookahead == 'l') ADVANCE(126);
      END_STATE();
    case 112:
      if (lookahead == 'i') ADVANCE(127);
      END_STATE();
    case 113:
      if (lookahead == 'n') ADVANCE(128);
      END_STATE();
    case 114:
      if (lookahead == 't') ADVANCE(129);
      END_STATE();
    case 115:
      if (lookahead == 't') ADVANCE(130);
      END_STATE();
    case 116:
      ACCEPT_TOKEN(anon_sym_fan_in);
      END_STATE();
    case 117:
      if (lookahead == 'r') ADVANCE(131);
      END_STATE();
    case 118:
      if (lookahead == 'd') ADVANCE(132);
      END_STATE();
    case 119:
      if (lookahead == 'e') ADVANCE(133);
      END_STATE();
    case 120:
      if (lookahead == 'e') ADVANCE(134);
      END_STATE();
    case 121:
      if (lookahead == 't') ADVANCE(135);
      END_STATE();
    case 122:
      if (lookahead == 'w') ADVANCE(136);
      END_STATE();
    case 123:
      if (lookahead == 'h') ADVANCE(137);
      END_STATE();
    case 124:
      if (lookahead == 'p') ADVANCE(138);
      END_STATE();
    case 125:
      ACCEPT_TOKEN(anon_sym_weight);
      END_STATE();
    case 126:
      if (lookahead == 'o') ADVANCE(139);
      END_STATE();
    case 127:
      if (lookahead == 'o') ADVANCE(140);
      END_STATE();
    case 128:
      if (lookahead == 's') ADVANCE(141);
      END_STATE();
    case 129:
      if (lookahead == 's') ADVANCE(142);
      END_STATE();
    case 130:
      if (lookahead == 'h') ADVANCE(143);
      END_STATE();
    case 131:
      if (lookahead == '_') ADVANCE(144);
      END_STATE();
    case 132:
      if (lookahead == 'e') ADVANCE(145);
      END_STATE();
    case 133:
      if (lookahead == 'l') ADVANCE(146);
      END_STATE();
    case 134:
      if (lookahead == 's') ADVANCE(147);
      END_STATE();
    case 135:
      ACCEPT_TOKEN(anon_sym_restart);
      END_STATE();
    case 136:
      if (lookahead == 'i') ADVANCE(148);
      END_STATE();
    case 137:
      if (lookahead == 'e') ADVANCE(149);
      END_STATE();
    case 138:
      if (lookahead == 'h') ADVANCE(150);
      END_STATE();
    case 139:
      if (lookahead == 'w') ADVANCE(151);
      END_STATE();
    case 140:
      if (lookahead == 'n') ADVANCE(152);
      END_STATE();
    case 141:
      ACCEPT_TOKEN(anon_sym_contains);
      END_STATE();
    case 142:
      ACCEPT_TOKEN(anon_sym_defaults);
      END_STATE();
    case 143:
      ACCEPT_TOKEN(anon_sym_endswith);
      END_STATE();
    case 144:
      if (lookahead == 'l') ADVANCE(153);
      END_STATE();
    case 145:
      ACCEPT_TOKEN(anon_sym_override);
      END_STATE();
    case 146:
      ACCEPT_TOKEN(anon_sym_parallel);
      END_STATE();
    case 147:
      ACCEPT_TOKEN(anon_sym_requires);
      END_STATE();
    case 148:
      if (lookahead == 't') ADVANCE(154);
      END_STATE();
    case 149:
      if (lookahead == 'e') ADVANCE(155);
      END_STATE();
    case 150:
      ACCEPT_TOKEN(anon_sym_subgraph);
      END_STATE();
    case 151:
      ACCEPT_TOKEN(anon_sym_workflow);
      END_STATE();
    case 152:
      if (lookahead == 'a') ADVANCE(156);
      END_STATE();
    case 153:
      if (lookahead == 'o') ADVANCE(157);
      END_STATE();
    case 154:
      if (lookahead == 'h') ADVANCE(158);
      END_STATE();
    case 155:
      if (lookahead == 't') ADVANCE(159);
      END_STATE();
    case 156:
      if (lookahead == 'l') ADVANCE(160);
      END_STATE();
    case 157:
      if (lookahead == 'o') ADVANCE(161);
      END_STATE();
    case 158:
      ACCEPT_TOKEN(anon_sym_startswith);
      END_STATE();
    case 159:
      ACCEPT_TOKEN(anon_sym_stylesheet);
      END_STATE();
    case 160:
      ACCEPT_TOKEN(anon_sym_conditional);
      END_STATE();
    case 161:
      if (lookahead == 'p') ADVANCE(162);
      END_STATE();
    case 162:
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
  [39] = {.lex_state = 9, .external_lex_state = 3},
  [40] = {.lex_state = 9, .external_lex_state = 3},
  [41] = {.lex_state = 9, .external_lex_state = 3},
  [42] = {.lex_state = 9},
  [43] = {.lex_state = 9, .external_lex_state = 3},
  [44] = {.lex_state = 9, .external_lex_state = 3},
  [45] = {.lex_state = 0, .external_lex_state = 3},
  [46] = {.lex_state = 0, .external_lex_state = 3},
  [47] = {.lex_state = 0, .external_lex_state = 2},
  [48] = {.lex_state = 9},
  [49] = {.lex_state = 9},
  [50] = {.lex_state = 1, .external_lex_state = 4},
  [51] = {.lex_state = 1, .external_lex_state = 4},
  [52] = {.lex_state = 1, .external_lex_state = 4},
  [53] = {.lex_state = 1, .external_lex_state = 4},
  [54] = {.lex_state = 9},
  [55] = {.lex_state = 1, .external_lex_state = 4},
  [56] = {.lex_state = 9, .external_lex_state = 3},
  [57] = {.lex_state = 9, .external_lex_state = 3},
  [58] = {.lex_state = 9, .external_lex_state = 3},
  [59] = {.lex_state = 9, .external_lex_state = 3},
  [60] = {.lex_state = 9, .external_lex_state = 3},
  [61] = {.lex_state = 0, .external_lex_state = 3},
  [62] = {.lex_state = 9},
  [63] = {.lex_state = 9, .external_lex_state = 3},
  [64] = {.lex_state = 9, .external_lex_state = 3},
  [65] = {.lex_state = 9, .external_lex_state = 3},
  [66] = {.lex_state = 9},
  [67] = {.lex_state = 9, .external_lex_state = 3},
  [68] = {.lex_state = 9, .external_lex_state = 3},
  [69] = {.lex_state = 9, .external_lex_state = 2},
  [70] = {.lex_state = 9, .external_lex_state = 2},
  [71] = {.lex_state = 9, .external_lex_state = 3},
  [72] = {.lex_state = 9, .external_lex_state = 3},
  [73] = {.lex_state = 9, .external_lex_state = 2},
  [74] = {.lex_state = 9, .external_lex_state = 2},
  [75] = {.lex_state = 9, .external_lex_state = 2},
  [76] = {.lex_state = 9, .external_lex_state = 3},
  [77] = {.lex_state = 9, .external_lex_state = 2},
  [78] = {.lex_state = 9, .external_lex_state = 2},
  [79] = {.lex_state = 9, .external_lex_state = 2},
  [80] = {.lex_state = 9, .external_lex_state = 3},
  [81] = {.lex_state = 2},
  [82] = {.lex_state = 2},
  [83] = {.lex_state = 3, .external_lex_state = 3},
  [84] = {.lex_state = 4},
  [85] = {.lex_state = 4},
  [86] = {.lex_state = 9, .external_lex_state = 2},
  [87] = {.lex_state = 3, .external_lex_state = 2},
  [88] = {.lex_state = 9, .external_lex_state = 5},
  [89] = {.lex_state = 9, .external_lex_state = 2},
  [90] = {.lex_state = 9, .external_lex_state = 2},
  [91] = {.lex_state = 9, .external_lex_state = 5},
  [92] = {.lex_state = 2},
  [93] = {.lex_state = 4},
  [94] = {.lex_state = 3, .external_lex_state = 3},
  [95] = {.lex_state = 9, .external_lex_state = 5},
  [96] = {.lex_state = 9, .external_lex_state = 2},
  [97] = {.lex_state = 9, .external_lex_state = 5},
  [98] = {.lex_state = 9, .external_lex_state = 3},
  [99] = {.lex_state = 9, .external_lex_state = 3},
  [100] = {.lex_state = 9},
  [101] = {.lex_state = 9, .external_lex_state = 5},
  [102] = {.lex_state = 9, .external_lex_state = 3},
  [103] = {.lex_state = 9, .external_lex_state = 5},
  [104] = {.lex_state = 9},
  [105] = {.lex_state = 9},
  [106] = {.lex_state = 9},
  [107] = {.lex_state = 9, .external_lex_state = 4},
  [108] = {.lex_state = 9},
  [109] = {.lex_state = 9},
  [110] = {.lex_state = 9},
  [111] = {.lex_state = 9},
  [112] = {.lex_state = 9},
  [113] = {.lex_state = 9},
  [114] = {.lex_state = 9},
  [115] = {.lex_state = 9},
  [116] = {.lex_state = 9},
  [117] = {.lex_state = 9},
  [118] = {.lex_state = 9},
  [119] = {.lex_state = 9},
  [120] = {.lex_state = 9, .external_lex_state = 4},
  [121] = {.lex_state = 9, .external_lex_state = 6},
  [122] = {.lex_state = 9},
  [123] = {.lex_state = 9},
  [124] = {.lex_state = 9},
  [125] = {.lex_state = 9},
  [126] = {.lex_state = 9, .external_lex_state = 4},
  [127] = {.lex_state = 9},
  [128] = {.lex_state = 9, .external_lex_state = 6},
  [129] = {.lex_state = 9},
  [130] = {.lex_state = 9},
  [131] = {.lex_state = 9},
  [132] = {.lex_state = 9},
  [133] = {.lex_state = 9},
  [134] = {.lex_state = 9, .external_lex_state = 4},
  [135] = {.lex_state = 9, .external_lex_state = 4},
  [136] = {.lex_state = 9, .external_lex_state = 4},
  [137] = {.lex_state = 9, .external_lex_state = 4},
  [138] = {.lex_state = 9},
  [139] = {.lex_state = 9, .external_lex_state = 4},
  [140] = {.lex_state = 9},
  [141] = {.lex_state = 9},
  [142] = {.lex_state = 9},
  [143] = {.lex_state = 9, .external_lex_state = 4},
  [144] = {.lex_state = 9, .external_lex_state = 4},
  [145] = {.lex_state = 9},
  [146] = {.lex_state = 9, .external_lex_state = 4},
  [147] = {.lex_state = 9, .external_lex_state = 4},
  [148] = {.lex_state = 9, .external_lex_state = 4},
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
    [anon_sym_loop] = ACTIONS(1),
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
    [aux_sym_source_file_repeat1] = STATE(90),
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
    ACTIONS(11), 31,
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
      anon_sym_loop,
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
  [43] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(17), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(15), 31,
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
      anon_sym_loop,
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
  [86] = 17,
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
  [153] = 17,
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
  [220] = 17,
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
  [287] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(92), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(90), 23,
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
      anon_sym_loop,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_stylesheet,
      sym_identifier,
  [320] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(96), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(94), 23,
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
      anon_sym_loop,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_stylesheet,
      sym_identifier,
  [353] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(102), 1,
      anon_sym_DOT,
    ACTIONS(100), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(98), 16,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
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
  [384] = 7,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(106), 1,
      anon_sym_not,
    STATE(62), 1,
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
    ACTIONS(104), 10,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      anon_sym_and,
      sym_identifier,
  [421] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(100), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(98), 16,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
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
  [449] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(116), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(114), 16,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
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
  [477] = 2,
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
  [500] = 2,
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
  [523] = 2,
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
  [546] = 2,
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
  [569] = 2,
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
  [592] = 2,
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
  [615] = 2,
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
  [638] = 2,
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
  [661] = 2,
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
  [684] = 2,
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
  [707] = 2,
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
  [730] = 2,
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
  [753] = 2,
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
  [776] = 2,
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
  [799] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(148), 1,
      anon_sym_and,
    STATE(27), 1,
      aux_sym_and_expr_repeat1,
    ACTIONS(151), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(146), 9,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      sym_identifier,
  [824] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(155), 1,
      anon_sym_and,
    STATE(27), 1,
      aux_sym_and_expr_repeat1,
    ACTIONS(157), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(153), 9,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      sym_identifier,
  [849] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(155), 1,
      anon_sym_and,
    STATE(28), 1,
      aux_sym_and_expr_repeat1,
    ACTIONS(161), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(159), 9,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      sym_identifier,
  [874] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(165), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(163), 10,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      anon_sym_and,
      sym_identifier,
  [894] = 8,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(167), 1,
      sym_identifier,
    ACTIONS(169), 1,
      anon_sym_when,
    ACTIONS(171), 1,
      anon_sym_on,
    ACTIONS(173), 1,
      anon_sym_loop,
    ACTIONS(177), 2,
      sym__dedent,
      sym__newline,
    STATE(32), 2,
      sym_edge_attr,
      aux_sym_edge_entry_repeat1,
    ACTIONS(175), 4,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
  [924] = 8,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(169), 1,
      anon_sym_when,
    ACTIONS(171), 1,
      anon_sym_on,
    ACTIONS(173), 1,
      anon_sym_loop,
    ACTIONS(179), 1,
      sym_identifier,
    ACTIONS(181), 2,
      sym__dedent,
      sym__newline,
    STATE(34), 2,
      sym_edge_attr,
      aux_sym_edge_entry_repeat1,
    ACTIONS(175), 4,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
  [954] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(185), 1,
      anon_sym_or,
    STATE(35), 1,
      aux_sym_or_expr_repeat1,
    ACTIONS(187), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(183), 8,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [978] = 8,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(189), 1,
      sym_identifier,
    ACTIONS(191), 1,
      anon_sym_when,
    ACTIONS(194), 1,
      anon_sym_on,
    ACTIONS(197), 1,
      anon_sym_loop,
    ACTIONS(203), 2,
      sym__dedent,
      sym__newline,
    STATE(34), 2,
      sym_edge_attr,
      aux_sym_edge_entry_repeat1,
    ACTIONS(200), 4,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
  [1008] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(207), 1,
      anon_sym_or,
    STATE(35), 1,
      aux_sym_or_expr_repeat1,
    ACTIONS(210), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(205), 8,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [1032] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(151), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(146), 10,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      anon_sym_and,
      sym_identifier,
  [1052] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(185), 1,
      anon_sym_or,
    STATE(33), 1,
      aux_sym_or_expr_repeat1,
    ACTIONS(214), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(212), 8,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [1076] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(218), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(216), 10,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      anon_sym_and,
      sym_identifier,
  [1096] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(210), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(205), 9,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_or,
      sym_identifier,
  [1115] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(222), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(220), 8,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [1133] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(226), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(224), 8,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [1151] = 10,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(228), 1,
      sym_identifier,
    ACTIONS(230), 1,
      anon_sym_DQUOTE,
    ACTIONS(232), 1,
      anon_sym_SQUOTE,
    STATE(10), 1,
      sym_operand,
    STATE(29), 1,
      sym_compare_expr,
    STATE(37), 1,
      sym_and_expr,
    STATE(40), 1,
      sym_or_expr,
    STATE(41), 1,
      sym_condition,
    STATE(11), 2,
      sym_variable,
      sym_string,
  [1183] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(236), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(234), 8,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [1201] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(240), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(238), 8,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [1219] = 8,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(244), 1,
      anon_sym_DOT,
    ACTIONS(246), 1,
      anon_sym_POUND,
    ACTIONS(248), 1,
      sym__dedent,
    ACTIONS(250), 1,
      sym__newline,
    STATE(147), 1,
      sym_selector,
    ACTIONS(242), 2,
      anon_sym_STAR,
      sym_identifier,
    STATE(46), 2,
      sym_stylesheet_rule,
      aux_sym_stylesheet_section_repeat1,
  [1246] = 8,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(255), 1,
      anon_sym_DOT,
    ACTIONS(258), 1,
      anon_sym_POUND,
    ACTIONS(261), 1,
      sym__dedent,
    ACTIONS(263), 1,
      sym__newline,
    STATE(147), 1,
      sym_selector,
    ACTIONS(252), 2,
      anon_sym_STAR,
      sym_identifier,
    STATE(46), 2,
      sym_stylesheet_rule,
      aux_sym_stylesheet_section_repeat1,
  [1273] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(244), 1,
      anon_sym_DOT,
    ACTIONS(246), 1,
      anon_sym_POUND,
    ACTIONS(266), 1,
      sym__newline,
    STATE(147), 1,
      sym_selector,
    ACTIONS(242), 2,
      anon_sym_STAR,
      sym_identifier,
    STATE(45), 2,
      sym_stylesheet_rule,
      aux_sym_stylesheet_section_repeat1,
  [1297] = 8,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(228), 1,
      sym_identifier,
    ACTIONS(230), 1,
      anon_sym_DQUOTE,
    ACTIONS(232), 1,
      anon_sym_SQUOTE,
    STATE(10), 1,
      sym_operand,
    STATE(29), 1,
      sym_compare_expr,
    STATE(39), 1,
      sym_and_expr,
    STATE(11), 2,
      sym_variable,
      sym_string,
  [1323] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(110), 1,
      anon_sym_EQ,
    STATE(66), 1,
      sym_compare_op,
    ACTIONS(108), 6,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
      anon_sym_contains,
      anon_sym_startswith,
      anon_sym_endswith,
      anon_sym_in,
  [1341] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(268), 1,
      sym_raw_inline,
    ACTIONS(270), 1,
      anon_sym_DQUOTE,
    ACTIONS(272), 1,
      anon_sym_SQUOTE,
    ACTIONS(274), 1,
      sym__indent,
    STATE(18), 1,
      sym_field_value,
    STATE(8), 2,
      sym_multiline_block,
      sym_string,
  [1364] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(268), 1,
      sym_raw_inline,
    ACTIONS(270), 1,
      anon_sym_DQUOTE,
    ACTIONS(272), 1,
      anon_sym_SQUOTE,
    ACTIONS(274), 1,
      sym__indent,
    STATE(44), 1,
      sym_field_value,
    STATE(8), 2,
      sym_multiline_block,
      sym_string,
  [1387] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(268), 1,
      sym_raw_inline,
    ACTIONS(270), 1,
      anon_sym_DQUOTE,
    ACTIONS(272), 1,
      anon_sym_SQUOTE,
    ACTIONS(274), 1,
      sym__indent,
    STATE(98), 1,
      sym_field_value,
    STATE(8), 2,
      sym_multiline_block,
      sym_string,
  [1410] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(268), 1,
      sym_raw_inline,
    ACTIONS(270), 1,
      anon_sym_DQUOTE,
    ACTIONS(272), 1,
      anon_sym_SQUOTE,
    ACTIONS(274), 1,
      sym__indent,
    STATE(102), 1,
      sym_field_value,
    STATE(8), 2,
      sym_multiline_block,
      sym_string,
  [1433] = 7,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(228), 1,
      sym_identifier,
    ACTIONS(230), 1,
      anon_sym_DQUOTE,
    ACTIONS(232), 1,
      anon_sym_SQUOTE,
    STATE(10), 1,
      sym_operand,
    STATE(36), 1,
      sym_compare_expr,
    STATE(11), 2,
      sym_variable,
      sym_string,
  [1456] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(268), 1,
      sym_raw_inline,
    ACTIONS(270), 1,
      anon_sym_DQUOTE,
    ACTIONS(272), 1,
      anon_sym_SQUOTE,
    ACTIONS(274), 1,
      sym__indent,
    STATE(99), 1,
      sym_field_value,
    STATE(8), 2,
      sym_multiline_block,
      sym_string,
  [1479] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(276), 1,
      sym_identifier,
    ACTIONS(278), 1,
      sym__dedent,
    ACTIONS(280), 1,
      sym__newline,
    STATE(129), 1,
      sym_field_name,
    STATE(63), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1499] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(276), 1,
      sym_identifier,
    ACTIONS(280), 1,
      sym__newline,
    ACTIONS(282), 1,
      sym__dedent,
    STATE(129), 1,
      sym_field_name,
    STATE(63), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1519] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(276), 1,
      sym_identifier,
    ACTIONS(280), 1,
      sym__newline,
    ACTIONS(284), 1,
      sym__dedent,
    STATE(129), 1,
      sym_field_name,
    STATE(63), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1539] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(276), 1,
      sym_identifier,
    ACTIONS(280), 1,
      sym__newline,
    ACTIONS(286), 1,
      sym__dedent,
    STATE(129), 1,
      sym_field_name,
    STATE(63), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1559] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(276), 1,
      sym_identifier,
    ACTIONS(280), 1,
      sym__newline,
    ACTIONS(288), 1,
      sym__dedent,
    STATE(129), 1,
      sym_field_name,
    STATE(63), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1579] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(292), 1,
      anon_sym_POUND,
    ACTIONS(290), 5,
      sym__dedent,
      sym__newline,
      anon_sym_DOT,
      anon_sym_STAR,
      sym_identifier,
  [1593] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(228), 1,
      sym_identifier,
    ACTIONS(230), 1,
      anon_sym_DQUOTE,
    ACTIONS(232), 1,
      anon_sym_SQUOTE,
    STATE(38), 1,
      sym_operand,
    STATE(11), 2,
      sym_variable,
      sym_string,
  [1613] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(294), 1,
      sym_identifier,
    ACTIONS(297), 1,
      sym__dedent,
    ACTIONS(299), 1,
      sym__newline,
    STATE(129), 1,
      sym_field_name,
    STATE(63), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1633] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(302), 1,
      sym_identifier,
    ACTIONS(305), 1,
      sym__dedent,
    ACTIONS(307), 1,
      sym__newline,
    STATE(125), 1,
      sym_field_name,
    STATE(64), 2,
      sym_defaults_field,
      aux_sym_defaults_section_repeat1,
  [1653] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(276), 1,
      sym_identifier,
    ACTIONS(280), 1,
      sym__newline,
    ACTIONS(310), 1,
      sym__dedent,
    STATE(129), 1,
      sym_field_name,
    STATE(63), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1673] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(228), 1,
      sym_identifier,
    ACTIONS(230), 1,
      anon_sym_DQUOTE,
    ACTIONS(232), 1,
      anon_sym_SQUOTE,
    STATE(30), 1,
      sym_operand,
    STATE(11), 2,
      sym_variable,
      sym_string,
  [1693] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(276), 1,
      sym_identifier,
    ACTIONS(280), 1,
      sym__newline,
    ACTIONS(312), 1,
      sym__dedent,
    STATE(129), 1,
      sym_field_name,
    STATE(63), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1713] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(276), 1,
      sym_identifier,
    ACTIONS(314), 1,
      sym__dedent,
    ACTIONS(316), 1,
      sym__newline,
    STATE(125), 1,
      sym_field_name,
    STATE(64), 2,
      sym_defaults_field,
      aux_sym_defaults_section_repeat1,
  [1733] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(276), 1,
      sym_identifier,
    ACTIONS(318), 1,
      sym__newline,
    STATE(129), 1,
      sym_field_name,
    STATE(60), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1750] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(276), 1,
      sym_identifier,
    ACTIONS(320), 1,
      sym__newline,
    STATE(129), 1,
      sym_field_name,
    STATE(65), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1767] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(322), 1,
      sym_identifier,
    ACTIONS(325), 1,
      sym__dedent,
    ACTIONS(327), 1,
      sym__newline,
    STATE(71), 2,
      sym_edge_entry,
      aux_sym_edges_section_repeat1,
  [1784] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(276), 1,
      sym_identifier,
    ACTIONS(330), 1,
      sym__dedent,
    ACTIONS(332), 1,
      sym__newline,
    STATE(76), 1,
      aux_sym_stylesheet_rule_repeat1,
    STATE(115), 1,
      sym_field_name,
  [1803] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(276), 1,
      sym_identifier,
    ACTIONS(334), 1,
      sym__newline,
    STATE(129), 1,
      sym_field_name,
    STATE(58), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1820] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(276), 1,
      sym_identifier,
    ACTIONS(336), 1,
      sym__newline,
    STATE(129), 1,
      sym_field_name,
    STATE(57), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1837] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(276), 1,
      sym_identifier,
    ACTIONS(338), 1,
      sym__newline,
    STATE(125), 1,
      sym_field_name,
    STATE(68), 2,
      sym_defaults_field,
      aux_sym_defaults_section_repeat1,
  [1854] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(340), 1,
      sym_identifier,
    ACTIONS(343), 1,
      sym__dedent,
    ACTIONS(345), 1,
      sym__newline,
    STATE(76), 1,
      aux_sym_stylesheet_rule_repeat1,
    STATE(115), 1,
      sym_field_name,
  [1873] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(276), 1,
      sym_identifier,
    ACTIONS(348), 1,
      sym__newline,
    STATE(129), 1,
      sym_field_name,
    STATE(59), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1890] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(276), 1,
      sym_identifier,
    ACTIONS(350), 1,
      sym__newline,
    STATE(129), 1,
      sym_field_name,
    STATE(67), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1907] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(276), 1,
      sym_identifier,
    ACTIONS(352), 1,
      sym__newline,
    STATE(129), 1,
      sym_field_name,
    STATE(56), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1924] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(354), 1,
      sym_identifier,
    ACTIONS(356), 1,
      sym__dedent,
    ACTIONS(358), 1,
      sym__newline,
    STATE(71), 2,
      sym_edge_entry,
      aux_sym_edges_section_repeat1,
  [1941] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(360), 1,
      anon_sym_DQUOTE,
    STATE(81), 1,
      aux_sym_string_repeat1,
    ACTIONS(362), 2,
      aux_sym_string_token1,
      aux_sym_string_token2,
  [1955] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(365), 1,
      anon_sym_DQUOTE,
    STATE(92), 1,
      aux_sym_string_repeat1,
    ACTIONS(367), 2,
      aux_sym_string_token1,
      aux_sym_string_token2,
  [1969] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(372), 1,
      sym__dedent,
    STATE(83), 1,
      aux_sym_block_content_repeat1,
    ACTIONS(369), 2,
      sym__newline,
      sym_block_line,
  [1983] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(374), 1,
      anon_sym_SQUOTE,
    STATE(84), 1,
      aux_sym_string_repeat2,
    ACTIONS(376), 2,
      aux_sym_string_token3,
      anon_sym_SQUOTE_SQUOTE,
  [1997] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(365), 1,
      anon_sym_SQUOTE,
    STATE(93), 1,
      aux_sym_string_repeat2,
    ACTIONS(379), 2,
      aux_sym_string_token3,
      anon_sym_SQUOTE_SQUOTE,
  [2011] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(354), 1,
      sym_identifier,
    ACTIONS(381), 1,
      sym__newline,
    STATE(80), 2,
      sym_edge_entry,
      aux_sym_edges_section_repeat1,
  [2025] = 4,
    ACTIONS(3), 1,
      sym_comment,
    STATE(94), 1,
      aux_sym_block_content_repeat1,
    STATE(121), 1,
      sym_block_content,
    ACTIONS(383), 2,
      sym__newline,
      sym_block_line,
  [2039] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(385), 1,
      anon_sym_COMMA,
    STATE(91), 1,
      aux_sym_identifier_list_repeat1,
    ACTIONS(387), 2,
      sym__indent,
      sym__newline,
  [2053] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(276), 1,
      sym_identifier,
    ACTIONS(389), 1,
      sym__newline,
    STATE(72), 1,
      aux_sym_stylesheet_rule_repeat1,
    STATE(115), 1,
      sym_field_name,
  [2069] = 5,
    ACTIONS(5), 1,
      anon_sym_workflow,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(391), 1,
      sym__newline,
    STATE(96), 1,
      aux_sym_source_file_repeat1,
    STATE(117), 1,
      sym_workflow_decl,
  [2085] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(393), 1,
      anon_sym_COMMA,
    STATE(91), 1,
      aux_sym_identifier_list_repeat1,
    ACTIONS(396), 2,
      sym__indent,
      sym__newline,
  [2099] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(398), 1,
      anon_sym_DQUOTE,
    STATE(81), 1,
      aux_sym_string_repeat1,
    ACTIONS(400), 2,
      aux_sym_string_token1,
      aux_sym_string_token2,
  [2113] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(398), 1,
      anon_sym_SQUOTE,
    STATE(84), 1,
      aux_sym_string_repeat2,
    ACTIONS(402), 2,
      aux_sym_string_token3,
      anon_sym_SQUOTE_SQUOTE,
  [2127] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(406), 1,
      sym__dedent,
    STATE(83), 1,
      aux_sym_block_content_repeat1,
    ACTIONS(404), 2,
      sym__newline,
      sym_block_line,
  [2141] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(385), 1,
      anon_sym_COMMA,
    STATE(88), 1,
      aux_sym_identifier_list_repeat1,
    ACTIONS(408), 2,
      sym__indent,
      sym__newline,
  [2155] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(410), 1,
      anon_sym_workflow,
    ACTIONS(412), 1,
      sym__newline,
    STATE(96), 1,
      aux_sym_source_file_repeat1,
  [2168] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(415), 1,
      sym__indent,
    ACTIONS(417), 1,
      sym__newline,
    STATE(24), 1,
      sym_node_attr_block,
  [2181] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(419), 3,
      sym__dedent,
      sym__newline,
      sym_identifier,
  [2190] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(421), 3,
      sym__dedent,
      sym__newline,
      sym_identifier,
  [2199] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(423), 3,
      sym_identifier,
      anon_sym_DQUOTE,
      anon_sym_SQUOTE,
  [2208] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(396), 3,
      sym__indent,
      sym__newline,
      anon_sym_COMMA,
  [2217] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(425), 3,
      sym__dedent,
      sym__newline,
      sym_identifier,
  [2226] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(415), 1,
      sym__indent,
    ACTIONS(427), 1,
      sym__newline,
    STATE(23), 1,
      sym_node_attr_block,
  [2239] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(429), 1,
      sym_identifier,
    STATE(103), 1,
      sym_identifier_list,
  [2249] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(429), 1,
      sym_identifier,
    STATE(97), 1,
      sym_identifier_list,
  [2259] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(431), 1,
      sym_identifier,
  [2266] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(433), 1,
      sym__indent,
  [2273] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(435), 1,
      ts_builtin_sym_end,
  [2280] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(437), 1,
      sym_identifier,
  [2287] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(439), 1,
      sym_identifier,
  [2294] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(441), 1,
      sym_identifier,
  [2301] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(443), 1,
      sym_identifier,
  [2308] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(445), 1,
      anon_sym_COLON,
  [2315] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(447), 1,
      sym_identifier,
  [2322] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(449), 1,
      anon_sym_COLON,
  [2329] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(451), 1,
      sym_identifier,
  [2336] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(453), 1,
      ts_builtin_sym_end,
  [2343] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(455), 1,
      anon_sym_DASH_GT,
  [2350] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(457), 1,
      sym_identifier,
  [2357] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(459), 1,
      sym__indent,
  [2364] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(461), 1,
      sym__dedent,
  [2371] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(463), 1,
      sym_identifier,
  [2378] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(465), 1,
      sym_identifier,
  [2385] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(467), 1,
      anon_sym_COLON,
  [2392] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(469), 1,
      anon_sym_COLON,
  [2399] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(471), 1,
      sym__indent,
  [2406] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(473), 1,
      anon_sym_COLON,
  [2413] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(475), 1,
      sym__dedent,
  [2420] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(477), 1,
      anon_sym_COLON,
  [2427] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(479), 1,
      sym_identifier,
  [2434] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(481), 1,
      ts_builtin_sym_end,
  [2441] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(483), 1,
      ts_builtin_sym_end,
  [2448] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(485), 1,
      anon_sym_COLON,
  [2455] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(487), 1,
      sym__indent,
  [2462] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(489), 1,
      sym__indent,
  [2469] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(491), 1,
      sym__indent,
  [2476] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(493), 1,
      sym__indent,
  [2483] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(495), 1,
      anon_sym_LT_DASH,
  [2490] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(497), 1,
      sym__indent,
  [2497] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(499), 1,
      sym_identifier,
  [2504] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(501), 1,
      anon_sym_DASH_GT,
  [2511] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(503), 1,
      sym_identifier,
  [2518] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(505), 1,
      sym__indent,
  [2525] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(507), 1,
      sym__indent,
  [2532] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(509), 1,
      sym_identifier,
  [2539] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(511), 1,
      sym__indent,
  [2546] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(513), 1,
      sym__indent,
  [2553] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(515), 1,
      sym__indent,
};

static const uint32_t ts_small_parse_table_map[] = {
  [SMALL_STATE(2)] = 0,
  [SMALL_STATE(3)] = 43,
  [SMALL_STATE(4)] = 86,
  [SMALL_STATE(5)] = 153,
  [SMALL_STATE(6)] = 220,
  [SMALL_STATE(7)] = 287,
  [SMALL_STATE(8)] = 320,
  [SMALL_STATE(9)] = 353,
  [SMALL_STATE(10)] = 384,
  [SMALL_STATE(11)] = 421,
  [SMALL_STATE(12)] = 449,
  [SMALL_STATE(13)] = 477,
  [SMALL_STATE(14)] = 500,
  [SMALL_STATE(15)] = 523,
  [SMALL_STATE(16)] = 546,
  [SMALL_STATE(17)] = 569,
  [SMALL_STATE(18)] = 592,
  [SMALL_STATE(19)] = 615,
  [SMALL_STATE(20)] = 638,
  [SMALL_STATE(21)] = 661,
  [SMALL_STATE(22)] = 684,
  [SMALL_STATE(23)] = 707,
  [SMALL_STATE(24)] = 730,
  [SMALL_STATE(25)] = 753,
  [SMALL_STATE(26)] = 776,
  [SMALL_STATE(27)] = 799,
  [SMALL_STATE(28)] = 824,
  [SMALL_STATE(29)] = 849,
  [SMALL_STATE(30)] = 874,
  [SMALL_STATE(31)] = 894,
  [SMALL_STATE(32)] = 924,
  [SMALL_STATE(33)] = 954,
  [SMALL_STATE(34)] = 978,
  [SMALL_STATE(35)] = 1008,
  [SMALL_STATE(36)] = 1032,
  [SMALL_STATE(37)] = 1052,
  [SMALL_STATE(38)] = 1076,
  [SMALL_STATE(39)] = 1096,
  [SMALL_STATE(40)] = 1115,
  [SMALL_STATE(41)] = 1133,
  [SMALL_STATE(42)] = 1151,
  [SMALL_STATE(43)] = 1183,
  [SMALL_STATE(44)] = 1201,
  [SMALL_STATE(45)] = 1219,
  [SMALL_STATE(46)] = 1246,
  [SMALL_STATE(47)] = 1273,
  [SMALL_STATE(48)] = 1297,
  [SMALL_STATE(49)] = 1323,
  [SMALL_STATE(50)] = 1341,
  [SMALL_STATE(51)] = 1364,
  [SMALL_STATE(52)] = 1387,
  [SMALL_STATE(53)] = 1410,
  [SMALL_STATE(54)] = 1433,
  [SMALL_STATE(55)] = 1456,
  [SMALL_STATE(56)] = 1479,
  [SMALL_STATE(57)] = 1499,
  [SMALL_STATE(58)] = 1519,
  [SMALL_STATE(59)] = 1539,
  [SMALL_STATE(60)] = 1559,
  [SMALL_STATE(61)] = 1579,
  [SMALL_STATE(62)] = 1593,
  [SMALL_STATE(63)] = 1613,
  [SMALL_STATE(64)] = 1633,
  [SMALL_STATE(65)] = 1653,
  [SMALL_STATE(66)] = 1673,
  [SMALL_STATE(67)] = 1693,
  [SMALL_STATE(68)] = 1713,
  [SMALL_STATE(69)] = 1733,
  [SMALL_STATE(70)] = 1750,
  [SMALL_STATE(71)] = 1767,
  [SMALL_STATE(72)] = 1784,
  [SMALL_STATE(73)] = 1803,
  [SMALL_STATE(74)] = 1820,
  [SMALL_STATE(75)] = 1837,
  [SMALL_STATE(76)] = 1854,
  [SMALL_STATE(77)] = 1873,
  [SMALL_STATE(78)] = 1890,
  [SMALL_STATE(79)] = 1907,
  [SMALL_STATE(80)] = 1924,
  [SMALL_STATE(81)] = 1941,
  [SMALL_STATE(82)] = 1955,
  [SMALL_STATE(83)] = 1969,
  [SMALL_STATE(84)] = 1983,
  [SMALL_STATE(85)] = 1997,
  [SMALL_STATE(86)] = 2011,
  [SMALL_STATE(87)] = 2025,
  [SMALL_STATE(88)] = 2039,
  [SMALL_STATE(89)] = 2053,
  [SMALL_STATE(90)] = 2069,
  [SMALL_STATE(91)] = 2085,
  [SMALL_STATE(92)] = 2099,
  [SMALL_STATE(93)] = 2113,
  [SMALL_STATE(94)] = 2127,
  [SMALL_STATE(95)] = 2141,
  [SMALL_STATE(96)] = 2155,
  [SMALL_STATE(97)] = 2168,
  [SMALL_STATE(98)] = 2181,
  [SMALL_STATE(99)] = 2190,
  [SMALL_STATE(100)] = 2199,
  [SMALL_STATE(101)] = 2208,
  [SMALL_STATE(102)] = 2217,
  [SMALL_STATE(103)] = 2226,
  [SMALL_STATE(104)] = 2239,
  [SMALL_STATE(105)] = 2249,
  [SMALL_STATE(106)] = 2259,
  [SMALL_STATE(107)] = 2266,
  [SMALL_STATE(108)] = 2273,
  [SMALL_STATE(109)] = 2280,
  [SMALL_STATE(110)] = 2287,
  [SMALL_STATE(111)] = 2294,
  [SMALL_STATE(112)] = 2301,
  [SMALL_STATE(113)] = 2308,
  [SMALL_STATE(114)] = 2315,
  [SMALL_STATE(115)] = 2322,
  [SMALL_STATE(116)] = 2329,
  [SMALL_STATE(117)] = 2336,
  [SMALL_STATE(118)] = 2343,
  [SMALL_STATE(119)] = 2350,
  [SMALL_STATE(120)] = 2357,
  [SMALL_STATE(121)] = 2364,
  [SMALL_STATE(122)] = 2371,
  [SMALL_STATE(123)] = 2378,
  [SMALL_STATE(124)] = 2385,
  [SMALL_STATE(125)] = 2392,
  [SMALL_STATE(126)] = 2399,
  [SMALL_STATE(127)] = 2406,
  [SMALL_STATE(128)] = 2413,
  [SMALL_STATE(129)] = 2420,
  [SMALL_STATE(130)] = 2427,
  [SMALL_STATE(131)] = 2434,
  [SMALL_STATE(132)] = 2441,
  [SMALL_STATE(133)] = 2448,
  [SMALL_STATE(134)] = 2455,
  [SMALL_STATE(135)] = 2462,
  [SMALL_STATE(136)] = 2469,
  [SMALL_STATE(137)] = 2476,
  [SMALL_STATE(138)] = 2483,
  [SMALL_STATE(139)] = 2490,
  [SMALL_STATE(140)] = 2497,
  [SMALL_STATE(141)] = 2504,
  [SMALL_STATE(142)] = 2511,
  [SMALL_STATE(143)] = 2518,
  [SMALL_STATE(144)] = 2525,
  [SMALL_STATE(145)] = 2532,
  [SMALL_STATE(146)] = 2539,
  [SMALL_STATE(147)] = 2546,
  [SMALL_STATE(148)] = 2553,
};

static const TSParseActionEntry ts_parse_actions[] = {
  [0] = {.entry = {.count = 0, .reusable = false}},
  [1] = {.entry = {.count = 1, .reusable = false}}, RECOVER(),
  [3] = {.entry = {.count = 1, .reusable = false}}, SHIFT_EXTRA(),
  [5] = {.entry = {.count = 1, .reusable = true}}, SHIFT(122),
  [7] = {.entry = {.count = 1, .reusable = true}}, SHIFT_EXTRA(),
  [9] = {.entry = {.count = 1, .reusable = true}}, SHIFT(90),
  [11] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_string, 2, 0, 0),
  [13] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_string, 2, 0, 0),
  [15] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_string, 3, 0, 0),
  [17] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_string, 3, 0, 0),
  [19] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(133),
  [22] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(139),
  [25] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(140),
  [28] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(109),
  [31] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(110),
  [34] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(112),
  [37] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(114),
  [40] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(119),
  [43] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(106),
  [46] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(123),
  [49] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(148),
  [52] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(127),
  [55] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0),
  [57] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(4),
  [60] = {.entry = {.count = 1, .reusable = true}}, SHIFT(133),
  [62] = {.entry = {.count = 1, .reusable = true}}, SHIFT(139),
  [64] = {.entry = {.count = 1, .reusable = true}}, SHIFT(140),
  [66] = {.entry = {.count = 1, .reusable = true}}, SHIFT(109),
  [68] = {.entry = {.count = 1, .reusable = true}}, SHIFT(110),
  [70] = {.entry = {.count = 1, .reusable = true}}, SHIFT(112),
  [72] = {.entry = {.count = 1, .reusable = true}}, SHIFT(114),
  [74] = {.entry = {.count = 1, .reusable = true}}, SHIFT(119),
  [76] = {.entry = {.count = 1, .reusable = true}}, SHIFT(106),
  [78] = {.entry = {.count = 1, .reusable = true}}, SHIFT(123),
  [80] = {.entry = {.count = 1, .reusable = true}}, SHIFT(148),
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
  [102] = {.entry = {.count = 1, .reusable = true}}, SHIFT(130),
  [104] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_compare_expr, 1, 0, 0),
  [106] = {.entry = {.count = 1, .reusable = false}}, SHIFT(49),
  [108] = {.entry = {.count = 1, .reusable = true}}, SHIFT(100),
  [110] = {.entry = {.count = 1, .reusable = false}}, SHIFT(100),
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
  [142] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_stylesheet_section, 5, 0, 0),
  [144] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_node_attr_block, 3, 0, 0),
  [146] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_and_expr_repeat1, 2, 0, 0),
  [148] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_and_expr_repeat1, 2, 0, 0), SHIFT_REPEAT(54),
  [151] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_and_expr_repeat1, 2, 0, 0),
  [153] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_and_expr, 2, 0, 0),
  [155] = {.entry = {.count = 1, .reusable = false}}, SHIFT(54),
  [157] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_and_expr, 2, 0, 0),
  [159] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_and_expr, 1, 0, 0),
  [161] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_and_expr, 1, 0, 0),
  [163] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_compare_expr, 4, 0, 0),
  [165] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_expr, 4, 0, 0),
  [167] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_entry, 3, 0, 0),
  [169] = {.entry = {.count = 1, .reusable = false}}, SHIFT(42),
  [171] = {.entry = {.count = 1, .reusable = false}}, SHIFT(111),
  [173] = {.entry = {.count = 1, .reusable = false}}, SHIFT(43),
  [175] = {.entry = {.count = 1, .reusable = false}}, SHIFT(113),
  [177] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_entry, 3, 0, 0),
  [179] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_entry, 4, 0, 0),
  [181] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_entry, 4, 0, 0),
  [183] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_or_expr, 2, 0, 0),
  [185] = {.entry = {.count = 1, .reusable = false}}, SHIFT(48),
  [187] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_or_expr, 2, 0, 0),
  [189] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0),
  [191] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0), SHIFT_REPEAT(42),
  [194] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0), SHIFT_REPEAT(111),
  [197] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0), SHIFT_REPEAT(43),
  [200] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0), SHIFT_REPEAT(113),
  [203] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0),
  [205] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_or_expr_repeat1, 2, 0, 0),
  [207] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_or_expr_repeat1, 2, 0, 0), SHIFT_REPEAT(48),
  [210] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_or_expr_repeat1, 2, 0, 0),
  [212] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_or_expr, 1, 0, 0),
  [214] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_or_expr, 1, 0, 0),
  [216] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_compare_expr, 3, 0, 0),
  [218] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_expr, 3, 0, 0),
  [220] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_condition, 1, 0, 0),
  [222] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_condition, 1, 0, 0),
  [224] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_attr, 2, 0, 0),
  [226] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_attr, 2, 0, 0),
  [228] = {.entry = {.count = 1, .reusable = true}}, SHIFT(9),
  [230] = {.entry = {.count = 1, .reusable = true}}, SHIFT(82),
  [232] = {.entry = {.count = 1, .reusable = true}}, SHIFT(85),
  [234] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_attr, 1, 0, 0),
  [236] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_attr, 1, 0, 0),
  [238] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_attr, 3, 0, 0),
  [240] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_attr, 3, 0, 0),
  [242] = {.entry = {.count = 1, .reusable = true}}, SHIFT(146),
  [244] = {.entry = {.count = 1, .reusable = true}}, SHIFT(145),
  [246] = {.entry = {.count = 1, .reusable = false}}, SHIFT(145),
  [248] = {.entry = {.count = 1, .reusable = true}}, SHIFT(25),
  [250] = {.entry = {.count = 1, .reusable = true}}, SHIFT(46),
  [252] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(146),
  [255] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(145),
  [258] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(145),
  [261] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0),
  [263] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(46),
  [266] = {.entry = {.count = 1, .reusable = true}}, SHIFT(45),
  [268] = {.entry = {.count = 1, .reusable = false}}, SHIFT(8),
  [270] = {.entry = {.count = 1, .reusable = false}}, SHIFT(82),
  [272] = {.entry = {.count = 1, .reusable = false}}, SHIFT(85),
  [274] = {.entry = {.count = 1, .reusable = true}}, SHIFT(87),
  [276] = {.entry = {.count = 1, .reusable = true}}, SHIFT(124),
  [278] = {.entry = {.count = 1, .reusable = true}}, SHIFT(22),
  [280] = {.entry = {.count = 1, .reusable = true}}, SHIFT(63),
  [282] = {.entry = {.count = 1, .reusable = true}}, SHIFT(13),
  [284] = {.entry = {.count = 1, .reusable = true}}, SHIFT(17),
  [286] = {.entry = {.count = 1, .reusable = true}}, SHIFT(26),
  [288] = {.entry = {.count = 1, .reusable = true}}, SHIFT(21),
  [290] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_stylesheet_rule, 4, 0, 0),
  [292] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_stylesheet_rule, 4, 0, 0),
  [294] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_agent_node_repeat1, 2, 0, 0), SHIFT_REPEAT(124),
  [297] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_agent_node_repeat1, 2, 0, 0),
  [299] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_agent_node_repeat1, 2, 0, 0), SHIFT_REPEAT(63),
  [302] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_defaults_section_repeat1, 2, 0, 0), SHIFT_REPEAT(124),
  [305] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_defaults_section_repeat1, 2, 0, 0),
  [307] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_defaults_section_repeat1, 2, 0, 0), SHIFT_REPEAT(64),
  [310] = {.entry = {.count = 1, .reusable = true}}, SHIFT(19),
  [312] = {.entry = {.count = 1, .reusable = true}}, SHIFT(20),
  [314] = {.entry = {.count = 1, .reusable = true}}, SHIFT(15),
  [316] = {.entry = {.count = 1, .reusable = true}}, SHIFT(64),
  [318] = {.entry = {.count = 1, .reusable = true}}, SHIFT(60),
  [320] = {.entry = {.count = 1, .reusable = true}}, SHIFT(65),
  [322] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_edges_section_repeat1, 2, 0, 0), SHIFT_REPEAT(118),
  [325] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_edges_section_repeat1, 2, 0, 0),
  [327] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_edges_section_repeat1, 2, 0, 0), SHIFT_REPEAT(71),
  [330] = {.entry = {.count = 1, .reusable = true}}, SHIFT(61),
  [332] = {.entry = {.count = 1, .reusable = true}}, SHIFT(76),
  [334] = {.entry = {.count = 1, .reusable = true}}, SHIFT(58),
  [336] = {.entry = {.count = 1, .reusable = true}}, SHIFT(57),
  [338] = {.entry = {.count = 1, .reusable = true}}, SHIFT(68),
  [340] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 2, 0, 0), SHIFT_REPEAT(124),
  [343] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 2, 0, 0),
  [345] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 2, 0, 0), SHIFT_REPEAT(76),
  [348] = {.entry = {.count = 1, .reusable = true}}, SHIFT(59),
  [350] = {.entry = {.count = 1, .reusable = true}}, SHIFT(67),
  [352] = {.entry = {.count = 1, .reusable = true}}, SHIFT(56),
  [354] = {.entry = {.count = 1, .reusable = true}}, SHIFT(118),
  [356] = {.entry = {.count = 1, .reusable = true}}, SHIFT(16),
  [358] = {.entry = {.count = 1, .reusable = true}}, SHIFT(71),
  [360] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_string_repeat1, 2, 0, 0),
  [362] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_string_repeat1, 2, 0, 0), SHIFT_REPEAT(81),
  [365] = {.entry = {.count = 1, .reusable = false}}, SHIFT(2),
  [367] = {.entry = {.count = 1, .reusable = false}}, SHIFT(92),
  [369] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_block_content_repeat1, 2, 0, 0), SHIFT_REPEAT(83),
  [372] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_block_content_repeat1, 2, 0, 0),
  [374] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_string_repeat2, 2, 0, 0),
  [376] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_string_repeat2, 2, 0, 0), SHIFT_REPEAT(84),
  [379] = {.entry = {.count = 1, .reusable = false}}, SHIFT(93),
  [381] = {.entry = {.count = 1, .reusable = true}}, SHIFT(80),
  [383] = {.entry = {.count = 1, .reusable = true}}, SHIFT(94),
  [385] = {.entry = {.count = 1, .reusable = true}}, SHIFT(116),
  [387] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_identifier_list, 2, 0, 0),
  [389] = {.entry = {.count = 1, .reusable = true}}, SHIFT(72),
  [391] = {.entry = {.count = 1, .reusable = true}}, SHIFT(96),
  [393] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_identifier_list_repeat1, 2, 0, 0), SHIFT_REPEAT(116),
  [396] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_identifier_list_repeat1, 2, 0, 0),
  [398] = {.entry = {.count = 1, .reusable = false}}, SHIFT(3),
  [400] = {.entry = {.count = 1, .reusable = false}}, SHIFT(81),
  [402] = {.entry = {.count = 1, .reusable = false}}, SHIFT(84),
  [404] = {.entry = {.count = 1, .reusable = true}}, SHIFT(83),
  [406] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_block_content, 1, 0, 0),
  [408] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_identifier_list, 1, 0, 0),
  [410] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0),
  [412] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(96),
  [415] = {.entry = {.count = 1, .reusable = true}}, SHIFT(77),
  [417] = {.entry = {.count = 1, .reusable = true}}, SHIFT(24),
  [419] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_defaults_field, 3, 0, 0),
  [421] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_node_field, 3, 0, 0),
  [423] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_op, 1, 0, 0),
  [425] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 3, 0, 0),
  [427] = {.entry = {.count = 1, .reusable = true}}, SHIFT(23),
  [429] = {.entry = {.count = 1, .reusable = true}}, SHIFT(95),
  [431] = {.entry = {.count = 1, .reusable = true}}, SHIFT(141),
  [433] = {.entry = {.count = 1, .reusable = true}}, SHIFT(47),
  [435] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_workflow_decl, 5, 0, 0),
  [437] = {.entry = {.count = 1, .reusable = true}}, SHIFT(137),
  [439] = {.entry = {.count = 1, .reusable = true}}, SHIFT(143),
  [441] = {.entry = {.count = 1, .reusable = true}}, SHIFT(41),
  [443] = {.entry = {.count = 1, .reusable = true}}, SHIFT(144),
  [445] = {.entry = {.count = 1, .reusable = true}}, SHIFT(51),
  [447] = {.entry = {.count = 1, .reusable = true}}, SHIFT(120),
  [449] = {.entry = {.count = 1, .reusable = true}}, SHIFT(53),
  [451] = {.entry = {.count = 1, .reusable = true}}, SHIFT(101),
  [453] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 2, 0, 0),
  [455] = {.entry = {.count = 1, .reusable = true}}, SHIFT(142),
  [457] = {.entry = {.count = 1, .reusable = true}}, SHIFT(126),
  [459] = {.entry = {.count = 1, .reusable = true}}, SHIFT(74),
  [461] = {.entry = {.count = 1, .reusable = true}}, SHIFT(7),
  [463] = {.entry = {.count = 1, .reusable = true}}, SHIFT(134),
  [465] = {.entry = {.count = 1, .reusable = true}}, SHIFT(138),
  [467] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_field_name, 1, 0, 0),
  [469] = {.entry = {.count = 1, .reusable = true}}, SHIFT(52),
  [471] = {.entry = {.count = 1, .reusable = true}}, SHIFT(79),
  [473] = {.entry = {.count = 1, .reusable = true}}, SHIFT(107),
  [475] = {.entry = {.count = 1, .reusable = true}}, SHIFT(108),
  [477] = {.entry = {.count = 1, .reusable = true}}, SHIFT(55),
  [479] = {.entry = {.count = 1, .reusable = true}}, SHIFT(12),
  [481] = {.entry = {.count = 1, .reusable = true}},  ACCEPT_INPUT(),
  [483] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 1, 0, 0),
  [485] = {.entry = {.count = 1, .reusable = true}}, SHIFT(50),
  [487] = {.entry = {.count = 1, .reusable = true}}, SHIFT(6),
  [489] = {.entry = {.count = 1, .reusable = true}}, SHIFT(70),
  [491] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_selector, 2, 0, 0),
  [493] = {.entry = {.count = 1, .reusable = true}}, SHIFT(78),
  [495] = {.entry = {.count = 1, .reusable = true}}, SHIFT(105),
  [497] = {.entry = {.count = 1, .reusable = true}}, SHIFT(75),
  [499] = {.entry = {.count = 1, .reusable = true}}, SHIFT(135),
  [501] = {.entry = {.count = 1, .reusable = true}}, SHIFT(104),
  [503] = {.entry = {.count = 1, .reusable = true}}, SHIFT(31),
  [505] = {.entry = {.count = 1, .reusable = true}}, SHIFT(69),
  [507] = {.entry = {.count = 1, .reusable = true}}, SHIFT(73),
  [509] = {.entry = {.count = 1, .reusable = true}}, SHIFT(136),
  [511] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_selector, 1, 0, 0),
  [513] = {.entry = {.count = 1, .reusable = true}}, SHIFT(89),
  [515] = {.entry = {.count = 1, .reusable = true}}, SHIFT(86),
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
