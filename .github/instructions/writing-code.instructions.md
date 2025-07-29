---
applyTo: '**'
---

- When writing code, ensure that the core logic remains intact unless explicitly instructed to modify it. Always seek permission before altering the core logic.
- dont have migrattion phases. this a library, not an application and doesn't have a migration phase. users of the library will use the latest version.
- plans shouldnt span multiple weeks. they are to be done in a single day or a few hours.
- when coming up with a plan, focus on the immediate next steps rather than long-term goals.
- each major feature or functionality should be broken down into small, manageable tasks that can be completed quickly.
- on completing a plan task, immediately test it with pre-commit-check script(fast mode) to ensure it works as expected. each major task or phase should run the full mode of the pre-commit-check script to ensure everything is functioning correctly.