Read the plan.md file in this workspace root. Based on the plan, create br tasks with proper structure.

For each task:
1. Use `br create -t task "Title" -d "description" -p N` where N is priority (1=critical, 2=important, 3=nice-to-have)
2. The description (-d flag) must include:
   - **Why**: context and motivation
   - **What**: specific deliverables
   - **Acceptance criteria**: checkboxes for done-ness
   - **Key files**: paths that will be created or modified
3. Set dependencies with `br dep add <id> <blocker-id>` where appropriate
4. Keep tasks small enough for one focused session (~30min-2hr)
5. Use specific titles (e.g., "Add SQLite FTS index for note search" not "implement search")

First read plan.md, then create the tasks in dependency order (foundations first).
