#include "tree_sitter/parser.h"

#if defined(__GNUC__) || defined(__clang__)
#pragma GCC diagnostic ignored "-Wmissing-field-initializers"
#endif

#define LANGUAGE_VERSION 14
#define STATE_COUNT 144
#define LARGE_STATE_COUNT 2
#define SYMBOL_COUNT 97
#define ALIAS_COUNT 0
#define TOKEN_COUNT 48
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
  sym_comment = 44,
  sym__indent = 45,
  sym__dedent = 46,
  sym__newline = 47,
  sym_source_file = 48,
  sym_workflow_decl = 49,
  sym_workflow_body = 50,
  sym_workflow_field = 51,
  sym_defaults_section = 52,
  sym_defaults_field = 53,
  sym_node_decl = 54,
  sym_agent_node = 55,
  sym_human_node = 56,
  sym_tool_node = 57,
  sym_subgraph_node = 58,
  sym_conditional_node = 59,
  sym_manager_loop_node = 60,
  sym_parallel_node = 61,
  sym_fan_in_node = 62,
  sym_node_attr_block = 63,
  sym_node_field = 64,
  sym_edges_section = 65,
  sym_edge_entry = 66,
  sym_edge_attr = 67,
  sym_condition = 68,
  sym_or_expr = 69,
  sym_and_expr = 70,
  sym_compare_expr = 71,
  sym_compare_op = 72,
  sym_operand = 73,
  sym_variable = 74,
  sym_stylesheet_section = 75,
  sym_stylesheet_rule = 76,
  sym_selector = 77,
  sym_field_name = 78,
  sym_field_value = 79,
  sym_multiline_block = 80,
  sym_block_content = 81,
  sym_identifier_list = 82,
  sym_string = 83,
  aux_sym_source_file_repeat1 = 84,
  aux_sym_workflow_body_repeat1 = 85,
  aux_sym_defaults_section_repeat1 = 86,
  aux_sym_agent_node_repeat1 = 87,
  aux_sym_edges_section_repeat1 = 88,
  aux_sym_edge_entry_repeat1 = 89,
  aux_sym_or_expr_repeat1 = 90,
  aux_sym_and_expr_repeat1 = 91,
  aux_sym_stylesheet_section_repeat1 = 92,
  aux_sym_stylesheet_rule_repeat1 = 93,
  aux_sym_block_content_repeat1 = 94,
  aux_sym_identifier_list_repeat1 = 95,
  aux_sym_string_repeat1 = 96,
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
};

