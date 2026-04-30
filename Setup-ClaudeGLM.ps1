# ================================================================
# Setup-ClaudeGLM.ps1
#
# Sets up multi-profile Claude Code:
#   claude       -> Claude.ai login (Opus) - all phases except implement
#   g-claude     -> GLM via Z.ai API key   - speckit implement only
#   q-claude     -> Qwen 3.5 Plus via DashScope (ANTHROPIC_MODEL=qwen3.5-plus)
#   k-claude     -> Kimi K2.6 via Moonshot API (api.moonshot.ai/v1)
#   qq-claude    -> Qwen 3.6 Max Preview via DashScope Intl compatible-mode
#   c-claude     -> OpenAI Codex CLI (@openai/codex) with model gpt-5.4
#   init-claude  -> NEW project: creates CLAUDE.md + read/ directory
#   adopt-claude -> EXISTING project: adds read/ without touching CLAUDE.md
#
# Usage:
#   .\Setup-ClaudeGLM.ps1
#
# Note: Some IDE terminals run PowerShell with -NoProfile, so profile functions
#   (k-claude, qq-claude, …) are skipped. Step 3b writes k-claude.cmd / qq-claude.cmd
#   under %AppData%\npm and ~/.local/bin so aliases still resolve from PATH.
# ================================================================

function Write-Header  { param($msg) Write-Host "`n-- $msg" -ForegroundColor Cyan }
function Write-Success { param($msg) Write-Host "  [OK] $msg" -ForegroundColor Green }
function Write-Info    { param($msg) Write-Host "  [..] $msg" -ForegroundColor Yellow }
function Write-Err     { param($msg) Write-Host "  [!!] $msg" -ForegroundColor Red }

Write-Host @'
  Claude Code Multi-Profile Setup
  claude -> Opus  |  g-claude -> GLM  |  q-claude -> Qwen 3.5 Plus  |  k-claude -> Kimi K2.6  |  qq-claude -> Qwen 3.6 Max Preview  |  c-claude -> Codex (gpt-5.4)  |  init-claude / adopt-claude
'@ -ForegroundColor Cyan

