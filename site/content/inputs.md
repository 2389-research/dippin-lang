---
title: "Inputs"
description: "Typed workflow inputs: the inputs block, input types, attributes, the closed ${inputs.*} namespace, validation, and the dippin inputs introspection contract."
section_label: "Language"
subtitle: "Typed workflow inputs and their host-facing contract."
---

## The Inputs Block

The optional `inputs` block declares the workflow's callee-side signature — the named values a caller (a human at the entry point, or a parent workflow via a `subgraph` node's call-site binding — `inputs:` in `dip 2`, `params:` in `dip 1`) must or may supply at run start:

```
  inputs
    idea: text
      required: true
      prompt: "What do you want built?"
      description: "One or two sentences describing the change."
      max_length: 4000
      multiline: true
    target_branch: text
      default: main
      pattern: "^[A-Za-z0-9._/-]+$"
    risk: enum
      default: medium
      options: low, medium, high
```

Each entry is `name: type` with an optional indented block of attributes. Declaration order is significant — a host renders inputs as an ordered form — so the formatter never reorders entries.

## Input Types

The six v1 types are:

| Type | Description |
|------|-------------|
| `text` | Free text value |
| `number` | Numeric value |
| `bool` | Boolean value |
| `enum` | One of a fixed set of `options` |
| `file` | A file |
| `secret` | A secret value |

## Attributes

Each entry may carry an indented block of attributes:

| Attribute | Description |
|-----------|-------------|
| `required` | Whether the caller must supply this input |
| `prompt` | Text a host shows when collecting the value |
| `description` | Human-readable description of the input |
| `default` | Value used when the caller supplies none |
| `options` | The allowed values for an `enum` |
| `pattern` | A regex the value must match |
| `min` | Minimum value |
| `max` | Maximum value |
| `max_length` | Maximum length |
| `multiline` | Whether the value spans multiple lines |

## Referencing Inputs

Reference a declared input as `${inputs.name}` in prompts. The `inputs` namespace is **closed**: it contains exactly the names declared in the `inputs` block, unlike the open `ctx` namespace.

## Validation

Because the namespace is closed, references and declarations are checked:

- An unrecognized type is **DIP155**.
- A reference to an undeclared input is **DIP156**.
- A reference inside a tool node's `command:` is **DIP157** (inputs never interpolate into a shell).

## Introspection

The declared `inputs` schema is the typed, introspectable contract a host uses to collect values before a run. Print it with:

```
dippin inputs [--format text|json] <file>
```

`--format json` emits a stable array in declaration order — each entry with `name`, `type`, `required`, and any declared attributes. Defaults are typed (a `number` default is a JSON number), and a workflow with no inputs emits `[]`, never `null`.

The same schema is surfaced by `dippin inspect --format json` for a `.dipx` bundle: the verified payload includes the entry workflow's declared `inputs` schema, so a host can enumerate what to collect without unpacking.
