package notes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// builtinTemplates are the fallbacks used when the vault has no override.
var builtinTemplates = map[string]string{
	"daily": `---
date: {{date}}
---
# {{date_long}}

## Tasks

## Notes

## Standup
`,

	"task": `---
title: "{{title}}"
linear_id: {{linear_id}}
linear_url: {{linear_url}}
status: {{status}}
project: "{{project}}"
tags: [task]
---
# {{linear_id}}: {{title}}

## Notes

## Log
`,

	"project": `---
title: "{{name}}"
linear_project_id:
tags: [project]
---
# {{name}}

## Overview

## Approach

## Rejected Alternatives

## Tasks

## Out of scope

## Open
`,

	"investigation": `---
title: "{{title}}"
date: {{date}}
scratch: "{{scratch}}"
scopes: [{{scopes}}]
project: "{{project}}"
task: "{{task}}"
tags: [investigation]
---
# {{title}}

## Question

## Log
`,
}

// TemplatePath returns where a vault override for the named template would live.
func TemplatePath(vaultPath, name string) string {
	return filepath.Join(vaultPath, "_templates", name+".md")
}

// render fills the named template with vars, preferring a vault override in
// _templates/ over the built-in. The second return value names the source that
// was used, for reporting. Unknown {{placeholders}} are left in place so a typo
// in a hand-written template is visible rather than silently blank.
func render(vaultPath, name string, vars map[string]string) (string, string) {
	tmpl, source := builtinTemplates[name], "built-in"

	if data, err := os.ReadFile(TemplatePath(vaultPath, name)); err == nil {
		tmpl = string(data)
		source = filepath.Join("_templates", name+".md")
	}

	pairs := make([]string, 0, len(vars)*2)
	for k, v := range vars {
		pairs = append(pairs, "{{"+k+"}}", v)
	}
	return strings.NewReplacer(pairs...).Replace(tmpl), source
}

// reportTemplate tells the user which template a freshly created note came from.
func reportTemplate(path, source string) {
	fmt.Fprintf(os.Stderr, "created %s from %s template\n", path, source)
}

// RenderDailyTemplate renders the daily note for date.
func RenderDailyTemplate(vaultPath string, date time.Time) (string, string) {
	return render(vaultPath, "daily", map[string]string{
		"date":      date.Format("2006-01-02"),
		"date_long": date.Format("Monday, January 2, 2006"),
	})
}