# Shared permissions fragment: deny destructive rm, IDE/graph MCP, scoped Bash (no blanket Bash)
$permissionsDenyAndBashAllows = @"
    "deny": [
      "Bash(rm -rf /*)"
    ],
    "allow": [
      "mcp__ide__getDiagnostics",
      "mcp__plugin_code-review-graph_code-review-graph__build_or_update_graph_tool",
      "Bash(git:*)",
      "Bash(npm:*)",
      "Bash(date:*)",
      "Bash(echo:*)",
      "Bash(cat:*)",
      "Bash(ls:*)",
      "Bash(cd *)",
      "Bash(cd:*)",
      "Bash(mkdir:*)",
      "Bash(wc:*)",
      "Bash(head:*)",
      "Bash(tail:*)",
      "Bash(sort:*)",
      "Bash(grep:*)",
      "Bash(tr:*)",
      "Bash(git add:*)",
      "Bash(git commit:*)",
      "Bash(git status:*)",
      "Bash(git log:*)",
      "Bash(git diff:*)",
      "Bash(git tag:*)",
      "Bash(php artisan:*)",
      "Bash(php -l:*)",
      "Bash(php -r:*)",
      "Bash(cd * && php -r:*)",
      "Bash(npx playwright-cli:*)",
      "Bash(playwright-cli:*)",
      "Bash(playwright-cli run-code:*)",
      "Bash(dir:*)",
      "Bash(npx eslint:*)",
      "Bash(php artisan tinker --execute=\":*)",
      "Bash(npx tsc:*)",
      "Bash(npx vite:*)",
      "Bash(sqlite3:*)",
      "Bash(cd * && git *)",
      "Bash(cd:*&& git:*)",
      "Bash(cd *&& git *)",
      "Bash(cd * && git diff *)",
      "Bash(cd * && git log *)",
      "Bash(cd * && git status *)",
      "Bash(cd * && git show *)",
"@


# -- Step 1: API Keys (hardcoded) ──────────────────────────────
Write-Header "Step 1: API Keys"
Write-Info "claude uses your Claude.ai browser login - no API key needed for Opus."

$ZaiApiKey = "986cf7cd0f2840fb97a6223ded350268.pGbtSMBrUz3jzgby"
Write-Success "Z.ai API key loaded (hardcoded)."

$QwenApiKey = "sk-sp-5e918826caa24811b6fda891df88ad30"
Write-Success "Qwen/DashScope API key loaded (hardcoded)."

$KimiApiKey = "sk-kimi-NVHuylj0z6QMhmth0ZSD361x7jlpkciGDhjefncT22N2wK7vH7U4sb41tU7meVmV"
Write-Success "Kimi Code (Moderato) API key loaded (hardcoded)."

$QqDashScopeApiKey = "sk-34d9b9d3dfef4891884ed292a7c5e984"
Write-Success "DashScope Intl (qq-claude) API key loaded (hardcoded)."

$Mem0ApiKey = "m0-FOzf3SEJII1RtDA6q6oP5hNZUMaj3xHZk9rDxSCX"
Write-Success "Mem0 API key loaded (hardcoded)."


# -- Step 2: Resolve Claude CLI binary path ──────────────────
Write-Header "Step 2: Locating Claude CLI"

$claudeBin = (Get-Command claude -CommandType Application -ErrorAction SilentlyContinue |
              Select-Object -First 1).Source

if ($claudeBin) {
    Write-Success "Found: $claudeBin"
} else {
    Write-Err "Claude CLI not found. Install it first:"
    Write-Info "npm install -g @anthropic-ai/claude-code"
    Write-Info "Then re-run this script."
    exit 1
}

$codexCmd = Get-Command codex -ErrorAction SilentlyContinue | Select-Object -First 1
$codexBin = if ($codexCmd) { $codexCmd.Source } else { $null }
if ($codexBin) {
    Write-Success "Codex CLI: $codexBin"
} else {
    $codexBin = "codex"
    Write-Info "Codex CLI not in PATH; c-claude will call 'codex'. Install: npm install -g @openai/codex"
}


# -- Step 3: Create Profile Directories ──────────────────────
Write-Header "Step 3: Creating Profile Directories"

$primaryDir = "$HOME\.claude"
$glmDir     = "$HOME\.claude-glm"
$qwenDir    = "$HOME\.claude-qwen"
$kimiDir    = "$HOME\.claude-kimi"
$qqDir      = "$HOME\.claude-qq"

New-Item -ItemType Directory -Force -Path $primaryDir | Out-Null
New-Item -ItemType Directory -Force -Path $glmDir     | Out-Null
New-Item -ItemType Directory -Force -Path $qwenDir    | Out-Null
New-Item -ItemType Directory -Force -Path $kimiDir   | Out-Null
New-Item -ItemType Directory -Force -Path $qqDir    | Out-Null

Write-Success "Opus profile : $primaryDir"
Write-Success "GLM profile  : $glmDir"
Write-Success "Qwen profile : $qwenDir"
Write-Success "Kimi profile : $kimiDir"
Write-Success "QQ (Qwen 3.6 Max) profile : $qqDir"


# -- Step 3b: CMD shims on PATH (IDE terminals often use -NoProfile) ──
Write-Header "Step 3b: PATH shims (NoProfile-safe)"

$claudeExeForShim = Join-Path $HOME '.local\bin\claude.exe'
if (-not (Test-Path -LiteralPath $claudeExeForShim)) {
    $claudeExeForShim = $claudeBin
}

function Write-ClaudeAliasShim {
    param(
        [Parameter(Mandatory = $true)][string]$OutPath,
        [Parameter(Mandatory = $true)][string]$SettingsUnderProfile
    )
    # %USERPROFILE% + relative path → settings json (works for any Windows user)
    $invoke = "`"$claudeExeForShim`" --setting-sources user,project --settings `"%USERPROFILE%$SettingsUnderProfile`" %*"
    @(
        '@echo off'
        'REM Claude Code alias shim — generated by Setup-ClaudeGLM.ps1'
        $invoke
    ) -join "`r`n" | Set-Content -LiteralPath $OutPath -Encoding ascii
}

$shimDirs = @(
    (Join-Path $HOME 'AppData\Roaming\npm'),
    (Join-Path $HOME '.local\bin')
)
$aliasShims = [ordered]@{
    'k-claude.cmd'  = '\.claude-kimi\settings.json'
    'qq-claude.cmd' = '\.claude-qq\settings.json'
}
foreach ($dir in $shimDirs) {
    New-Item -ItemType Directory -Force -Path $dir | Out-Null
    foreach ($name in $aliasShims.Keys) {
        Write-ClaudeAliasShim -OutPath (Join-Path $dir $name) -SettingsUnderProfile $aliasShims[$name]
    }
}
Write-Success "Wrote k-claude.cmd + qq-claude.cmd to npm + .local\bin (usable when profile is not loaded)"


# -- Step 4: Primary Profile - Claude.ai Login (Opus) ────────
Write-Header "Step 4: Opus Profile (Claude.ai login)"

$opusJson = @"
{
  "permissions": {
$permissionsDenyAndBashAllows
      "Read",
      "Edit",
      "Write",
      "Fetch",
      "Web_search",
      "WebSearch",
      "WebFetch",
      "Skill(playwright-cli)",
      "Skill(get-api-docs)",
      "Skill(prisma)",
      "Skill(working-with-claude-code)",
      "Skill(developing-claude-code-plugins)",
      "mcp__plugin_playwright_playwright__browser_navigate",
      "mcp__plugin_playwright_playwright__browser_click",
      "mcp__plugin_context7_context7__resolve-library-id",
      "mcp__plugin_context7_context7__query-docs",
      "mcp__plugin_playwright_playwright__browser_console_messages",
      "mcp__plugin_playwright_playwright__browser_network_requests",
      "mcp__plugin_playwright_playwright__browser_wait_for",
      "mcp__plugin_playwright_playwright__browser_fill_form",
      "mcp__plugin_playwright_playwright__browser_snapshot",
      "mcp__plugin_playwright_playwright__browser_close"
    ]
  },
  "env": {
    "MEM0_API_KEY": "$Mem0ApiKey"
  },
  "enabledPlugins": {
    "superpowers@claude-plugins-official": true,
    "frontend-design@claude-plugins-official": true,
    "mem0@mem0-plugins": true
  },
  "extraKnownMarketplaces": {
    "mem0-plugins": {
      "source": {
        "source": "github",
        "repo": "mem0ai/mem0"
      }
    }
  },
  "skipDangerousModePermissionPrompt": true
}
"@
$opusJson | Out-File -FilePath "$primaryDir\settings.json" -Encoding UTF8

