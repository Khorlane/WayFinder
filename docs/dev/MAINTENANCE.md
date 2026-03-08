# Maintenance

## Regenerate Go Source Index

Purpose: refresh `docs/dev/go_source_index.md` from current `*.go` files.

From repo root (`C:\Projects\WayFinder`), run:

```powershell
@'
$ErrorActionPreference = 'Stop'
$files = rg --files -g "*.go" | Sort-Object

$sb = New-Object System.Text.StringBuilder
[void]$sb.AppendLine('# Go Source Index')
[void]$sb.AppendLine()
[void]$sb.AppendLine('Generated from current `.go` files. Use this as a quick technical map for chat/session continuity.')
[void]$sb.AppendLine()
[void]$sb.AppendLine('## Go File Tree')
[void]$sb.AppendLine()
foreach ($f in $files) {
    [void]$sb.AppendLine('- `' + $f + '`')
}

foreach ($f in $files) {
    [void]$sb.AppendLine()
    [void]$sb.AppendLine('## `' + $f + '`')
    $content = Get-Content $f

    $types = @()
    $funcs = @()

    for ($i = 0; $i -lt $content.Count; $i++) {
        $line = $content[$i]
        if ($line -match '^\s*type\s+([A-Za-z_][A-Za-z0-9_]*)\s') {
            $types += [PSCustomObject]@{ Name = $matches[1]; Line = $i + 1 }
            continue
        }
        if ($line -match '^\s*func\s*\(([^\)]+)\)\s*([A-Za-z_][A-Za-z0-9_]*)\s*\(') {
            $recv = $matches[1].Trim()
            $name = $matches[2]
            $funcs += [PSCustomObject]@{ Name = "$name (method on $recv)"; Line = $i + 1 }
            continue
        }
        if ($line -match '^\s*func\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(') {
            $funcs += [PSCustomObject]@{ Name = $matches[1]; Line = $i + 1 }
            continue
        }
    }

    [void]$sb.AppendLine()
    [void]$sb.AppendLine('Types:')
    if ($types.Count -eq 0) {
        [void]$sb.AppendLine('- (none)')
    } else {
        foreach ($t in $types) {
            [void]$sb.AppendLine('- ' + $t.Name + ' (line ' + $t.Line + ')')
        }
    }

    [void]$sb.AppendLine()
    [void]$sb.AppendLine('Functions:')
    if ($funcs.Count -eq 0) {
        [void]$sb.AppendLine('- (none)')
    } else {
        foreach ($fn in $funcs) {
            [void]$sb.AppendLine('- ' + $fn.Name + ' (line ' + $fn.Line + ')')
        }
    }
}

Set-Content -Path 'docs/dev/go_source_index.md' -Value $sb.ToString() -NoNewline
'@ | powershell -NoProfile -Command -
```

Notes:
- `rg` is required (`ripgrep`).
- This index is intentionally high-level and only tracks Go source files.
