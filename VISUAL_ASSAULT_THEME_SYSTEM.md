# Visual Assault Theme System

**A unified, cross-platform design token system for managing themes across CSS, Tkinter, Flet, and Flask applications.**

## Overview

Visual Assault is a centralized theme management system designed to keep design tokens (colors, typography, spacing) synchronized across multiple frameworks and platforms. Instead of maintaining themes separately in each project, Visual Assault provides a single source of truth that generates theme outputs for any framework.

## Current State

Themes currently exist in three separate repositories:

- **CSS Themes**: [`card-judge`](https://github.com/gerp93/card-judge) repo — `src/static/css/colors.css` — **19 complete themes**
- **Tkinter Themes**: [`KVG_Themes`](https://github.com/gerp93/KVG_Themes) repo — subset of themes, some out of sync
- **Flet Themes**: [`KVG_Themes_Flet`](https://github.com/gerp93/KVG_Themes_Flet) repo — subset of themes, some out of sync

**Problem**: Updates to one framework require manual edits in all others; new themes require triplication of effort.

## Vision

A single `visual-assault-themes` repository that:
1. Defines all themes in a language-agnostic format (JSON/YAML)
2. Generates theme outputs for CSS, Tkinter, Flet, and Flask
3. Validates theme completeness and contrast ratios
4. Auto-syncs changes to dependent projects via CI/webhooks

## Proposed Architecture

```
visual-assault-themes/
├── themes/
│   ├── base/
│   │   ├── dark.json
│   │   ├── light.json
│   │   └── ... (8 "original" themes)
│   ├── crazy/
│   │   ├── hawkeye.json
│   │   ├── green-acres.json
│   │   ├── red-barn.json
│   │   └── ... (11 "visually assaulting" themes)
│   └── schema.json (JSON Schema validation)
├── generators/
│   ├── generate_css.py
│   ├── generate_tkinter.py
│   ├── generate_flet.py
│   ├── generate_flask.py
│   └── generate_all.py (master orchestrator)
├── src/
│   ├── theme.py (Python dataclass for theme parsing)
│   └── validators.py (contrast checks, completeness)
├── output/
│   ├── css/
│   │   └── colors.css
│   ├── tkinter/
│   │   └── themes.json (or Python dict)
│   ├── flet/
│   │   └── themes.json
│   └── flask/
│       └── config.py
├── .github/workflows/
│   ├── validate.yml (run on PR)
│   ├── generate.yml (run on merge to main)
│   └── sync.yml (push outputs to dependent repos)
├── tests/
│   ├── test_theme_completeness.py
│   ├── test_contrast.py
│   └── test_generators.py
├── README.md
├── THEME_SPEC.md (theme JSON format documentation)
└── CONTRIBUTOR_GUIDE.md
```

## Theme JSON Format

**Example: `themes/crazy/hawkeye.json`**

```json
{
  "id": "hawkeye",
  "name": "Hawkeye",
  "displayName": "🦅 Hawkeye",
  "description": "Iowa Hawkeyes - Black and Old Gold",
  "category": "crazy",
  "colors": {
    "topBarBg": "rgb(0, 0, 0)",
    "topBarHover": "rgb(40, 40, 0)",
    "text": "rgb(255, 205, 0)",
    "bg": "rgb(0, 0, 0)",
    "bgHover": "rgb(30, 25, 0)",
    "primaryAction": "rgb(40, 40, 40)",
    "primaryActionHover": "rgb(60, 60, 60)",
    "primaryActionText": null,
    "accentGreen": "rgb(255, 205, 0)",
    "accentRed": "rgb(255, 150, 0)",
    "accentBlue": "rgb(255, 230, 100)"
  },
  "typography": {
    "fontFamily": "'Arial Black', 'Helvetica Bold', sans-serif",
    "fontFamilyMono": "'Consolas', 'Courier New', monospace",
    "fontSizeBase": "16px",
    "fontWeightNormal": 400,
    "fontWeightBold": 700
  },
  "metadata": {
    "inspiration": "Iowa Hawkeyes athletic colors",
    "author": "Grantford Barnes",
    "createdAt": "2025-12-07",
    "wcagAACompliant": false
  }
}
```

## Workflow

### 1. Add/Modify a Theme
Edit the JSON in `themes/` → commit → CI validates → generates outputs

### 2. Automatic Generation (CI/CD)
```yaml
# .github/workflows/generate.yml
on: [push to main]
jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: python generators/generate_all.py
      - commit & push changes to /output
      - trigger sync workflow
```

### 3. Sync to Dependent Repos
```yaml
# .github/workflows/sync.yml
on: [workflow_run from generate.yml]
jobs:
  sync-css:
    - copy output/css/colors.css → card-judge/src/static/css/
    - PR or direct push to dependent repo
  sync-tkinter:
    - copy output/tkinter/ → KVG_Themes/
    - PR or direct push
  sync-flet:
    - copy output/flet/ → KVG_Themes_Flet/
    - PR or direct push
```

## Generators (High-Level)

### CSS Generator
```python
# generators/generate_css.py
def generate_css(themes_dir):
    # Load all .json files from themes/
    # For each theme:
    #   body.{theme-id}-theme {
    #     --color-top-bar-bg: ...
    #     --color-text: ...
    #     etc.
    #   }
    # Output: src/static/css/colors.css
```

### Tkinter Generator
```python
# generators/generate_tkinter.py
def generate_tkinter(themes_dir):
    # Load all .json files
    # Convert to Python dict or Tkinter color schemes
    # Output: src/themes.py or themes.json
```

### Flet Generator
```python
# generators/generate_flet.py
def generate_flet(themes_dir):
    # Convert to Flet color/theme format
    # Output: flet theme dicts
```

### Flask Generator
```python
# generators/generate_flask.py
def generate_flask(themes_dir):
    # Create Flask config templates
    # Output: app/config/themes.py or YAML
```

## Validators

- **Completeness**: All required color keys present?
- **Contrast**: Text/background contrast WCAG AA/AAA?
- **Format**: Valid RGB/Hex/Named colors?
- **Schema**: JSON conforms to `schema.json`?

## Documentation

### THEME_SPEC.md
- Color key definitions and purposes
- Typography variable meanings
- RGB/Hex conversion rules
- Platform-specific notes (e.g., Tkinter doesn't support CSS variables)

### CONTRIBUTOR_GUIDE.md
- How to add a new theme
- How to audit themes for contrast
- How to test generated outputs locally
- How to propose theme changes

## Phase 1 (MVP)

1. Create repo: `visual-assault-themes`
2. Extract all 19 themes from `card-judge/src/static/css/colors.css` → JSON
3. Write CSS generator
4. Validate CSS output matches original
5. Document theme JSON format

## Phase 2

1. Write Tkinter generator
2. Write Flet generator
3. Consolidate/audit themes across all three repos
4. Merge themes from KVG_Themes and KVG_Themes_Flet (resolve conflicts)

## Phase 3

1. Add CI/CD validation & generation
2. Add contrast checking
3. Set up sync workflows to dependent repos
4. Write comprehensive documentation

## Future Enhancements

- Design token library (spacing, shadows, etc.)
- Theme editor UI (web-based preview)
- Figma integration (sync from Figma design tokens)
- Auto-generated theme preview page
- NPM package for web projects
- PyPI package for Python projects

## Key Benefits

✅ **Single source of truth** — Edit once, deploy everywhere  
✅ **Consistency** — No more out-of-sync themes  
✅ **Scalability** — Add new frameworks without duplicating effort  
✅ **Validation** — CI ensures quality & completeness  
✅ **Collaboration** — Clear theme spec for contributors  
✅ **Branding** — Visual Assault becomes a design system

## Reference Repositories

When implementing this system, review the existing theme definitions in these repos:

- **CSS**: https://github.com/gerp93/card-judge/blob/main/src/static/css/colors.css
- **Tkinter**: https://github.com/gerp93/KVG_Themes
- **Flet**: https://github.com/gerp93/KVG_Themes_Flet

---

## Next Steps for Implementation

1. Create `visual-assault-themes` repo
2. Convert existing CSS themes to JSON (Phase 1)
3. Write & test CSS generator
4. Document theme format
5. Share with team for feedback
6. Plan Phase 2 timeline