Write-Success "Written: $primaryDir\settings.json (Bash, Read, Edit, Write, fetch, web_search, WebFetch, Mem0 plugins)"
Write-Info "Uses Claude.ai browser auth - no API key stored."


# -- Step 5: GLM Profile - Z.ai ──────────────────────────────
Write-Header "Step 5: GLM Profile - Z.ai"

# Update ANTHROPIC_DEFAULT_xxx_MODEL values when Z.ai publishes glm-5.1 model string
$glmJson = @"
{
  "permissions": {
$permissionsDenyAndBashAllows
      "Read",
      "Edit",
      "Write",
      "Fetch",
      "Web_search",
      "WebSearch",
      "WebFetch",
      "Skill(playwright-cli)",
      "Skill(get-api-docs)",
      "Skill(prisma)",
      "Skill(working-with-claude-code)",
      "Skill(developing-claude-code-plugins)",
      "mcp__plugin_playwright_playwright__browser_navigate",
      "mcp__plugin_playwright_playwright__browser_click",
      "mcp__plugin_context7_context7__resolve-library-id",
      "mcp__plugin_context7_context7__query-docs",
      "mcp__plugin_playwright_playwright__browser_console_messages",
      "mcp__plugin_playwright_playwright__browser_network_requests",
      "mcp__plugin_playwright_playwright__browser_wait_for",
      "mcp__plugin_playwright_playwright__browser_fill_form",
      "mcp__plugin_playwright_playwright__browser_snapshot",
      "mcp__plugin_playwright_playwright__browser_close"
    ]
  },
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "$ZaiApiKey",
    "ANTHROPIC_BASE_URL": "https://api.z.ai/api/anthropic",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "glm-5.1",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-5.1",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "glm-5.1"
  }
}
"@
$glmJson | Out-File -FilePath "$glmDir\settings.json" -Encoding UTF8

Write-Success "Written: $glmDir\settings.json (Bash, Read, Edit, Write, fetch, web_search, WebFetch)"
Write-Info "Update model strings to glm-5.1 in the file above once Z.ai confirms the name."


# -- Step 5a: Qwen Profile - DashScope ────────────────────────
Write-Header "Step 5a: Qwen Profile - DashScope (qwen3.5-plus)"
# ANTHROPIC_MODEL must match Coding Plan / Model Studio model list.

