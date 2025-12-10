# sqd | A SQL-like document editor

Traditional Unix tools (grep, sed, awk) are powerful but have inconsistent syntax and steep learning curves. `sqd` (pronounced like squad) provides a familiar SQL interface for common text operations.

## Useful Commands

Count all the open todos in your markdown files

```bash
sqd 'SELECT count(*) FROM *.md WHERE content LIKE "%- [ ]%"'
```

Count all the LaTeX formulas in your notes

```bash
sqd 'SELECT count(*) FROM * WHERE content LIKE "%$$"'
```

Refactor your markdown title hierarchy

```bash
sqd 'UPDATE *.md SET content="### " WHERE content LIKE "## %"'
```

## The power of sqd

You have a file with multiple similar titles, but you only want to change specific ones. With sed or awk, you'd need complex regex or multiple commands. With sqd, you can target exact lines and batch multiple replacements in a single command.

```markdown
## Title 1 to be updated

## Title 1 not to be updated

## Title 1 TO be updated

## Title 2 to be updated

## Title 2 not to be updated

## Title 2 TO be updated
```

With only one sqd command

```bash
sqd 'UPDATE example.md 
SET content="## Title 1 UPDATED" WHERE content = "## Title 1 to be updated",
SET content="## Title 2 UPDATED" WHERE content = "## Title 2 TO be updated"'
```

You will obtain the following result

```markdown
## Title 1 UPDATED

## Title 1 not to be updated

## Title 1 TO be updated

## Title 2 to be updated

## Title 2 not to be updated

## Title 2 UPDATED
```