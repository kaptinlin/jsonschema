# Compilation and Reference Lifecycles

## Completed graphs

`Compile` and `CompileBatch` return only after initialization, dialect handling,
reference binding, and compilation checks succeed for every newly loaded
resource. A private resource view supports batch meta-schemas, forward references,
and legal cycles. Loader dependencies join the same compilation. Publication
occurs under the registry lock; loaders and caller callbacks run outside it.

Published graphs are never rebound or reinitialized by later compilation or
constructor composition. Concurrent validation and compilation are supported
when callers do not mutate schemas, instances, or unsynchronized configuration
and callbacks are concurrency-safe. This does not freeze format registrations,
caller assertion policy, custom codecs, or callback state.

## References and errors

References that cannot be processed fail compilation with
`ErrReferenceResolution`. Errors identify the owning resource, resource-relative
keyword location, reference text, and underlying cause. Missing anchors do not
fall back across resource boundaries. JSON Pointers start at the resource root;
array indices use their canonical decimal representation. An explicit empty
reference is normalized to `#`, preserving self-reference through serialization.

`$dynamicRef` requires a resolved initial target and retains runtime dynamic-scope
selection. Legacy keyword paths remain addressable, including subschemas that
are inert as applicators but are referenced elsewhere.

`Compile` does not perform full meta-validation. `ValidateSchema` remains the
separate entry point for that task. `Schema.Unmarshal` applies defaults without
validating instances.

## Resource identity and failure

Each resolved resource URI identifies one definition per compiler. `Compile` and
`CompileBatch` reject explicit duplicate definitions with
`ErrSchemaConflict`, including repeated submission of the same document, nested
resource conflicts, and concurrent competing submissions. An empty URI fragment
does not establish a distinct identity. Retrieval addresses and declared `$id`
addresses can alias the same resource. `Schema` retrieves existing definitions;
replacement definitions use a new compiler.

Concurrent retrieval of the same missing dependency reuses the first successfully
published resource. A losing private build is discarded and rebuilt against the
registry; its references must not retain a different copy of that resource. This
applies to `Schema`, `Compile`, and `CompileBatch`, including legal loader cycles.
A retry requires observable publication of a previously missing retrieval URI.
Loaders may run concurrently or more than once and should return stable content
for a resource URI. If concurrent responses differ, the first complete graph
published determines the shared definition; content equivalence is not computed.

A failed operation does not publish incomplete resources, overwrite existing
bindings, or mutate old graphs. Batch resources are published together. There is
no rollback promise for loader side effects or independently completed caches.

## In-memory entry points

Compilation accepts standalone JSON documents. To compile a document assembled
with constructors, call `Schema.MarshalJSON` followed by `Compiler.Compile`.
Serialization transfers document content, not a compiled node's parent resource,
inherited dialect, or compiler policy. It is not a graph clone or import API.

Reuse a compiled root or subschema directly in constructor composition. It keeps
its original resource scope, resolved targets, dialect, and effective compiler.
There is no `SetSchema` API: silently serializing arbitrary nodes made registration
ambiguous and lost the context of references in extracted subschemas. Cross-compiler
graph import is not needed for direct composition and is not provided.

Constructors remain convenient `*Schema` builders. Direct validation of an
uncompiled graph checks reference readiness before evaluating applicators, so
unresolved references cannot be hidden by `not` or an unused alternative. It
returns an `unresolved_reference` evaluation error without loading or binding.
Struct-tag generation propagates reference errors instead of caching an
unresolved result.

Public field edits belong to construction. Editing compiled fields does not
rebuild caches or rebind references; obtain a new compiled result after editing.
No automatic delayed-binding method or unresolved-reference waiting queue is
part of the lifecycle.

## Mutable builders and validation cost

Direct validation of an uncompiled builder traverses its graph to check reference
readiness on every call. This includes unused definitions: invalid schema state
cannot become acceptable merely because an instance does not visit that branch.
Public fields are editable, so caching readiness would allow stale bindings to
bypass this check. Compiled graphs skip the traversal.

Keep this distinction explicit rather than adding automatic freezing or cache
invalidation. For repeated validation of large constructor documents, serialize
and compile once, then reuse the returned graph. The constructor lifecycle
benchmark compares construction, compilation, and both validation paths using
the same constraints with increasing numbers of unused definitions.
