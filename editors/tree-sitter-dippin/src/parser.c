#include "tree_sitter/parser.h"

#if defined(__GNUC__) || defined(__clang__)
#pragma GCC diagnostic ignored "-Wmissing-field-initializers"
#endif

#define LANGUAGE_VERSION 14
#define STATE_COUNT 155
#define LARGE_STATE_COUNT 2
#define SYMBOL_COUNT 107
#define ALIAS_COUNT 0
#define TOKEN_COUNT 56
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
  anon_sym_when = 21,
  anon_sym_on = 22,
  anon_sym_loop = 23,
  anon_sym_label = 24,
  anon_sym_choice = 25,
  anon_sym_weight = 26,
  anon_sym_restart = 27,
  anon_sym_override = 28,
  anon_sym_or = 29,
  anon_sym_and = 30,
  anon_sym_not = 31,
  anon_sym_EQ_EQ = 32,
  anon_sym_BANG_EQ = 33,
  anon_sym_EQ = 34,
  anon_sym_contains = 35,
  anon_sym_startswith = 36,
  anon_sym_endswith = 37,
  anon_sym_in = 38,
  anon_sym_DOT = 39,
  anon_sym_stylesheet = 40,
  anon_sym_STAR = 41,
  anon_sym_POUND = 42,
  sym_raw_inline = 43,
  sym_block_line = 44,
  anon_sym_COMMA = 45,
  anon_sym_DQUOTE = 46,
  aux_sym_string_token1 = 47,
  aux_sym_string_token2 = 48,
  anon_sym_SQUOTE = 49,
  aux_sym_string_token3 = 50,
  anon_sym_SQUOTE_SQUOTE = 51,
  sym_comment = 52,
  sym__indent = 53,
  sym__dedent = 54,
  sym__newline = 55,
  sym_source_file = 56,
  sym_version_decl = 57,
  sym_workflow_decl = 58,
  sym_workflow_body = 59,
  sym_workflow_field = 60,
  sym_defaults_section = 61,
  sym_defaults_field = 62,
  sym_node_decl = 63,
  sym_agent_node = 64,
  sym_human_node = 65,
  sym_tool_node = 66,
  sym_subgraph_node = 67,
  sym_conditional_node = 68,
  sym_manager_loop_node = 69,
  sym_parallel_node = 70,
  sym_fan_in_node = 71,
  sym_node_attr_block = 72,
  sym_node_field = 73,
  sym_edges_section = 74,
  sym_edge_entry = 75,
  sym_edge_attr = 76,
  sym_condition = 77,
  sym_or_expr = 78,
  sym_and_expr = 79,
  sym_compare_expr = 80,
  sym_compare_op = 81,
  sym_operand = 82,
  sym_variable = 83,
  sym_stylesheet_section = 84,
  sym_stylesheet_rule = 85,
  sym_selector = 86,
  sym_field_name = 87,
  sym_field_value = 88,
  sym_multiline_block = 89,
  sym_block_content = 90,
  sym_identifier_list = 91,
  sym_string = 92,
  aux_sym_source_file_repeat1 = 93,
  aux_sym_workflow_body_repeat1 = 94,
  aux_sym_defaults_section_repeat1 = 95,
  aux_sym_agent_node_repeat1 = 96,
  aux_sym_edges_section_repeat1 = 97,
  aux_sym_edge_entry_repeat1 = 98,
  aux_sym_or_expr_repeat1 = 99,
  aux_sym_and_expr_repeat1 = 100,
  aux_sym_stylesheet_section_repeat1 = 101,
  aux_sym_stylesheet_rule_repeat1 = 102,
  aux_sym_block_content_repeat1 = 103,
  aux_sym_identifier_list_repeat1 = 104,
  aux_sym_string_repeat1 = 105,
  aux_sym_string_repeat2 = 106,
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
      if (lookahead == '#') ADVANCE(33);
      if (lookahead == '\'') ADVANCE(32);
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') ADVANCE(34);
      if (lookahead != 0) ADVANCE(35);
      END_STATE();
    case 4:
      if (lookahead == '#') ADVANCE(23);
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') ADVANCE(22);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n') ADVANCE(23);
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
      if (lookahead == 'n') ADVANCE(25);
      if (lookahead == 'x') ADVANCE(26);
      END_STATE();
    case 5:
      if (lookahead == 'a') ADVANCE(27);
      END_STATE();
    case 6:
      if (lookahead == 'o') ADVANCE(28);
      END_STATE();
    case 7:
      if (lookahead == 'u') ADVANCE(29);
      END_STATE();
    case 8:
      if (lookahead == 'n') ADVANCE(30);
      END_STATE();
    case 9:
      if (lookahead == 'a') ADVANCE(31);
      if (lookahead == 'o') ADVANCE(32);
      END_STATE();
    case 10:
      if (lookahead == 'a') ADVANCE(33);
      END_STATE();
    case 11:
      if (lookahead == 'o') ADVANCE(34);
      END_STATE();
    case 12:
      if (lookahead == 'n') ADVANCE(35);
      if (lookahead == 'r') ADVANCE(36);
      if (lookahead == 'v') ADVANCE(37);
      END_STATE();
    case 13:
      if (lookahead == 'a') ADVANCE(38);
      END_STATE();
    case 14:
      if (lookahead == 'e') ADVANCE(39);
      END_STATE();
    case 15:
      if (lookahead == 't') ADVANCE(40);
      if (lookahead == 'u') ADVANCE(41);
      END_STATE();
    case 16:
      if (lookahead == 'o') ADVANCE(42);
      END_STATE();
    case 17:
      if (lookahead == 'e') ADVANCE(43);
      if (lookahead == 'h') ADVANCE(44);
      if (lookahead == 'o') ADVANCE(45);
      END_STATE();
    case 18:
      if (lookahead == 'e') ADVANCE(46);
      END_STATE();
    case 19:
      if (lookahead == 'd') ADVANCE(47);
      END_STATE();
    case 20:
      if (lookahead == 'o') ADVANCE(48);
      END_STATE();
    case 21:
      if (lookahead == 'n') ADVANCE(49);
      END_STATE();
    case 22:
      if (lookahead == 'f') ADVANCE(50);
      END_STATE();
    case 23:
      if (lookahead == 'p') ADVANCE(51);
      END_STATE();
    case 24:
      if (lookahead == 'g') ADVANCE(52);
      END_STATE();
    case 25:
      if (lookahead == 'd') ADVANCE(53);
      END_STATE();
    case 26:
      if (lookahead == 'i') ADVANCE(54);
      END_STATE();
    case 27:
      if (lookahead == 'n') ADVANCE(55);
      END_STATE();
    case 28:
      if (lookahead == 'a') ADVANCE(56);
      END_STATE();
    case 29:
      if (lookahead == 'm') ADVANCE(57);
      END_STATE();
    case 30:
      ACCEPT_TOKEN(anon_sym_in);
      END_STATE();
    case 31:
      if (lookahead == 'b') ADVANCE(58);
      END_STATE();
    case 32:
      if (lookahead == 'o') ADVANCE(59);
      END_STATE();
    case 33:
      if (lookahead == 'n') ADVANCE(60);
      END_STATE();
    case 34:
      if (lookahead == 't') ADVANCE(61);
      END_STATE();
    case 35:
      ACCEPT_TOKEN(anon_sym_on);
      END_STATE();
    case 36:
      ACCEPT_TOKEN(anon_sym_or);
      END_STATE();
    case 37:
      if (lookahead == 'e') ADVANCE(62);
      END_STATE();
    case 38:
      if (lookahead == 'r') ADVANCE(63);
      END_STATE();
    case 39:
      if (lookahead == 'q') ADVANCE(64);
      if (lookahead == 's') ADVANCE(65);
      END_STATE();
    case 40:
      if (lookahead == 'a') ADVANCE(66);
      if (lookahead == 'y') ADVANCE(67);
      END_STATE();
    case 41:
      if (lookahead == 'b') ADVANCE(68);
      END_STATE();
    case 42:
      if (lookahead == 'o') ADVANCE(69);
      END_STATE();
    case 43:
      if (lookahead == 'i') ADVANCE(70);
      END_STATE();
    case 44:
      if (lookahead == 'e') ADVANCE(71);
      END_STATE();
    case 45:
      if (lookahead == 'r') ADVANCE(72);
      END_STATE();
    case 46:
      if (lookahead == 'n') ADVANCE(73);
      END_STATE();
    case 47:
      ACCEPT_TOKEN(anon_sym_and);
      END_STATE();
    case 48:
      if (lookahead == 'i') ADVANCE(74);
      END_STATE();
    case 49:
      if (lookahead == 'd') ADVANCE(75);
      if (lookahead == 't') ADVANCE(76);
      END_STATE();
    case 50:
      if (lookahead == 'a') ADVANCE(77);
      END_STATE();
    case 51:
      ACCEPT_TOKEN(anon_sym_dip);
      END_STATE();
    case 52:
      if (lookahead == 'e') ADVANCE(78);
      END_STATE();
    case 53:
      if (lookahead == 's') ADVANCE(79);
      END_STATE();
    case 54:
      if (lookahead == 't') ADVANCE(80);
      END_STATE();
    case 55:
      if (lookahead == '_') ADVANCE(81);
      END_STATE();
    case 56:
      if (lookahead == 'l') ADVANCE(82);
      END_STATE();
    case 57:
      if (lookahead == 'a') ADVANCE(83);
      END_STATE();
    case 58:
      if (lookahead == 'e') ADVANCE(84);
      END_STATE();
    case 59:
      if (lookahead == 'p') ADVANCE(85);
      END_STATE();
    case 60:
      if (lookahead == 'a') ADVANCE(86);
      END_STATE();
    case 61:
      ACCEPT_TOKEN(anon_sym_not);
      END_STATE();
    case 62:
      if (lookahead == 'r') ADVANCE(87);
      END_STATE();
    case 63:
      if (lookahead == 'a') ADVANCE(88);
      END_STATE();
    case 64:
      if (lookahead == 'u') ADVANCE(89);
      END_STATE();
    case 65:
      if (lookahead == 't') ADVANCE(90);
      END_STATE();
    case 66:
      if (lookahead == 'r') ADVANCE(91);
      END_STATE();
    case 67:
      if (lookahead == 'l') ADVANCE(92);
      END_STATE();
    case 68:
      if (lookahead == 'g') ADVANCE(93);
      END_STATE();
    case 69:
      if (lookahead == 'l') ADVANCE(94);
      END_STATE();
    case 70:
      if (lookahead == 'g') ADVANCE(95);
      END_STATE();
    case 71:
      if (lookahead == 'n') ADVANCE(96);
      END_STATE();
    case 72:
      if (lookahead == 'k') ADVANCE(97);
      END_STATE();
    case 73:
      if (lookahead == 't') ADVANCE(98);
      END_STATE();
    case 74:
      if (lookahead == 'c') ADVANCE(99);
      END_STATE();
    case 75:
      if (lookahead == 'i') ADVANCE(100);
      END_STATE();
    case 76:
      if (lookahead == 'a') ADVANCE(101);
      END_STATE();
    case 77:
      if (lookahead == 'u') ADVANCE(102);
      END_STATE();
    case 78:
      if (lookahead == 's') ADVANCE(103);
      END_STATE();
    case 79:
      if (lookahead == 'w') ADVANCE(104);
      END_STATE();
    case 80:
      ACCEPT_TOKEN(anon_sym_exit);
      END_STATE();
    case 81:
      if (lookahead == 'i') ADVANCE(105);
      END_STATE();
    case 82:
      ACCEPT_TOKEN(anon_sym_goal);
      END_STATE();
    case 83:
      if (lookahead == 'n') ADVANCE(106);
      END_STATE();
    case 84:
      if (lookahead == 'l') ADVANCE(107);
      END_STATE();
    case 85:
      ACCEPT_TOKEN(anon_sym_loop);
      END_STATE();
    case 86:
      if (lookahead == 'g') ADVANCE(108);
      END_STATE();
    case 87:
      if (lookahead == 'r') ADVANCE(109);
      END_STATE();
    case 88:
      if (lookahead == 'l') ADVANCE(110);
      END_STATE();
    case 89:
      if (lookahead == 'i') ADVANCE(111);
      END_STATE();
    case 90:
      if (lookahead == 'a') ADVANCE(112);
      END_STATE();
    case 91:
      if (lookahead == 't') ADVANCE(113);
      END_STATE();
    case 92:
      if (lookahead == 'e') ADVANCE(114);
      END_STATE();
    case 93:
      if (lookahead == 'r') ADVANCE(115);
      END_STATE();
    case 94:
      ACCEPT_TOKEN(anon_sym_tool);
      END_STATE();
    case 95:
      if (lookahead == 'h') ADVANCE(116);
      END_STATE();
    case 96:
      ACCEPT_TOKEN(anon_sym_when);
      END_STATE();
    case 97:
      if (lookahead == 'f') ADVANCE(117);
      END_STATE();
    case 98:
      ACCEPT_TOKEN(anon_sym_agent);
      END_STATE();
    case 99:
      if (lookahead == 'e') ADVANCE(118);
      END_STATE();
    case 100:
      if (lookahead == 't') ADVANCE(119);
      END_STATE();
    case 101:
      if (lookahead == 'i') ADVANCE(120);
      END_STATE();
    case 102:
      if (lookahead == 'l') ADVANCE(121);
      END_STATE();
    case 103:
      ACCEPT_TOKEN(anon_sym_edges);
      END_STATE();
    case 104:
      if (lookahead == 'i') ADVANCE(122);
      END_STATE();
    case 105:
      if (lookahead == 'n') ADVANCE(123);
      END_STATE();
    case 106:
      ACCEPT_TOKEN(anon_sym_human);
      END_STATE();
    case 107:
      ACCEPT_TOKEN(anon_sym_label);
      END_STATE();
    case 108:
      if (lookahead == 'e') ADVANCE(124);
      END_STATE();
    case 109:
      if (lookahead == 'i') ADVANCE(125);
      END_STATE();
    case 110:
      if (lookahead == 'l') ADVANCE(126);
      END_STATE();
    case 111:
      if (lookahead == 'r') ADVANCE(127);
      END_STATE();
    case 112:
      if (lookahead == 'r') ADVANCE(128);
      END_STATE();
    case 113:
      ACCEPT_TOKEN(anon_sym_start);
      if (lookahead == 's') ADVANCE(129);
      END_STATE();
    case 114:
      if (lookahead == 's') ADVANCE(130);
      END_STATE();
    case 115:
      if (lookahead == 'a') ADVANCE(131);
      END_STATE();
    case 116:
      if (lookahead == 't') ADVANCE(132);
      END_STATE();
    case 117:
      if (lookahead == 'l') ADVANCE(133);
      END_STATE();
    case 118:
      ACCEPT_TOKEN(anon_sym_choice);
      END_STATE();
    case 119:
      if (lookahead == 'i') ADVANCE(134);
      END_STATE();
    case 120:
      if (lookahead == 'n') ADVANCE(135);
      END_STATE();
    case 121:
      if (lookahead == 't') ADVANCE(136);
      END_STATE();
    case 122:
      if (lookahead == 't') ADVANCE(137);
      END_STATE();
    case 123:
      ACCEPT_TOKEN(anon_sym_fan_in);
      END_STATE();
    case 124:
      if (lookahead == 'r') ADVANCE(138);
      END_STATE();
    case 125:
      if (lookahead == 'd') ADVANCE(139);
      END_STATE();
    case 126:
      if (lookahead == 'e') ADVANCE(140);
      END_STATE();
    case 127:
      if (lookahead == 'e') ADVANCE(141);
      END_STATE();
    case 128:
      if (lookahead == 't') ADVANCE(142);
      END_STATE();
    case 129:
      if (lookahead == 'w') ADVANCE(143);
      END_STATE();
    case 130:
      if (lookahead == 'h') ADVANCE(144);
      END_STATE();
    case 131:
      if (lookahead == 'p') ADVANCE(145);
      END_STATE();
    case 132:
      ACCEPT_TOKEN(anon_sym_weight);
      END_STATE();
    case 133:
      if (lookahead == 'o') ADVANCE(146);
      END_STATE();
    case 134:
      if (lookahead == 'o') ADVANCE(147);
      END_STATE();
    case 135:
      if (lookahead == 's') ADVANCE(148);
      END_STATE();
    case 136:
      if (lookahead == 's') ADVANCE(149);
      END_STATE();
    case 137:
      if (lookahead == 'h') ADVANCE(150);
      END_STATE();
    case 138:
      if (lookahead == '_') ADVANCE(151);
      END_STATE();
    case 139:
      if (lookahead == 'e') ADVANCE(152);
      END_STATE();
    case 140:
      if (lookahead == 'l') ADVANCE(153);
      END_STATE();
    case 141:
      if (lookahead == 's') ADVANCE(154);
      END_STATE();
    case 142:
      ACCEPT_TOKEN(anon_sym_restart);
      END_STATE();
    case 143:
      if (lookahead == 'i') ADVANCE(155);
      END_STATE();
    case 144:
      if (lookahead == 'e') ADVANCE(156);
      END_STATE();
    case 145:
      if (lookahead == 'h') ADVANCE(157);
      END_STATE();
    case 146:
      if (lookahead == 'w') ADVANCE(158);
      END_STATE();
    case 147:
      if (lookahead == 'n') ADVANCE(159);
      END_STATE();
    case 148:
      ACCEPT_TOKEN(anon_sym_contains);
      END_STATE();
    case 149:
      ACCEPT_TOKEN(anon_sym_defaults);
      END_STATE();
    case 150:
      ACCEPT_TOKEN(anon_sym_endswith);
      END_STATE();
    case 151:
      if (lookahead == 'l') ADVANCE(160);
      END_STATE();
    case 152:
      ACCEPT_TOKEN(anon_sym_override);
      END_STATE();
    case 153:
      ACCEPT_TOKEN(anon_sym_parallel);
      END_STATE();
    case 154:
      ACCEPT_TOKEN(anon_sym_requires);
      END_STATE();
    case 155:
      if (lookahead == 't') ADVANCE(161);
      END_STATE();
    case 156:
      if (lookahead == 'e') ADVANCE(162);
      END_STATE();
    case 157:
      ACCEPT_TOKEN(anon_sym_subgraph);
      END_STATE();
    case 158:
      ACCEPT_TOKEN(anon_sym_workflow);
      END_STATE();
    case 159:
      if (lookahead == 'a') ADVANCE(163);
      END_STATE();
    case 160:
      if (lookahead == 'o') ADVANCE(164);
      END_STATE();
    case 161:
      if (lookahead == 'h') ADVANCE(165);
      END_STATE();
    case 162:
      if (lookahead == 't') ADVANCE(166);
      END_STATE();
    case 163:
      if (lookahead == 'l') ADVANCE(167);
      END_STATE();
    case 164:
      if (lookahead == 'o') ADVANCE(168);
      END_STATE();
    case 165:
      ACCEPT_TOKEN(anon_sym_startswith);
      END_STATE();
    case 166:
      ACCEPT_TOKEN(anon_sym_stylesheet);
      END_STATE();
    case 167:
      ACCEPT_TOKEN(anon_sym_conditional);
      END_STATE();
    case 168:
      if (lookahead == 'p') ADVANCE(169);
      END_STATE();
    case 169:
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
  [47] = {.lex_state = 0, .external_lex_state = 2},
  [48] = {.lex_state = 9},
  [49] = {.lex_state = 9},
  [50] = {.lex_state = 1, .external_lex_state = 4},
  [51] = {.lex_state = 1, .external_lex_state = 4},
  [52] = {.lex_state = 9},
  [53] = {.lex_state = 1, .external_lex_state = 4},
  [54] = {.lex_state = 1, .external_lex_state = 4},
  [55] = {.lex_state = 1, .external_lex_state = 4},
  [56] = {.lex_state = 9},
  [57] = {.lex_state = 9, .external_lex_state = 3},
  [58] = {.lex_state = 9, .external_lex_state = 3},
  [59] = {.lex_state = 9, .external_lex_state = 3},
  [60] = {.lex_state = 9, .external_lex_state = 3},
  [61] = {.lex_state = 9, .external_lex_state = 3},
  [62] = {.lex_state = 9, .external_lex_state = 3},
  [63] = {.lex_state = 9, .external_lex_state = 3},
  [64] = {.lex_state = 9, .external_lex_state = 3},
  [65] = {.lex_state = 9, .external_lex_state = 3},
  [66] = {.lex_state = 0, .external_lex_state = 3},
  [67] = {.lex_state = 9, .external_lex_state = 3},
  [68] = {.lex_state = 9},
  [69] = {.lex_state = 9, .external_lex_state = 2},
  [70] = {.lex_state = 9, .external_lex_state = 3},
  [71] = {.lex_state = 9, .external_lex_state = 2},
  [72] = {.lex_state = 9, .external_lex_state = 2},
  [73] = {.lex_state = 9, .external_lex_state = 2},
  [74] = {.lex_state = 9, .external_lex_state = 2},
  [75] = {.lex_state = 9, .external_lex_state = 2},
  [76] = {.lex_state = 9, .external_lex_state = 2},
  [77] = {.lex_state = 9, .external_lex_state = 3},
  [78] = {.lex_state = 9, .external_lex_state = 2},
  [79] = {.lex_state = 9, .external_lex_state = 2},
  [80] = {.lex_state = 9, .external_lex_state = 3},
  [81] = {.lex_state = 9, .external_lex_state = 3},
  [82] = {.lex_state = 9, .external_lex_state = 2},
  [83] = {.lex_state = 3},
  [84] = {.lex_state = 2},
  [85] = {.lex_state = 4, .external_lex_state = 3},
  [86] = {.lex_state = 9, .external_lex_state = 2},
  [87] = {.lex_state = 9, .external_lex_state = 5},
  [88] = {.lex_state = 9, .external_lex_state = 5},
  [89] = {.lex_state = 9, .external_lex_state = 2},
  [90] = {.lex_state = 4, .external_lex_state = 3},
  [91] = {.lex_state = 2},
  [92] = {.lex_state = 4, .external_lex_state = 2},
  [93] = {.lex_state = 2},
  [94] = {.lex_state = 3},
  [95] = {.lex_state = 9, .external_lex_state = 5},
  [96] = {.lex_state = 3},
  [97] = {.lex_state = 9, .external_lex_state = 3},
  [98] = {.lex_state = 9, .external_lex_state = 3},
  [99] = {.lex_state = 9, .external_lex_state = 5},
  [100] = {.lex_state = 9, .external_lex_state = 5},
  [101] = {.lex_state = 9, .external_lex_state = 3},
  [102] = {.lex_state = 9, .external_lex_state = 5},
  [103] = {.lex_state = 9},
  [104] = {.lex_state = 9},
  [105] = {.lex_state = 9},
  [106] = {.lex_state = 9},
  [107] = {.lex_state = 9},
  [108] = {.lex_state = 9, .external_lex_state = 4},
  [109] = {.lex_state = 9},
  [110] = {.lex_state = 9},
  [111] = {.lex_state = 9, .external_lex_state = 4},
  [112] = {.lex_state = 9},
  [113] = {.lex_state = 9},
  [114] = {.lex_state = 9},
  [115] = {.lex_state = 9},
  [116] = {.lex_state = 9, .external_lex_state = 4},
  [117] = {.lex_state = 9},
  [118] = {.lex_state = 9, .external_lex_state = 4},
  [119] = {.lex_state = 9},
  [120] = {.lex_state = 9, .external_lex_state = 4},
  [121] = {.lex_state = 9},
  [122] = {.lex_state = 9},
  [123] = {.lex_state = 9},
  [124] = {.lex_state = 9},
  [125] = {.lex_state = 9},
  [126] = {.lex_state = 9, .external_lex_state = 4},
  [127] = {.lex_state = 9},
  [128] = {.lex_state = 9},
  [129] = {.lex_state = 9},
  [130] = {.lex_state = 9},
  [131] = {.lex_state = 9, .external_lex_state = 4},
  [132] = {.lex_state = 9, .external_lex_state = 4},
  [133] = {.lex_state = 9, .external_lex_state = 6},
  [134] = {.lex_state = 9},
  [135] = {.lex_state = 9, .external_lex_state = 2},
  [136] = {.lex_state = 9},
  [137] = {.lex_state = 9, .external_lex_state = 4},
  [138] = {.lex_state = 9},
  [139] = {.lex_state = 9},
  [140] = {.lex_state = 9, .external_lex_state = 4},
  [141] = {.lex_state = 9},
  [142] = {.lex_state = 9},
  [143] = {.lex_state = 9, .external_lex_state = 4},
  [144] = {.lex_state = 9},
  [145] = {.lex_state = 9},
  [146] = {.lex_state = 9},
  [147] = {.lex_state = 9},
  [148] = {.lex_state = 9, .external_lex_state = 4},
  [149] = {.lex_state = 9, .external_lex_state = 6},
  [150] = {.lex_state = 9},
  [151] = {.lex_state = 9, .external_lex_state = 4},
  [152] = {.lex_state = 9},
  [153] = {.lex_state = 9},
  [154] = {.lex_state = 9},
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
    [sym_source_file] = STATE(144),
    [sym_version_decl] = STATE(106),
    [sym_workflow_decl] = STATE(147),
    [aux_sym_source_file_repeat1] = STATE(69),
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
    ACTIONS(13), 32,
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
  [44] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(19), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(17), 32,
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
  [88] = 17,
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
    STATE(13), 8,
      sym_agent_node,
      sym_human_node,
      sym_tool_node,
      sym_subgraph_node,
      sym_conditional_node,
      sym_manager_loop_node,
      sym_parallel_node,
      sym_fan_in_node,
  [155] = 17,
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
    STATE(149), 1,
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
    STATE(13), 8,
      sym_agent_node,
      sym_human_node,
      sym_tool_node,
      sym_subgraph_node,
      sym_conditional_node,
      sym_manager_loop_node,
      sym_parallel_node,
      sym_fan_in_node,
  [222] = 17,
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
    STATE(13), 8,
      sym_agent_node,
      sym_human_node,
      sym_tool_node,
      sym_subgraph_node,
      sym_conditional_node,
      sym_manager_loop_node,
      sym_parallel_node,
      sym_fan_in_node,
  [289] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(94), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(92), 24,
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
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_stylesheet,
      sym_identifier,
  [323] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(98), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(96), 24,
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
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      anon_sym_stylesheet,
      sym_identifier,
  [357] = 7,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(102), 1,
      anon_sym_not,
    STATE(68), 1,
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
    ACTIONS(100), 11,
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
  [395] = 4,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(114), 1,
      anon_sym_DOT,
    ACTIONS(112), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(110), 17,
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
  [427] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(112), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(110), 17,
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
  [456] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(118), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(116), 17,
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
  [485] = 2,
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
  [508] = 2,
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
  [531] = 2,
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
  [554] = 2,
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
  [577] = 2,
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
  [600] = 2,
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
  [623] = 2,
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
  [646] = 2,
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
  [669] = 2,
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
  [692] = 2,
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
  [715] = 2,
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
  [738] = 2,
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
  [761] = 2,
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
  [784] = 2,
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
  [807] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(150), 1,
      anon_sym_and,
    STATE(29), 1,
      aux_sym_and_expr_repeat1,
    ACTIONS(152), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(148), 10,
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
  [833] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(150), 1,
      anon_sym_and,
    STATE(27), 1,
      aux_sym_and_expr_repeat1,
    ACTIONS(156), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(154), 10,
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
  [859] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(160), 1,
      anon_sym_and,
    STATE(29), 1,
      aux_sym_and_expr_repeat1,
    ACTIONS(163), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(158), 10,
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
  [885] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(167), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(165), 11,
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
  [906] = 8,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(169), 1,
      sym_identifier,
    ACTIONS(171), 1,
      anon_sym_when,
    ACTIONS(173), 1,
      anon_sym_on,
    ACTIONS(175), 1,
      anon_sym_loop,
    ACTIONS(179), 2,
      sym__dedent,
      sym__newline,
    STATE(34), 2,
      sym_edge_attr,
      aux_sym_edge_entry_repeat1,
    ACTIONS(177), 5,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
  [937] = 8,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(171), 1,
      anon_sym_when,
    ACTIONS(173), 1,
      anon_sym_on,
    ACTIONS(175), 1,
      anon_sym_loop,
    ACTIONS(181), 1,
      sym_identifier,
    ACTIONS(183), 2,
      sym__dedent,
      sym__newline,
    STATE(31), 2,
      sym_edge_attr,
      aux_sym_edge_entry_repeat1,
    ACTIONS(177), 5,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
  [968] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(187), 1,
      anon_sym_or,
    STATE(36), 1,
      aux_sym_or_expr_repeat1,
    ACTIONS(189), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(185), 9,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [993] = 8,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(191), 1,
      sym_identifier,
    ACTIONS(193), 1,
      anon_sym_when,
    ACTIONS(196), 1,
      anon_sym_on,
    ACTIONS(199), 1,
      anon_sym_loop,
    ACTIONS(205), 2,
      sym__dedent,
      sym__newline,
    STATE(34), 2,
      sym_edge_attr,
      aux_sym_edge_entry_repeat1,
    ACTIONS(202), 5,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
  [1024] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(163), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(158), 11,
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
  [1045] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(187), 1,
      anon_sym_or,
    STATE(37), 1,
      aux_sym_or_expr_repeat1,
    ACTIONS(209), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(207), 9,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [1070] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(213), 1,
      anon_sym_or,
    STATE(37), 1,
      aux_sym_or_expr_repeat1,
    ACTIONS(216), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(211), 9,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [1095] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(220), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(218), 11,
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
  [1116] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(216), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(211), 10,
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
  [1136] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(224), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(222), 9,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [1155] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(228), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(226), 9,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [1174] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(232), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(230), 9,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [1193] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(236), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(234), 9,
      anon_sym_when,
      anon_sym_on,
      anon_sym_loop,
      anon_sym_label,
      anon_sym_choice,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_override,
      sym_identifier,
  [1212] = 10,
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
    STATE(28), 1,
      sym_compare_expr,
    STATE(33), 1,
      sym_and_expr,
    STATE(40), 1,
      sym_condition,
    STATE(42), 1,
      sym_or_expr,
    STATE(11), 2,
      sym_variable,
      sym_string,
  [1244] = 8,
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
    STATE(132), 1,
      sym_selector,
    ACTIONS(244), 2,
      anon_sym_STAR,
      sym_identifier,
    STATE(46), 2,
      sym_stylesheet_rule,
      aux_sym_stylesheet_section_repeat1,
  [1271] = 8,
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
    STATE(132), 1,
      sym_selector,
    ACTIONS(254), 2,
      anon_sym_STAR,
      sym_identifier,
    STATE(46), 2,
      sym_stylesheet_rule,
      aux_sym_stylesheet_section_repeat1,
  [1298] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(246), 1,
      anon_sym_DOT,
    ACTIONS(248), 1,
      anon_sym_POUND,
    ACTIONS(268), 1,
      sym__newline,
    STATE(132), 1,
      sym_selector,
    ACTIONS(244), 2,
      anon_sym_STAR,
      sym_identifier,
    STATE(45), 2,
      sym_stylesheet_rule,
      aux_sym_stylesheet_section_repeat1,
  [1322] = 8,
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
    STATE(28), 1,
      sym_compare_expr,
    STATE(39), 1,
      sym_and_expr,
    STATE(11), 2,
      sym_variable,
      sym_string,
  [1348] = 4,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(106), 1,
      anon_sym_EQ,
    STATE(56), 1,
      sym_compare_op,
    ACTIONS(104), 6,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
      anon_sym_contains,
      anon_sym_startswith,
      anon_sym_endswith,
      anon_sym_in,
  [1366] = 7,
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
    STATE(101), 1,
      sym_field_value,
    STATE(8), 2,
      sym_multiline_block,
      sym_string,
  [1389] = 7,
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
    STATE(8), 2,
      sym_multiline_block,
      sym_string,
  [1412] = 7,
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
    STATE(35), 1,
      sym_compare_expr,
    STATE(11), 2,
      sym_variable,
      sym_string,
  [1435] = 7,
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
    STATE(16), 1,
      sym_field_value,
    STATE(8), 2,
      sym_multiline_block,
      sym_string,
  [1458] = 7,
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
    STATE(97), 1,
      sym_field_value,
    STATE(8), 2,
      sym_multiline_block,
      sym_string,
  [1481] = 7,
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
    STATE(43), 1,
      sym_field_value,
    STATE(8), 2,
      sym_multiline_block,
      sym_string,
  [1504] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(238), 1,
      sym_identifier,
    ACTIONS(240), 1,
      anon_sym_DQUOTE,
    ACTIONS(242), 1,
      anon_sym_SQUOTE,
    STATE(30), 1,
      sym_operand,
    STATE(11), 2,
      sym_variable,
      sym_string,
  [1524] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(278), 1,
      sym_identifier,
    ACTIONS(281), 1,
      sym__dedent,
    ACTIONS(283), 1,
      sym__newline,
    STATE(153), 1,
      sym_field_name,
    STATE(57), 2,
      sym_defaults_field,
      aux_sym_defaults_section_repeat1,
  [1544] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(286), 1,
      sym_identifier,
    ACTIONS(288), 1,
      sym__dedent,
    ACTIONS(290), 1,
      sym__newline,
    STATE(115), 1,
      sym_field_name,
    STATE(65), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1564] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(286), 1,
      sym_identifier,
    ACTIONS(290), 1,
      sym__newline,
    ACTIONS(292), 1,
      sym__dedent,
    STATE(115), 1,
      sym_field_name,
    STATE(65), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1584] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(286), 1,
      sym_identifier,
    ACTIONS(290), 1,
      sym__newline,
    ACTIONS(294), 1,
      sym__dedent,
    STATE(115), 1,
      sym_field_name,
    STATE(65), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1604] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(286), 1,
      sym_identifier,
    ACTIONS(290), 1,
      sym__newline,
    ACTIONS(296), 1,
      sym__dedent,
    STATE(115), 1,
      sym_field_name,
    STATE(65), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1624] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(286), 1,
      sym_identifier,
    ACTIONS(290), 1,
      sym__newline,
    ACTIONS(298), 1,
      sym__dedent,
    STATE(115), 1,
      sym_field_name,
    STATE(65), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1644] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(286), 1,
      sym_identifier,
    ACTIONS(290), 1,
      sym__newline,
    ACTIONS(300), 1,
      sym__dedent,
    STATE(115), 1,
      sym_field_name,
    STATE(65), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1664] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(286), 1,
      sym_identifier,
    ACTIONS(290), 1,
      sym__newline,
    ACTIONS(302), 1,
      sym__dedent,
    STATE(115), 1,
      sym_field_name,
    STATE(65), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1684] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(304), 1,
      sym_identifier,
    ACTIONS(307), 1,
      sym__dedent,
    ACTIONS(309), 1,
      sym__newline,
    STATE(115), 1,
      sym_field_name,
    STATE(65), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1704] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(314), 1,
      anon_sym_POUND,
    ACTIONS(312), 5,
      sym__dedent,
      sym__newline,
      anon_sym_DOT,
      anon_sym_STAR,
      sym_identifier,
  [1718] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(286), 1,
      sym_identifier,
    ACTIONS(316), 1,
      sym__dedent,
    ACTIONS(318), 1,
      sym__newline,
    STATE(153), 1,
      sym_field_name,
    STATE(57), 2,
      sym_defaults_field,
      aux_sym_defaults_section_repeat1,
  [1738] = 6,
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
  [1758] = 7,
    ACTIONS(5), 1,
      anon_sym_dip,
    ACTIONS(7), 1,
      anon_sym_workflow,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(320), 1,
      sym__newline,
    STATE(82), 1,
      aux_sym_source_file_repeat1,
    STATE(104), 1,
      sym_version_decl,
    STATE(114), 1,
      sym_workflow_decl,
  [1780] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(322), 1,
      sym_identifier,
    ACTIONS(325), 1,
      sym__dedent,
    ACTIONS(327), 1,
      sym__newline,
    STATE(70), 2,
      sym_edge_entry,
      aux_sym_edges_section_repeat1,
  [1797] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(286), 1,
      sym_identifier,
    ACTIONS(330), 1,
      sym__newline,
    STATE(115), 1,
      sym_field_name,
    STATE(58), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1814] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(286), 1,
      sym_identifier,
    ACTIONS(332), 1,
      sym__newline,
    STATE(115), 1,
      sym_field_name,
    STATE(59), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1831] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(286), 1,
      sym_identifier,
    ACTIONS(334), 1,
      sym__newline,
    STATE(153), 1,
      sym_field_name,
    STATE(67), 2,
      sym_defaults_field,
      aux_sym_defaults_section_repeat1,
  [1848] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(286), 1,
      sym_identifier,
    ACTIONS(336), 1,
      sym__newline,
    STATE(115), 1,
      sym_field_name,
    STATE(60), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1865] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(286), 1,
      sym_identifier,
    ACTIONS(338), 1,
      sym__newline,
    STATE(115), 1,
      sym_field_name,
    STATE(64), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1882] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(286), 1,
      sym_identifier,
    ACTIONS(340), 1,
      sym__newline,
    STATE(115), 1,
      sym_field_name,
    STATE(61), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1899] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(286), 1,
      sym_identifier,
    ACTIONS(342), 1,
      sym__dedent,
    ACTIONS(344), 1,
      sym__newline,
    STATE(81), 1,
      aux_sym_stylesheet_rule_repeat1,
    STATE(121), 1,
      sym_field_name,
  [1918] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(286), 1,
      sym_identifier,
    ACTIONS(346), 1,
      sym__newline,
    STATE(115), 1,
      sym_field_name,
    STATE(62), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1935] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(286), 1,
      sym_identifier,
    ACTIONS(348), 1,
      sym__newline,
    STATE(115), 1,
      sym_field_name,
    STATE(63), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1952] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(350), 1,
      sym_identifier,
    ACTIONS(352), 1,
      sym__dedent,
    ACTIONS(354), 1,
      sym__newline,
    STATE(70), 2,
      sym_edge_entry,
      aux_sym_edges_section_repeat1,
  [1969] = 6,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(356), 1,
      sym_identifier,
    ACTIONS(359), 1,
      sym__dedent,
    ACTIONS(361), 1,
      sym__newline,
    STATE(81), 1,
      aux_sym_stylesheet_rule_repeat1,
    STATE(121), 1,
      sym_field_name,
  [1988] = 4,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(366), 1,
      sym__newline,
    STATE(82), 1,
      aux_sym_source_file_repeat1,
    ACTIONS(364), 2,
      anon_sym_dip,
      anon_sym_workflow,
  [2002] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(369), 1,
      anon_sym_SQUOTE,
    STATE(83), 1,
      aux_sym_string_repeat2,
    ACTIONS(371), 2,
      aux_sym_string_token3,
      anon_sym_SQUOTE_SQUOTE,
  [2016] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(374), 1,
      anon_sym_DQUOTE,
    STATE(84), 1,
      aux_sym_string_repeat1,
    ACTIONS(376), 2,
      aux_sym_string_token1,
      aux_sym_string_token2,
  [2030] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(382), 1,
      sym__dedent,
    STATE(85), 1,
      aux_sym_block_content_repeat1,
    ACTIONS(379), 2,
      sym__newline,
      sym_block_line,
  [2044] = 5,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(286), 1,
      sym_identifier,
    ACTIONS(384), 1,
      sym__newline,
    STATE(77), 1,
      aux_sym_stylesheet_rule_repeat1,
    STATE(121), 1,
      sym_field_name,
  [2060] = 4,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(386), 1,
      anon_sym_COMMA,
    STATE(88), 1,
      aux_sym_identifier_list_repeat1,
    ACTIONS(388), 2,
      sym__indent,
      sym__newline,
  [2074] = 4,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(390), 1,
      anon_sym_COMMA,
    STATE(88), 1,
      aux_sym_identifier_list_repeat1,
    ACTIONS(393), 2,
      sym__indent,
      sym__newline,
  [2088] = 4,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(350), 1,
      sym_identifier,
    ACTIONS(395), 1,
      sym__newline,
    STATE(80), 2,
      sym_edge_entry,
      aux_sym_edges_section_repeat1,
  [2102] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(399), 1,
      sym__dedent,
    STATE(85), 1,
      aux_sym_block_content_repeat1,
    ACTIONS(397), 2,
      sym__newline,
      sym_block_line,
  [2116] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(401), 1,
      anon_sym_DQUOTE,
    STATE(84), 1,
      aux_sym_string_repeat1,
    ACTIONS(403), 2,
      aux_sym_string_token1,
      aux_sym_string_token2,
  [2130] = 4,
    ACTIONS(3), 1,
      sym_comment,
    STATE(90), 1,
      aux_sym_block_content_repeat1,
    STATE(133), 1,
      sym_block_content,
    ACTIONS(405), 2,
      sym__newline,
      sym_block_line,
  [2144] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(407), 1,
      anon_sym_DQUOTE,
    STATE(91), 1,
      aux_sym_string_repeat1,
    ACTIONS(409), 2,
      aux_sym_string_token1,
      aux_sym_string_token2,
  [2158] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(407), 1,
      anon_sym_SQUOTE,
    STATE(96), 1,
      aux_sym_string_repeat2,
    ACTIONS(411), 2,
      aux_sym_string_token3,
      anon_sym_SQUOTE_SQUOTE,
  [2172] = 4,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(386), 1,
      anon_sym_COMMA,
    STATE(87), 1,
      aux_sym_identifier_list_repeat1,
    ACTIONS(413), 2,
      sym__indent,
      sym__newline,
  [2186] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(401), 1,
      anon_sym_SQUOTE,
    STATE(83), 1,
      aux_sym_string_repeat2,
    ACTIONS(415), 2,
      aux_sym_string_token3,
      anon_sym_SQUOTE_SQUOTE,
  [2200] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(417), 3,
      sym__dedent,
      sym__newline,
      sym_identifier,
  [2209] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(419), 3,
      sym__dedent,
      sym__newline,
      sym_identifier,
  [2218] = 4,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(421), 1,
      sym__indent,
    ACTIONS(423), 1,
      sym__newline,
    STATE(25), 1,
      sym_node_attr_block,
  [2231] = 4,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(421), 1,
      sym__indent,
    ACTIONS(425), 1,
      sym__newline,
    STATE(20), 1,
      sym_node_attr_block,
  [2244] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(427), 3,
      sym__dedent,
      sym__newline,
      sym_identifier,
  [2253] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(393), 3,
      sym__indent,
      sym__newline,
      anon_sym_COMMA,
  [2262] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(429), 3,
      sym_identifier,
      anon_sym_DQUOTE,
      anon_sym_SQUOTE,
  [2271] = 3,
    ACTIONS(7), 1,
      anon_sym_workflow,
    ACTIONS(9), 1,
      sym_comment,
    STATE(128), 1,
      sym_workflow_decl,
  [2281] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(431), 1,
      sym_identifier,
    STATE(100), 1,
      sym_identifier_list,
  [2291] = 3,
    ACTIONS(7), 1,
      anon_sym_workflow,
    ACTIONS(9), 1,
      sym_comment,
    STATE(114), 1,
      sym_workflow_decl,
  [2301] = 3,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(431), 1,
      sym_identifier,
    STATE(99), 1,
      sym_identifier_list,
  [2311] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(433), 1,
      sym__indent,
  [2318] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(435), 1,
      sym_identifier,
  [2325] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(437), 1,
      sym_identifier,
  [2332] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(439), 1,
      sym__indent,
  [2339] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(441), 1,
      sym_identifier,
  [2346] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(443), 1,
      sym_identifier,
  [2353] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(445), 1,
      ts_builtin_sym_end,
  [2360] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(447), 1,
      anon_sym_COLON,
  [2367] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(449), 1,
      sym__indent,
  [2374] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(451), 1,
      sym_identifier,
  [2381] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(453), 1,
      sym__indent,
  [2388] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(455), 1,
      anon_sym_COLON,
  [2395] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(457), 1,
      sym__indent,
  [2402] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(459), 1,
      anon_sym_COLON,
  [2409] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(461), 1,
      sym_identifier,
  [2416] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(463), 1,
      sym_identifier,
  [2423] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(465), 1,
      sym_identifier,
  [2430] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(467), 1,
      sym_identifier,
  [2437] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(469), 1,
      sym__indent,
  [2444] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(471), 1,
      anon_sym_DASH_GT,
  [2451] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(473), 1,
      ts_builtin_sym_end,
  [2458] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(475), 1,
      sym_identifier,
  [2465] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(477), 1,
      sym_identifier,
  [2472] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(479), 1,
      sym__indent,
  [2479] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(481), 1,
      sym__indent,
  [2486] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(483), 1,
      sym__dedent,
  [2493] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(485), 1,
      anon_sym_COLON,
  [2500] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(487), 1,
      sym__newline,
  [2507] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(489), 1,
      sym_identifier,
  [2514] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(491), 1,
      sym__indent,
  [2521] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(493), 1,
      ts_builtin_sym_end,
  [2528] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(495), 1,
      anon_sym_DASH_GT,
  [2535] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(497), 1,
      sym__indent,
  [2542] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(499), 1,
      sym_identifier,
  [2549] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(501), 1,
      sym_identifier,
  [2556] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(503), 1,
      sym__indent,
  [2563] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(505), 1,
      ts_builtin_sym_end,
  [2570] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(507), 1,
      sym_identifier,
  [2577] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(509), 1,
      anon_sym_COLON,
  [2584] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(511), 1,
      ts_builtin_sym_end,
  [2591] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(513), 1,
      sym__indent,
  [2598] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(515), 1,
      sym__dedent,
  [2605] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(517), 1,
      anon_sym_workflow,
  [2612] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(519), 1,
      sym__indent,
  [2619] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(521), 1,
      anon_sym_COLON,
  [2626] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(523), 1,
      anon_sym_COLON,
  [2633] = 2,
    ACTIONS(9), 1,
      sym_comment,
    ACTIONS(525), 1,
      anon_sym_LT_DASH,
};