$qwenJson = @"
{
  "permissions": {
$permissionsDenyAndBashAllows
      "Read",
      "Edit",
      "Write",
      "Fetch",
      "Web_search",
      "WebSearch",
      "WebFetch",
      "Skill(playwright-cli)",
      "Skill(get-api-docs)",
      "Skill(prisma)",
      "Skill(working-with-claude-code)",
      "Skill(developing-claude-code-plugins)",
      "mcp__plugin_playwright_playwright__browser_navigate",
      "mcp__plugin_playwright_playwright__browser_click",
      "mcp__plugin_context7_context7__resolve-library-id",
      "mcp__plugin_context7_context7__query-docs",
      "mcp__plugin_playwright_playwright__browser_console_messages",
      "mcp__plugin_playwright_playwright__browser_network_requests",
      "mcp__plugin_playwright_playwright__browser_wait_for",
      "mcp__plugin_playwright_playwright__browser_fill_form",
      "mcp__plugin_playwright_playwright__browser_snapshot",
      "mcp__plugin_playwright_playwright__browser_close"
    ]
  },
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "$QwenApiKey",
    "ANTHROPIC_BASE_URL": "https://coding-intl.dashscope.aliyuncs.com/apps/anthropic",
    "ANTHROPIC_MODEL": "qwen3.5-plus"
  }
}
"@
$qwenJson | Out-File -FilePath "$qwenDir\settings.json" -Encoding UTF8

Write-Success "Written: $qwenDir\settings.json (Qwen 3.5 Plus via DashScope)"


# -- Step 5b: Kimi Profile - Moonshot ───────────────────────────
Write-Header "Step 5b: Kimi Profile - Moonshot (kimi-k2.6)"

$kimiJson = @"
{
  "permissions": {
$permissionsDenyAndBashAllows
      "Read",
      "Edit",
      "Write",
      "Fetch",
      "Web_search",
      "WebSearch",
      "WebFetch",
      "Skill(playwright-cli)",
      "Skill(get-api-docs)",
      "Skill(prisma)",
      "Skill(working-with-claude-code)",
      "Skill(developing-claude-code-plugins)",
      "mcp__plugin_playwright_playwright__browser_navigate",
      "mcp__plugin_playwright_playwright__browser_click",
      "mcp__plugin_context7_context7__resolve-library-id",
      "mcp__plugin_context7_context7__query-docs",
      "mcp__plugin_playwright_playwright__browser_console_messages",
      "mcp__plugin_playwright_playwright__browser_network_requests",
      "mcp__plugin_playwright_playwright__browser_wait_for",
      "mcp__plugin_playwright_playwright__browser_fill_form",
      "mcp__plugin_playwright_playwright__browser_snapshot",
      "mcp__plugin_playwright_playwright__browser_close"
    ]
  },
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "$KimiApiKey",
    "ANTHROPIC_API_KEY": "$KimiApiKey",
    "ANTHROPIC_BASE_URL": "https://api.kimi.com/coding/",
    "ANTHROPIC_MODEL": "kimi-k2.6",
    "ANTHROPIC_SMALL_FAST_MODEL": "kimi-k2.6",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "kimi-k2.6",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "kimi-k2.6",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "kimi-k2.6",
    "CLAUDE_CODE_SUBAGENT_MODEL": "kimi-k2.6",
    "ENABLE_TOOL_SEARCH": "FALSE",
    "API_TIMEOUT_MS": "600000"
  }
}
"@
$kimiJson | Out-File -FilePath "$kimiDir\settings.json" -Encoding UTF8

Write-Success "Written: $kimiDir\settings.json (Kimi Code / Moderato - api.kimi.com/coding)"


# -- Step 5c: Qwen 3.6 Max Preview - DashScope Intl ────────────
Write-Header "Step 5c: Qwen 3.6 Max Preview - DashScope Intl (compatible-mode)"