static bool ts_lex(TSLexer *lexer, TSStateId state) {
  START_LEXER();
  eof = lexer->eof(lexer);
  switch (state) {
    case 0:
      if (eof) ADVANCE(9);
      ADVANCE_MAP(
        '!', 5,
        '"', 25,
        '#', 18,
        '*', 17,
        ',', 23,
        '-', 6,
        '.', 16,
        ':', 10,
        '<', 4,
        '=', 15,
        '\\', 7,
      );
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') SKIP(0);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(24);
      END_STATE();
    case 1:
      if (lookahead == '"') ADVANCE(25);
      if (lookahead == '#') ADVANCE(30);
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') ADVANCE(19);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n') ADVANCE(20);
      END_STATE();
    case 2:
      if (lookahead == '"') ADVANCE(25);
      if (lookahead == '#') ADVANCE(26);
      if (lookahead == '\\') ADVANCE(7);
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') ADVANCE(27);
      if (lookahead != 0) ADVANCE(28);
      END_STATE();
    case 3:
      if (lookahead == '#') ADVANCE(22);
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') ADVANCE(21);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n') ADVANCE(22);
      END_STATE();
    case 4:
      if (lookahead == '-') ADVANCE(12);
      END_STATE();
    case 5:
      if (lookahead == '=') ADVANCE(14);
      END_STATE();
    case 6:
      if (lookahead == '>') ADVANCE(11);
      END_STATE();
    case 7:
      if (lookahead != 0 &&
          lookahead != '\n') ADVANCE(29);
      END_STATE();
    case 8:
      if (eof) ADVANCE(9);
      ADVANCE_MAP(
        '!', 5,
        '"', 25,
        '#', 30,
        ',', 23,
        '-', 6,
        '.', 16,
        ':', 10,
        '<', 4,
        '=', 15,
      );
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') SKIP(8);
      if (('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(24);
      END_STATE();
    case 9:
      ACCEPT_TOKEN(ts_builtin_sym_end);
      END_STATE();
    case 10:
      ACCEPT_TOKEN(anon_sym_COLON);
      END_STATE();
    case 11:
      ACCEPT_TOKEN(anon_sym_DASH_GT);
      END_STATE();
    case 12:
      ACCEPT_TOKEN(anon_sym_LT_DASH);
      END_STATE();
    case 13:
      ACCEPT_TOKEN(anon_sym_EQ_EQ);
      END_STATE();
    case 14:
      ACCEPT_TOKEN(anon_sym_BANG_EQ);
      END_STATE();
    case 15:
      ACCEPT_TOKEN(anon_sym_EQ);
      if (lookahead == '=') ADVANCE(13);
      END_STATE();
    case 16:
      ACCEPT_TOKEN(anon_sym_DOT);
      END_STATE();
    case 17:
      ACCEPT_TOKEN(anon_sym_STAR);
      END_STATE();
    case 18:
      ACCEPT_TOKEN(anon_sym_POUND);
      if (lookahead != 0 &&
          lookahead != '\n') ADVANCE(30);
      END_STATE();
    case 19:
      ACCEPT_TOKEN(sym_raw_inline);
      if (lookahead == '"') ADVANCE(25);
      if (lookahead == '#') ADVANCE(30);
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') ADVANCE(19);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n') ADVANCE(20);
      END_STATE();
    case 20:
      ACCEPT_TOKEN(sym_raw_inline);
      if (lookahead != 0 &&
          lookahead != '\n') ADVANCE(20);
      END_STATE();
    case 21:
      ACCEPT_TOKEN(sym_block_line);
      if (lookahead == '#') ADVANCE(22);
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') ADVANCE(21);
      if (lookahead != 0 &&
          lookahead != '\t' &&
          lookahead != '\n') ADVANCE(22);
      END_STATE();
    case 22:
      ACCEPT_TOKEN(sym_block_line);
      if (lookahead != 0 &&
          lookahead != '\n') ADVANCE(22);
      END_STATE();
    case 23:
      ACCEPT_TOKEN(anon_sym_COMMA);
      END_STATE();
    case 24:
      ACCEPT_TOKEN(sym_identifier);
      if (lookahead == '-' ||
          ('0' <= lookahead && lookahead <= '9') ||
          ('A' <= lookahead && lookahead <= 'Z') ||
          lookahead == '_' ||
          ('a' <= lookahead && lookahead <= 'z')) ADVANCE(24);
      END_STATE();
    case 25:
      ACCEPT_TOKEN(anon_sym_DQUOTE);
      END_STATE();
    case 26:
      ACCEPT_TOKEN(aux_sym_string_token1);
      if (lookahead == '\n') ADVANCE(28);
      if (lookahead == '"' ||
          lookahead == '\\') ADVANCE(30);
      if (lookahead != 0) ADVANCE(26);
      END_STATE();
    case 27:
      ACCEPT_TOKEN(aux_sym_string_token1);
      if (lookahead == '#') ADVANCE(26);
      if (lookahead == '\t' ||
          lookahead == '\r' ||
          lookahead == ' ') ADVANCE(27);
      if (lookahead != 0 &&
          lookahead != '"' &&
          lookahead != '#' &&
          lookahead != '\\') ADVANCE(28);
      END_STATE();
    case 28:
      ACCEPT_TOKEN(aux_sym_string_token1);
      if (lookahead != 0 &&
          lookahead != '"' &&
          lookahead != '\\') ADVANCE(28);
      END_STATE();
    case 29:
      ACCEPT_TOKEN(aux_sym_string_token2);
      END_STATE();
    case 30:
      ACCEPT_TOKEN(sym_comment);
      if (lookahead != 0 &&
          lookahead != '\n') ADVANCE(30);
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
  [1] = {.lex_state = 8, .external_lex_state = 2},
  [2] = {.lex_state = 8, .external_lex_state = 3},
  [3] = {.lex_state = 8, .external_lex_state = 3},
  [4] = {.lex_state = 8, .external_lex_state = 3},
  [5] = {.lex_state = 8, .external_lex_state = 2},
  [6] = {.lex_state = 8, .external_lex_state = 3},
  [7] = {.lex_state = 8, .external_lex_state = 3},
  [8] = {.lex_state = 8, .external_lex_state = 3},
  [9] = {.lex_state = 8, .external_lex_state = 3},
  [10] = {.lex_state = 8, .external_lex_state = 3},
  [11] = {.lex_state = 8, .external_lex_state = 3},
  [12] = {.lex_state = 8, .external_lex_state = 3},
  [13] = {.lex_state = 8, .external_lex_state = 3},
  [14] = {.lex_state = 8, .external_lex_state = 3},
  [15] = {.lex_state = 8, .external_lex_state = 3},
  [16] = {.lex_state = 8, .external_lex_state = 3},
  [17] = {.lex_state = 8, .external_lex_state = 3},
  [18] = {.lex_state = 8, .external_lex_state = 3},
  [19] = {.lex_state = 8, .external_lex_state = 3},
  [20] = {.lex_state = 8, .external_lex_state = 3},
  [21] = {.lex_state = 8, .external_lex_state = 3},
  [22] = {.lex_state = 8, .external_lex_state = 3},
  [23] = {.lex_state = 8, .external_lex_state = 3},
  [24] = {.lex_state = 8, .external_lex_state = 3},
  [25] = {.lex_state = 8, .external_lex_state = 3},
  [26] = {.lex_state = 8, .external_lex_state = 3},
  [27] = {.lex_state = 8, .external_lex_state = 3},
  [28] = {.lex_state = 8, .external_lex_state = 3},
  [29] = {.lex_state = 8, .external_lex_state = 3},
  [30] = {.lex_state = 8},
  [31] = {.lex_state = 0, .external_lex_state = 3},
  [32] = {.lex_state = 8, .external_lex_state = 3},
  [33] = {.lex_state = 8, .external_lex_state = 3},
  [34] = {.lex_state = 8, .external_lex_state = 3},
  [35] = {.lex_state = 8, .external_lex_state = 3},
  [36] = {.lex_state = 8, .external_lex_state = 3},
  [37] = {.lex_state = 0, .external_lex_state = 3},
  [38] = {.lex_state = 8, .external_lex_state = 3},
  [39] = {.lex_state = 8, .external_lex_state = 3},
  [40] = {.lex_state = 8, .external_lex_state = 3},
  [41] = {.lex_state = 8, .external_lex_state = 3},
  [42] = {.lex_state = 0, .external_lex_state = 2},
  [43] = {.lex_state = 8, .external_lex_state = 3},
  [44] = {.lex_state = 8},
  [45] = {.lex_state = 8, .external_lex_state = 3},
  [46] = {.lex_state = 8, .external_lex_state = 3},
  [47] = {.lex_state = 8},
  [48] = {.lex_state = 8, .external_lex_state = 3},
  [49] = {.lex_state = 1, .external_lex_state = 4},
  [50] = {.lex_state = 8, .external_lex_state = 3},
  [51] = {.lex_state = 1, .external_lex_state = 4},
  [52] = {.lex_state = 8, .external_lex_state = 3},
  [53] = {.lex_state = 8, .external_lex_state = 3},
  [54] = {.lex_state = 8, .external_lex_state = 3},
  [55] = {.lex_state = 8, .external_lex_state = 3},
  [56] = {.lex_state = 8, .external_lex_state = 3},
  [57] = {.lex_state = 8, .external_lex_state = 3},
  [58] = {.lex_state = 8, .external_lex_state = 3},
  [59] = {.lex_state = 8, .external_lex_state = 3},
  [60] = {.lex_state = 1, .external_lex_state = 4},
  [61] = {.lex_state = 8, .external_lex_state = 3},
  [62] = {.lex_state = 1, .external_lex_state = 4},
  [63] = {.lex_state = 1, .external_lex_state = 4},
  [64] = {.lex_state = 0, .external_lex_state = 3},
  [65] = {.lex_state = 8},
  [66] = {.lex_state = 8, .external_lex_state = 2},
  [67] = {.lex_state = 8, .external_lex_state = 3},
  [68] = {.lex_state = 8, .external_lex_state = 2},
  [69] = {.lex_state = 8, .external_lex_state = 2},
  [70] = {.lex_state = 8, .external_lex_state = 2},
  [71] = {.lex_state = 8, .external_lex_state = 3},
  [72] = {.lex_state = 8, .external_lex_state = 2},
  [73] = {.lex_state = 8, .external_lex_state = 3},
  [74] = {.lex_state = 8, .external_lex_state = 2},
  [75] = {.lex_state = 8, .external_lex_state = 3},
  [76] = {.lex_state = 8, .external_lex_state = 2},
  [77] = {.lex_state = 8, .external_lex_state = 2},
  [78] = {.lex_state = 8},
  [79] = {.lex_state = 8},
  [80] = {.lex_state = 3, .external_lex_state = 3},
  [81] = {.lex_state = 3, .external_lex_state = 3},
  [82] = {.lex_state = 8, .external_lex_state = 5},
  [83] = {.lex_state = 2},
  [84] = {.lex_state = 8, .external_lex_state = 2},
  [85] = {.lex_state = 2},
  [86] = {.lex_state = 8, .external_lex_state = 5},
  [87] = {.lex_state = 8, .external_lex_state = 2},
  [88] = {.lex_state = 8, .external_lex_state = 5},
  [89] = {.lex_state = 2},
  [90] = {.lex_state = 3, .external_lex_state = 2},
  [91] = {.lex_state = 8, .external_lex_state = 2},
  [92] = {.lex_state = 8, .external_lex_state = 3},
  [93] = {.lex_state = 8, .external_lex_state = 5},
  [94] = {.lex_state = 8, .external_lex_state = 5},
  [95] = {.lex_state = 8, .external_lex_state = 5},
  [96] = {.lex_state = 8, .external_lex_state = 2},
  [97] = {.lex_state = 8, .external_lex_state = 3},
  [98] = {.lex_state = 8, .external_lex_state = 3},
  [99] = {.lex_state = 8},
  [100] = {.lex_state = 8},
  [101] = {.lex_state = 8},
  [102] = {.lex_state = 8},
  [103] = {.lex_state = 8},
  [104] = {.lex_state = 8},
  [105] = {.lex_state = 8, .external_lex_state = 6},
  [106] = {.lex_state = 8},
  [107] = {.lex_state = 8, .external_lex_state = 4},
  [108] = {.lex_state = 8},
  [109] = {.lex_state = 8, .external_lex_state = 4},
  [110] = {.lex_state = 8},
  [111] = {.lex_state = 8},
  [112] = {.lex_state = 8, .external_lex_state = 4},
  [113] = {.lex_state = 8, .external_lex_state = 4},
  [114] = {.lex_state = 8, .external_lex_state = 4},
  [115] = {.lex_state = 8, .external_lex_state = 4},
  [116] = {.lex_state = 8},
  [117] = {.lex_state = 8},
  [118] = {.lex_state = 8},
  [119] = {.lex_state = 8, .external_lex_state = 4},
  [120] = {.lex_state = 8},
  [121] = {.lex_state = 8},
  [122] = {.lex_state = 8, .external_lex_state = 4},
  [123] = {.lex_state = 8},
  [124] = {.lex_state = 8},
  [125] = {.lex_state = 8},
  [126] = {.lex_state = 8},
  [127] = {.lex_state = 8},
  [128] = {.lex_state = 8, .external_lex_state = 4},
  [129] = {.lex_state = 8, .external_lex_state = 4},
  [130] = {.lex_state = 8},
  [131] = {.lex_state = 8},
  [132] = {.lex_state = 8},
  [133] = {.lex_state = 8},
  [134] = {.lex_state = 8, .external_lex_state = 4},
  [135] = {.lex_state = 8},
  [136] = {.lex_state = 8},
  [137] = {.lex_state = 8, .external_lex_state = 6},
  [138] = {.lex_state = 8},
  [139] = {.lex_state = 8},
  [140] = {.lex_state = 8, .external_lex_state = 4},
  [141] = {.lex_state = 8},
  [142] = {.lex_state = 8, .external_lex_state = 4},
  [143] = {.lex_state = 8},
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
    [sym_comment] = ACTIONS(3),
    [sym__indent] = ACTIONS(1),
    [sym__dedent] = ACTIONS(1),
    [sym__newline] = ACTIONS(1),
  },
  [1] = {
    [sym_source_file] = STATE(102),
    [sym_workflow_decl] = STATE(139),
    [aux_sym_source_file_repeat1] = STATE(87),
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
    STATE(13), 8,
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
      sym__newline,
    STATE(137), 1,
      sym_workflow_body,
    ACTIONS(60), 4,
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
  [214] = 17,
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
    ACTIONS(86), 1,
      sym__dedent,
    ACTIONS(88), 1,
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
    STATE(13), 8,
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
    STATE(79), 1,
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
  [403] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(116), 4,
      sym__dedent,
      sym__newline,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
    ACTIONS(114), 13,
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
  [428] = 3,
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
  [453] = 2,
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
  [476] = 2,
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
  [499] = 2,
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
  [522] = 2,
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
  [545] = 2,
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
  [568] = 2,
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
  [591] = 2,
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
  [614] = 2,
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
  [637] = 2,
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
  [660] = 2,
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
  [683] = 2,
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
  [706] = 2,
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
  [729] = 2,
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
  [752] = 2,
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
  [775] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(148), 1,
      anon_sym_and,
    STATE(29), 1,
      aux_sym_and_expr_repeat1,
    ACTIONS(150), 2,
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
    ACTIONS(154), 1,
      anon_sym_and,
    STATE(28), 1,
      aux_sym_and_expr_repeat1,
    ACTIONS(157), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(152), 6,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_or,
      sym_identifier,
  [819] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(148), 1,
      anon_sym_and,
    STATE(28), 1,
      aux_sym_and_expr_repeat1,
    ACTIONS(161), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(159), 6,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_or,
      sym_identifier,
  [841] = 9,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(163), 1,
      sym_identifier,
    ACTIONS(165), 1,
      anon_sym_DQUOTE,
    STATE(10), 1,
      sym_operand,
    STATE(27), 1,
      sym_compare_expr,
    STATE(32), 1,
      sym_and_expr,
    STATE(45), 1,
      sym_condition,
    STATE(46), 1,
      sym_or_expr,
    STATE(12), 2,
      sym_variable,
      sym_string,
  [870] = 8,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(170), 1,
      anon_sym_DOT,
    ACTIONS(173), 1,
      anon_sym_POUND,
    ACTIONS(176), 1,
      sym__dedent,
    ACTIONS(178), 1,
      sym__newline,
    STATE(122), 1,
      sym_selector,
    ACTIONS(167), 2,
      anon_sym_STAR,
      sym_identifier,
    STATE(31), 2,
      sym_stylesheet_rule,
      aux_sym_stylesheet_section_repeat1,
  [897] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(183), 1,
      anon_sym_or,
    STATE(36), 1,
      aux_sym_or_expr_repeat1,
    ACTIONS(185), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(181), 5,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      sym_identifier,
  [918] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(187), 1,
      sym_identifier,
    ACTIONS(189), 1,
      anon_sym_when,
    ACTIONS(193), 2,
      sym__dedent,
      sym__newline,
    STATE(34), 2,
      sym_edge_attr,
      aux_sym_edge_entry_repeat1,
    ACTIONS(191), 3,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
  [941] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(189), 1,
      anon_sym_when,
    ACTIONS(195), 1,
      sym_identifier,
    ACTIONS(197), 2,
      sym__dedent,
      sym__newline,
    STATE(35), 2,
      sym_edge_attr,
      aux_sym_edge_entry_repeat1,
    ACTIONS(191), 3,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
  [964] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(199), 1,
      sym_identifier,
    ACTIONS(201), 1,
      anon_sym_when,
    ACTIONS(207), 2,
      sym__dedent,
      sym__newline,
    STATE(35), 2,
      sym_edge_attr,
      aux_sym_edge_entry_repeat1,
    ACTIONS(204), 3,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
  [987] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(183), 1,
      anon_sym_or,
    STATE(38), 1,
      aux_sym_or_expr_repeat1,
    ACTIONS(211), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(209), 5,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      sym_identifier,
  [1008] = 8,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(215), 1,
      anon_sym_DOT,
    ACTIONS(217), 1,
      anon_sym_POUND,
    ACTIONS(219), 1,
      sym__dedent,
    ACTIONS(221), 1,
      sym__newline,
    STATE(122), 1,
      sym_selector,
    ACTIONS(213), 2,
      anon_sym_STAR,
      sym_identifier,
    STATE(31), 2,
      sym_stylesheet_rule,
      aux_sym_stylesheet_section_repeat1,
  [1035] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(225), 1,
      anon_sym_or,
    STATE(38), 1,
      aux_sym_or_expr_repeat1,
    ACTIONS(228), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(223), 5,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      sym_identifier,
  [1056] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(157), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(152), 7,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_or,
      anon_sym_and,
      sym_identifier,
  [1073] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(232), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(230), 7,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_or,
      anon_sym_and,
      sym_identifier,
  [1090] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(236), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(234), 7,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_or,
      anon_sym_and,
      sym_identifier,
  [1107] = 7,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(215), 1,
      anon_sym_DOT,
    ACTIONS(217), 1,
      anon_sym_POUND,
    ACTIONS(238), 1,
      sym__newline,
    STATE(122), 1,
      sym_selector,
    ACTIONS(213), 2,
      anon_sym_STAR,
      sym_identifier,
    STATE(37), 2,
      sym_stylesheet_rule,
      aux_sym_stylesheet_section_repeat1,
  [1131] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(228), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(223), 6,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      anon_sym_or,
      sym_identifier,
  [1147] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(110), 1,
      anon_sym_EQ,
    STATE(78), 1,
      sym_compare_op,
    ACTIONS(108), 6,
      anon_sym_EQ_EQ,
      anon_sym_BANG_EQ,
      anon_sym_contains,
      anon_sym_startswith,
      anon_sym_endswith,
      anon_sym_in,
  [1165] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(242), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(240), 5,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      sym_identifier,
  [1180] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(246), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(244), 5,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      sym_identifier,
  [1195] = 7,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(163), 1,
      sym_identifier,
    ACTIONS(165), 1,
      anon_sym_DQUOTE,
    STATE(10), 1,
      sym_operand,
    STATE(27), 1,
      sym_compare_expr,
    STATE(43), 1,
      sym_and_expr,
    STATE(12), 2,
      sym_variable,
      sym_string,
  [1218] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(250), 2,
      sym__dedent,
      sym__newline,
    ACTIONS(248), 5,
      anon_sym_when,
      anon_sym_label,
      anon_sym_weight,
      anon_sym_restart,
      sym_identifier,
  [1233] = 6,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(252), 1,
      sym_raw_inline,
    ACTIONS(254), 1,
      anon_sym_DQUOTE,
    ACTIONS(256), 1,
      sym__indent,
    STATE(97), 1,
      sym_field_value,
    STATE(7), 2,
      sym_multiline_block,
      sym_string,
  [1253] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(258), 1,
      sym_identifier,
    ACTIONS(260), 1,
      sym__dedent,
    ACTIONS(262), 1,
      sym__newline,
    STATE(136), 1,
      sym_field_name,
    STATE(52), 2,
      sym_defaults_field,
      aux_sym_defaults_section_repeat1,
  [1273] = 6,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(252), 1,
      sym_raw_inline,
    ACTIONS(254), 1,
      anon_sym_DQUOTE,
    ACTIONS(256), 1,
      sym__indent,
    STATE(98), 1,
      sym_field_value,
    STATE(7), 2,
      sym_multiline_block,
      sym_string,
  [1293] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(264), 1,
      sym_identifier,
    ACTIONS(267), 1,
      sym__dedent,
    ACTIONS(269), 1,
      sym__newline,
    STATE(136), 1,
      sym_field_name,
    STATE(52), 2,
      sym_defaults_field,
      aux_sym_defaults_section_repeat1,
  [1313] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(258), 1,
      sym_identifier,
    ACTIONS(272), 1,
      sym__dedent,
    ACTIONS(274), 1,
      sym__newline,
    STATE(138), 1,
      sym_field_name,
    STATE(59), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1333] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(258), 1,
      sym_identifier,
    ACTIONS(274), 1,
      sym__newline,
    ACTIONS(276), 1,
      sym__dedent,
    STATE(138), 1,
      sym_field_name,
    STATE(59), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1353] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(258), 1,
      sym_identifier,
    ACTIONS(274), 1,
      sym__newline,
    ACTIONS(278), 1,
      sym__dedent,
    STATE(138), 1,
      sym_field_name,
    STATE(59), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1373] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(258), 1,
      sym_identifier,
    ACTIONS(274), 1,
      sym__newline,
    ACTIONS(280), 1,
      sym__dedent,
    STATE(138), 1,
      sym_field_name,
    STATE(59), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1393] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(258), 1,
      sym_identifier,
    ACTIONS(274), 1,
      sym__newline,
    ACTIONS(282), 1,
      sym__dedent,
    STATE(138), 1,
      sym_field_name,
    STATE(59), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1413] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(258), 1,
      sym_identifier,
    ACTIONS(274), 1,
      sym__newline,
    ACTIONS(284), 1,
      sym__dedent,
    STATE(138), 1,
      sym_field_name,
    STATE(59), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1433] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(286), 1,
      sym_identifier,
    ACTIONS(289), 1,
      sym__dedent,
    ACTIONS(291), 1,
      sym__newline,
    STATE(138), 1,
      sym_field_name,
    STATE(59), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1453] = 6,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(252), 1,
      sym_raw_inline,
    ACTIONS(254), 1,
      anon_sym_DQUOTE,
    ACTIONS(256), 1,
      sym__indent,
    STATE(26), 1,
      sym_field_value,
    STATE(7), 2,
      sym_multiline_block,
      sym_string,
  [1473] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(258), 1,
      sym_identifier,
    ACTIONS(274), 1,
      sym__newline,
    ACTIONS(294), 1,
      sym__dedent,
    STATE(138), 1,
      sym_field_name,
    STATE(59), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1493] = 6,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(252), 1,
      sym_raw_inline,
    ACTIONS(254), 1,
      anon_sym_DQUOTE,
    ACTIONS(256), 1,
      sym__indent,
    STATE(48), 1,
      sym_field_value,
    STATE(7), 2,
      sym_multiline_block,
      sym_string,
  [1513] = 6,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(252), 1,
      sym_raw_inline,
    ACTIONS(254), 1,
      anon_sym_DQUOTE,
    ACTIONS(256), 1,
      sym__indent,
    STATE(92), 1,
      sym_field_value,
    STATE(7), 2,
      sym_multiline_block,
      sym_string,
  [1533] = 3,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(298), 1,
      anon_sym_POUND,
    ACTIONS(296), 5,
      sym__dedent,
      sym__newline,
      anon_sym_DOT,
      anon_sym_STAR,
      sym_identifier,
  [1547] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(163), 1,
      sym_identifier,
    ACTIONS(165), 1,
      anon_sym_DQUOTE,
    STATE(10), 1,
      sym_operand,
    STATE(39), 1,
      sym_compare_expr,
    STATE(12), 2,
      sym_variable,
      sym_string,
  [1567] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(258), 1,
      sym_identifier,
    ACTIONS(300), 1,
      sym__newline,
    STATE(138), 1,
      sym_field_name,
    STATE(58), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1584] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(302), 1,
      sym_identifier,
    ACTIONS(305), 1,
      sym__dedent,
    ACTIONS(307), 1,
      sym__newline,
    STATE(67), 2,
      sym_edge_entry,
      aux_sym_edges_section_repeat1,
  [1601] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(258), 1,
      sym_identifier,
    ACTIONS(310), 1,
      sym__newline,
    STATE(138), 1,
      sym_field_name,
    STATE(57), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1618] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(258), 1,
      sym_identifier,
    ACTIONS(312), 1,
      sym__newline,
    STATE(138), 1,
      sym_field_name,
    STATE(61), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1635] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(258), 1,
      sym_identifier,
    ACTIONS(314), 1,
      sym__newline,
    STATE(136), 1,
      sym_field_name,
    STATE(50), 2,
      sym_defaults_field,
      aux_sym_defaults_section_repeat1,
  [1652] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(258), 1,
      sym_identifier,
    ACTIONS(316), 1,
      sym__dedent,
    ACTIONS(318), 1,
      sym__newline,
    STATE(75), 1,
      aux_sym_stylesheet_rule_repeat1,
    STATE(110), 1,
      sym_field_name,
  [1671] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(258), 1,
      sym_identifier,
    ACTIONS(320), 1,
      sym__newline,
    STATE(138), 1,
      sym_field_name,
    STATE(56), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1688] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(322), 1,
      sym_identifier,
    ACTIONS(324), 1,
      sym__dedent,
    ACTIONS(326), 1,
      sym__newline,
    STATE(67), 2,
      sym_edge_entry,
      aux_sym_edges_section_repeat1,
  [1705] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(258), 1,
      sym_identifier,
    ACTIONS(328), 1,
      sym__newline,
    STATE(138), 1,
      sym_field_name,
    STATE(54), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1722] = 6,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(330), 1,
      sym_identifier,
    ACTIONS(333), 1,
      sym__dedent,
    ACTIONS(335), 1,
      sym__newline,
    STATE(75), 1,
      aux_sym_stylesheet_rule_repeat1,
    STATE(110), 1,
      sym_field_name,
  [1741] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(258), 1,
      sym_identifier,
    ACTIONS(338), 1,
      sym__newline,
    STATE(138), 1,
      sym_field_name,
    STATE(53), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1758] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(258), 1,
      sym_identifier,
    ACTIONS(340), 1,
      sym__newline,
    STATE(138), 1,
      sym_field_name,
    STATE(55), 2,
      sym_node_field,
      aux_sym_agent_node_repeat1,
  [1775] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(163), 1,
      sym_identifier,
    ACTIONS(165), 1,
      anon_sym_DQUOTE,
    STATE(41), 1,
      sym_operand,
    STATE(12), 2,
      sym_variable,
      sym_string,
  [1792] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(163), 1,
      sym_identifier,
    ACTIONS(165), 1,
      anon_sym_DQUOTE,
    STATE(40), 1,
      sym_operand,
    STATE(12), 2,
      sym_variable,
      sym_string,
  [1809] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(344), 1,
      sym__dedent,
    STATE(81), 1,
      aux_sym_block_content_repeat1,
    ACTIONS(342), 2,
      sym__newline,
      sym_block_line,
  [1823] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(349), 1,
      sym__dedent,
    STATE(81), 1,
      aux_sym_block_content_repeat1,
    ACTIONS(346), 2,
      sym__newline,
      sym_block_line,
  [1837] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(351), 1,
      anon_sym_COMMA,
    STATE(86), 1,
      aux_sym_identifier_list_repeat1,
    ACTIONS(353), 2,
      sym__indent,
      sym__newline,
  [1851] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(355), 1,
      anon_sym_DQUOTE,
    STATE(83), 1,
      aux_sym_string_repeat1,
    ACTIONS(357), 2,
      aux_sym_string_token1,
      aux_sym_string_token2,
  [1865] = 5,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(258), 1,
      sym_identifier,
    ACTIONS(360), 1,
      sym__newline,
    STATE(71), 1,
      aux_sym_stylesheet_rule_repeat1,
    STATE(110), 1,
      sym_field_name,
  [1881] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(362), 1,
      anon_sym_DQUOTE,
    STATE(83), 1,
      aux_sym_string_repeat1,
    ACTIONS(364), 2,
      aux_sym_string_token1,
      aux_sym_string_token2,
  [1895] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(366), 1,
      anon_sym_COMMA,
    STATE(86), 1,
      aux_sym_identifier_list_repeat1,
    ACTIONS(369), 2,
      sym__indent,
      sym__newline,
  [1909] = 5,
    ACTIONS(5), 1,
      anon_sym_workflow,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(371), 1,
      sym__newline,
    STATE(96), 1,
      aux_sym_source_file_repeat1,
    STATE(133), 1,
      sym_workflow_decl,
  [1925] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(351), 1,
      anon_sym_COMMA,
    STATE(82), 1,
      aux_sym_identifier_list_repeat1,
    ACTIONS(373), 2,
      sym__indent,
      sym__newline,
  [1939] = 4,
    ACTIONS(3), 1,
      sym_comment,
    ACTIONS(375), 1,
      anon_sym_DQUOTE,
    STATE(85), 1,
      aux_sym_string_repeat1,
    ACTIONS(377), 2,
      aux_sym_string_token1,
      aux_sym_string_token2,
  [1953] = 4,
    ACTIONS(3), 1,
      sym_comment,
    STATE(80), 1,
      aux_sym_block_content_repeat1,
    STATE(105), 1,
      sym_block_content,
    ACTIONS(379), 2,
      sym__newline,
      sym_block_line,
  [1967] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(322), 1,
      sym_identifier,
    ACTIONS(381), 1,
      sym__newline,
    STATE(73), 2,
      sym_edge_entry,
      aux_sym_edges_section_repeat1,
  [1981] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(383), 3,
      sym__dedent,
      sym__newline,
      sym_identifier,
  [1990] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(385), 1,
      sym__indent,
    ACTIONS(387), 1,
      sym__newline,
    STATE(21), 1,
      sym_node_attr_block,
  [2003] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(385), 1,
      sym__indent,
    ACTIONS(389), 1,
      sym__newline,
    STATE(20), 1,
      sym_node_attr_block,
  [2016] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(369), 3,
      sym__indent,
      sym__newline,
      anon_sym_COMMA,
  [2025] = 4,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(391), 1,
      anon_sym_workflow,
    ACTIONS(393), 1,
      sym__newline,
    STATE(96), 1,
      aux_sym_source_file_repeat1,
  [2038] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(396), 3,
      sym__dedent,
      sym__newline,
      sym_identifier,
  [2047] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(398), 3,
      sym__dedent,
      sym__newline,
      sym_identifier,
  [2056] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(400), 1,
      sym_identifier,
    STATE(94), 1,
      sym_identifier_list,
  [2066] = 3,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(400), 1,
      sym_identifier,
    STATE(93), 1,
      sym_identifier_list,
  [2076] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(402), 2,
      sym_identifier,
      anon_sym_DQUOTE,
  [2084] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(404), 1,
      ts_builtin_sym_end,
  [2091] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(406), 1,
      sym_identifier,
  [2098] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(408), 1,
      sym_identifier,
  [2105] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(410), 1,
      sym__dedent,
  [2112] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(412), 1,
      sym_identifier,
  [2119] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(414), 1,
      sym__indent,
  [2126] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(416), 1,
      anon_sym_COLON,
  [2133] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(418), 1,
      sym__indent,
  [2140] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(420), 1,
      anon_sym_COLON,
  [2147] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(422), 1,
      anon_sym_DASH_GT,
  [2154] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(424), 1,
      sym__indent,
  [2161] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(426), 1,
      sym__indent,
  [2168] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(428), 1,
      sym__indent,
  [2175] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(430), 1,
      sym__indent,
  [2182] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(432), 1,
      anon_sym_DASH_GT,
  [2189] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(434), 1,
      anon_sym_LT_DASH,
  [2196] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(436), 1,
      anon_sym_COLON,
  [2203] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(438), 1,
      sym__indent,
  [2210] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(440), 1,
      sym_identifier,
  [2217] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(442), 1,
      ts_builtin_sym_end,
  [2224] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(444), 1,
      sym__indent,
  [2231] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(446), 1,
      sym_identifier,
  [2238] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(448), 1,
      sym_identifier,
  [2245] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(450), 1,
      sym_identifier,
  [2252] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(452), 1,
      sym_identifier,
  [2259] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(454), 1,
      sym_identifier,
  [2266] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(456), 1,
      sym__indent,
  [2273] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(458), 1,
      sym__indent,
  [2280] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(460), 1,
      anon_sym_COLON,
  [2287] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(462), 1,
      sym_identifier,
  [2294] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(464), 1,
      sym_identifier,
  [2301] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(466), 1,
      ts_builtin_sym_end,
  [2308] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(468), 1,
      sym__indent,
  [2315] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(470), 1,
      anon_sym_COLON,
  [2322] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(472), 1,
      anon_sym_COLON,
  [2329] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(474), 1,
      sym__dedent,
  [2336] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(476), 1,
      anon_sym_COLON,
  [2343] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(478), 1,
      ts_builtin_sym_end,
  [2350] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(480), 1,
      sym__indent,
  [2357] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(482), 1,
      sym_identifier,
  [2364] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(484), 1,
      sym__indent,
  [2371] = 2,
    ACTIONS(7), 1,
      sym_comment,
    ACTIONS(486), 1,
      sym_identifier,
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
  [SMALL_STATE(12)] = 428,
  [SMALL_STATE(13)] = 453,
  [SMALL_STATE(14)] = 476,
  [SMALL_STATE(15)] = 499,
  [SMALL_STATE(16)] = 522,
  [SMALL_STATE(17)] = 545,
  [SMALL_STATE(18)] = 568,
  [SMALL_STATE(19)] = 591,
  [SMALL_STATE(20)] = 614,
  [SMALL_STATE(21)] = 637,
  [SMALL_STATE(22)] = 660,
  [SMALL_STATE(23)] = 683,
  [SMALL_STATE(24)] = 706,
  [SMALL_STATE(25)] = 729,
  [SMALL_STATE(26)] = 752,
  [SMALL_STATE(27)] = 775,
  [SMALL_STATE(28)] = 797,
  [SMALL_STATE(29)] = 819,
  [SMALL_STATE(30)] = 841,
  [SMALL_STATE(31)] = 870,
  [SMALL_STATE(32)] = 897,
  [SMALL_STATE(33)] = 918,
  [SMALL_STATE(34)] = 941,
  [SMALL_STATE(35)] = 964,
  [SMALL_STATE(36)] = 987,
  [SMALL_STATE(37)] = 1008,
  [SMALL_STATE(38)] = 1035,
  [SMALL_STATE(39)] = 1056,
  [SMALL_STATE(40)] = 1073,
  [SMALL_STATE(41)] = 1090,
  [SMALL_STATE(42)] = 1107,
  [SMALL_STATE(43)] = 1131,
  [SMALL_STATE(44)] = 1147,
  [SMALL_STATE(45)] = 1165,
  [SMALL_STATE(46)] = 1180,
  [SMALL_STATE(47)] = 1195,
  [SMALL_STATE(48)] = 1218,
  [SMALL_STATE(49)] = 1233,
  [SMALL_STATE(50)] = 1253,
  [SMALL_STATE(51)] = 1273,
  [SMALL_STATE(52)] = 1293,
  [SMALL_STATE(53)] = 1313,
  [SMALL_STATE(54)] = 1333,
  [SMALL_STATE(55)] = 1353,
  [SMALL_STATE(56)] = 1373,
  [SMALL_STATE(57)] = 1393,
  [SMALL_STATE(58)] = 1413,
  [SMALL_STATE(59)] = 1433,
  [SMALL_STATE(60)] = 1453,
  [SMALL_STATE(61)] = 1473,
  [SMALL_STATE(62)] = 1493,
  [SMALL_STATE(63)] = 1513,
  [SMALL_STATE(64)] = 1533,
  [SMALL_STATE(65)] = 1547,
  [SMALL_STATE(66)] = 1567,
  [SMALL_STATE(67)] = 1584,
  [SMALL_STATE(68)] = 1601,
  [SMALL_STATE(69)] = 1618,
  [SMALL_STATE(70)] = 1635,
  [SMALL_STATE(71)] = 1652,
  [SMALL_STATE(72)] = 1671,
  [SMALL_STATE(73)] = 1688,
  [SMALL_STATE(74)] = 1705,
  [SMALL_STATE(75)] = 1722,
  [SMALL_STATE(76)] = 1741,
  [SMALL_STATE(77)] = 1758,
  [SMALL_STATE(78)] = 1775,
  [SMALL_STATE(79)] = 1792,
  [SMALL_STATE(80)] = 1809,
  [SMALL_STATE(81)] = 1823,
  [SMALL_STATE(82)] = 1837,
  [SMALL_STATE(83)] = 1851,
  [SMALL_STATE(84)] = 1865,
  [SMALL_STATE(85)] = 1881,
  [SMALL_STATE(86)] = 1895,
  [SMALL_STATE(87)] = 1909,
  [SMALL_STATE(88)] = 1925,
  [SMALL_STATE(89)] = 1939,
  [SMALL_STATE(90)] = 1953,
  [SMALL_STATE(91)] = 1967,
  [SMALL_STATE(92)] = 1981,
  [SMALL_STATE(93)] = 1990,
  [SMALL_STATE(94)] = 2003,
  [SMALL_STATE(95)] = 2016,
  [SMALL_STATE(96)] = 2025,
  [SMALL_STATE(97)] = 2038,
  [SMALL_STATE(98)] = 2047,
  [SMALL_STATE(99)] = 2056,
  [SMALL_STATE(100)] = 2066,
  [SMALL_STATE(101)] = 2076,
  [SMALL_STATE(102)] = 2084,
  [SMALL_STATE(103)] = 2091,
  [SMALL_STATE(104)] = 2098,
  [SMALL_STATE(105)] = 2105,
  [SMALL_STATE(106)] = 2112,
  [SMALL_STATE(107)] = 2119,
  [SMALL_STATE(108)] = 2126,
  [SMALL_STATE(109)] = 2133,
  [SMALL_STATE(110)] = 2140,
  [SMALL_STATE(111)] = 2147,
  [SMALL_STATE(112)] = 2154,
  [SMALL_STATE(113)] = 2161,
  [SMALL_STATE(114)] = 2168,
  [SMALL_STATE(115)] = 2175,
  [SMALL_STATE(116)] = 2182,
  [SMALL_STATE(117)] = 2189,
  [SMALL_STATE(118)] = 2196,
  [SMALL_STATE(119)] = 2203,
  [SMALL_STATE(120)] = 2210,
  [SMALL_STATE(121)] = 2217,
  [SMALL_STATE(122)] = 2224,
  [SMALL_STATE(123)] = 2231,
  [SMALL_STATE(124)] = 2238,
  [SMALL_STATE(125)] = 2245,
  [SMALL_STATE(126)] = 2252,
  [SMALL_STATE(127)] = 2259,
  [SMALL_STATE(128)] = 2266,
  [SMALL_STATE(129)] = 2273,
  [SMALL_STATE(130)] = 2280,
  [SMALL_STATE(131)] = 2287,
  [SMALL_STATE(132)] = 2294,
  [SMALL_STATE(133)] = 2301,
  [SMALL_STATE(134)] = 2308,
  [SMALL_STATE(135)] = 2315,
  [SMALL_STATE(136)] = 2322,
  [SMALL_STATE(137)] = 2329,
  [SMALL_STATE(138)] = 2336,
  [SMALL_STATE(139)] = 2343,
  [SMALL_STATE(140)] = 2350,
  [SMALL_STATE(141)] = 2357,
  [SMALL_STATE(142)] = 2364,
  [SMALL_STATE(143)] = 2371,
};