static const uint32_t ts_small_parse_table_map[] = {
  [SMALL_STATE(2)] = 0,
  [SMALL_STATE(3)] = 44,
  [SMALL_STATE(4)] = 88,
  [SMALL_STATE(5)] = 155,
  [SMALL_STATE(6)] = 222,
  [SMALL_STATE(7)] = 289,
  [SMALL_STATE(8)] = 323,
  [SMALL_STATE(9)] = 357,
  [SMALL_STATE(10)] = 395,
  [SMALL_STATE(11)] = 427,
  [SMALL_STATE(12)] = 456,
  [SMALL_STATE(13)] = 485,
  [SMALL_STATE(14)] = 508,
  [SMALL_STATE(15)] = 531,
  [SMALL_STATE(16)] = 554,
  [SMALL_STATE(17)] = 577,
  [SMALL_STATE(18)] = 600,
  [SMALL_STATE(19)] = 623,
  [SMALL_STATE(20)] = 646,
  [SMALL_STATE(21)] = 669,
  [SMALL_STATE(22)] = 692,
  [SMALL_STATE(23)] = 715,
  [SMALL_STATE(24)] = 738,
  [SMALL_STATE(25)] = 761,
  [SMALL_STATE(26)] = 784,
  [SMALL_STATE(27)] = 807,
  [SMALL_STATE(28)] = 833,
  [SMALL_STATE(29)] = 859,
  [SMALL_STATE(30)] = 885,
  [SMALL_STATE(31)] = 906,
  [SMALL_STATE(32)] = 937,
  [SMALL_STATE(33)] = 968,
  [SMALL_STATE(34)] = 993,
  [SMALL_STATE(35)] = 1024,
  [SMALL_STATE(36)] = 1045,
  [SMALL_STATE(37)] = 1070,
  [SMALL_STATE(38)] = 1095,
  [SMALL_STATE(39)] = 1116,
  [SMALL_STATE(40)] = 1136,
  [SMALL_STATE(41)] = 1155,
  [SMALL_STATE(42)] = 1174,
  [SMALL_STATE(43)] = 1193,
  [SMALL_STATE(44)] = 1212,
  [SMALL_STATE(45)] = 1244,
  [SMALL_STATE(46)] = 1271,
  [SMALL_STATE(47)] = 1298,
  [SMALL_STATE(48)] = 1322,
  [SMALL_STATE(49)] = 1348,
  [SMALL_STATE(50)] = 1366,
  [SMALL_STATE(51)] = 1389,
  [SMALL_STATE(52)] = 1412,
  [SMALL_STATE(53)] = 1435,
  [SMALL_STATE(54)] = 1458,
  [SMALL_STATE(55)] = 1481,
  [SMALL_STATE(56)] = 1504,
  [SMALL_STATE(57)] = 1524,
  [SMALL_STATE(58)] = 1544,
  [SMALL_STATE(59)] = 1564,
  [SMALL_STATE(60)] = 1584,
  [SMALL_STATE(61)] = 1604,
  [SMALL_STATE(62)] = 1624,
  [SMALL_STATE(63)] = 1644,
  [SMALL_STATE(64)] = 1664,
  [SMALL_STATE(65)] = 1684,
  [SMALL_STATE(66)] = 1704,
  [SMALL_STATE(67)] = 1718,
  [SMALL_STATE(68)] = 1738,
  [SMALL_STATE(69)] = 1758,
  [SMALL_STATE(70)] = 1780,
  [SMALL_STATE(71)] = 1797,
  [SMALL_STATE(72)] = 1814,
  [SMALL_STATE(73)] = 1831,
  [SMALL_STATE(74)] = 1848,
  [SMALL_STATE(75)] = 1865,
  [SMALL_STATE(76)] = 1882,
  [SMALL_STATE(77)] = 1899,
  [SMALL_STATE(78)] = 1918,
  [SMALL_STATE(79)] = 1935,
  [SMALL_STATE(80)] = 1952,
  [SMALL_STATE(81)] = 1969,
  [SMALL_STATE(82)] = 1988,
  [SMALL_STATE(83)] = 2002,
  [SMALL_STATE(84)] = 2016,
  [SMALL_STATE(85)] = 2030,
  [SMALL_STATE(86)] = 2044,
  [SMALL_STATE(87)] = 2060,
  [SMALL_STATE(88)] = 2074,
  [SMALL_STATE(89)] = 2088,
  [SMALL_STATE(90)] = 2102,
  [SMALL_STATE(91)] = 2116,
  [SMALL_STATE(92)] = 2130,
  [SMALL_STATE(93)] = 2144,
  [SMALL_STATE(94)] = 2158,
  [SMALL_STATE(95)] = 2172,
  [SMALL_STATE(96)] = 2186,
  [SMALL_STATE(97)] = 2200,
  [SMALL_STATE(98)] = 2209,
  [SMALL_STATE(99)] = 2218,
  [SMALL_STATE(100)] = 2231,
  [SMALL_STATE(101)] = 2244,
  [SMALL_STATE(102)] = 2253,
  [SMALL_STATE(103)] = 2262,
  [SMALL_STATE(104)] = 2271,
  [SMALL_STATE(105)] = 2281,
  [SMALL_STATE(106)] = 2291,
  [SMALL_STATE(107)] = 2301,
  [SMALL_STATE(108)] = 2311,
  [SMALL_STATE(109)] = 2318,
  [SMALL_STATE(110)] = 2325,
  [SMALL_STATE(111)] = 2332,
  [SMALL_STATE(112)] = 2339,
  [SMALL_STATE(113)] = 2346,
  [SMALL_STATE(114)] = 2353,
  [SMALL_STATE(115)] = 2360,
  [SMALL_STATE(116)] = 2367,
  [SMALL_STATE(117)] = 2374,
  [SMALL_STATE(118)] = 2381,
  [SMALL_STATE(119)] = 2388,
  [SMALL_STATE(120)] = 2395,
  [SMALL_STATE(121)] = 2402,
  [SMALL_STATE(122)] = 2409,
  [SMALL_STATE(123)] = 2416,
  [SMALL_STATE(124)] = 2423,
  [SMALL_STATE(125)] = 2430,
  [SMALL_STATE(126)] = 2437,
  [SMALL_STATE(127)] = 2444,
  [SMALL_STATE(128)] = 2451,
  [SMALL_STATE(129)] = 2458,
  [SMALL_STATE(130)] = 2465,
  [SMALL_STATE(131)] = 2472,
  [SMALL_STATE(132)] = 2479,
  [SMALL_STATE(133)] = 2486,
  [SMALL_STATE(134)] = 2493,
  [SMALL_STATE(135)] = 2500,
  [SMALL_STATE(136)] = 2507,
  [SMALL_STATE(137)] = 2514,
  [SMALL_STATE(138)] = 2521,
  [SMALL_STATE(139)] = 2528,
  [SMALL_STATE(140)] = 2535,
  [SMALL_STATE(141)] = 2542,
  [SMALL_STATE(142)] = 2549,
  [SMALL_STATE(143)] = 2556,
  [SMALL_STATE(144)] = 2563,
  [SMALL_STATE(145)] = 2570,
  [SMALL_STATE(146)] = 2577,
  [SMALL_STATE(147)] = 2584,
  [SMALL_STATE(148)] = 2591,
  [SMALL_STATE(149)] = 2598,
  [SMALL_STATE(150)] = 2605,
  [SMALL_STATE(151)] = 2612,
  [SMALL_STATE(152)] = 2619,
  [SMALL_STATE(153)] = 2626,
  [SMALL_STATE(154)] = 2633,
};

