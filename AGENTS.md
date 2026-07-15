# AGENTS.md

Guidance for coding agents working in **flowgraph**.

## Commenting style

- Two lines max per comment block. If it needs more, the identifier is
  probably named wrong -- fix the name, not the comment.
- Never reference issue or PR numbers in source comments. Rationale that
  matters belongs in `DESIGN.md`, not tied to a tracker entry that outlives
  the code differently than the code does.
- Don't restate what the identifier already says. `// Pipe returns a pipe
  by index` teaches nothing `Pipe(n int) Pipe` didn't already say -- delete
  it or say something the signature can't.
- Comment the non-obvious only: a naming decision, a subtle invariant, a
  reason something is shaped the way it is. See `pipe.go`'s Stream->Pipe
  note for the target length and tone.
- Source is the documentation. See https://wiki.c2.com/?StudyTheSourceWithaDebugger.

## Where design rationale goes

Naming and architecture decisions that need more than two lines go in
`DESIGN.md`, not inline. Keep entries as short as the reasoning allows.
