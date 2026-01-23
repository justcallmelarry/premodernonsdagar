---
applyTo: '**'
---

# premodernonsdagar

This is a website for tracking and displaying leaderboards for a gaming community, playing the game Magic: The Gathering in the format Premodern. It aggregates match data, computes player statistics, and renders HTML pages to present the information.

## Testing and Debugging

- You can use `go run cmd/main/main.go --build` to only build templates without starting the server, this is useful for testing the data aggregation and template rendering processes.

## Project Structure

- `files/` contains generated/aggregated data (players, events, leaderboards)
- `input/` contains source data (raw events and decklists)
- `pages/html/` contains rendered templates (dev mode only)
- All data files use JSON format with consistent date format (YYYY-MM-DD)

## Development Workflow

- Use `DEVENV=1 go run cmd/main/main.go` for development (re-renders templates on startup)
- Use `go run cmd/main/main.go --build` to test data aggregation without starting server
- Run `npm run tailwind` in background while working on templates
- Use `npm run format` to format templates before committing

## General Guidelines

- Favor clarity and simplicity over cleverness
- Follow the principle of least surprise
- Write self-documenting code with clear, descriptive names
- Avoid excess comments, the code should be clear enough on its own
- Avoid using emojis in code, comments, or documentation

---
applyTo: '**/*.go'
---

# Go Code Guidelines

- Keep the happy path left-aligned (minimize indentation)
- Prefer early return over if-else chains, is `if condition { return }` pattern to avoid else blocks
- Leverage the Go standard library instead of reinventing the wheel (e.g. use `strings.Builder` for string concatenation, `filepath.Join` for path construction)
- Prefer standard library solutions over custom solutions when functionality exists
- Use structs for JSON serialization/deserialization instead of unspecified structs or slices
- All templates must extend `base.tmpl`
- Color/styling should use the centralized `ColorScheme()` function, not hardcoded Tailwind classes

---
applyTo: '**/*.tmpl'
---

## Template Guidelines

- Always use tailwind classes for styling, avoid custom CSS unless absolutely necessary
- Do loops and conditionals using Go template syntax
- Prefer looping over similar items instead of duplicating code