$qqJson = @"
{
  "permissions": {
$permissionsDenyAndBashAllows
      "Read",
      "Edit",
      "Write",
      "Fetch",
      "Web_search",
      "WebSearch",
      "WebFetch",
      "Skill(playwright-cli)",
      "Skill(get-api-docs)",
      "Skill(prisma)",
      "Skill(working-with-claude-code)",
      "Skill(developing-claude-code-plugins)",
      "mcp__plugin_playwright_playwright__browser_navigate",
      "mcp__plugin_playwright_playwright__browser_click",
      "mcp__plugin_context7_context7__resolve-library-id",
      "mcp__plugin_context7_context7__query-docs",
      "mcp__plugin_playwright_playwright__browser_console_messages",
      "mcp__plugin_playwright_playwright__browser_network_requests",
      "mcp__plugin_playwright_playwright__browser_wait_for",
      "mcp__plugin_playwright_playwright__browser_fill_form",
      "mcp__plugin_playwright_playwright__browser_snapshot",
      "mcp__plugin_playwright_playwright__browser_close"
    ]
  },
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "$QqDashScopeApiKey",
    "ANTHROPIC_BASE_URL": "https://dashscope-intl.aliyuncs.com/apps/anthropic",
    "ANTHROPIC_MODEL": "qwen3.6-max-preview",
    "ANTHROPIC_SMALL_FAST_MODEL": "qwen3.6-max-preview",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "qwen3.6-max-preview",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "qwen3.6-max-preview",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "qwen3.6-max-preview",
    "CLAUDE_CODE_SUBAGENT_MODEL": "qwen3.6-max-preview",
    "API_TIMEOUT_MS": "600000"
  }
}
"@
$qqJson | Out-File -FilePath "$qqDir\settings.json" -Encoding UTF8

Write-Success "Written: $qqDir\settings.json (qwen3.6-max-preview via DashScope Intl Anthropic route)"


# -- Step 5d: Set hasCompletedOnboarding ───────────────────────
Write-Header "Step 5d: Onboarding flag"

$claudeJsonPath = "$HOME\.claude.json"
if (Test-Path $claudeJsonPath) {
    try {
        $claudeJson = Get-Content $claudeJsonPath -Raw -Encoding UTF8 | ConvertFrom-Json
        $claudeJson.hasCompletedOnboarding = $true
        $claudeJson | ConvertTo-Json -Depth 5 | Set-Content $claudeJsonPath -Encoding UTF8 -NoNewline
        Write-Success "Updated hasCompletedOnboarding = true in $claudeJsonPath"
    } catch {
        Write-Err "Could not update $claudeJsonPath : $_"
    }
} else {
    '{ "hasCompletedOnboarding": true }' | Out-File -FilePath $claudeJsonPath -Encoding UTF8
    Write-Success "Created $claudeJsonPath with hasCompletedOnboarding = true"
}


# -- Step 5e: Project-level .claude/settings.json ────────────
Write-Header "Step 5e: Project-level .claude/settings.json template"
Write-Info "This will be copied into projects by init-claude / adopt-claude."

$projectClaudeJson = @"
{
  "permissions": {
$permissionsDenyAndBashAllows
      "Read",
      "Edit",
      "Write",
      "Fetch",
      "Web_search",
      "WebSearch",
      "WebFetch",
      "Skill(playwright-cli)",
      "Skill(get-api-docs)",
      "Skill(prisma)",
      "Skill(working-with-claude-code)",
      "Skill(developing-claude-code-plugins)",
      "mcp__plugin_playwright_playwright__browser_navigate",
      "mcp__plugin_playwright_playwright__browser_click",
      "mcp__plugin_context7_context7__resolve-library-id",
      "mcp__plugin_context7_context7__query-docs",
      "mcp__plugin_playwright_playwright__browser_console_messages",
      "mcp__plugin_playwright_playwright__browser_network_requests",
      "mcp__plugin_playwright_playwright__browser_wait_for",
      "mcp__plugin_playwright_playwright__browser_fill_form",
      "mcp__plugin_playwright_playwright__browser_snapshot",
      "mcp__plugin_playwright_playwright__browser_close"
    ]
  }
}
"@

New-Item -ItemType Directory -Force -Path "$HOME\.claude-templates" | Out-Null
$projectClaudeJson | Out-File -FilePath "$HOME\.claude-templates\.claude-settings.json" -Encoding UTF8

Write-Success "Saved project .claude/settings.json template to ~/.claude-templates/"


