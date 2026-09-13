# Format Vocabularies

## Overview

The package treats `format` according to the active schema resource's dialect.
Standard Draft 2020-12 uses Format-Annotation, while a custom meta-schema can
declare Format-Assertion. Callers can separately request best-effort assertion
through `Compiler.SetAssertFormat`.

This specification owns format assertion policy, format lookup precedence,
unknown-format behavior, annotation output, and inheritance across schema
resources. Individual format validators remain owned by `formats.go` and the
standards named in the public format table.

## Terminology

| Term | Meaning |
|------|---------|
| Annotation mode | `format` records its value but does not affect validity. |
| Best-effort assertion | Caller policy enabled by `SetAssertFormat(true)`; recognized formats assert and unknown names remain annotations. |
| Vocabulary assertion | Required behavior when the active dialect declares the Draft 2020-12 Format-Assertion vocabulary. |
| Recognized format | A name with a compiler-specific validator or a package-level validator. |
| Schema resource | A schema scope that can select its own dialect with `$schema`; nested resources inherit no dialect behavior past an explicit `$schema`. |

## Contracts

### Default and caller policy

- A standard Draft 2020-12 schema treats `format` as an annotation by default.
- `SetAssertFormat(true)` asserts every recognized format without changing the
  schema's dialect.
- Unknown format names remain annotations in best-effort mode.
- `SetAssertFormat(false)` does not disable behavior required by an active
  Format-Assertion vocabulary.

### Vocabulary assertion

- A custom dialect declares Format-Assertion with
  `https://json-schema.org/draft/2020-12/vocab/format-assertion` in
  `$vocabulary`.
- Presence of that URI activates its semantics. The associated boolean says
  whether an implementation that does not understand the vocabulary must reject
  the schema; it is not an enablement switch.
- Consequently, declarations with values `true` and `false` behave identically
  once this package recognizes Format-Assertion.
- Declaring Format-Annotation and Format-Assertion together has the same
  assertion behavior as declaring Format-Assertion alone.
- Every recognized format validates only its applicable instance types.
  Draft 2020-12 standard formats apply to strings, so non-string values pass.

### Lookup and failures

- Compiler-specific registrations take precedence over package-level entries in
  `Formats`.
- A compiler-registered custom format is recognized by Format-Assertion and its
  validator determines validity.
- An unknown format under Format-Assertion is a schema-processing error. `Compile`
  and `CompileBatch` return an error wrapping `ErrUnknownFormat` and naming the
  format.
- If a recognized format is removed after compilation, evaluation fails
  defensively with the `unknown_format` code.
- A required unrelated unknown vocabulary still fails with
  `ErrUnsupportedVocabulary`; an optional unknown vocabulary is ignored.

### Resource scope and references

- Format-Assertion state is resolved during compilation and stored on each
  schema node.
- Subschemas inherit the state of their containing schema resource.
- A nested resource with its own `$schema` recomputes the state from that
  meta-schema. Selecting standard Draft 2020-12 resets to annotation mode;
  selecting an assertion dialect enables assertion for that resource and its
  descendants.
- `$ref` and `$dynamicRef` evaluate the already-compiled target schema, so the
  target resource retains its own format behavior.
- `Compile` and `CompileBatch` resolve the same behavior. Batch compilation
  makes supplied resources visible in a private compilation view before resolving
  their vocabularies, so a schema can select a meta-schema in the same batch.
  The public registry receives only fully initialized and bound graphs.

### Annotation output

- Evaluating a schema object that contains `format` records the schema's format
  value in `EvaluationResult.Annotations["format"]`.
- The annotation is present in annotation, best-effort, and vocabulary assertion
  modes, including when format assertion fails.

### Standard format syntax

- Every Draft 2020-12 standard format has a syntactic validator and ignores
  non-string instances.
- `email` and `idn-email` perform the minimal syntactic validation permitted by
  JSON Schema Validation section 7.2.2. Mailbox structure is parsed with
  `net/mail`; domains are checked as ASCII hostnames or with `x/net/idna`.
- IDN hostname validation applies the `x/net/idna` registration, DNS-length,
  and bidirectional checks. It does not claim exhaustive RFC 5892 ContextO or
  derived-property enforcement.
- URI and IRI validation use distinct character repertoires. IRI validation
  admits RFC 3987 `ucschar` values and admits `iprivate` values only in query
  components.
- URI Templates are parsed as RFC 6570 Level 4 templates.
- The `regex` format accepts the interoperable expression subset listed by JSON
  Schema Core section 6.4. Syntax is parsed with `regexp/syntax` in Perl mode,
  which may also accept Go/RE2 extensions. Full ECMA-262 syntax, including
  lookaround and backreferences, is not supported.

## Prior Decisions

- **Decision**: Keep caller policy separate from dialect requirements. **Why**:
  compiler settings are mutable shared policy, while vocabulary behavior belongs
  to a compiled schema resource. **Rejected**: mutating `Compiler.AssertFormat`
  while compiling a schema.
- **Decision**: Reject unknown formats at compilation in vocabulary mode. **Why**:
  a schema with an unknown required assertion cannot be processed correctly.
  **Rejected**: silently annotating or waiting for instance evaluation.
- **Decision**: Keep one evaluation path. **Why**: JSON, map, and struct entry
  points must not drift. **Rejected**: separate format engines per input type.
- **Decision**: Compose maintained standards implementations for specialized
  grammars and use the standard library for the interoperable regex subset.
  **Why**: IDNA processing and URI Template syntax require dedicated parsers,
  while full ECMA-262 regex support is recommended but not required. The regex
  validator normalizes counted repetitions only when `regexp/syntax` reaches
  its implementation limit, then reparses with the same standard parser.
  **Rejected**: private Unicode tables and embedding a JavaScript runtime solely
  for full ECMA-262 parsing.

## Forbidden

- Do not make standard Draft 2020-12 assert formats by default.
- Do not interpret `$vocabulary` boolean values as vocabulary on/off switches.
- Do not allow caller settings to disable vocabulary-required assertions.
- Do not add synthetic always-valid entries to `Formats` for unknown names.
- Do not duplicate format validation across public validation entry points.
- Do not claim Format-Assertion support without syntactic validators for every
  standard Draft 2020-12 format.
