# Gas City backend store conformance

`GascityBackendStore` is a subprocess-backed implementation of the same
`beads.Store` contract used by the in-memory, file, exec, and native stores.
Its protocol adapter must preserve object-model semantics even when the backend
stores some values relationally instead of as issue columns.

The executable contract is `TestGascityBackendStoreConformance`, which runs the
shared `beadstest.RunStoreTests` suite through a real newline-delimited JSON
subprocess boundary. The protocol oracle uses `MemStore` so the expected
behavior comes from Gas City's canonical store model rather than from the
DoltLite plugin implementation.

| Store surface | Protocol operation | Contract notes |
|---|---|---|
| `Create` | `create_issue` | IDs, defaults, sender/from, labels, metadata, parent and dependency records round-trip. |
| `Get` | `get_issue` | Missing IDs map to `ErrNotFound`; parent is derived from `parent-child` when necessary. |
| `Update` | `update_issue` | Row fields, label additions/removals, metadata patches and parent replacement retain distinct semantics. |
| `Close` | `close_issue` | Idempotent; removes work from ready views. |
| `Reopen` | `reopen_issue` | Restores closed work to open state. |
| `CloseAll` | composed operations | Already-closed and missing beads are skipped and not counted. |
| `List` | `search_issues`, `list_wisps` | Fetches a safe superset, unions physical tiers, then applies canonical filtering, sorting and limiting. |
| `ListOpen` | composed `List` | Uses canonical closed-status filtering. |
| `Ready` | `ready_work` | Applies canonical assignee, tier and limit filtering after backend retrieval. |
| `Children` | composed `List` | Matches derived `parent-child` relationships. |
| `ListByLabel` | composed `List` | Exact-match and limit behavior follow `ListQuery`. |
| `ListByAssignee` | composed `List` | Status and assignee are both enforced. |
| `ListByMetadata` | composed `List` | All metadata fields use AND semantics. |
| `SetMetadata` | `slot_set` | Empty strings remain observable as clears. |
| `SetMetadataBatch` | repeated `slot_set` | Sequential external-store semantics are explicit. |
| `Tx` | sequential transaction facade | Callback and error propagation match non-atomic external-store rules. |
| `Delete` | `delete_issue` | Missing IDs map to `ErrNotFound`. |
| `Ping` | `ping` | Performs a real subprocess round-trip rather than returning a constant success. |
| `DepAdd` | `add_dependency` | Preserves arbitrary dependency types. |
| `DepRemove` | `remove_dependency` | Idempotent relationship removal. |
| `DepList` | `get_dependencies`, `get_dependents` | Returns canonical dependency records; never reconstructs or hard-codes types. |

Plugin-specific linked integration additionally verifies that a mixed field,
label-add, label-remove and parent update commits atomically against DoltLite.
The integration test must link the same DoltLite format version used by the
target city; a mismatched static library can report a valid database as
malformed.