# -- Step 5f: Add WebFetch + MCP permissions to this project's .claude/settings.json if missing ─
Write-Header "Step 5f: Current project .claude/settings.json (add WebFetch + MCP if missing)"
$runDir = if ($PSScriptRoot) { $PSScriptRoot } else { Split-Path -Parent $MyInvocation.MyCommand.Path }
$projectSettingsPath = Join-Path $runDir ".claude\settings.json"
$permissionsToEnsure = @(
    "Bash(cd *)",
    "Bash(cd:*)",
    "WebSearch",
    "Skill(playwright-cli)",
    "Skill(get-api-docs)",
    "Skill(prisma)",
    "Skill(working-with-claude-code)",
    "Skill(developing-claude-code-plugins)",
    "WebFetch",
    "mcp__ide__getDiagnostics",
    "mcp__plugin_code-review-graph_code-review-graph__build_or_update_graph_tool",
    "mcp__plugin_playwright_playwright__browser_navigate",
    "mcp__plugin_playwright_playwright__browser_click",
    "mcp__plugin_context7_context7__resolve-library-id",
    "mcp__plugin_context7_context7__query-docs",
    "mcp__plugin_playwright_playwright__browser_console_messages",
    "mcp__plugin_playwright_playwright__browser_network_requests",
    "mcp__plugin_playwright_playwright__browser_wait_for",
    "mcp__plugin_playwright_playwright__browser_fill_form",
    "mcp__plugin_playwright_playwright__browser_snapshot",
    "mcp__plugin_playwright_playwright__browser_close"
)
$denyToEnsure = @("Bash(rm -rf /*)")
$bashAllowsToEnsure = @(
    "Bash(git:*)", "Bash(npm:*)", "Bash(date:*)", "Bash(echo:*)", "Bash(cat:*)", "Bash(ls:*)",
    "Bash(cd *)", "Bash(cd:*)", "Bash(mkdir:*)", "Bash(wc:*)", "Bash(head:*)", "Bash(tail:*)", "Bash(sort:*)", "Bash(grep:*)", "Bash(tr:*)",
    "Bash(git add:*)", "Bash(git commit:*)", "Bash(git status:*)", "Bash(git log:*)", "Bash(git diff:*)", "Bash(git tag:*)",
    "Bash(php artisan:*)",
    "Bash(php -l:*)",
    "Bash(php -r:*)",
    "Bash(cd * && php -r:*)",
    "Bash(npx playwright-cli:*)",
    "Bash(playwright-cli:*)",
    "Bash(playwright-cli run-code:*)",
    "Bash(dir:*)",
    "Bash(npx eslint:*)",
    'Bash(php artisan tinker --execute=":*)',
    "Bash(npx tsc:*)",
    "Bash(npx vite:*)",
    "Bash(sqlite3:*)",
    "Bash(cd * && git *)",
    "Bash(cd:*&& git:*)",
    "Bash(cd *&& git *)",
    "Bash(cd * && git diff *)",
    "Bash(cd * && git log *)",
    "Bash(cd * && git status *)",
    "Bash(cd * && git show *)"
)
if (Test-Path $projectSettingsPath) {
    try {
        $projectSettings = Get-Content $projectSettingsPath -Raw -Encoding UTF8 | ConvertFrom-Json
        if (-not $projectSettings.permissions) {
            $projectSettings | Add-Member -NotePropertyName permissions -NotePropertyValue ([pscustomobject]@{ allow = @(); deny = @() }) -Force
        } else {
            if (-not ($projectSettings.permissions.PSObject.Properties.Name -contains 'allow')) {
                $projectSettings.permissions | Add-Member -NotePropertyName allow -NotePropertyValue @() -Force
            }
            if (-not ($projectSettings.permissions.PSObject.Properties.Name -contains 'deny')) {
                $projectSettings.permissions | Add-Member -NotePropertyName deny -NotePropertyValue @() -Force
            }
        }
        $allow = [System.Collections.ArrayList]@()
        if ($projectSettings.permissions.allow) {
            $allow.AddRange(@($projectSettings.permissions.allow))
        }
        $added = @()
        foreach ($perm in $permissionsToEnsure) {
            if ($allow -notcontains $perm) {
                $allow.Add($perm) | Out-Null
                $added += $perm
            }
        }
        foreach ($perm in $bashAllowsToEnsure) {
            if ($allow -notcontains $perm) {
                $allow.Add($perm) | Out-Null
                $added += $perm
            }
        }
        $denyList = [System.Collections.ArrayList]@()
        if ($projectSettings.permissions.deny) {
            $denyList.AddRange(@($projectSettings.permissions.deny))
        }
        foreach ($d in $denyToEnsure) {
            if ($denyList -notcontains $d) {
                $denyList.Add($d) | Out-Null
                $added += "deny:$d"
            }
        }
        $changed = $added.Count -gt 0
        if ($changed) {
            $projectSettings.permissions.allow = @($allow)
            $projectSettings.permissions.deny = @($denyList)
            $projectSettings | ConvertTo-Json -Depth 5 | Set-Content $projectSettingsPath -Encoding UTF8 -NoNewline
            Write-Success "Updated $projectSettingsPath : $($added -join ', ')"
        } else {
            Write-Info "WebFetch + MCP + deny/bash permissions already present - skipped."
        }
    } catch {
        Write-Err "Could not update project settings: $_"
    }
} else {
    Write-Info "No .claude/settings.json in current project - skip."
}


