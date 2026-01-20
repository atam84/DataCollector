# Vibe-coding Workflow (Docs + Dev Tracking)

This workflow ensures you always know:
- **what** you are building (PRD),
- **how** it should work (Architecture),
- **how** it will be implemented (Implementation Guides + ADRs),
- **what to do next** (Tasks),
- **how to execute quickly with AI** (Prompts).

## Canonical artifacts (the rule)

1. **PRD** is the source of truth for scope and behavior.
2. **ADRs** capture decisions that change architecture/implementation.
3. **Architecture** describes the agreed structure and contracts.
4. **Implementation Guides** provide step-by-step changes for developers.
5. **TASK-TO-DO** is the single backlog.
6. **PROMPTS** are pre-written instructions to generate code and deliverables.

If something is unclear: create an ADR and update PRD/Architecture.

## Folder structure

```
docs/
  00-Overview/
  01-PRD/
  02-Architecture/
  03-Implementation/
  04-API/
  05-Prompts/
  06-Dev-Tracking/
  07-ADRs/
  08-Quality/
```

## How to track progress (simple)

- Each sprint produces:
  - 1 PRD delta (if scope changed)
  - 0..N ADRs (only if decisions changed)
  - 1 or more concrete deliverables (code, tests, migration, docs)
  - updated TASK-TO-DO status

### Definition of Done (per sprint)

- Endpoints are implemented + tested.
- DB changes are migrated.
- Observability is in place (logs/metrics for new flows).
- Docs updated (PRD/Architecture/Implementation as needed).

## Vibe-coding prompt hygiene

- Each prompt must specify:
  - **goal**, **constraints**, **inputs/outputs**, **acceptance criteria**.
- Prefer prompts per sprint (not per milestone), so outputs are small and testable.

