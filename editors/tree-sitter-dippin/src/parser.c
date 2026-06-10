#include "tree_sitter/parser.h"

#if defined(__GNUC__) || defined(__clang__)
#pragma GCC diagnostic ignored "-Wmissing-field-initializers"
#endif

#define LANGUAGE_VERSION 14
#define STATE_COUNT 147
#define LARGE_STATE_COUNT 2
#define SYMBOL_COUNT 101
#define ALIAS_COUNT 0
#define TOKEN_COUNT 51
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
  anon_sym_label = 21,
  anon_sym_weight = 22,
  anon_sym_restart = 23,
  anon_sym_or = 24,
  anon_sym_and = 25,
  anon_sym_not = 26,
  anon_sym_EQ_EQ = 27,
  anon_sym_BANG_EQ = 28,
  anon_sym_EQ = 29,
  anon_sym_contains = 30,
  anon_sym_startswith = 31,
  anon_sym_endswith = 32,
  anon_sym_in = 33,
  anon_sym_DOT = 34,
  anon_sym_stylesheet = 35,
  anon_sym_STAR = 36,
  anon_sym_POUND = 37,
  sym_raw_inline = 38,
  sym_block_line = 39,
  anon_sym_COMMA = 40,
  anon_sym_DQUOTE = 41,
  aux_sym_string_token1 = 42,
  aux_sym_string_token2 = 43,
  anon_sym_SQUOTE = 44,
  aux_sym_string_token3 = 45,
  anon_sym_SQUOTE_SQUOTE = 46,
  sym_comment = 47,
  sym__indent = 48,
  sym__dedent = 49,
  sym__newline = 50,
  sym_source_file = 51,
  sym_workflow_decl = 52,
  sym_workflow_body = 53,
  sym_workflow_field = 54,
  sym_defaults_section = 55,
  sym_defaults_field = 56,
  sym_node_decl = 57,
  sym_agent_node = 58,
  sym_human_node = 59,
  sym_tool_node = 60,
  sym_subgraph_node = 61,
  sym_conditional_node = 62,
  sym_manager_loop_node = 63,
  sym_parallel_node = 64,
  sym_fan_in_node = 65,
  sym_node_attr_block = 66,
  sym_node_field = 67,
  sym_edges_section = 68,
  sym_edge_entry = 69,
  sym_edge_attr = 70,
  sym_condition = 71,
  sym_or_expr = 72,
  sym_and_expr = 73,
  sym_compare_expr = 74,
  sym_compare_op = 75,
  sym_operand = 76,
  sym_variable = 77,
  sym_stylesheet_section = 78,
  sym_stylesheet_rule = 79,
  sym_selector = 80,
  sym_field_name = 81,
  sym_field_value = 82,
  sym_multiline_block = 83,
  sym_block_content = 84,
  sym_identifier_list = 85,
  sym_string = 86,
  aux_sym_source_file_repeat1 = 87,
  aux_sym_workflow_body_repeat1 = 88,
  aux_sym_defaults_section_repeat1 = 89,
  aux_sym_agent_node_repeat1 = 90,
  aux_sym_edges_section_repeat1 = 91,
  aux_sym_edge_entry_repeat1 = 92,
  aux_sym_or_expr_repeat1 = 93,
  aux_sym_and_expr_repeat1 = 94,
  aux_sym_stylesheet_section_repeat1 = 95,
  aux_sym_stylesheet_rule_repeat1 = 96,
  aux_sym_block_content_repeat1 = 97,
  aux_sym_identifier_list_repeat1 = 98,
  aux_sym_string_repeat1 = 99,
  aux_sym_string_repeat2 = 100,
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
  [anon_sym_label] = "label",
  [anon_sym_weight] = "weight",
  [anon_sym_restart] = "restart",
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
  [anon_sym_label] = anon_sym_label,
  [anon_sym_weight] = anon_sym_weight,
  [anon_sym_restart] = anon_sym_restart,
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
      if (lookahead == 'r') ADVANCE(32);
      END_STATE();
    case 13:
      if (lookahead == 'a') ADVANCE(33);
      END_STATE();
    case 14:
      if (lookahead == 'e') ADVANCE(34);
      END_STATE();
    case 15:
      if (lookahead == 't') ADVANCE(35);
      if (lookahead == 'u') ADVANCE(36);
      END_STATE();
    case 16:
      if (lookahead == 'o') ADVANCE(37);
      END_STATE();
    case 17:
      if (lookahead == 'e') ADVANCE(38);
      if (lookahead == 'h') ADVANCE(39);
      if (lookahead == 'o') ADVANCE(40);
      END_STATE();
    case 18:
      if (lookahead == 'e') ADVANCE(41);
      END_STATE();
    case 19:
      if (lookahead == 'd') ADVANCE(42);
      END_STATE();
    case 20:
      if (lookahead == 'n') ADVANCE(43);
      END_STATE();
    case 21:
      if (lookahead == 'f') ADVANCE(44);
      END_STATE();
    case 22:
      if (lookahead == 'g') ADVANCE(45);
      END_STATE();
    case 23:
      if (lookahead == 'd') ADVANCE(46);
      END_STATE();
    case 24:
      if (lookahead == 'i') ADVANCE(47);
      END_STATE();
    case 25:
      if (lookahead == 'n') ADVANCE(48);
      END_STATE();
    case 26:
      if (lookahead == 'a') ADVANCE(49);
      END_STATE();
    case 27:
      if (lookahead == 'm') ADVANCE(50);
      END_STATE();
    case 28:
      ACCEPT_TOKEN(anon_sym_in);
      END_STATE();
    case 29:
      if (lookahead == 'b') ADVANCE(51);
      END_STATE();
    case 30:
      if (lookahead == 'n') ADVANCE(52);
      END_STATE();
    case 31:
      if (lookahead == 't') ADVANCE(53);
      END_STATE();
    case 32:
      ACCEPT_TOKEN(anon_sym_or);
      END_STATE();
    case 33:
      if (lookahead == 'r') ADVANCE(54);
      END_STATE();
    case 34:
      if (lookahead == 'q') ADVANCE(55);
      if (lookahead == 's') ADVANCE(56);
      END_STATE();
    case 35:
      if (lookahead == 'a') ADVANCE(57);
      if (lookahead == 'y') ADVANCE(58);
      END_STATE();
    case 36:
      if (lookahead == 'b') ADVANCE(59);
      END_STATE();
    case 37:
      if (lookahead == 'o') ADVANCE(60);
      END_STATE();
    case 38:
      if (lookahead == 'i') ADVANCE(61);
      END_STATE();
    case 39:
      if (lookahead == 'e') ADVANCE(62);
      END_STATE();
    case 40:
      if (lookahead == 'r') ADVANCE(63);
      END_STATE();
    case 41:
      if (lookahead == 'n') ADVANCE(64);
      END_STATE();
    case 42:
      ACCEPT_TOKEN(anon_sym_and);
      END_STATE();
    case 43:
      if (lookahead == 'd') ADVANCE(65);
      if (lookahead == 't') ADVANCE(66);
      END_STATE();
    case 44:
      if (lookahead == 'a') ADVANCE(67);
      END_STATE();
    case 45:
      if (lookahead == 'e') ADVANCE(68);
      END_STATE();
    case 46:
      if (lookahead == 's') ADVANCE(69);
      END_STATE();
    case 47:
      if (lookahead == 't') ADVANCE(70);
      END_STATE();
    case 48:
      if (lookahead == '_') ADVANCE(71);
      END_STATE();
    case 49:
      if (lookahead == 'l') ADVANCE(72);
      END_STATE();
    case 50:
      if (lookahead == 'a') ADVANCE(73);
      END_STATE();
    case 51:
      if (lookahead == 'e') ADVANCE(74);
      END_STATE();
    case 52:
      if (lookahead == 'a') ADVANCE(75);
      END_STATE();
    case 53:
      ACCEPT_TOKEN(anon_sym_not);
      END_STATE();
    case 54:
      if (lookahead == 'a') ADVANCE(76);
      END_STATE();
    case 55:
      if (lookahead == 'u') ADVANCE(77);
      END_STATE();
    case 56:
      if (lookahead == 't') ADVANCE(78);
      END_STATE();
    case 57:
      if (lookahead == 'r') ADVANCE(79);
      END_STATE();
    case 58:
      if (lookahead == 'l') ADVANCE(80);
      END_STATE();
    case 59:
      if (lookahead == 'g') ADVANCE(81);
      END_STATE();
    case 60:
      if (lookahead == 'l') ADVANCE(82);
      END_STATE();
    case 61:
      if (lookahead == 'g') ADVANCE(83);
      END_STATE();
    case 62:
      if (lookahead == 'n') ADVANCE(84);
      END_STATE();
    case 63:
      if (lookahead == 'k') ADVANCE(85);
      END_STATE();
    case 64:
      if (lookahead == 't') ADVANCE(86);
      END_STATE();
    case 65:
      if (lookahead == 'i') ADVANCE(87);
      END_STATE();
    case 66:
      if (lookahead == 'a') ADVANCE(88);
      END_STATE();
    case 67:
      if (lookahead == 'u') ADVANCE(89);
      END_STATE();
    case 68:
      if (lookahead == 's') ADVANCE(90);
      END_STATE();
    case 69:
      if (lookahead == 'w') ADVANCE(91);
      END_STATE();
    case 70:
      ACCEPT_TOKEN(anon_sym_exit);
      END_STATE();
    case 71:
      if (lookahead == 'i') ADVANCE(92);
      END_STATE();
    case 72:
      ACCEPT_TOKEN(anon_sym_goal);
      END_STATE();
    case 73:
      if (lookahead == 'n') ADVANCE(93);
      END_STATE();
    case 74:
      if (lookahead == 'l') ADVANCE(94);
      END_STATE();
    case 75:
      if (lookahead == 'g') ADVANCE(95);
      END_STATE();
    case 76:
      if (lookahead == 'l') ADVANCE(96);
      END_STATE();
    case 77:
      if (lookahead == 'i') ADVANCE(97);
      END_STATE();
    case 78:
      if (lookahead == 'a') ADVANCE(98);
      END_STATE();
    case 79:
      if (lookahead == 't') ADVANCE(99);
      END_STATE();
    case 80:
      if (lookahead == 'e') ADVANCE(100);
      END_STATE();
    case 81:
      if (lookahead == 'r') ADVANCE(101);
      END_STATE();
    case 82:
      ACCEPT_TOKEN(anon_sym_tool);
      END_STATE();
    case 83:
      if (lookahead == 'h') ADVANCE(102);
      END_STATE();
    case 84:
      ACCEPT_TOKEN(anon_sym_when);
      END_STATE();
    case 85:
      if (lookahead == 'f') ADVANCE(103);
      END_STATE();
    case 86:
      ACCEPT_TOKEN(anon_sym_agent);
      END_STATE();
    case 87:
      if (lookahead == 't') ADVANCE(104);
      END_STATE();
    case 88:
      if (lookahead == 'i') ADVANCE(105);
      END_STATE();
    case 89:
      if (lookahead == 'l') ADVANCE(106);
      END_STATE();
    case 90:
      ACCEPT_TOKEN(anon_sym_edges);
      END_STATE();
    case 91:
      if (lookahead == 'i') ADVANCE(107);
      END_STATE();
    case 92:
      if (lookahead == 'n') ADVANCE(108);
      END_STATE();
    case 93:
      ACCEPT_TOKEN(anon_sym_human);
      END_STATE();
    case 94:
      ACCEPT_TOKEN(anon_sym_label);
      END_STATE();
    case 95:
      if (lookahead == 'e') ADVANCE(109);
      END_STATE();
    case 96:
      if (lookahead == 'l') ADVANCE(110);
      END_STATE();
    case 97:
      if (lookahead == 'r') ADVANCE(111);
      END_STATE();
    case 98:
      if (lookahead == 'r') ADVANCE(112);
      END_STATE();
    case 99:
      ACCEPT_TOKEN(anon_sym_start);
      if (lookahead == 's') ADVANCE(113);
      END_STATE();
    case 100:
      if (lookahead == 's') ADVANCE(114);
      END_STATE();
    case 101:
      if (lookahead == 'a') ADVANCE(115);
      END_STATE();
    case 102:
      if (lookahead == 't') ADVANCE(116);
      END_STATE();
    case 103:
      if (lookahead == 'l') ADVANCE(117);
      END_STATE();
    case 104:
      if (lookahead == 'i') ADVANCE(118);
      END_STATE();
    case 105:
      if (lookahead == 'n') ADVANCE(119);
      END_STATE();
    case 106:
      if (lookahead == 't') ADVANCE(120);
      END_STATE();
    case 107:
      if (lookahead == 't') ADVANCE(121);
      END_STATE();
    case 108:
      ACCEPT_TOKEN(anon_sym_fan_in);
      END_STATE();
    case 109:
      if (lookahead == 'r') ADVANCE(122);
      END_STATE();
    case 110:
      if (lookahead == 'e') ADVANCE(123);
      END_STATE();
    case 111:
      if (lookahead == 'e') ADVANCE(124);
      END_STATE();
    case 112:
      if (lookahead == 't') ADVANCE(125);
      END_STATE();
    case 113:
      if (lookahead == 'w') ADVANCE(126);
      END_STATE();
    case 114:
      if (lookahead == 'h') ADVANCE(127);
      END_STATE();
    case 115:
      if (lookahead == 'p') ADVANCE(128);
      END_STATE();
    case 116:
      ACCEPT_TOKEN(anon_sym_weight);
      END_STATE();
    case 117:
      if (lookahead == 'o') ADVANCE(129);
      END_STATE();
    case 118:
      if (lookahead == 'o') ADVANCE(130);
      END_STATE();
    case 119:
      if (lookahead == 's') ADVANCE(131);
      END_STATE();
    case 120:
      if (lookahead == 's') ADVANCE(132);
      END_STATE();
    case 121:
      if (lookahead == 'h') ADVANCE(133);
      END_STATE();
    case 122:
      if (lookahead == '_') ADVANCE(134);
      END_STATE();
    case 123:
      if (lookahead == 'l') ADVANCE(135);
      END_STATE();
    case 124:
      if (lookahead == 's') ADVANCE(136);
      END_STATE();
    case 125:
      ACCEPT_TOKEN(anon_sym_restart);
      END_STATE();
    case 126:
      if (lookahead == 'i') ADVANCE(137);
      END_STATE();
    case 127:
      if (lookahead == 'e') ADVANCE(138);
      END_STATE();
    case 128:
      if (lookahead == 'h') ADVANCE(139);
      END_STATE();
    case 129:
      if (lookahead == 'w') ADVANCE(140);
      END_STATE();
    case 130:
      if (lookahead == 'n') ADVANCE(141);
      END_STATE();
    case 131:
      ACCEPT_TOKEN(anon_sym_contains);
      END_STATE();
    case 132:
      ACCEPT_TOKEN(anon_sym_defaults);
      END_STATE();
    case 133:
      ACCEPT_TOKEN(anon_sym_endswith);
      END_STATE();
    case 134:
      if (lookahead == 'l') ADVANCE(142);
      END_STATE();
    case 135:
      ACCEPT_TOKEN(anon_sym_parallel);
      END_STATE();
    case 136:
      ACCEPT_TOKEN(anon_sym_requires);
      END_STATE();
    case 137:
      if (lookahead == 't') ADVANCE(143);
      END_STATE();
    case 138:
      if (lookahead == 'e') ADVANCE(144);
      END_STATE();
    case 139:
      ACCEPT_TOKEN(anon_sym_subgraph);
      END_STATE();
    case 140:
      ACCEPT_TOKEN(anon_sym_workflow);
      END_STATE();
    case 141:
      if (lookahead == 'a') ADVANCE(145);
      END_STATE();
    case 142:
      if (lookahead == 'o') ADVANCE(146);
      END_STATE();
    case 143:
      if (lookahead == 'h') ADVANCE(147);
      END_STATE();
    case 144:
      if (lookahead == 't') ADVANCE(148);
      END_STATE();
    case 145:
      if (lookahead == 'l') ADVANCE(149);
      END_STATE();
    case 146:
      if (lookahead == 'o') ADVANCE(150);
      END_STATE();
    case 147:
      ACCEPT_TOKEN(anon_sym_startswith);
      END_STATE();
    case 148:
      ACCEPT_TOKEN(anon_sym_stylesheet);
      END_STATE();
    case 149:
      ACCEPT_TOKEN(anon_sym_conditional);
      END_STATE();
    case 150:
      if (lookahead == 'p') ADVANCE(151);
      END_STATE();
    case 151:
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
  [4] = {.lex_state = 9, .external_lex_state = 2},
  [5] = {.lex_state = 9, .external_lex_state = 3},
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
  [29] = {.lex_state = 9},
  [30] = {.lex_state = 9, .external_lex_state = 3},
  [31] = {.lex_state = 9, .external_lex_state = 3},
  [32] = {.lex_state = 9, .external_lex_state = 3},
  [33] = {.lex_state = 9, .external_lex_state = 3},
  [34] = {.lex_state = 9, .external_lex_state = 3},
  [35] = {.lex_state = 0, .external_lex_state = 3},
  [36] = {.lex_state = 9, .external_lex_state = 3},
  [37] = {.lex_state = 9, .external_lex_state = 3},
  [38] = {.lex_state = 9, .external_lex_state = 3},
  [39] = {.lex_state = 0, .external_lex_state = 3},
  [40] = {.lex_state = 9, .external_lex_state = 3},
  [41] = {.lex_state = 9, .external_lex_state = 3},
  [42] = {.lex_state = 0, .external_lex_state = 2},
  [43] = {.lex_state = 9},
  [44] = {.lex_state = 9},
  [45] = {.lex_state = 9, .external_lex_state = 3},
  [46] = {.lex_state = 9, .external_lex_state = 3},
  [47] = {.lex_state = 1, .external_lex_state = 4},
  [48] = {.lex_state = 9, .external_lex_state = 3},
  [49] = {.lex_state = 1, .external_lex_state = 4},
  [50] = {.lex_state = 1, .external_lex_state = 4},
  [51] = {.lex_state = 9, .external_lex_state = 3},
  [52] = {.lex_state = 9},
  [53] = {.lex_state = 1, .external_lex_state = 4},
  [54] = {.lex_state = 1, .external_lex_state = 4},
  [55] = {.lex_state = 9, .external_lex_state = 3},
  [56] = {.lex_state = 9, .external_lex_state = 3},
  [57] = {.lex_state = 9, .external_lex_state = 3},
  [58] = {.lex_state = 9, .external_lex_state = 3},
  [59] = {.lex_state = 0, .external_lex_state = 3},
  [60] = {.lex_state = 9, .external_lex_state = 3},
  [61] = {.lex_state = 9, .external_lex_state = 3},
  [62] = {.lex_state = 9},
  [63] = {.lex_state = 9, .external_lex_state = 3},
  [64] = {.lex_state = 9, .external_lex_state = 3},
  [65] = {.lex_state = 9},
  [66] = {.lex_state = 9, .external_lex_state = 3},
  [67] = {.lex_state = 9, .external_lex_state = 3},
  [68] = {.lex_state = 9, .external_lex_state = 2},
  [69] = {.lex_state = 9, .external_lex_state = 2},
  [70] = {.lex_state = 9, .external_lex_state = 2},
  [71] = {.lex_state = 9, .external_lex_state = 2},
  [72] = {.lex_state = 9, .external_lex_state = 3},
  [73] = {.lex_state = 9, .external_lex_state = 3},
  [74] = {.lex_state = 9, .external_lex_state = 2},
  [75] = {.lex_state = 9, .external_lex_state = 2},
  [76] = {.lex_state = 9, .external_lex_state = 2},
  [77] = {.lex_state = 9, .external_lex_state = 3},
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
  [115] = {.lex_state = 9, .external_lex_state = 4},
  [116] = {.lex_state = 9},
  [117] = {.lex_state = 9},
  [118] = {.lex_state = 9},
  [119] = {.lex_state = 9, .external_lex_state = 4},
  [120] = {.lex_state = 9, .external_lex_state = 6},
  [121] = {.lex_state = 9},
  [122] = {.lex_state = 9},
  [123] = {.lex_state = 9},
  [124] = {.lex_state = 9},
  [125] = {.lex_state = 9, .external_lex_state = 4},
  [126] = {.lex_state = 9},
  [127] = {.lex_state = 9, .external_lex_state = 6},
  [128] = {.lex_state = 9},
  [129] = {.lex_state = 9},
  [130] = {.lex_state = 9},
  [131] = {.lex_state = 9},
  [132] = {.lex_state = 9},
  [133] = {.lex_state = 9, .external_lex_state = 4},
  [134] = {.lex_state = 9},
  [135] = {.lex_state = 9, .external_lex_state = 4},
  [136] = {.lex_state = 9, .external_lex_state = 4},
  [137] = {.lex_state = 9, .external_lex_state = 4},
  [138] = {.lex_state = 9},
  [139] = {.lex_state = 9, .external_lex_state = 4},
  [140] = {.lex_state = 9},
  [141] = {.lex_state = 9},
  [142] = {.lex_state = 9, .external_lex_state = 4},
  [143] = {.lex_state = 9, .external_lex_state = 4},
  [144] = {.lex_state = 9},
  [145] = {.lex_state = 9, .external_lex_state = 4},
  [146] = {.lex_state = 9, .external_lex_state = 4},
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
    [anon_sym_label] = ACTIONS(1),
    [anon_sym_weight] = ACTIONS(1),
    [anon_sym_restart] = ACTIONS(1),
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
    [sym_source_file] = STATE(130),
    [sym_workflow_decl] = STATE(131),
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
    ACTIONS(11), 28,
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
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
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
  [40] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(17), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(15), 28,
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
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
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
  [80] = 17,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(21), 1,
      anon_sym_defaults,
    ACTIONS(23), 1,
      anon_sym_agent,
    ACTIONS(25), 1,
      anon_sym_human,
    ACTIONS(27), 1,
      anon_sym_tool,
    ACTIONS(29), 1,
      anon_sym_subgraph,
    ACTIONS(31), 1,
      anon_sym_conditional,
    ACTIONS(33), 1,
      anon_sym_manager_loop,
    ACTIONS(35), 1,
      anon_sym_parallel,
    ACTIONS(37), 1,
      anon_sym_fan_in,
    ACTIONS(39), 1,
      anon_sym_edges,
    ACTIONS(41), 1,
      anon_sym_stylesheet,
    ACTIONS(43), 1,
      sym__newline,
    STATE(127), 1,
      sym_workflow_body,
    ACTIONS(19), 4,
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
    STATE(12), 8,
      sym_agent_node,
      sym_human_node,
      sym_tool_node,
      sym_subgraph_node,
      sym_conditional_node,
      sym_manager_loop_node,
      sym_parallel_node,
      sym_fan_in_node,
  [147] = 17,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(48), 1,
      anon_sym_defaults,
    ACTIONS(51), 1,
      anon_sym_agent,
    ACTIONS(54), 1,
      anon_sym_human,
    ACTIONS(57), 1,
      anon_sym_tool,
    ACTIONS(60), 1,
      anon_sym_subgraph,
    ACTIONS(63), 1,
      anon_sym_conditional,
    ACTIONS(66), 1,
      anon_sym_manager_loop,
    ACTIONS(69), 1,
      anon_sym_parallel,
    ACTIONS(72), 1,
      anon_sym_fan_in,
    ACTIONS(75), 1,
      anon_sym_edges,
    ACTIONS(78), 1,
      anon_sym_stylesheet,
    ACTIONS(81), 1,
      sym__dedent,
    ACTIONS(83), 1,
      sym__newline,
    ACTIONS(45), 4,
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
    STATE(12), 8,
      sym_agent_node,
      sym_human_node,
      sym_tool_node,
      sym_subgraph_node,
      sym_conditional_node,
      sym_manager_loop_node,
      sym_parallel_node,
      sym_fan_in_node,
  [214] = 17,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(21), 1,
      anon_sym_defaults,
    ACTIONS(23), 1,
      anon_sym_agent,
    ACTIONS(25), 1,
      anon_sym_human,
    ACTIONS(27), 1,
      anon_sym_tool,
    ACTIONS(29), 1,
      anon_sym_subgraph,
    ACTIONS(31), 1,
      anon_sym_conditional,
    ACTIONS(33), 1,
      anon_sym_manager_loop,
    ACTIONS(35), 1,
      anon_sym_parallel,
    ACTIONS(37), 1,
      anon_sym_fan_in,
    ACTIONS(39), 1,
      anon_sym_edges,
    ACTIONS(41), 1,
      anon_sym_stylesheet,
    ACTIONS(86), 1,
      sym__dedent,
    ACTIONS(88), 1,
      sym__newline,
    ACTIONS(19), 4,
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
    STATE(12), 8,
      sym_agent_node,
      sym_human_node,
      sym_tool_node,
      sym_subgraph_node,
      sym_conditional_node,
      sym_manager_loop_node,
      sym_parallel_node,
      sym_fan_in_node,
  [281] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(92), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(90), 20,
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
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_stylesheet,
      sym_identifier,
  [311] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(96), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(94), 20,
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
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_stylesheet,
      sym_identifier,
  [341] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(102), 1,
      anon_sym_DOT,
    ACTIONS(100), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(98), 13,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_or,
      anon_sym_and,
      anon_sym_not,
      anon_sym_EQ,
      anon_sym_contains,
      anon_sym_startswith,
      anon_sym_endswith,
      anon_sym_in,
      sym_identifier,
  [369] = 7,
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
    ACTIONS(104), 7,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_or,
      anon_sym_and,
      sym_identifier,
  [403] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(114), 17,
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
  [426] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(116), 17,
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
  [449] = 2,
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
  [472] = 2,
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
  [495] = 2,
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
  [518] = 2,
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
  [541] = 2,
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
  [564] = 2,
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
  [587] = 2,
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
  [610] = 2,
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
  [633] = 2,
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
  [656] = 2,
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
  [679] = 2,
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
  [702] = 2,
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
  [725] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(100), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(98), 13,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_or,
      anon_sym_and,
      anon_sym_not,
      anon_sym_EQ,
      anon_sym_contains,
      anon_sym_startswith,
      anon_sym_endswith,
      anon_sym_in,
      sym_identifier,
  [750] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(144), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(142), 13,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_or,
      anon_sym_and,
      anon_sym_not,
      anon_sym_EQ,
      anon_sym_contains,
      anon_sym_startswith,
      anon_sym_endswith,
      anon_sym_in,
      sym_identifier,
  [775] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(148), 1,
      anon_sym_and,
    STATE(27), 1,
      aux_sym_and_expr_repeat1,
    ACTIONS(151), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(146), 6,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_or,
      sym_identifier,
  [797] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(155), 1,
      anon_sym_and,
    STATE(30), 1,
      aux_sym_and_expr_repeat1,
    ACTIONS(157), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(153), 6,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_or,
      sym_identifier,
  [819] = 10,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(159), 1,
      sym_identifier,
    ACTIONS(161), 1,
      anon_sym_DQUOTE,
    ACTIONS(163), 1,
      anon_sym_SQUOTE,
    STATE(10), 1,
      sym_operand,
    STATE(28), 1,
      sym_compare_expr,
    STATE(34), 1,
      sym_and_expr,
    STATE(46), 1,
      sym_condition,
    STATE(48), 1,
      sym_or_expr,
    STATE(25), 2,
      sym_variable,
      sym_string,
  [851] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(155), 1,
      anon_sym_and,
    STATE(27), 1,
      aux_sym_and_expr_repeat1,
    ACTIONS(167), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(165), 6,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_or,
      sym_identifier,
  [873] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(171), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(169), 7,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_or,
      anon_sym_and,
      sym_identifier,
  [890] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(173), 1,
      sym_identifier,
    ACTIONS(175), 1,
      anon_sym_when,
    ACTIONS(179), 2,
      sym__dedent,
      sym__newline,
    STATE(33), 2,
      sym_edge_attr,
      aux_sym_edge_entry_repeat1,
    ACTIONS(177), 3,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
  [913] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(175), 1,
      anon_sym_when,
    ACTIONS(181), 1,
      sym_identifier,
    ACTIONS(183), 2,
      sym__dedent,
      sym__newline,
    STATE(36), 2,
      sym_edge_attr,
      aux_sym_edge_entry_repeat1,
    ACTIONS(177), 3,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
  [936] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(187), 1,
      anon_sym_or,
    STATE(38), 1,
      aux_sym_or_expr_repeat1,
    ACTIONS(189), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(185), 5,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      sym_identifier,
  [957] = 8,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(193), 1,
      anon_sym_DOT,
    ACTIONS(195), 1,
      anon_sym_POUND,
    ACTIONS(197), 1,
      sym__dedent,
    ACTIONS(199), 1,
      sym__newline,
    STATE(119), 1,
      sym_selector,
    ACTIONS(191), 2,
      anon_sym_STAR,
      sym_identifier,
    STATE(39), 2,
      sym_stylesheet_rule,
      aux_sym_stylesheet_section_repeat1,
  [984] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(201), 1,
      sym_identifier,
    ACTIONS(203), 1,
      anon_sym_when,
    ACTIONS(209), 2,
      sym__dedent,
      sym__newline,
    STATE(36), 2,
      sym_edge_attr,
      aux_sym_edge_entry_repeat1,
    ACTIONS(206), 3,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
  [1007] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(213), 1,
      anon_sym_or,
    STATE(37), 1,
      aux_sym_or_expr_repeat1,
    ACTIONS(216), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(211), 5,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      sym_identifier,
  [1028] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(187), 1,
      anon_sym_or,
    STATE(37), 1,
      aux_sym_or_expr_repeat1,
    ACTIONS(220), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(218), 5,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      sym_identifier,
  [1049] = 8,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(225), 1,
      anon_sym_DOT,
    ACTIONS(228), 1,
      anon_sym_POUND,
    ACTIONS(231), 1,
      sym__dedent,
    ACTIONS(233), 1,
      sym__newline,
    STATE(119), 1,
      sym_selector,
    ACTIONS(222), 2,
      anon_sym_STAR,
      sym_identifier,
    STATE(39), 2,
      sym_stylesheet_rule,
      aux_sym_stylesheet_section_repeat1,
  [1076] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(151), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(146), 7,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_or,
      anon_sym_and,
      sym_identifier,
  [1093] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(238), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(236), 7,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_or,
      anon_sym_and,
      sym_identifier,
  [1110] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(193), 1,
      anon_sym_DOT,
    ACTIONS(195), 1,
      anon_sym_POUND,
    ACTIONS(240), 1,
      sym__newline,
    STATE(119), 1,
      sym_selector,
    ACTIONS(191), 2,
      anon_sym_STAR,
      sym_identifier,
    STATE(35), 2,
      sym_stylesheet_rule,
      aux_sym_stylesheet_section_repeat1,
  [1134] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(110), 1,
      anon_sym_EQ,
    STATE(65), 1,
      sym_compare_op,
    ACTIONS(108), 6,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
      anon_sym_contains,
      anon_sym_startswith,
      anon_sym_endswith,
      anon_sym_in,
  [1152] = 8,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(159), 1,
      sym_identifier,
    ACTIONS(161), 1,
      anon_sym_DQUOTE,
    ACTIONS(163), 1,
      anon_sym_SQUOTE,
    STATE(10), 1,
      sym_operand,
    STATE(28), 1,
      sym_compare_expr,
    STATE(45), 1,
      sym_and_expr,
    STATE(25), 2,
      sym_variable,
      sym_string,
  [1178] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(216), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(211), 6,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_or,
      sym_identifier,
  [1194] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(244), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(242), 5,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      sym_identifier,
  [1209] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(246), 1,
      sym_raw_inline,
    ACTIONS(248), 1,
      anon_sym_DQUOTE,
    ACTIONS(250), 1,
      anon_sym_SQUOTE,
    ACTIONS(252), 1,
      sym__indent,
    STATE(16), 1,
      sym_field_value,
    STATE(8), 2,
      sym_multiline_block,
      sym_string,
  [1232] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(256), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(254), 5,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      sym_identifier,
  [1247] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(246), 1,
      sym_raw_inline,
    ACTIONS(248), 1,
      anon_sym_DQUOTE,
    ACTIONS(250), 1,
      anon_sym_SQUOTE,
    ACTIONS(252), 1,
      sym__indent,
    STATE(51), 1,
      sym_field_value,
    STATE(8), 2,
      sym_multiline_block,
      sym_string,
  [1270] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(246), 1,
      sym_raw_inline,
    ACTIONS(248), 1,
      anon_sym_DQUOTE,
    ACTIONS(250), 1,
      anon_sym_SQUOTE,
    ACTIONS(252), 1,
      sym__indent,
    STATE(101), 1,
      sym_field_value,
    STATE(8), 2,
      sym_multiline_block,
      sym_string,
  [1293] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(260), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(258), 5,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      sym_identifier,
  [1308] = 7,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(159), 1,
      sym_identifier,
    ACTIONS(161), 1,
      anon_sym_DQUOTE,
    ACTIONS(163), 1,
      anon_sym_SQUOTE,
    STATE(10), 1,
      sym_operand,
    STATE(40), 1,
      sym_compare_expr,
    STATE(25), 2,
      sym_variable,
      sym_string,
  [1331] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(246), 1,
      sym_raw_inline,
    ACTIONS(248), 1,
      anon_sym_DQUOTE,
    ACTIONS(250), 1,
      anon_sym_SQUOTE,
    ACTIONS(252), 1,
      sym__indent,
    STATE(98), 1,
      sym_field_value,
    STATE(8), 2,
      sym_multiline_block,
      sym_string,
  [1354] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(246), 1,
      sym_raw_inline,
    ACTIONS(248), 1,
      anon_sym_DQUOTE,
    ACTIONS(250), 1,
      anon_sym_SQUOTE,
    ACTIONS(252), 1,
      sym__indent,
    STATE(97), 1,
      sym_field_value,
    STATE(8), 2,
      sym_multiline_block,
      sym_string,
  [1377] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(262), 1,
      sym_identifier,
    ACTIONS(264), 1,
      sym__dedent,
    ACTIONS(266), 1,
      sym__newline,
    STATE(129), 1,
      sym_field_name,
    STATE(61), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1397] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(262), 1,
      sym_identifier,
    ACTIONS(268), 1,
      sym__dedent,
    ACTIONS(270), 1,
      sym__newline,
    STATE(132), 1,
      sym_field_name,
    STATE(60), 2,
      sym_defaults_field,
      aux_sym_defaults_section_repeat1,
  [1417] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(262), 1,
      sym_identifier,
    ACTIONS(266), 1,
      sym__newline,
    ACTIONS(272), 1,
      sym__dedent,
    STATE(129), 1,
      sym_field_name,
    STATE(61), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1437] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(262), 1,
      sym_identifier,
    ACTIONS(266), 1,
      sym__newline,
    ACTIONS(274), 1,
      sym__dedent,
    STATE(129), 1,
      sym_field_name,
    STATE(61), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1457] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(278), 1,
      anon_sym_POUND,
    ACTIONS(276), 5,
      sym__dedent,
      sym__newline,
      anon_sym_DOT,
      anon_sym_STAR,
      sym_identifier,
  [1471] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(280), 1,
      sym_identifier,
    ACTIONS(283), 1,
      sym__dedent,
    ACTIONS(285), 1,
      sym__newline,
    STATE(132), 1,
      sym_field_name,
    STATE(60), 2,
      sym_defaults_field,
      aux_sym_defaults_section_repeat1,
  [1491] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(288), 1,
      sym_identifier,
    ACTIONS(291), 1,
      sym__dedent,
    ACTIONS(293), 1,
      sym__newline,
    STATE(129), 1,
      sym_field_name,
    STATE(61), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1511] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(159), 1,
      sym_identifier,
    ACTIONS(161), 1,
      anon_sym_DQUOTE,
    ACTIONS(163), 1,
      anon_sym_SQUOTE,
    STATE(41), 1,
      sym_operand,
    STATE(25), 2,
      sym_variable,
      sym_string,
  [1531] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(262), 1,
      sym_identifier,
    ACTIONS(266), 1,
      sym__newline,
    ACTIONS(296), 1,
      sym__dedent,
    STATE(129), 1,
      sym_field_name,
    STATE(61), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1551] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(262), 1,
      sym_identifier,
    ACTIONS(266), 1,
      sym__newline,
    ACTIONS(298), 1,
      sym__dedent,
    STATE(129), 1,
      sym_field_name,
    STATE(61), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1571] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(159), 1,
      sym_identifier,
    ACTIONS(161), 1,
      anon_sym_DQUOTE,
    ACTIONS(163), 1,
      anon_sym_SQUOTE,
    STATE(31), 1,
      sym_operand,
    STATE(25), 2,
      sym_variable,
      sym_string,
  [1591] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(262), 1,
      sym_identifier,
    ACTIONS(266), 1,
      sym__newline,
    ACTIONS(300), 1,
      sym__dedent,
    STATE(129), 1,
      sym_field_name,
    STATE(61), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1611] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(262), 1,
      sym_identifier,
    ACTIONS(266), 1,
      sym__newline,
    ACTIONS(302), 1,
      sym__dedent,
    STATE(129), 1,
      sym_field_name,
    STATE(61), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1631] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(262), 1,
      sym_identifier,
    ACTIONS(304), 1,
      sym__newline,
    STATE(129), 1,
      sym_field_name,
    STATE(63), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1648] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(262), 1,
      sym_identifier,
    ACTIONS(306), 1,
      sym__newline,
    STATE(129), 1,
      sym_field_name,
    STATE(67), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1665] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(262), 1,
      sym_identifier,
    ACTIONS(308), 1,
      sym__newline,
    STATE(129), 1,
      sym_field_name,
    STATE(57), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1682] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(262), 1,
      sym_identifier,
    ACTIONS(310), 1,
      sym__newline,
    STATE(129), 1,
      sym_field_name,
    STATE(55), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1699] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(312), 1,
      sym_identifier,
    ACTIONS(315), 1,
      sym__dedent,
    ACTIONS(317), 1,
      sym__newline,
    STATE(72), 1,
      aux_sym_stylesheet_rule_repeat1,
    STATE(113), 1,
      sym_field_name,
  [1718] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(320), 1,
      sym_identifier,
    ACTIONS(323), 1,
      sym__dedent,
    ACTIONS(325), 1,
      sym__newline,
    STATE(73), 2,
      sym_edge_entry,
      aux_sym_edges_section_repeat1,
  [1735] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(262), 1,
      sym_identifier,
    ACTIONS(328), 1,
      sym__newline,
    STATE(132), 1,
      sym_field_name,
    STATE(56), 2,
      sym_defaults_field,
      aux_sym_defaults_section_repeat1,
  [1752] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(262), 1,
      sym_identifier,
    ACTIONS(330), 1,
      sym__newline,
    STATE(129), 1,
      sym_field_name,
    STATE(58), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1769] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(262), 1,
      sym_identifier,
    ACTIONS(332), 1,
      sym__newline,
    STATE(129), 1,
      sym_field_name,
    STATE(66), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1786] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(262), 1,
      sym_identifier,
    ACTIONS(334), 1,
      sym__dedent,
    ACTIONS(336), 1,
      sym__newline,
    STATE(72), 1,
      aux_sym_stylesheet_rule_repeat1,
    STATE(113), 1,
      sym_field_name,
  [1805] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(262), 1,
      sym_identifier,
    ACTIONS(338), 1,
      sym__newline,
    STATE(129), 1,
      sym_field_name,
    STATE(64), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1822] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(340), 1,
      sym_identifier,
    ACTIONS(342), 1,
      sym__dedent,
    ACTIONS(344), 1,
      sym__newline,
    STATE(73), 2,
      sym_edge_entry,
      aux_sym_edges_section_repeat1,
  [1839] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(346), 1,
      anon_sym_DQUOTE,
    STATE(80), 1,
      aux_sym_string_repeat1,
    ACTIONS(348), 2,
      aux_sym_string_token1,
      aux_sym_string_token2,
  [1853] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(351), 1,
      anon_sym_DQUOTE,
    STATE(91), 1,
      aux_sym_string_repeat1,
    ACTIONS(353), 2,
      aux_sym_string_token1,
      aux_sym_string_token2,
  [1867] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(358), 1,
      sym__dedent,
    STATE(82), 1,
      aux_sym_block_content_repeat1,
    ACTIONS(355), 2,
      sym__newline,
      sym_block_line,
  [1881] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(360), 1,
      anon_sym_SQUOTE,
    STATE(83), 1,
      aux_sym_string_repeat2,
    ACTIONS(362), 2,
      aux_sym_string_token3,
      anon_sym_SQUOTE_SQUOTE,
  [1895] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(351), 1,
      anon_sym_SQUOTE,
    STATE(92), 1,
      aux_sym_string_repeat2,
    ACTIONS(365), 2,
      aux_sym_string_token3,
      anon_sym_SQUOTE_SQUOTE,
  [1909] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(340), 1,
      sym_identifier,
    ACTIONS(367), 1,
      sym__newline,
    STATE(79), 2,
      sym_edge_entry,
      aux_sym_edges_section_repeat1,
  [1923] = 4,
    ACTIONS(3), 1,
      sym_comment,
    STATE(93), 1,
      aux_sym_block_content_repeat1,
    STATE(120), 1,
      sym_block_content,
    ACTIONS(369), 2,
      sym__newline,
      sym_block_line,
  [1937] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(371), 1,
      anon_sym_COMMA,
    STATE(90), 1,
      aux_sym_identifier_list_repeat1,
    ACTIONS(373), 2,
      sym__indent,
      sym__newline,
  [1951] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(262), 1,
      sym_identifier,
    ACTIONS(375), 1,
      sym__newline,
    STATE(77), 1,
      aux_sym_stylesheet_rule_repeat1,
    STATE(113), 1,
      sym_field_name,
  [1967] = 5,
    ACTIONS(5), 1,
      anon_sym_workflow,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(377), 1,
      sym__newline,
    STATE(95), 1,
      aux_sym_source_file_repeat1,
    STATE(138), 1,
      sym_workflow_decl,
  [1983] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(379), 1,
      anon_sym_COMMA,
    STATE(90), 1,
      aux_sym_identifier_list_repeat1,
    ACTIONS(382), 2,
      sym__indent,
      sym__newline,
  [1997] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(384), 1,
      anon_sym_DQUOTE,
    STATE(80), 1,
      aux_sym_string_repeat1,
    ACTIONS(386), 2,
      aux_sym_string_token1,
      aux_sym_string_token2,
  [2011] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(384), 1,
      anon_sym_SQUOTE,
    STATE(83), 1,
      aux_sym_string_repeat2,
    ACTIONS(388), 2,
      aux_sym_string_token3,
      anon_sym_SQUOTE_SQUOTE,
  [2025] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(392), 1,
      sym__dedent,
    STATE(82), 1,
      aux_sym_block_content_repeat1,
    ACTIONS(390), 2,
      sym__newline,
      sym_block_line,
  [2039] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(371), 1,
      anon_sym_COMMA,
    STATE(87), 1,
      aux_sym_identifier_list_repeat1,
    ACTIONS(394), 2,
      sym__indent,
      sym__newline,
  [2053] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(396), 1,
      anon_sym_workflow,
    ACTIONS(398), 1,
      sym__newline,
    STATE(95), 1,
      aux_sym_source_file_repeat1,
  [2066] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(401), 1,
      sym__indent,
    ACTIONS(403), 1,
      sym__newline,
    STATE(22), 1,
      sym_node_attr_block,
  [2079] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(405), 3,
      sym__dedent,
      sym__newline,
      sym_identifier,
  [2088] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(407), 3,
      sym__dedent,
      sym__newline,
      sym_identifier,
  [2097] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(409), 3,
      sym_identifier,
      anon_sym_DQUOTE,
      anon_sym_SQUOTE,
  [2106] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(382), 3,
      sym__indent,
      sym__newline,
      anon_sym_COMMA,
  [2115] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(411), 3,
      sym__dedent,
      sym__newline,
      sym_identifier,
  [2124] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(401), 1,
      sym__indent,
    ACTIONS(413), 1,
      sym__newline,
    STATE(21), 1,
      sym_node_attr_block,
  [2137] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(415), 1,
      sym_identifier,
    STATE(102), 1,
      sym_identifier_list,
  [2147] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(415), 1,
      sym_identifier,
    STATE(96), 1,
      sym_identifier_list,
  [2157] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(417), 1,
      sym_identifier,
  [2164] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(419), 1,
      sym__indent,
  [2171] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(421), 1,
      anon_sym_DASH_GT,
  [2178] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(423), 1,
      anon_sym_LT_DASH,
  [2185] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(425), 1,
      sym_identifier,
  [2192] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(427), 1,
      sym_identifier,
  [2199] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(429), 1,
      anon_sym_COLON,
  [2206] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(431), 1,
      sym_identifier,
  [2213] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(433), 1,
      anon_sym_COLON,
  [2220] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(435), 1,
      sym_identifier,
  [2227] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(437), 1,
      sym__indent,
  [2234] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(439), 1,
      sym_identifier,
  [2241] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(441), 1,
      anon_sym_DASH_GT,
  [2248] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(443), 1,
      sym_identifier,
  [2255] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(445), 1,
      sym__indent,
  [2262] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(447), 1,
      sym__dedent,
  [2269] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(449), 1,
      ts_builtin_sym_end,
  [2276] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(451), 1,
      sym_identifier,
  [2283] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(453), 1,
      sym_identifier,
  [2290] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(455), 1,
      anon_sym_COLON,
  [2297] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(457), 1,
      sym__indent,
  [2304] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(459), 1,
      anon_sym_COLON,
  [2311] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(461), 1,
      sym__dedent,
  [2318] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(463), 1,
      sym_identifier,
  [2325] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(465), 1,
      anon_sym_COLON,
  [2332] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(467), 1,
      ts_builtin_sym_end,
  [2339] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(469), 1,
      ts_builtin_sym_end,
  [2346] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(471), 1,
      anon_sym_COLON,
  [2353] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(473), 1,
      sym__indent,
  [2360] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(475), 1,
      anon_sym_COLON,
  [2367] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(477), 1,
      sym__indent,
  [2374] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(479), 1,
      sym__indent,
  [2381] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(481), 1,
      sym__indent,
  [2388] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(483), 1,
      ts_builtin_sym_end,
  [2395] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(485), 1,
      sym__indent,
  [2402] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(487), 1,
      sym_identifier,
  [2409] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(489), 1,
      sym_identifier,
  [2416] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(491), 1,
      sym__indent,
  [2423] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(493), 1,
      sym__indent,
  [2430] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(495), 1,
      sym_identifier,
  [2437] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(497), 1,
      sym__indent,
  [2444] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(499), 1,
      sym__indent,
};

