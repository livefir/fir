---
applyTo: '**'
---

- While fixing tests, ensure core logic is preserved until explicitly requested to change it. Ask for permission before making any changes to the core logic.
- If a test is failing, first check if the test itself is correct. If the test is incorrect, fix the test rather than the code it tests.
- If a test is failing due to a missing feature, implement the feature in the `fir/internal` package first. This allows for migrating features to the public API when they are stable and ready for external use.