# -- Step 6: Save global CLAUDE.md template to user home ─────
Write-Header "Step 6: Saving global AI memory templates to ~\.claude-templates"

$templatesDir = "$HOME\.claude-templates"
$rulesDir     = "$templatesDir\rules"
New-Item -ItemType Directory -Force -Path $rulesDir | Out-Null

# Load templates from embedded files (same dir as script)
$scriptDir = if ($PSScriptRoot) { $PSScriptRoot } else { Split-Path -Parent $MyInvocation.MyCommand.Path }
$templatesSrc = Join-Path $scriptDir "Setup-ClaudeGLM-templates"
if (Test-Path $templatesSrc) {
    Copy-Item (Join-Path $templatesSrc "*") -Destination $templatesDir -Recurse -Force
    Copy-Item (Join-Path $templatesSrc "rules\*") -Destination $rulesDir -Force -ErrorAction SilentlyContinue
    Write-Success "Templates copied from $templatesSrc"
} else {
    # Fallback: create minimal templates inline
    $claudeMd = "# CLAUDE.md - Fill per project. Run init-claude for full template."
    $claudeMd | Out-File "$templatesDir\CLAUDE.md" -Encoding UTF8
    "read/project-context.md" | Out-File "$templatesDir\project-context.md" -Encoding UTF8
    
    $minimalRule = @"
# Placeholder rule file
# Replace with actual content
"@
    
    $minimalRule | Out-File "$rulesDir\workflow.md" -Encoding UTF8
    $minimalRule | Out-File "$rulesDir\speckit.md" -Encoding UTF8
    $minimalRule | Out-File "$rulesDir\code-quality.md" -Encoding UTF8
    $minimalRule | Out-File "$rulesDir\verification.md" -Encoding UTF8
    $minimalRule | Out-File "$rulesDir\lessons.md" -Encoding UTF8
    Write-Info "Minimal templates created. For full templates, add Setup-ClaudeGLM-templates folder."
}
Write-Success "Templates ready at: $templatesDir"


# -- Step 7: Inject Functions into PowerShell Profile ────────
Write-Header "Step 7: PowerShell Profile Functions"

# Build the profile functions block with proper path substitution
$functionsBlock = @"

