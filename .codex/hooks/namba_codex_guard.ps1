# NambaAI Codex lifecycle hook guard wrapper for Windows PowerShell 5.
$ErrorActionPreference = "Stop"

$scriptPath = Join-Path $PSScriptRoot "namba_codex_guard.py"
$rawInput = [Console]::In.ReadToEnd()
$eventName = "Unknown"
$failureMessage = "NambaAI hook guard failed: Python 3 was not found or could not run .codex/hooks/namba_codex_guard.py. Install Python 3 or the Python launcher, then run namba regen and review /hooks again. Hooks are guardrails, so continue only with explicit validation."

function Write-NambaHookFailure {
    param(
        [string]$Message,
        [string]$HookEventName
    )

    [Console]::Error.WriteLine($Message)
    $output = [ordered]@{
        hookSpecificOutput = [ordered]@{
            hookEventName = $HookEventName
            additionalContext = $Message
        }
    }
    $output | ConvertTo-Json -Compress -Depth 5
}

try {
    $trimmedInput = $rawInput.Trim()
    if ($trimmedInput.Length -gt 0) {
        $payload = $trimmedInput | ConvertFrom-Json -ErrorAction Stop
        if ($payload.hook_event_name) {
            $eventName = [string]$payload.hook_event_name
        }
    }
} catch {
    $eventName = "Unknown"
}

try {
    function Quote-NambaArgument {
        param([string]$Value)
        return '"' + ($Value -replace '"', '\"') + '"'
    }

    function Invoke-NambaPythonHook {
        param(
            [string]$Executable,
            [string[]]$Arguments,
            [string]$InputText
        )

        $startInfo = New-Object System.Diagnostics.ProcessStartInfo
        $startInfo.FileName = $Executable
        $startInfo.Arguments = ($Arguments | ForEach-Object { Quote-NambaArgument $_ }) -join " "
        $startInfo.UseShellExecute = $false
        $startInfo.RedirectStandardInput = $true
        $startInfo.RedirectStandardOutput = $true
        $startInfo.RedirectStandardError = $true
        $startInfo.CreateNoWindow = $true

        $process = New-Object System.Diagnostics.Process
        $process.StartInfo = $startInfo

        try {
            [void]$process.Start()
            $process.StandardInput.Write($InputText)
            $process.StandardInput.Close()
            $stdout = $process.StandardOutput.ReadToEnd()
            $stderr = $process.StandardError.ReadToEnd()
            $process.WaitForExit()
            return @{
                ExitCode = $process.ExitCode
                Stdout = $stdout
                Stderr = $stderr
            }
        } catch {
            return @{
                ExitCode = -1
                Stdout = ""
                Stderr = $_.Exception.Message
            }
        } finally {
            if ($process -ne $null) {
                $process.Dispose()
            }
        }
    }

    $candidates = @(
        @{ Name = "py.exe"; Args = @("-3", $scriptPath) },
        @{ Name = "py"; Args = @("-3", $scriptPath) },
        @{ Name = "python.exe"; Args = @($scriptPath) },
        @{ Name = "python"; Args = @($scriptPath) },
        @{ Name = "python3.exe"; Args = @($scriptPath) },
        @{ Name = "python3"; Args = @($scriptPath) }
    )

    $failedCandidates = @()
    foreach ($candidate in $candidates) {
        $resolved = Get-Command $candidate.Name -ErrorAction SilentlyContinue
        if (-not $resolved) {
            continue
        }

        $executable = $candidate.Name
        if ($resolved.Path) {
            $executable = $resolved.Path
        } elseif ($resolved.Source) {
            $executable = $resolved.Source
        }

        $result = Invoke-NambaPythonHook -Executable $executable -Arguments $candidate.Args -InputText $rawInput
        if ($result.Stderr.Trim().Length -gt 0) {
            [Console]::Error.Write($result.Stderr)
        }
        if ($result.ExitCode -eq 0) {
            if ($result.Stdout.Trim().Length -gt 0) {
                [Console]::Out.Write($result.Stdout)
            }
            exit 0
        }

        $summary = $candidate.Name + " exited " + $result.ExitCode
        if ($result.Stderr.Trim().Length -gt 0) {
            $summary = $summary + ": " + $result.Stderr.Trim()
        }
        $failedCandidates += $summary
    }

    if ($failedCandidates.Count -gt 0) {
        [Console]::Error.WriteLine("NambaAI hook launcher candidates failed: " + ($failedCandidates -join "; "))
    }

    Write-NambaHookFailure -Message $failureMessage -HookEventName $eventName
    exit 1
} catch {
    Write-NambaHookFailure -Message ("NambaAI hook launcher failed with an unhandled exception: " + $_.Exception.Message) -HookEventName $eventName
    exit 1
}
