# sqd | A SQL-like document editor

Traditional Unix tools (grep, sed, awk) are powerful but have inconsistent syntax and steep learning curves. `sqd` (pronounced like squad) provides a familiar SQL interface for common text operations.

## Getting Started

This project requires **Go >= 1.27**. Make sure you have a compatible version installed. If needed, download the latest version from [https://go.dev/dl/](https://go.dev/dl/)

1. **Installation**: Install sqd on your system

    ```bash
    go install github.com/overthinkinglabs/sqd@latest
    ```

2. **Start using sqd**: See all the open todos in your markdown files

    ```bash
    sqd 'SELECT * FROM *.md WHERE content LIKE "%- [ ]%"'
    ```

## Initialization

Run `sqd init` to create the global configuration file at `~/.config/sqd/config.json` and a local `.sqdignore` to exclude files and folders. Existing files are never overwritten. Here an example of a `config.json` file with the relative explaination:

```json
{
  "output": {
    "color": "blue",
    "show_stats": true
  },
  "from_aliases": {
    "obsidian": "/home/user/Documents/obsidian"
  }
}
```

- `output.color`: highlight color for matches. Choose `blue`, `green`, `red`, `yellow`, `cyan`, `purple`, or `none`.
- `output.show_stats`: print processing stats after text output. Stats are omitted automatically when using `--json` or `--csv`.
- `from_aliases`: map short names to file patterns or absolute paths. Use them in `FROM`:

    ```bash
    sqd 'SELECT content FROM obsidian WHERE content LIKE "- [ ]%"'
    ```

**N.B.** Aliases also work as prefixes. If `"notes": "/home/user/Documents/notes"` is defined, `FROM notes/*.md` resolves to `/home/user/Documents/notes/*.md`.

## Useful Commands

Count all the LaTeX formulas in your notes

```bash
sqd 'SELECT COUNT(*) FROM *.md WHERE content LIKE "$$%"'
```

Refactor your markdown title hierarchy

```bash
sqd 'UPDATE *.md SET content="### " WHERE content LIKE "## %"'
```

Remove all DEBUG logs

```bash
sqd 'DELETE FROM *.log WHERE content LIKE "%DEBUG%"'
```

## Columns

You can reference the following columns in the `SELECT` clause:

- `name`: the file name.
- `line`: the line number.
- `content`: the content of each line.
- `*`: file name, line number, and content.

Examples:

```bash
sqd 'SELECT COUNT(name) FROM *.md WHERE content LIKE "### %"' 
# Counts the number of files that contain at least one matching line.

sqd 'SELECT COUNT(content) FROM *.md WHERE content LIKE "### %"' 
# Counts the total number of matching lines across all files.
```

## Ordering

You can control the order of results using the `ORDER BY` clause. This is useful when you want to sort lines by file name, by content, or by a combination of both. You can specify the sorting direction:

- `ASC`: ascending order (default). Values are sorted from A to Z.

- `DESC`: descending order. Values are sorted from Z to A.

```bash
# Order by content (ascending by default)
sqd 'SELECT name FROM *.md WHERE content LIKE "- [ ]%" ORDER BY content'

# Order by file name descending
sqd 'SELECT name FROM *.md WHERE content LIKE "- [ ]%" ORDER BY name DESC'
```

You can also apply multiple ordering rules. The first column is used as the primary key, and the following ones are used to break ties:

```bash
sqd 'SELECT * FROM *.md WHERE content LIKE "- [ ]%" ORDER BY content ASC, name DESC'
```

In this example, lines are sorted alphabetically by `content`, and when two lines have the same content, they are ordered by file name in descending order.

## Limit

Use `LIMIT n` to return only the first `n` results.

```bash
sqd 'SELECT content FROM *.md WHERE content LIKE "- [ ]%" ORDER BY content LIMIT 10'
```

## Flags

- `-f`, `--file`: Runs all queries from a file. Useful for refactoring and repetitive tasks.
- `-t`, `--transaction`: Apply changes atomically. If any operation fails, all changes are rolled back.
- `--json`: Print `SELECT` and `COUNT` results as JSON.
- `--csv`: Print `SELECT` and `COUNT` results as CSV.

## Dry Mode

The `dry` command shows what changes would be made without modifying any files:

```bash
sqd dry 'UPDATE *.md SET content="new" WHERE content = "old"'
# Dry run: pass | fail
```

Use the `-c` or `--complete` flag to see detailed information for each modified line. Dry mode supports all the same flags as regular commands.

## The power of sqd

Let's suppose we have a file with multiple similar titles, but we only want to change specific ones. With sed or awk, we need complex regex or multiple commands. With sqd, we can target exact lines and batch multiple replacements in a single command.

```markdown
## Title 1 to be updated

## Title 1 not to be updated

## Title 1 TO be updated

## Title 2 to be updated

## Title 2 not to be updated

## Title 2 TO be updated
```

With a single sqd command

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

## Before pushing

1. **See if you have any rebase to do** (you must have the updated commits history before pushing to avoid conflicts between main and your branch):

    ```sh
    git fetch
    git pull origin main --rebase
    ```

2. **Build the project locally to avoid compiling errors**:

    ```sh
    goreleaser release --snapshot --clean
    ```

3. **Run the tests to avoid your code breaks anything**:

    ```sh
    go test -v -race ./...
    ```

4. **Lint your code** (if the following command return errors or warnings you must resolve them before pushing):

    ```sh
    golangci-lint run ./...
    ```

5. **Format your code**:

    ```sh
    gofumpt -w ./
    ```

## Contributing

If you want to contribute to **sqd**, follow these steps:

1. Create a new branch for your changes (`git checkout -b your-branch-name`).
2. Make your changes and commit them (`git commit -m 'Change something'`).
3. Push your branch (`git push origin your-branch-name`).
4. Open a pull request.