# --- Claude Multi Profile (do not remove this line) ---
function claude {
    `$opusSettings = "`$HOME\.claude\settings.json"
    & '$claudeBin' --setting-sources user,project --settings `$opusSettings @args
}
function g-claude {
    `$glmSettings = "`$HOME\.claude-glm\settings.json"
    & '$claudeBin' --setting-sources user,project --settings `$glmSettings @args
}
function q-claude {
    `$qwenSettings = "`$HOME\.claude-qwen\settings.json"
    & '$claudeBin' --setting-sources user,project --settings `$qwenSettings @args
}
function k-claude {
    `$kimiSettings = "`$HOME\.claude-kimi\settings.json"
    & '$claudeBin' --setting-sources user,project --settings `$kimiSettings @args
}
function qq-claude {
    `$qqSettings = "`$HOME\.claude-qq\settings.json"
    & '$claudeBin' --setting-sources user,project --settings `$qqSettings @args
}
function c-claude {
    # OpenAI Codex CLI — fixed model gpt-5.4 (override: codex -m other-model ...)
    & '$codexBin' --model gpt-5.4 @args
}
function init-claude {
    `$tpl = "`$HOME\.claude-templates"
    if (-not (Test-Path "CLAUDE.md")) {
        Copy-Item "`$tpl\CLAUDE.md" -Destination "CLAUDE.md" -Force -ErrorAction SilentlyContinue
        if (Test-Path "CLAUDE.md") { Write-Host "Created CLAUDE.md" -ForegroundColor Green }
    }
    if (-not (Test-Path ".claude")) { New-Item -ItemType Directory -Path ".claude" -Force | Out-Null }
    if (Test-Path "`$tpl\.claude-settings.json") {
        Copy-Item "`$tpl\.claude-settings.json" -Destination ".claude\settings.json" -Force
        Write-Host "Created .claude/settings.json" -ForegroundColor Green
    }
    if (-not (Test-Path "read")) { New-Item -ItemType Directory -Path "read" -Force | Out-Null }
    Get-ChildItem "`$tpl\rules" -ErrorAction SilentlyContinue | ForEach-Object {
        Copy-Item `$_.FullName -Destination "read\" -Force
    }
    if (Test-Path "read") { Write-Host "Created read/ directory" -ForegroundColor Green }
}
function adopt-claude {
    `$tpl = "`$HOME\.claude-templates"
    if (-not (Test-Path ".claude")) { New-Item -ItemType Directory -Path ".claude" -Force | Out-Null }
    if (Test-Path "`$tpl\.claude-settings.json") {
        Copy-Item "`$tpl\.claude-settings.json" -Destination ".claude\settings.json" -Force
        Write-Host "Created .claude/settings.json" -ForegroundColor Green
    }
    if (-not (Test-Path "read")) { New-Item -ItemType Directory -Path "read" -Force | Out-Null }
    Get-ChildItem "`$tpl\rules" -ErrorAction SilentlyContinue | ForEach-Object {
        Copy-Item `$_.FullName -Destination "read\" -Force
    }
    if (Test-Path "read") { Write-Host "Added read/ directory" -ForegroundColor Green }
}
"@

# Resolve profile path
$profilePath = if ($PROFILE) { $PROFILE } else {
    "$HOME\Documents\PowerShell\Microsoft.PowerShell_profile.ps1"
}

$profileParent = Split-Path $profilePath -Parent
if (-not (Test-Path $profileParent)) {
    New-Item -ItemType Directory -Force -Path $profileParent | Out-Null
}
if (-not (Test-Path $profilePath)) {
    New-Item -ItemType File -Force -Path $profilePath | Out-Null
    Write-Info "Created new profile file: $profilePath"
}

$existing = Get-Content $profilePath -Raw -ErrorAction SilentlyContinue
$marker = 'Claude Multi Profile'

function Append-ClaudeAliasPatch {
    param(
        [Parameter(Mandatory=$true)][string]$Path,
        [Parameter(Mandatory=$true)][string]$ClaudeExe,
        [Parameter(Mandatory=$true)][string]$Raw
    )
    $needsK = $Raw -notmatch 'function\s+k-claude\b'
    $needsQq = $Raw -notmatch 'function\s+qq-claude\b'
    if (-not $needsK -and -not $needsQq) { return $false }
    $aliasPatchSb = [System.Text.StringBuilder]::new()
    [void]$aliasPatchSb.AppendLine('')
    [void]$aliasPatchSb.AppendLine('# --- Setup-ClaudeGLM: k-claude / qq-claude patch ---')
    if ($needsK) {
        [void]$aliasPatchSb.Append(@"

function k-claude {
    `$kimiSettings = "`$HOME\.claude-kimi\settings.json"
    & '$ClaudeExe' --setting-sources user,project --settings `$kimiSettings @args
}
"@)
    }
    if ($needsQq) {
        [void]$aliasPatchSb.Append(@"

function qq-claude {
    `$qqSettings = "`$HOME\.claude-qq\settings.json"
    & '$ClaudeExe' --setting-sources user,project --settings `$qqSettings @args
}
"@)
    }
    Add-Content -Path $Path -Value $aliasPatchSb.ToString() -Encoding UTF8
    return $true
}

if ($existing -and $existing.Contains($marker)) {
    if (Append-ClaudeAliasPatch -Path $profilePath -ClaudeExe $claudeBin -Raw $existing) {
        Write-Success "Appended k-claude / qq-claude (were missing) to: $profilePath"
    } else {
        Write-Info "Functions already present in profile - skipping injection."
    }
} else {
    Add-Content -Path $profilePath -Value $functionsBlock -Encoding UTF8
    Write-Success "Functions injected into: $profilePath"
}


# -- Done
Write-Host "`n" -NoNewline
Write-Host "  Setup Complete!" -ForegroundColor Green
Write-Host "  Reload your profile: . `$PROFILE" -ForegroundColor Yellow
Write-Host "  For NEW projects: init-claude | For EXISTING: adopt-claude" -ForegroundColor Yellow
Write-Host "  claude = Opus | g-claude = GLM | q-claude = Qwen 3.5 Plus | k-claude = Kimi K2.6 | qq-claude = Qwen 3.6 Max Preview | c-claude = Codex gpt-5.4" -ForegroundColor Cyan
Write-Host ""