static const TSParseActionEntry ts_parse_actions[] = {
  [0] = {.entry = {.count = 0, .reusable = false}},
  [1] = {.entry = {.count = 1, .reusable = false}}, RECOVER(),
  [3] = {.entry = {.count = 1, .reusable = false}}, SHIFT_EXTRA(),
  [5] = {.entry = {.count = 1, .reusable = true}}, SHIFT(123),
  [7] = {.entry = {.count = 1, .reusable = true}}, SHIFT_EXTRA(),
  [9] = {.entry = {.count = 1, .reusable = true}}, SHIFT(87),
  [11] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_string, 2, 0, 0),
  [13] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_string, 2, 0, 0),
  [15] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_string, 3, 0, 0),
  [17] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_string, 3, 0, 0),
  [19] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(118),
  [22] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(134),
  [25] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(103),
  [28] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(104),
  [31] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(106),
  [34] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(120),
  [37] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(143),
  [40] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(124),
  [43] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(126),
  [46] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(127),
  [49] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(128),
  [52] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(130),
  [55] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0),
  [57] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_workflow_body_repeat1, 2, 0, 0), SHIFT_REPEAT(4),
  [60] = {.entry = {.count = 1, .reusable = true}}, SHIFT(118),
  [62] = {.entry = {.count = 1, .reusable = true}}, SHIFT(134),
  [64] = {.entry = {.count = 1, .reusable = true}}, SHIFT(103),
  [66] = {.entry = {.count = 1, .reusable = true}}, SHIFT(104),
  [68] = {.entry = {.count = 1, .reusable = true}}, SHIFT(106),
  [70] = {.entry = {.count = 1, .reusable = true}}, SHIFT(120),
  [72] = {.entry = {.count = 1, .reusable = true}}, SHIFT(143),
  [74] = {.entry = {.count = 1, .reusable = true}}, SHIFT(124),
  [76] = {.entry = {.count = 1, .reusable = true}}, SHIFT(126),
  [78] = {.entry = {.count = 1, .reusable = true}}, SHIFT(127),
  [80] = {.entry = {.count = 1, .reusable = true}}, SHIFT(128),
  [82] = {.entry = {.count = 1, .reusable = true}}, SHIFT(130),
  [84] = {.entry = {.count = 1, .reusable = true}}, SHIFT(6),
  [86] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_workflow_body, 1, 0, 0),
  [88] = {.entry = {.count = 1, .reusable = true}}, SHIFT(4),
  [90] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_field_value, 1, 0, 0),
  [92] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_field_value, 1, 0, 0),
  [94] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_multiline_block, 3, 0, 0),
  [96] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_multiline_block, 3, 0, 0),
  [98] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_operand, 1, 0, 0),
  [100] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_operand, 1, 0, 0),
  [102] = {.entry = {.count = 1, .reusable = true}}, SHIFT(125),
  [104] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_compare_expr, 1, 0, 0),
  [106] = {.entry = {.count = 1, .reusable = false}}, SHIFT(44),
  [108] = {.entry = {.count = 1, .reusable = true}}, SHIFT(101),
  [110] = {.entry = {.count = 1, .reusable = false}}, SHIFT(101),
  [112] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_expr, 1, 0, 0),
  [114] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_variable, 3, 0, 0),
  [116] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_variable, 3, 0, 0),
  [118] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_node_decl, 1, 0, 0),
  [120] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_agent_node, 5, 0, 0),
  [122] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_human_node, 5, 0, 0),
  [124] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_tool_node, 5, 0, 0),
  [126] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_subgraph_node, 5, 0, 0),
  [128] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_conditional_node, 5, 0, 0),
  [130] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_manager_loop_node, 5, 0, 0),
  [132] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_parallel_node, 5, 0, 0),
  [134] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_fan_in_node, 5, 0, 0),
  [136] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_stylesheet_section, 5, 0, 0),
  [138] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_node_attr_block, 3, 0, 0),
  [140] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_defaults_section, 4, 0, 0),
  [142] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edges_section, 4, 0, 0),
  [144] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_workflow_field, 3, 0, 0),
  [146] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_and_expr, 1, 0, 0),
  [148] = {.entry = {.count = 1, .reusable = false}}, SHIFT(65),
  [150] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_and_expr, 1, 0, 0),
  [152] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_and_expr_repeat1, 2, 0, 0),
  [154] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_and_expr_repeat1, 2, 0, 0), SHIFT_REPEAT(65),
  [157] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_and_expr_repeat1, 2, 0, 0),
  [159] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_and_expr, 2, 0, 0),
  [161] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_and_expr, 2, 0, 0),
  [163] = {.entry = {.count = 1, .reusable = true}}, SHIFT(9),
  [165] = {.entry = {.count = 1, .reusable = true}}, SHIFT(89),
  [167] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(142),
  [170] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(141),
  [173] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(141),
  [176] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0),
  [178] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_section_repeat1, 2, 0, 0), SHIFT_REPEAT(31),
  [181] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_or_expr, 1, 0, 0),
  [183] = {.entry = {.count = 1, .reusable = false}}, SHIFT(47),
  [185] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_or_expr, 1, 0, 0),
  [187] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_entry, 3, 0, 0),
  [189] = {.entry = {.count = 1, .reusable = false}}, SHIFT(30),
  [191] = {.entry = {.count = 1, .reusable = false}}, SHIFT(108),
  [193] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_entry, 3, 0, 0),
  [195] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_entry, 4, 0, 0),
  [197] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_entry, 4, 0, 0),
  [199] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0),
  [201] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0), SHIFT_REPEAT(30),
  [204] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0), SHIFT_REPEAT(108),
  [207] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_edge_entry_repeat1, 2, 0, 0),
  [209] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_or_expr, 2, 0, 0),
  [211] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_or_expr, 2, 0, 0),
  [213] = {.entry = {.count = 1, .reusable = true}}, SHIFT(142),
  [215] = {.entry = {.count = 1, .reusable = true}}, SHIFT(141),
  [217] = {.entry = {.count = 1, .reusable = false}}, SHIFT(141),
  [219] = {.entry = {.count = 1, .reusable = true}}, SHIFT(22),
  [221] = {.entry = {.count = 1, .reusable = true}}, SHIFT(31),
  [223] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_or_expr_repeat1, 2, 0, 0),
  [225] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_or_expr_repeat1, 2, 0, 0), SHIFT_REPEAT(47),
  [228] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_or_expr_repeat1, 2, 0, 0),
  [230] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_compare_expr, 3, 0, 0),
  [232] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_expr, 3, 0, 0),
  [234] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_compare_expr, 4, 0, 0),
  [236] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_expr, 4, 0, 0),
  [238] = {.entry = {.count = 1, .reusable = true}}, SHIFT(37),
  [240] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_attr, 2, 0, 0),
  [242] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_attr, 2, 0, 0),
  [244] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_condition, 1, 0, 0),
  [246] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_condition, 1, 0, 0),
  [248] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_edge_attr, 3, 0, 0),
  [250] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_edge_attr, 3, 0, 0),
  [252] = {.entry = {.count = 1, .reusable = false}}, SHIFT(7),
  [254] = {.entry = {.count = 1, .reusable = false}}, SHIFT(89),
  [256] = {.entry = {.count = 1, .reusable = true}}, SHIFT(90),
  [258] = {.entry = {.count = 1, .reusable = true}}, SHIFT(135),
  [260] = {.entry = {.count = 1, .reusable = true}}, SHIFT(24),
  [262] = {.entry = {.count = 1, .reusable = true}}, SHIFT(52),
  [264] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_defaults_section_repeat1, 2, 0, 0), SHIFT_REPEAT(135),
  [267] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_defaults_section_repeat1, 2, 0, 0),
  [269] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_defaults_section_repeat1, 2, 0, 0), SHIFT_REPEAT(52),
  [272] = {.entry = {.count = 1, .reusable = true}}, SHIFT(14),
  [274] = {.entry = {.count = 1, .reusable = true}}, SHIFT(59),
  [276] = {.entry = {.count = 1, .reusable = true}}, SHIFT(15),
  [278] = {.entry = {.count = 1, .reusable = true}}, SHIFT(16),
  [280] = {.entry = {.count = 1, .reusable = true}}, SHIFT(17),
  [282] = {.entry = {.count = 1, .reusable = true}}, SHIFT(18),
  [284] = {.entry = {.count = 1, .reusable = true}}, SHIFT(19),
  [286] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_agent_node_repeat1, 2, 0, 0), SHIFT_REPEAT(135),
  [289] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_agent_node_repeat1, 2, 0, 0),
  [291] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_agent_node_repeat1, 2, 0, 0), SHIFT_REPEAT(59),
  [294] = {.entry = {.count = 1, .reusable = true}}, SHIFT(23),
  [296] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_stylesheet_rule, 4, 0, 0),
  [298] = {.entry = {.count = 1, .reusable = false}}, REDUCE(sym_stylesheet_rule, 4, 0, 0),
  [300] = {.entry = {.count = 1, .reusable = true}}, SHIFT(58),
  [302] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_edges_section_repeat1, 2, 0, 0), SHIFT_REPEAT(111),
  [305] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_edges_section_repeat1, 2, 0, 0),
  [307] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_edges_section_repeat1, 2, 0, 0), SHIFT_REPEAT(67),
  [310] = {.entry = {.count = 1, .reusable = true}}, SHIFT(57),
  [312] = {.entry = {.count = 1, .reusable = true}}, SHIFT(61),
  [314] = {.entry = {.count = 1, .reusable = true}}, SHIFT(50),
  [316] = {.entry = {.count = 1, .reusable = true}}, SHIFT(64),
  [318] = {.entry = {.count = 1, .reusable = true}}, SHIFT(75),
  [320] = {.entry = {.count = 1, .reusable = true}}, SHIFT(56),
  [322] = {.entry = {.count = 1, .reusable = true}}, SHIFT(111),
  [324] = {.entry = {.count = 1, .reusable = true}}, SHIFT(25),
  [326] = {.entry = {.count = 1, .reusable = true}}, SHIFT(67),
  [328] = {.entry = {.count = 1, .reusable = true}}, SHIFT(54),
  [330] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 2, 0, 0), SHIFT_REPEAT(135),
  [333] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 2, 0, 0),
  [335] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 2, 0, 0), SHIFT_REPEAT(75),
  [338] = {.entry = {.count = 1, .reusable = true}}, SHIFT(53),
  [340] = {.entry = {.count = 1, .reusable = true}}, SHIFT(55),
  [342] = {.entry = {.count = 1, .reusable = true}}, SHIFT(81),
  [344] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_block_content, 1, 0, 0),
  [346] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_block_content_repeat1, 2, 0, 0), SHIFT_REPEAT(81),
  [349] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_block_content_repeat1, 2, 0, 0),
  [351] = {.entry = {.count = 1, .reusable = true}}, SHIFT(131),
  [353] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_identifier_list, 2, 0, 0),
  [355] = {.entry = {.count = 1, .reusable = false}}, REDUCE(aux_sym_string_repeat1, 2, 0, 0),
  [357] = {.entry = {.count = 2, .reusable = false}}, REDUCE(aux_sym_string_repeat1, 2, 0, 0), SHIFT_REPEAT(83),
  [360] = {.entry = {.count = 1, .reusable = true}}, SHIFT(71),
  [362] = {.entry = {.count = 1, .reusable = false}}, SHIFT(3),
  [364] = {.entry = {.count = 1, .reusable = false}}, SHIFT(83),
  [366] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_identifier_list_repeat1, 2, 0, 0), SHIFT_REPEAT(131),
  [369] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_identifier_list_repeat1, 2, 0, 0),
  [371] = {.entry = {.count = 1, .reusable = true}}, SHIFT(96),
  [373] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_identifier_list, 1, 0, 0),
  [375] = {.entry = {.count = 1, .reusable = false}}, SHIFT(2),
  [377] = {.entry = {.count = 1, .reusable = false}}, SHIFT(85),
  [379] = {.entry = {.count = 1, .reusable = true}}, SHIFT(80),
  [381] = {.entry = {.count = 1, .reusable = true}}, SHIFT(73),
  [383] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_stylesheet_rule_repeat1, 3, 0, 0),
  [385] = {.entry = {.count = 1, .reusable = true}}, SHIFT(69),
  [387] = {.entry = {.count = 1, .reusable = true}}, SHIFT(21),
  [389] = {.entry = {.count = 1, .reusable = true}}, SHIFT(20),
  [391] = {.entry = {.count = 1, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0),
  [393] = {.entry = {.count = 2, .reusable = true}}, REDUCE(aux_sym_source_file_repeat1, 2, 0, 0), SHIFT_REPEAT(96),
  [396] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_node_field, 3, 0, 0),
  [398] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_defaults_field, 3, 0, 0),
  [400] = {.entry = {.count = 1, .reusable = true}}, SHIFT(88),
  [402] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_compare_op, 1, 0, 0),
  [404] = {.entry = {.count = 1, .reusable = true}},  ACCEPT_INPUT(),
  [406] = {.entry = {.count = 1, .reusable = true}}, SHIFT(107),
  [408] = {.entry = {.count = 1, .reusable = true}}, SHIFT(109),
  [410] = {.entry = {.count = 1, .reusable = true}}, SHIFT(8),
  [412] = {.entry = {.count = 1, .reusable = true}}, SHIFT(112),
  [414] = {.entry = {.count = 1, .reusable = true}}, SHIFT(76),
  [416] = {.entry = {.count = 1, .reusable = true}}, SHIFT(62),
  [418] = {.entry = {.count = 1, .reusable = true}}, SHIFT(74),
  [420] = {.entry = {.count = 1, .reusable = true}}, SHIFT(63),
  [422] = {.entry = {.count = 1, .reusable = true}}, SHIFT(132),
  [424] = {.entry = {.count = 1, .reusable = true}}, SHIFT(77),
  [426] = {.entry = {.count = 1, .reusable = true}}, SHIFT(72),
  [428] = {.entry = {.count = 1, .reusable = true}}, SHIFT(68),
  [430] = {.entry = {.count = 1, .reusable = true}}, SHIFT(66),
  [432] = {.entry = {.count = 1, .reusable = true}}, SHIFT(99),
  [434] = {.entry = {.count = 1, .reusable = true}}, SHIFT(100),
  [436] = {.entry = {.count = 1, .reusable = true}}, SHIFT(60),
  [438] = {.entry = {.count = 1, .reusable = true}}, SHIFT(42),
  [440] = {.entry = {.count = 1, .reusable = true}}, SHIFT(113),
  [442] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_workflow_decl, 5, 0, 0),
  [444] = {.entry = {.count = 1, .reusable = true}}, SHIFT(84),
  [446] = {.entry = {.count = 1, .reusable = true}}, SHIFT(129),
  [448] = {.entry = {.count = 1, .reusable = true}}, SHIFT(115),
  [450] = {.entry = {.count = 1, .reusable = true}}, SHIFT(11),
  [452] = {.entry = {.count = 1, .reusable = true}}, SHIFT(116),
  [454] = {.entry = {.count = 1, .reusable = true}}, SHIFT(117),
  [456] = {.entry = {.count = 1, .reusable = true}}, SHIFT(91),
  [458] = {.entry = {.count = 1, .reusable = true}}, SHIFT(5),
  [460] = {.entry = {.count = 1, .reusable = true}}, SHIFT(119),
  [462] = {.entry = {.count = 1, .reusable = true}}, SHIFT(95),
  [464] = {.entry = {.count = 1, .reusable = true}}, SHIFT(33),
  [466] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 2, 0, 0),
  [468] = {.entry = {.count = 1, .reusable = true}}, SHIFT(70),
  [470] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_field_name, 1, 0, 0),
  [472] = {.entry = {.count = 1, .reusable = true}}, SHIFT(51),
  [474] = {.entry = {.count = 1, .reusable = true}}, SHIFT(121),
  [476] = {.entry = {.count = 1, .reusable = true}}, SHIFT(49),
  [478] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_source_file, 1, 0, 0),
  [480] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_selector, 2, 0, 0),
  [482] = {.entry = {.count = 1, .reusable = true}}, SHIFT(140),
  [484] = {.entry = {.count = 1, .reusable = true}}, REDUCE(sym_selector, 1, 0, 0),
  [486] = {.entry = {.count = 1, .reusable = true}}, SHIFT(114),
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
