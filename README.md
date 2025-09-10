# gh-gl-create-refs

A GitHub CLI extension to fetch GitLab merge request references and export them as CSV files.

## Installation

This is a GitHub CLI extension. To install it:

```bash
gh extension install amenocal/gh-gl-create-refs
```

## Usage

### Prerequisites

You need a GitLab access token to use this extension. You can create one in GitLab under **User Settings > Access Tokens**.

Set your token either as an environment variable:
```bash
export GITLAB_TOKEN=your_token_here
```

Or pass it via the `--token` flag.

### Fetch Merge Request References

Use the `fetch-ref` command to fetch all merge request references from a GitLab repository:

```bash
# Using environment variable for token
gh gl-create-refs fetch-ref group/project

# Using token flag
gh gl-create-refs fetch-ref --token your_token group/project

# With custom GitLab instance
gh gl-create-refs fetch-ref --base-url https://gitlab.example.com group/project

# With custom output file
gh gl-create-refs fetch-ref --output my_output.csv group/project
```

### Supported Repository Formats

The extension supports various GitLab repository path formats:

- **Group/Project**: `group/project`
- **Nested Subgroups**: `group/subgroup/project`
- **Multiple Nested Subgroups**: `group/sub1/sub2/sub3/project`
- **Full URLs**: `https://gitlab.com/group/project`
- **Custom GitLab Instance URLs**: `https://gitlab.example.com/group/project`

### Output Format

The extension generates a CSV file with two columns:
1. **Merge Request Number** (IID) - The merge request number as shown in GitLab UI
2. **Head SHA** - The head commit SHA from the merge request's diff_refs

Example output (`group-project.csv`):
```csv
1,e8a44ccde03fc255605d38aec8db81db176398eb
16,f70267410222c85b3ea62df436acef0de0e9bda3
17,d47c8f40a570e567e6672b54528a4cc34c29eb60
```

### Command Options

- `--token`, `-t`: GitLab access token (can also use `GITLAB_TOKEN` environment variable)
- `--base-url`, `-b`: GitLab base URL (default: https://gitlab.com)
- `--output`, `-o`: Custom output CSV file path (default: auto-generated from repository name)

## Examples

```bash
# Fetch from gitlab.com
gh gl-create-refs fetch-ref mygroup/myproject

# Fetch from custom GitLab instance
gh gl-create-refs fetch-ref --base-url https://gitlab.company.com company/team/project

# Fetch with custom output file
gh gl-create-refs fetch-ref --output results.csv mygroup/subgroup/project

# Fetch from full URL
gh gl-create-refs fetch-ref https://gitlab.com/mygroup/myproject
```

## Development

### Building

```bash
go build -o gh-gl-create-refs
```

### Testing

```bash
go test ./...
```

## Requirements

- Go 1.19 or later
- GitLab access token with read access to the target repository
- GitHub CLI (gh) for installation as an extension

## License

[View License](LICENSE)