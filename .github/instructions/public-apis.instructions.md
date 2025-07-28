---
applyTo: '**'
---

- Public APIs should be first implemented in the fir/internal package until explicitly requested to be in the fir package. This allows for migrating APIs to the public package when they are stable and ready for external use.