static const TSParseActionEntry ts_parse_actions[] = {
  [0] = {.entry = {.count = 0, .reusable = false}},
  [1] = {.entry = {.count = 1, .reusable = false}}, RECOVER(),
  [3] = {.entry = {.count = 1, .reusable = false}}, SHIFT_EXTRA(),
  [5] = {.entry = {.count = 1, .reusable = true}}, SHIFT(125),
  [7] = {.entry = {.count = 1, .reusable = true}}, SHIFT(145),
  [9] = {.entry = {.count = 1, .reusable = true}}, SHIFT_EXTRA(),
  [11] = {.entry = {.count = 1, .reusable = true}}, SHIFT(69),
  [13] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_string, 2, 0, 0),
  [15] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_string, 2, 0, 0),
  [17] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_string, 3, 0, 0),
  [19] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_string, 3, 0, 0),
  [21] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(134),
  [24] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(143),
  [27] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(112),
  [30] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(141),
  [33] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(142),
  [36] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(110),
  [39] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(123),
  [42] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(124),
  [45] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(129),
  [48] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(130),
  [51] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(140),
  [54] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(146),
  [57] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0),
  [59] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(4),
  [62] = {.entry = {.count = 1, .reusable = true}}, SHIFT(134),
  [64] = {.entry = {.count = 1, .reusable = true}}, SHIFT(143),
  [66] = {.entry = {.count = 1, .reusable = true}}, SHIFT(112),
  [68] = {.entry = {.count = 1, .reusable = true}}, SHIFT(141),
  [70] = {.entry = {.count = 1, .reusable = true}}, SHIFT(142),
  [72] = {.entry = {.count = 1, .reusable = true}}, SHIFT(110),
  [74] = {.entry = {.count = 1, .reusable = true}}, SHIFT(123),
  [76] = {.entry = {.count = 1, .reusable = true}}, SHIFT(124),
  [78] = {.entry = {.count = 1, .reusable = true}}, SHIFT(129),
  [80] = {.entry = {.count = 1, .reusable = true}}, SHIFT(130),
  [82] = {.entry = {.count = 1, .reusable = true}}, SHIFT(140),
  [84] = {.entry = {.count = 1, .reusable = true}}, SHIFT(146),
  [86] = {.entry = {.count = 1, .reusable = true}}, SHIFT(6),
  [88] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_workflow_body, 1, 0, 0),
  [90] = {.entry = {.count = 1, .reusable = true}}, SHIFT(4),
  [92] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_multiline_block, 3, 0, 0),
  [94] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_multiline_block, 3, 0, 0),
  [96] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_field_value, 1, 0, 0),
  [98] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_field_value, 1, 0, 0),
  [100] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_compare_expr, 1, 0, 0),
  [102] = {.entry = {.count = 1, .reusable = false}}, SHIFT(49),
  [104] = {.entry = {.count = 1, .reusable = true}}, SHIFT(103),
  [106] = {.entry = {.count = 1, .reusable = false}}, SHIFT(103),
  [108] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_expr, 1, 0, 0),
  [110] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_operand, 1, 0, 0),
  [112] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_operand, 1, 0, 0),
  [114] = {.entry = {.count = 1, .reusable = true}}, SHIFT(136),
  [116] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_variable, 3, 0, 0),
  [118] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_variable, 3, 0, 0),
  [120] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_node_decl, 1, 0, 0),
  [122] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edges_section, 4, 0, 0),
  [124] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_defaults_section, 4, 0, 0),
  [126] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_workflow_field, 3, 0, 0),
  [128] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_agent_node, 5, 0, 0),
  [130] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_conditional_node, 5, 0, 0),
  [132] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_manager_loop_node, 5, 0, 0),
  [134] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_fan_in_node, 5, 0, 0),
  [136] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_stylesheet_section, 5, 0, 0),
  [138] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_node_attr_block, 3, 0, 0),
  [140] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_subgraph_node, 5, 0, 0),
  [142] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_tool_node, 5, 0, 0),
  [144] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_parallel_node, 5, 0, 0),
  [146] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_human_node, 5, 0, 0),
  [148] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_and_expr, 2, 0, 0),
  [150] = {.entry = {.count = 1, .reusable = false}}, SHIFT(52),
  [152] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_and_expr, 2, 0, 0),
  [154] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_and_expr, 1, 0, 0),
  [156] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_and_expr, 1, 0, 0),
  [158] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_and_expr_repeat1, 2, 0, 0),
  [160] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_and_expr_repeat1, 2, 0, 0), SHIFT_REPEAT(52),
  [163] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_and_expr_repeat1, 2, 0, 0),
  [165] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_compare_expr, 4, 0, 0),
  [167] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_expr, 4, 0, 0),
  [169] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_entry, 4, 0, 0),
  [171] = {.entry = {.count = 1, .reusable = false}}, SHIFT(44),
  [173] = {.entry = {.count = 1, .reusable = false}}, SHIFT(117),
  [175] = {.entry = {.count = 1, .reusable = false}}, SHIFT(41),
  [177] = {.entry = {.count = 1, .reusable = false}}, SHIFT(119),
  [179] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_entry, 4, 0, 0),
  [181] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_entry, 3, 0, 0),
  [183] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_entry, 3, 0, 0),
  [185] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_or_expr, 1, 0, 0),
  [187] = {.entry = {.count = 1, .reusable = false}}, SHIFT(48),
  [189] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_or_expr, 1, 0, 0),
  [191] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0),
  [193] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0), SHIFT_REPEAT(44),
  [196] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0), SHIFT_REPEAT(117),
  [199] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0), SHIFT_REPEAT(41),
  [202] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0), SHIFT_REPEAT(119),
  [205] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0),
  [207] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_or_expr, 2, 0, 0),
  [209] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_or_expr, 2, 0, 0),
  [211] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_or_expr_repeat1, 2, 0, 0),
  [213] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_or_expr_repeat1, 2, 0, 0), SHIFT_REPEAT(48),
  [216] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_or_expr_repeat1, 2, 0, 0),
  [218] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_compare_expr, 3, 0, 0),
  [220] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_expr, 3, 0, 0),
  [222] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_attr, 2, 0, 0),
  [224] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_attr, 2, 0, 0),
  [226] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_attr, 1, 0, 0),
  [228] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_attr, 1, 0, 0),
  [230] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_condition, 1, 0, 0),
  [232] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_condition, 1, 0, 0),
  [234] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_attr, 3, 0, 0),
  [236] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_attr, 3, 0, 0),
  [238] = {.entry = {.count = 1, .reusable = true}}, SHIFT(10),
  [240] = {.entry = {.count = 1, .reusable = true}}, SHIFT(93),
  [242] = {.entry = {.count = 1, .reusable = true}}, SHIFT(94),
  [244] = {.entry = {.count = 1, .reusable = true}}, SHIFT(131),
  [246] = {.entry = {.count = 1, .reusable = true}}, SHIFT(122),
  [248] = {.entry = {.count = 1, .reusable = false}}, SHIFT(122),
  [250] = {.entry = {.count = 1, .reusable = true}}, SHIFT(21),
  [252] = {.entry = {.count = 1, .reusable = true}}, SHIFT(46),
  [254] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(131),
  [257] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(122),
  [260] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(122),
  [263] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0),
  [265] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(46),
  [268] = {.entry = {.count = 1, .reusable = true}}, SHIFT(45),
  [270] = {.entry = {.count = 1, .reusable = false}}, SHIFT(8),
  [272] = {.entry = {.count = 1, .reusable = false}}, SHIFT(93),
  [274] = {.entry = {.count = 1, .reusable = false}}, SHIFT(94),
  [276] = {.entry = {.count = 1, .reusable = true}}, SHIFT(92),
  [278] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_defaults_section_repeat1, 2, 0, 0), SHIFT_REPEAT(152),
  [281] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_defaults_section_repeat1, 2, 0, 0),
  [283] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_defaults_section_repeat1, 2, 0, 0), SHIFT_REPEAT(57),
  [286] = {.entry = {.count = 1, .reusable = true}}, SHIFT(152),
  [288] = {.entry = {.count = 1, .reusable = true}}, SHIFT(17),
  [290] = {.entry = {.count = 1, .reusable = true}}, SHIFT(65),
  [292] = {.entry = {.count = 1, .reusable = true}}, SHIFT(26),
  [294] = {.entry = {.count = 1, .reusable = true}}, SHIFT(24),
  [296] = {.entry = {.count = 1, .reusable = true}}, SHIFT(23),
  [298] = {.entry = {.count = 1, .reusable = true}}, SHIFT(18),
  [300] = {.entry = {.count = 1, .reusable = true}}, SHIFT(19),
  [302] = {.entry = {.count = 1, .reusable = true}}, SHIFT(22),
  [304] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_agent_node_repeat1, 2, 0, 0), SHIFT_REPEAT(152),
  [307] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_agent_node_repeat1, 2, 0, 0),
  [309] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_agent_node_repeat1, 2, 0, 0), SHIFT_REPEAT(65),
  [312] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_stylesheet_rule, 4, 0, 0),
  [314] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_stylesheet_rule, 4, 0, 0),
  [316] = {.entry = {.count = 1, .reusable = true}}, SHIFT(15),
  [318] = {.entry = {.count = 1, .reusable = true}}, SHIFT(57),
  [320] = {.entry = {.count = 1, .reusable = true}}, SHIFT(82),
  [322] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_edges_section_repeat1, 2, 0, 0), SHIFT_REPEAT(139),
  [325] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_edges_section_repeat1, 2, 0, 0),
  [327] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_edges_section_repeat1, 2, 0, 0), SHIFT_REPEAT(70),
  [330] = {.entry = {.count = 1, .reusable = true}}, SHIFT(58),
  [332] = {.entry = {.count = 1, .reusable = true}}, SHIFT(59),
  [334] = {.entry = {.count = 1, .reusable = true}}, SHIFT(67),
  [336] = {.entry = {.count = 1, .reusable = true}}, SHIFT(60),
  [338] = {.entry = {.count = 1, .reusable = true}}, SHIFT(64),
  [340] = {.entry = {.count = 1, .reusable = true}}, SHIFT(61),
  [342] = {.entry = {.count = 1, .reusable = true}}, SHIFT(66),
  [344] = {.entry = {.count = 1, .reusable = true}}, SHIFT(81),
  [346] = {.entry = {.count = 1, .reusable = true}}, SHIFT(62),
  [348] = {.entry = {.count = 1, .reusable = true}}, SHIFT(63),
  [350] = {.entry = {.count = 1, .reusable = true}}, SHIFT(139),
  [352] = {.entry = {.count = 1, .reusable = true}}, SHIFT(14),
  [354] = {.entry = {.count = 1, .reusable = true}}, SHIFT(70),
  [356] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 2, 0, 0), SHIFT_REPEAT(152),
  [359] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 2, 0, 0),
  [361] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 2, 0, 0), SHIFT_REPEAT(81),
  [364] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0),
  [366] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(82),
  [369] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_string_repeat2, 2, 0, 0),
  [371] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_string_repeat2, 2, 0, 0), SHIFT_REPEAT(83),
  [374] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_string_repeat1, 2, 0, 0),
  [376] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_string_repeat1, 2, 0, 0), SHIFT_REPEAT(84),
  [379] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_block_content_repeat1, 2, 0, 0), SHIFT_REPEAT(85),
  [382] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_block_content_repeat1, 2, 0, 0),
  [384] = {.entry = {.count = 1, .reusable = true}}, SHIFT(77),
  [386] = {.entry = {.count = 1, .reusable = true}}, SHIFT(113),
  [388] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_identifier_list, 2, 0, 0),
  [390] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_identifier_list_repeat1, 2, 0, 0), SHIFT_REPEAT(113),
  [393] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_identifier_list_repeat1, 2, 0, 0),
  [395] = {.entry = {.count = 1, .reusable = true}}, SHIFT(80),
  [397] = {.entry = {.count = 1, .reusable = true}}, SHIFT(85),
  [399] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_block_content, 1, 0, 0),
  [401] = {.entry = {.count = 1, .reusable = false}}, SHIFT(3),
  [403] = {.entry = {.count = 1, .reusable = false}}, SHIFT(84),
  [405] = {.entry = {.count = 1, .reusable = true}}, SHIFT(90),
  [407] = {.entry = {.count = 1, .reusable = false}}, SHIFT(2),
  [409] = {.entry = {.count = 1, .reusable = false}}, SHIFT(91),
  [411] = {.entry = {.count = 1, .reusable = false}}, SHIFT(96),
  [413] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_identifier_list, 1, 0, 0),
  [415] = {.entry = {.count = 1, .reusable = false}}, SHIFT(83),
  [417] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 3, 0, 0),
  [419] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_defaults_field, 3, 0, 0),
  [421] = {.entry = {.count = 1, .reusable = true}}, SHIFT(75),
  [423] = {.entry = {.count = 1, .reusable = true}}, SHIFT(25),
  [425] = {.entry = {.count = 1, .reusable = true}}, SHIFT(20),
  [427] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_node_field, 3, 0, 0),
  [429] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_op, 1, 0, 0),
  [431] = {.entry = {.count = 1, .reusable = true}}, SHIFT(95),
  [433] = {.entry = {.count = 1, .reusable = true}}, SHIFT(78),
  [435] = {.entry = {.count = 1, .reusable = true}}, SHIFT(32),
  [437] = {.entry = {.count = 1, .reusable = true}}, SHIFT(120),
  [439] = {.entry = {.count = 1, .reusable = true}}, SHIFT(71),
  [441] = {.entry = {.count = 1, .reusable = true}}, SHIFT(111),
  [443] = {.entry = {.count = 1, .reusable = true}}, SHIFT(102),
  [445] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 2, 0, 0),
  [447] = {.entry = {.count = 1, .reusable = true}}, SHIFT(50),
  [449] = {.entry = {.count = 1, .reusable = true}}, SHIFT(72),
  [451] = {.entry = {.count = 1, .reusable = true}}, SHIFT(40),
  [453] = {.entry = {.count = 1, .reusable = true}}, SHIFT(74),
  [455] = {.entry = {.count = 1, .reusable = true}}, SHIFT(55),
  [457] = {.entry = {.count = 1, .reusable = true}}, SHIFT(76),
  [459] = {.entry = {.count = 1, .reusable = true}}, SHIFT(54),
  [461] = {.entry = {.count = 1, .reusable = true}}, SHIFT(148),
  [463] = {.entry = {.count = 1, .reusable = true}}, SHIFT(108),
  [465] = {.entry = {.count = 1, .reusable = true}}, SHIFT(126),
  [467] = {.entry = {.count = 1, .reusable = true}}, SHIFT(135),
  [469] = {.entry = {.count = 1, .reusable = true}}, SHIFT(79),
  [471] = {.entry = {.count = 1, .reusable = true}}, SHIFT(107),
  [473] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 3, 0, 0),
  [475] = {.entry = {.count = 1, .reusable = true}}, SHIFT(127),
  [477] = {.entry = {.count = 1, .reusable = true}}, SHIFT(154),
  [479] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_selector, 1, 0, 0),
  [481] = {.entry = {.count = 1, .reusable = true}}, SHIFT(86),
  [483] = {.entry = {.count = 1, .reusable = true}}, SHIFT(7),
  [485] = {.entry = {.count = 1, .reusable = true}}, SHIFT(53),
  [487] = {.entry = {.count = 1, .reusable = true}}, SHIFT(150),
  [489] = {.entry = {.count = 1, .reusable = true}}, SHIFT(12),
  [491] = {.entry = {.count = 1, .reusable = true}}, SHIFT(47),
  [493] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_workflow_decl, 5, 0, 0),
  [495] = {.entry = {.count = 1, .reusable = true}}, SHIFT(109),
  [497] = {.entry = {.count = 1, .reusable = true}}, SHIFT(89),
  [499] = {.entry = {.count = 1, .reusable = true}}, SHIFT(116),
  [501] = {.entry = {.count = 1, .reusable = true}}, SHIFT(118),
  [503] = {.entry = {.count = 1, .reusable = true}}, SHIFT(73),
  [505] = {.entry = {.count = 1, .reusable = true}},  ACCEPT_INPUT(),
  [507] = {.entry = {.count = 1, .reusable = true}}, SHIFT(151),
  [509] = {.entry = {.count = 1, .reusable = true}}, SHIFT(137),
  [511] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 1, 0, 0),
  [513] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_selector, 2, 0, 0),
  [515] = {.entry = {.count = 1, .reusable = true}}, SHIFT(138),
  [517] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_version_decl, 3, 0, 0),
  [519] = {.entry = {.count = 1, .reusable = true}}, SHIFT(5),
  [521] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_field_name, 1, 0, 0),
  [523] = {.entry = {.count = 1, .reusable = true}}, SHIFT(51),
  [525] = {.entry = {.count = 1, .reusable = true}}, SHIFT(105),
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
