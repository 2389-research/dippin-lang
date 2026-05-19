# Dippin sitemap

A flat index of every page on this site, grouped for agent consumption. The canonical machine-readable sitemap is [sitemap.xml]({{ "sitemap.xml" | absURL }}); this `.md` mirror exists for LLM clients that prefer Markdown.

## Documentation
{{ range where .Site.RegularPages "Section" "" }}{{ if ne .File.LogicalName "_index.md" }}
- [{{ .Title }}]({{ .Permalink }}) — {{ .Description }}{{ end }}{{ end }}

## Blog
{{ range where .Site.RegularPages "Section" "blog" }}
- [{{ .Title }}]({{ .Permalink }}) — {{ .Description }}{{ end }}

## Agent skills

- [Claude Code skill]({{ "skill.md" | absURL }}) — instructions for Claude Code agents working with `.dip` files.
- [AGENTS.md]({{ "AGENTS.md" | absURL }}) — install / configure / use the toolchain.
- [llms.txt]({{ "llms.txt" | absURL }}) — short LLM-oriented index.
- [llms-full.txt]({{ "llms-full.txt" | absURL }}) — full spec for deeper context.