static const uint32_t ts_small_parse_table_map[] = {
  [SMALL_STATE(2)] = 0,
  [SMALL_STATE(3)] = 40,
  [SMALL_STATE(4)] = 80,
  [SMALL_STATE(5)] = 147,
  [SMALL_STATE(6)] = 214,
  [SMALL_STATE(7)] = 281,
  [SMALL_STATE(8)] = 311,
  [SMALL_STATE(9)] = 341,
  [SMALL_STATE(10)] = 369,
  [SMALL_STATE(11)] = 403,
  [SMALL_STATE(12)] = 426,
  [SMALL_STATE(13)] = 449,
  [SMALL_STATE(14)] = 472,
  [SMALL_STATE(15)] = 495,
  [SMALL_STATE(16)] = 518,
  [SMALL_STATE(17)] = 541,
  [SMALL_STATE(18)] = 564,
  [SMALL_STATE(19)] = 587,
  [SMALL_STATE(20)] = 610,
  [SMALL_STATE(21)] = 633,
  [SMALL_STATE(22)] = 656,
  [SMALL_STATE(23)] = 679,
  [SMALL_STATE(24)] = 702,
  [SMALL_STATE(25)] = 725,
  [SMALL_STATE(26)] = 750,
  [SMALL_STATE(27)] = 775,
  [SMALL_STATE(28)] = 797,
  [SMALL_STATE(29)] = 819,
  [SMALL_STATE(30)] = 851,
  [SMALL_STATE(31)] = 873,
  [SMALL_STATE(32)] = 890,
  [SMALL_STATE(33)] = 913,
  [SMALL_STATE(34)] = 936,
  [SMALL_STATE(35)] = 957,
  [SMALL_STATE(36)] = 984,
  [SMALL_STATE(37)] = 1007,
  [SMALL_STATE(38)] = 1028,
  [SMALL_STATE(39)] = 1049,
  [SMALL_STATE(40)] = 1076,
  [SMALL_STATE(41)] = 1093,
  [SMALL_STATE(42)] = 1110,
  [SMALL_STATE(43)] = 1134,
  [SMALL_STATE(44)] = 1152,
  [SMALL_STATE(45)] = 1178,
  [SMALL_STATE(46)] = 1194,
  [SMALL_STATE(47)] = 1209,
  [SMALL_STATE(48)] = 1232,
  [SMALL_STATE(49)] = 1247,
  [SMALL_STATE(50)] = 1270,
  [SMALL_STATE(51)] = 1293,
  [SMALL_STATE(52)] = 1308,
  [SMALL_STATE(53)] = 1331,
  [SMALL_STATE(54)] = 1354,
  [SMALL_STATE(55)] = 1377,
  [SMALL_STATE(56)] = 1397,
  [SMALL_STATE(57)] = 1417,
  [SMALL_STATE(58)] = 1437,
  [SMALL_STATE(59)] = 1457,
  [SMALL_STATE(60)] = 1471,
  [SMALL_STATE(61)] = 1491,
  [SMALL_STATE(62)] = 1511,
  [SMALL_STATE(63)] = 1531,
  [SMALL_STATE(64)] = 1551,
  [SMALL_STATE(65)] = 1571,
  [SMALL_STATE(66)] = 1591,
  [SMALL_STATE(67)] = 1611,
  [SMALL_STATE(68)] = 1631,
  [SMALL_STATE(69)] = 1648,
  [SMALL_STATE(70)] = 1665,
  [SMALL_STATE(71)] = 1682,
  [SMALL_STATE(72)] = 1699,
  [SMALL_STATE(73)] = 1718,
  [SMALL_STATE(74)] = 1735,
  [SMALL_STATE(75)] = 1752,
  [SMALL_STATE(76)] = 1769,
  [SMALL_STATE(77)] = 1786,
  [SMALL_STATE(78)] = 1805,
  [SMALL_STATE(79)] = 1822,
  [SMALL_STATE(80)] = 1839,
  [SMALL_STATE(81)] = 1853,
  [SMALL_STATE(82)] = 1867,
  [SMALL_STATE(83)] = 1881,
  [SMALL_STATE(84)] = 1895,
  [SMALL_STATE(85)] = 1909,
  [SMALL_STATE(86)] = 1923,
  [SMALL_STATE(87)] = 1937,
  [SMALL_STATE(88)] = 1951,
  [SMALL_STATE(89)] = 1967,
  [SMALL_STATE(90)] = 1983,
  [SMALL_STATE(91)] = 1997,
  [SMALL_STATE(92)] = 2011,
  [SMALL_STATE(93)] = 2025,
  [SMALL_STATE(94)] = 2039,
  [SMALL_STATE(95)] = 2053,
  [SMALL_STATE(96)] = 2066,
  [SMALL_STATE(97)] = 2079,
  [SMALL_STATE(98)] = 2088,
  [SMALL_STATE(99)] = 2097,
  [SMALL_STATE(100)] = 2106,
  [SMALL_STATE(101)] = 2115,
  [SMALL_STATE(102)] = 2124,
  [SMALL_STATE(103)] = 2137,
  [SMALL_STATE(104)] = 2147,
  [SMALL_STATE(105)] = 2157,
  [SMALL_STATE(106)] = 2164,
  [SMALL_STATE(107)] = 2171,
  [SMALL_STATE(108)] = 2178,
  [SMALL_STATE(109)] = 2185,
  [SMALL_STATE(110)] = 2192,
  [SMALL_STATE(111)] = 2199,
  [SMALL_STATE(112)] = 2206,
  [SMALL_STATE(113)] = 2213,
  [SMALL_STATE(114)] = 2220,
  [SMALL_STATE(115)] = 2227,
  [SMALL_STATE(116)] = 2234,
  [SMALL_STATE(117)] = 2241,
  [SMALL_STATE(118)] = 2248,
  [SMALL_STATE(119)] = 2255,
  [SMALL_STATE(120)] = 2262,
  [SMALL_STATE(121)] = 2269,
  [SMALL_STATE(122)] = 2276,
  [SMALL_STATE(123)] = 2283,
  [SMALL_STATE(124)] = 2290,
  [SMALL_STATE(125)] = 2297,
  [SMALL_STATE(126)] = 2304,
  [SMALL_STATE(127)] = 2311,
  [SMALL_STATE(128)] = 2318,
  [SMALL_STATE(129)] = 2325,
  [SMALL_STATE(130)] = 2332,
  [SMALL_STATE(131)] = 2339,
  [SMALL_STATE(132)] = 2346,
  [SMALL_STATE(133)] = 2353,
  [SMALL_STATE(134)] = 2360,
  [SMALL_STATE(135)] = 2367,
  [SMALL_STATE(136)] = 2374,
  [SMALL_STATE(137)] = 2381,
  [SMALL_STATE(138)] = 2388,
  [SMALL_STATE(139)] = 2395,
  [SMALL_STATE(140)] = 2402,
  [SMALL_STATE(141)] = 2409,
  [SMALL_STATE(142)] = 2416,
  [SMALL_STATE(143)] = 2423,
  [SMALL_STATE(144)] = 2430,
  [SMALL_STATE(145)] = 2437,
  [SMALL_STATE(146)] = 2444,
};

