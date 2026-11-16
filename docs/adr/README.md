# ADRs

Architecture Decision Records document decisions about the codebase and its architecture.

## Naming

Each ADR has a three-digit number, and the `name` field in its frontmatter matches its filename, `NNN-Title.md`.

For example, `name: 012-Monorepo` is the file `012-Monorepo.md`, and its heading is `# ADR 012: Monorepo`.

## Metadata

Each ADR starts with YAML frontmatter containing:
- `name`, `title`, and `created`
- a link to its PR
- the ADRs it `depends_on` or `supersedes`

## Lifecycle

An ADR is accepted when its PR is merged. From then on it is frozen. A later change is written as a new ADR that supersedes it.

## Sections

The template, [`000-template.md`](000-template.md), lays out these sections:

- **Context** - The Background section on why the ADR exists and what the current state is, followed by an optional Glossary section defining new language.
- **Proposal** - What is being proposed, split into whatever parts suit the subject. Its last part lists any open questions.
- **Alternatives** - Other options that were considered, and why they weren't chosen.
