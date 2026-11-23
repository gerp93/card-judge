# TODO

This fork follows [gerp93/KVG_Standards](https://github.com/gerp93/KVG_Standards)
(category: Go web app, deployed via DigitalOcean push). This file records the
gaps against that standard that need a human decision. `main` tracks upstream
[GrantFBarnes/card-judge](https://github.com/GrantFBarnes/card-judge); the
gameshell-framework port lives on `f-framework-breakout` (PR #12, not to be
merged as-is), so most items below are about how far `main` should diverge
from upstream.

## Needs a human

- [ ] **Tests**: there are no automated tests (no `*_test.go`), so there is no
  `test_command` to run. Not a violation while this stays true; revisit if
  tests are added.
- [ ] **CI**: `main` has no `ci.yml` calling KVG_Standards' `ci-go.yml`
  (build + vet). It existed on `f-framework-breakout`, not here.
- [ ] **Vestigial release pipeline**: `.github/workflows/release.yml` (GitHub
  Release binary on tag) and `version_bump.sh` come from upstream. A Go web
  app deployed on push doesn't need them, but removing them diverges from
  upstream; kept for now.
- [ ] **Theming**: `main` uses upstream's own palette in
  `src/static/css/colors.css`, not VisualAssault through gameshell-framework.
  The framework-based theming exists only on `f-framework-breakout`.
- [ ] **Docs**: `README.md` (upstream's) doesn't mention KVG_Standards; this
  file is the pointer for now.
- [ ] **Logo**: only a favicon, no `assets/logo.png` (low priority for a web app).
- [ ] **Secrets**: runtime config comes from `CARD_JUDGE_*` environment
  variables (no keys in the repo). `.gitignore` now carries the standard
  secrets lines from KVG_Standards' `secrets.md`.
