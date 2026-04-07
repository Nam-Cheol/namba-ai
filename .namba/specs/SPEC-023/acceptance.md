# Acceptance

- [ ] `namba <command> --help`, `namba <command> -h`, and `namba help <command>` all return command-specific usage/help output and exit without mutating repository state.
- [ ] `namba help <command>` is semantically aligned with `<command> --help`, with output that is either identical or deliberately equivalent and stable enough for regression tests.
- [ ] This read-only help contract is applied consistently across at least `init`, `doctor`, `status`, `project`, `regen`, `update`, `plan`, `harness`, `fix`, `run`, `sync`, `pr`, `land`, `release`, and `worktree`.
- [ ] Help parsing happens before any project-root lookup, git-repo check, GitHub auth check, validation run, file write, SPEC creation, PR action, merge, or network download path.
- [ ] Commands that do not accept extra args or flags do not silently ignore them; `--help` shows usage, and other stray args fail with a non-mutating error.
- [ ] `namba project --help` and `namba sync --help` are read-only and do not regenerate `.namba/project/*` or any other managed artifacts.
- [ ] `namba pr --help`, `namba land --help`, `namba release --help`, `namba update --help`, `namba run --help`, `namba worktree --help`, and `namba init --help` produce help output instead of falling through to `unknown flag`, repo/auth errors, or other environment-dependent failures.
- [ ] `namba plan --help`, `namba harness --help`, and `namba fix --help` remain read-only and do not create or mutate `.namba/specs/<SPEC>`.
- [ ] `plan`, `harness`, and `fix` support a delimiter form so flag-like text after `--` is preserved as literal description input rather than being reinterpreted as help or an option.
- [ ] Unknown flags, missing flag values, malformed subcommands, and invalid flag combinations fail without mutating repo state.
- [ ] Help output and malformed-invocation error output stay clearly distinguishable, so users can tell whether they asked for documentation or hit a usage failure.
- [ ] Regression tests explicitly cover:
  - help probing for each supported top-level command
  - commands that previously ignored args and still performed work
  - malformed invocations that must remain non-mutating
  - literal flag-like description text after `--`
- [ ] Any user-facing docs or generated guidance that describe help probing are updated to match the hardened CLI contract.
- [ ] Validation commands pass.