static const TSParseActionEntry ts_parse_actions[] = {
  [0] = {.entry = {.count = 0, .reusable = false}},
  [1] = {.entry = {.count = 1, .reusable = false}}, RECOVER(),
  [3] = {.entry = {.count = 1, .reusable = false}}, SHIFT_EXTRA(),
  [5] = {.entry = {.count = 1, .reusable = true}}, SHIFT(123),
  [7] = {.entry = {.count = 1, .reusable = true}}, SHIFT_EXTRA(),
  [9] = {.entry = {.count = 1, .reusable = true}}, SHIFT(89),
  [11] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_string, 3, 0, 0),
  [13] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_string, 3, 0, 0),
  [15] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_string, 2, 0, 0),
  [17] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_string, 2, 0, 0),
  [19] = {.entry = {.count = 1, .reusable = true}}, SHIFT(134),
  [21] = {.entry = {.count = 1, .reusable = true}}, SHIFT(139),
  [23] = {.entry = {.count = 1, .reusable = true}}, SHIFT(140),
  [25] = {.entry = {.count = 1, .reusable = true}}, SHIFT(109),
  [27] = {.entry = {.count = 1, .reusable = true}}, SHIFT(110),
  [29] = {.entry = {.count = 1, .reusable = true}}, SHIFT(112),
  [31] = {.entry = {.count = 1, .reusable = true}}, SHIFT(114),
  [33] = {.entry = {.count = 1, .reusable = true}}, SHIFT(118),
  [35] = {.entry = {.count = 1, .reusable = true}}, SHIFT(105),
  [37] = {.entry = {.count = 1, .reusable = true}}, SHIFT(122),
  [39] = {.entry = {.count = 1, .reusable = true}}, SHIFT(146),
  [41] = {.entry = {.count = 1, .reusable = true}}, SHIFT(126),
  [43] = {.entry = {.count = 1, .reusable = true}}, SHIFT(6),
  [45] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(134),
  [48] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(139),
  [51] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(140),
  [54] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(109),
  [57] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(110),
  [60] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(112),
  [63] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(114),
  [66] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(118),
  [69] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(105),
  [72] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(122),
  [75] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(146),
  [78] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(126),
  [81] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0),
  [83] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(5),
  [86] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_workflow_body, 1, 0, 0),
  [88] = {.entry = {.count = 1, .reusable = true}}, SHIFT(5),
  [90] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_multiline_block, 3, 0, 0),
  [92] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_multiline_block, 3, 0, 0),
  [94] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_field_value, 1, 0, 0),
  [96] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_field_value, 1, 0, 0),
  [98] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_operand, 1, 0, 0),
  [100] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_operand, 1, 0, 0),
  [102] = {.entry = {.count = 1, .reusable = true}}, SHIFT(128),
  [104] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_compare_expr, 1, 0, 0),
  [106] = {.entry = {.count = 1, .reusable = false}}, SHIFT(43),
  [108] = {.entry = {.count = 1, .reusable = true}}, SHIFT(99),
  [110] = {.entry = {.count = 1, .reusable = false}}, SHIFT(99),
  [112] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_expr, 1, 0, 0),
  [114] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_conditional_node, 5, 0, 0),
  [116] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_node_decl, 1, 0, 0),
  [118] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_defaults_section, 4, 0, 0),
  [120] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edges_section, 4, 0, 0),
  [122] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_subgraph_node, 5, 0, 0),
  [124] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_workflow_field, 3, 0, 0),
  [126] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_agent_node, 5, 0, 0),
  [128] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_human_node, 5, 0, 0),
  [130] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_tool_node, 5, 0, 0),
  [132] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_manager_loop_node, 5, 0, 0),
  [134] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_parallel_node, 5, 0, 0),
  [136] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_fan_in_node, 5, 0, 0),
  [138] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_stylesheet_section, 5, 0, 0),
  [140] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_node_attr_block, 3, 0, 0),
  [142] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_variable, 3, 0, 0),
  [144] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_variable, 3, 0, 0),
  [146] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_and_expr_repeat1, 2, 0, 0),
  [148] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_and_expr_repeat1, 2, 0, 0), SHIFT_REPEAT(52),
  [151] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_and_expr_repeat1, 2, 0, 0),
  [153] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_and_expr, 1, 0, 0),
  [155] = {.entry = {.count = 1, .reusable = false}}, SHIFT(52),
  [157] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_and_expr, 1, 0, 0),
  [159] = {.entry = {.count = 1, .reusable = true}}, SHIFT(9),
  [161] = {.entry = {.count = 1, .reusable = true}}, SHIFT(81),
  [163] = {.entry = {.count = 1, .reusable = true}}, SHIFT(84),
  [165] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_and_expr, 2, 0, 0),
  [167] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_and_expr, 2, 0, 0),
  [169] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_compare_expr, 4, 0, 0),
  [171] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_expr, 4, 0, 0),
  [173] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_entry, 3, 0, 0),
  [175] = {.entry = {.count = 1, .reusable = false}}, SHIFT(29),
  [177] = {.entry = {.count = 1, .reusable = false}}, SHIFT(111),
  [179] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_entry, 3, 0, 0),
  [181] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_entry, 4, 0, 0),
  [183] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_entry, 4, 0, 0),
  [185] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_or_expr, 1, 0, 0),
  [187] = {.entry = {.count = 1, .reusable = false}}, SHIFT(44),
  [189] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_or_expr, 1, 0, 0),
  [191] = {.entry = {.count = 1, .reusable = true}}, SHIFT(145),
  [193] = {.entry = {.count = 1, .reusable = true}}, SHIFT(144),
  [195] = {.entry = {.count = 1, .reusable = false}}, SHIFT(144),
  [197] = {.entry = {.count = 1, .reusable = true}}, SHIFT(23),
  [199] = {.entry = {.count = 1, .reusable = true}}, SHIFT(39),
  [201] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0),
  [203] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0), SHIFT_REPEAT(29),
  [206] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0), SHIFT_REPEAT(111),
  [209] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0),
  [211] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_or_expr_repeat1, 2, 0, 0),
  [213] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_or_expr_repeat1, 2, 0, 0), SHIFT_REPEAT(44),
  [216] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_or_expr_repeat1, 2, 0, 0),
  [218] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_or_expr, 2, 0, 0),
  [220] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_or_expr, 2, 0, 0),
  [222] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(145),
  [225] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(144),
  [228] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(144),
  [231] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0),
  [233] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(39),
  [236] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_compare_expr, 3, 0, 0),
  [238] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_expr, 3, 0, 0),
  [240] = {.entry = {.count = 1, .reusable = true}}, SHIFT(35),
  [242] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_attr, 2, 0, 0),
  [244] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_attr, 2, 0, 0),
  [246] = {.entry = {.count = 1, .reusable = false}}, SHIFT(8),
  [248] = {.entry = {.count = 1, .reusable = false}}, SHIFT(81),
  [250] = {.entry = {.count = 1, .reusable = false}}, SHIFT(84),
  [252] = {.entry = {.count = 1, .reusable = true}}, SHIFT(86),
  [254] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_condition, 1, 0, 0),
  [256] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_condition, 1, 0, 0),
  [258] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_attr, 3, 0, 0),
  [260] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_attr, 3, 0, 0),
  [262] = {.entry = {.count = 1, .reusable = true}}, SHIFT(124),
  [264] = {.entry = {.count = 1, .reusable = true}}, SHIFT(24),
  [266] = {.entry = {.count = 1, .reusable = true}}, SHIFT(61),
  [268] = {.entry = {.count = 1, .reusable = true}}, SHIFT(13),
  [270] = {.entry = {.count = 1, .reusable = true}}, SHIFT(60),
  [272] = {.entry = {.count = 1, .reusable = true}}, SHIFT(11),
  [274] = {.entry = {.count = 1, .reusable = true}}, SHIFT(20),
  [276] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_stylesheet_rule, 4, 0, 0),
  [278] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_stylesheet_rule, 4, 0, 0),
  [280] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_defaults_section_repeat1, 2, 0, 0), SHIFT_REPEAT(124),
  [283] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_defaults_section_repeat1, 2, 0, 0),
  [285] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_defaults_section_repeat1, 2, 0, 0), SHIFT_REPEAT(60),
  [288] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_agent_node_repeat1, 2, 0, 0), SHIFT_REPEAT(124),
  [291] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_agent_node_repeat1, 2, 0, 0),
  [293] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_agent_node_repeat1, 2, 0, 0), SHIFT_REPEAT(61),
  [296] = {.entry = {.count = 1, .reusable = true}}, SHIFT(18),
  [298] = {.entry = {.count = 1, .reusable = true}}, SHIFT(17),
  [300] = {.entry = {.count = 1, .reusable = true}}, SHIFT(19),
  [302] = {.entry = {.count = 1, .reusable = true}}, SHIFT(15),
  [304] = {.entry = {.count = 1, .reusable = true}}, SHIFT(63),
  [306] = {.entry = {.count = 1, .reusable = true}}, SHIFT(67),
  [308] = {.entry = {.count = 1, .reusable = true}}, SHIFT(57),
  [310] = {.entry = {.count = 1, .reusable = true}}, SHIFT(55),
  [312] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 2, 0, 0), SHIFT_REPEAT(124),
  [315] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 2, 0, 0),
  [317] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 2, 0, 0), SHIFT_REPEAT(72),
  [320] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_edges_section_repeat1, 2, 0, 0), SHIFT_REPEAT(117),
  [323] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_edges_section_repeat1, 2, 0, 0),
  [325] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_edges_section_repeat1, 2, 0, 0), SHIFT_REPEAT(73),
  [328] = {.entry = {.count = 1, .reusable = true}}, SHIFT(56),
  [330] = {.entry = {.count = 1, .reusable = true}}, SHIFT(58),
  [332] = {.entry = {.count = 1, .reusable = true}}, SHIFT(66),
  [334] = {.entry = {.count = 1, .reusable = true}}, SHIFT(59),
  [336] = {.entry = {.count = 1, .reusable = true}}, SHIFT(72),
  [338] = {.entry = {.count = 1, .reusable = true}}, SHIFT(64),
  [340] = {.entry = {.count = 1, .reusable = true}}, SHIFT(117),
  [342] = {.entry = {.count = 1, .reusable = true}}, SHIFT(14),
  [344] = {.entry = {.count = 1, .reusable = true}}, SHIFT(73),
  [346] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_string_repeat1, 2, 0, 0),
  [348] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_string_repeat1, 2, 0, 0), SHIFT_REPEAT(80),
  [351] = {.entry = {.count = 1, .reusable = false}}, SHIFT(3),
  [353] = {.entry = {.count = 1, .reusable = false}}, SHIFT(91),
  [355] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_block_content_repeat1, 2, 0, 0), SHIFT_REPEAT(82),
  [358] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_block_content_repeat1, 2, 0, 0),
  [360] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_string_repeat2, 2, 0, 0),
  [362] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_string_repeat2, 2, 0, 0), SHIFT_REPEAT(83),
  [365] = {.entry = {.count = 1, .reusable = false}}, SHIFT(92),
  [367] = {.entry = {.count = 1, .reusable = true}}, SHIFT(79),
  [369] = {.entry = {.count = 1, .reusable = true}}, SHIFT(93),
  [371] = {.entry = {.count = 1, .reusable = true}}, SHIFT(116),
  [373] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_identifier_list, 2, 0, 0),
  [375] = {.entry = {.count = 1, .reusable = true}}, SHIFT(77),
  [377] = {.entry = {.count = 1, .reusable = true}}, SHIFT(95),
  [379] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_identifier_list_repeat1, 2, 0, 0), SHIFT_REPEAT(116),
  [382] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_identifier_list_repeat1, 2, 0, 0),
  [384] = {.entry = {.count = 1, .reusable = false}}, SHIFT(2),
  [386] = {.entry = {.count = 1, .reusable = false}}, SHIFT(80),
  [388] = {.entry = {.count = 1, .reusable = false}}, SHIFT(83),
  [390] = {.entry = {.count = 1, .reusable = true}}, SHIFT(82),
  [392] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_block_content, 1, 0, 0),
  [394] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_identifier_list, 1, 0, 0),
  [396] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0),
  [398] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(95),
  [401] = {.entry = {.count = 1, .reusable = true}}, SHIFT(71),
  [403] = {.entry = {.count = 1, .reusable = true}}, SHIFT(22),
  [405] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_defaults_field, 3, 0, 0),
  [407] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_node_field, 3, 0, 0),
  [409] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_op, 1, 0, 0),
  [411] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 3, 0, 0),
  [413] = {.entry = {.count = 1, .reusable = true}}, SHIFT(21),
  [415] = {.entry = {.count = 1, .reusable = true}}, SHIFT(94),
  [417] = {.entry = {.count = 1, .reusable = true}}, SHIFT(107),
  [419] = {.entry = {.count = 1, .reusable = true}}, SHIFT(75),
  [421] = {.entry = {.count = 1, .reusable = true}}, SHIFT(103),
  [423] = {.entry = {.count = 1, .reusable = true}}, SHIFT(104),
  [425] = {.entry = {.count = 1, .reusable = true}}, SHIFT(136),
  [427] = {.entry = {.count = 1, .reusable = true}}, SHIFT(142),
  [429] = {.entry = {.count = 1, .reusable = true}}, SHIFT(49),
  [431] = {.entry = {.count = 1, .reusable = true}}, SHIFT(143),
  [433] = {.entry = {.count = 1, .reusable = true}}, SHIFT(50),
  [435] = {.entry = {.count = 1, .reusable = true}}, SHIFT(125),
  [437] = {.entry = {.count = 1, .reusable = true}}, SHIFT(42),
  [439] = {.entry = {.count = 1, .reusable = true}}, SHIFT(100),
  [441] = {.entry = {.count = 1, .reusable = true}}, SHIFT(141),
  [443] = {.entry = {.count = 1, .reusable = true}}, SHIFT(106),
  [445] = {.entry = {.count = 1, .reusable = true}}, SHIFT(88),
  [447] = {.entry = {.count = 1, .reusable = true}}, SHIFT(7),
  [449] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_workflow_decl, 5, 0, 0),
  [451] = {.entry = {.count = 1, .reusable = true}}, SHIFT(108),
  [453] = {.entry = {.count = 1, .reusable = true}}, SHIFT(133),
  [455] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_field_name, 1, 0, 0),
  [457] = {.entry = {.count = 1, .reusable = true}}, SHIFT(70),
  [459] = {.entry = {.count = 1, .reusable = true}}, SHIFT(115),
  [461] = {.entry = {.count = 1, .reusable = true}}, SHIFT(121),
  [463] = {.entry = {.count = 1, .reusable = true}}, SHIFT(26),
  [465] = {.entry = {.count = 1, .reusable = true}}, SHIFT(53),
  [467] = {.entry = {.count = 1, .reusable = true}},  ACCEPT_INPUT(),
  [469] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 1, 0, 0),
  [471] = {.entry = {.count = 1, .reusable = true}}, SHIFT(54),
  [473] = {.entry = {.count = 1, .reusable = true}}, SHIFT(4),
  [475] = {.entry = {.count = 1, .reusable = true}}, SHIFT(47),
  [477] = {.entry = {.count = 1, .reusable = true}}, SHIFT(78),
  [479] = {.entry = {.count = 1, .reusable = true}}, SHIFT(68),
  [481] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_selector, 2, 0, 0),
  [483] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 2, 0, 0),
  [485] = {.entry = {.count = 1, .reusable = true}}, SHIFT(74),
  [487] = {.entry = {.count = 1, .reusable = true}}, SHIFT(135),
  [489] = {.entry = {.count = 1, .reusable = true}}, SHIFT(32),
  [491] = {.entry = {.count = 1, .reusable = true}}, SHIFT(76),
  [493] = {.entry = {.count = 1, .reusable = true}}, SHIFT(69),
  [495] = {.entry = {.count = 1, .reusable = true}}, SHIFT(137),
  [497] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_selector, 1, 0, 0),
  [499] = {.entry = {.count = 1, .reusable = true}}, SHIFT(85),
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
