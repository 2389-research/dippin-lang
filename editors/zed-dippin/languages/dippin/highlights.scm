; Keywords
[
  "workflow"
  "defaults"
  "edges"
  "stylesheet"
] @keyword

; Node kinds
[
  "agent"
  "human"
  "tool"
  "subgraph"
  "conditional"
  "parallel"
  "fan_in"
  "manager_loop"
] @type

; Edge / condition keywords
[
  "when"
  "and"
  "or"
  "not"
] @keyword.operator

; Format version declaration
(version_decl
  "dip" @keyword)

; Comparison operators
(compare_op) @operator

; Operators
[
  "->"
  "<-"
  ":"
] @operator

; Selectors
(selector
  "#" @punctuation.special)
(selector
  "." @punctuation.special)
(selector
  "*" @operator)

; Field names
(defaults_field
  (field_name) @property)
(node_field
  (field_name) @property)
(edge_attr
  "label" @property)
(edge_attr
  "choice" @property)
(edge_attr
  "weight" @property)
(edge_attr
  "restart" @property)
(edge_attr
  "override" @property)
(edge_attr
  "on" @keyword.operator)
(edge_attr
  "loop" @keyword)

; Identifiers in declarations
(workflow_decl
  (identifier) @type.definition)
(agent_node
  (identifier) @function)
(human_node
  (identifier) @function)
(tool_node
  (identifier) @function)
(subgraph_node
  (identifier) @function)
(conditional_node
  (identifier) @function)
(parallel_node
  (identifier) @function)
(fan_in_node
  (identifier) @function)
(manager_loop_node
  (identifier) @function)

; Edge source/target
(edge_entry
  (identifier) @variable)

; Variables in conditions
(variable
  (identifier) @variable.parameter)

; Strings
(string) @string

; Comments
(comment) @comment

; Boolean and numeric values
((identifier) @boolean
  (#match? @boolean "^(true|false)$"))

; Workflow header fields
(workflow_field
  "goal" @keyword)
(workflow_field
  "requires" @keyword)
(workflow_field
  "start" @keyword)
(workflow_field
  "exit" @keyword)
