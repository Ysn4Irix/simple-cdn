<!-- antislop:start -->
## antislop
For UI, copy, people, mobile layout, or code comments work, read these installed skill files directly (use these paths even if a same-named global skill exists):
- Core filter, always on: `antislop`: `.agents/skills/antislop/SKILL.md`
- UI / visual: `antislop-ui`: `.agents/skills/antislop-ui/SKILL.md`
- Copy & text: `antislop-copywriting`: `.agents/skills/antislop-copywriting/SKILL.md`
- People: `antislop-human`: `.agents/skills/antislop-human/SKILL.md`
- Mobile / responsive: `antislop-layoutmobile`: `.agents/skills/antislop-layoutmobile/SKILL.md`
- Code comments: `antislop-code`: `.agents/skills/antislop-code/SKILL.md`
Before starting, follow the core's "Two Usage Modes" section in strict order: explicit session instruction first, then global preference, then ask. A session instruction always wins. For a resolved mode, say `antislop active: <mode> (session override).` or `antislop active: <mode> (global preference).` once before presenting findings or making edits, using the actual mode and source. Acknowledging the user's request without naming the source does not replace this notice.
Only an explicit choice of antislop during or after selects a session mode. A request to review, audit, or avoid file edits does not select a mode; read the global preference in that case. Another skill's mode does not select antislop's mode.
If the mode is unresolved, ask during/after and end the response; wait for the answer before any UI review, planning, or concept. For read-only tasks, put the active-mode notice only at the start of the final answer, never in progress messages. For editing tasks, announce before the first edit and omit it from the final answer.
To update antislop later: `npx antislop-ai --update`, or run `npx antislop-ai` and pick Overwrite them.
<!-- antislop:end -->

<!-- ag-init:rule:github-cicd:start -->
## GitHub CI/CD & Release Management

### Repository Setup
- Initialize a Git repository named after the project.
- Configure the appropriate remote repository URL.
- Follow Git best practices and maintain a clean, consistent commit history.

### CI/CD Automation
Implement production-ready GitHub Actions workflows that:

- Automate linting, formatting checks, tests, and production builds.
- Run validation checks on pull requests and pushes.
- Include relevant security and dependency vulnerability checks.
- Use caching to optimize workflow execution.
- Apply least-privilege permissions and secure secret management.
- Fail fast when critical checks fail.
- Publish tagged releases with changelogs and build artifacts.

**Requirements:**
- Never commit secrets, credentials, or sensitive configuration.
- Never bypass CI checks or ignore failing validation.
- Never publish releases containing broken builds.
- Document the CI/CD pipeline and release process in `README.md`.

### Git & Versioning Policy

Follow **Conventional Commits** and **Semantic Versioning** (`MAJOR.MINOR.PATCH`).

Apply version increments according to the following rules:

| Change Type | Version Bump | Examples |
|---|---|---|
| Breaking changes | **MAJOR** | Incompatible API or architecture changes |
| New features | **MINOR** | New functionality or capabilities |
| Fixes and other changes | **PATCH** | Bug fixes, refactoring, documentation, and minor edits |

For every release:
- Update the project version appropriately.
- Update the changelog with relevant changes.
- Ensure all required CI checks pass before creating a release.
- Use meaningful, descriptive Conventional Commits.

### Remote Operations & Authorization

- Commit changes locally when appropriate.
- **Never push, publish, or synchronize changes to any remote repository without explicit user authorization.**
- Never create or publish a remote release without explicit user authorization.
- Local commits and local repository maintenance are permitted without additional approval.
<!-- ag-init:rule:github-cicd:end -->

<!-- ag-init:rule:engineering-standards:start -->
Project-wide instructions for AI coding agents. These rules apply unless a more specific, applicable instruction in a subdirectory overrides them. Follow repository conventions and the project's actual technology stack; do not assume tools or capabilities that are not present.

## 1. Core Principles

- Prioritize correctness, security, maintainability, accessibility, and a professional user experience.
- Understand the project before changing it. Prefer the smallest reliable change that fully addresses the task.
- Preserve existing behavior unless the requested change explicitly requires otherwise.
- Do not invent requirements, functionality, APIs, configuration, or successful test results.
- Do not perform destructive, externally visible, or irreversible actions without explicit user authorization.
- If instructions conflict, follow the most specific applicable instruction; ask when the conflict materially affects the result.

## 2. Project Understanding

- Before implementation, inspect the project structure, `README.md`, relevant documentation, configuration, dependencies, and affected code.
- Identify the framework, language, package manager, architecture, existing patterns, and available scripts from the repository itself.
- Check for existing utilities, components, services, and tests before creating new ones.
- Follow established project conventions unless they are demonstrably unsafe or unsuitable.
- Do not rewrite working systems, migrate frameworks, or introduce architectural changes without a clear requirement and explanation.

## 3. Development Workflow

1. Understand the request, constraints, and expected behavior.
2. Inspect the relevant implementation and identify affected areas.
3. For substantial work, outline a concise implementation plan before editing.
4. Implement in focused, coherent changes.
5. Verify the result using the relevant tests and quality checks.
6. Update affected documentation and report the outcome.

- Keep changes scoped to the task; avoid unrelated cleanup and formatting churn.
- Preserve compatibility with existing data, interfaces, and integrations unless a breaking change is requested.
- Consider loading, empty, success, validation, error, and edge states where relevant.
- Ask a concise clarifying question only when a missing decision materially blocks safe implementation. Otherwise, use a conservative, clearly stated assumption.
- Do not stop at a visual mock-up or placeholder when working functionality was requested.

## 4. Architecture & Code Quality

- Write clear, modular, maintainable, and reusable code.
- Keep responsibilities separated and functions/components focused.
- Apply SOLID principles and established design patterns where they improve the solution; avoid unnecessary abstraction.
- Prefer readable code over clever or excessively compressed implementations.
- Use consistent naming, formatting, and conventions already established by the project.
- Use the strongest practical typing supported by the stack. Avoid `any`, unsafe casts, broad suppressions, and disabled checks unless justified and documented.
- Avoid duplicated logic, oversized modules, hidden side effects, and unnecessary global state.
- Reuse existing abstractions when appropriate; do not create generic frameworks for one-off needs.
- Remove unused code, temporary files, debug statements, and abandoned implementations introduced by the task.
- Do not leave TODOs, stubs, fake data, or nonfunctional controls in completed production features unless explicitly requested and clearly identified.

## 5. Testing & Quality Assurance

- Add or update tests when behavior changes; prioritize meaningful coverage of business logic and critical user flows.
- Cover relevant boundary conditions, invalid input, failure paths, and permission cases.
- Run the applicable tests, linting, formatting checks, type checks, and production build when available and relevant.
- Fix regressions and issues introduced by the changes.
- Do not weaken, delete, skip, or bypass tests or quality checks merely to obtain a passing result.
- Never claim a check passed unless it was actually executed and passed.
- If a check cannot be run, state what was not run and why.
- Do not claim complete verification based only on a successful build.

## 6. Security & Privacy

- Never hardcode credentials, API keys, tokens, private certificates, or other secrets.
- Use environment variables or the project's approved secret-management mechanism; update example configuration without including real secrets.
- Validate untrusted input at appropriate boundaries and enforce authorization on the server, not only in the client.
- Apply least privilege and secure defaults.
- Protect personal, confidential, and business data in storage, transport, logs, exports, and error responses.
- Avoid exposing stack traces, secrets, internal paths, or sensitive implementation details to end users.
- Use parameterized database operations and safe output encoding appropriate to the stack.
- Review authentication, session handling, access control, file handling, and external integrations for security implications.
- Do not introduce telemetry, advertising trackers, or data sharing beyond the project's documented requirements.
- Do not run destructive database, filesystem, infrastructure, or account operations without explicit authorization.

## 7. UI/UX & Accessibility

- Follow the project's established design system. Keep typography, spacing, colors, icons, and component behavior consistent.
- Build responsive interfaces appropriate to the product's target devices; prioritize mobile usability when the project is mobile-first.
- Use semantic structure, accessible names, keyboard navigation, visible focus, and sufficient contrast.
- Do not rely on color alone to communicate meaning.
- Provide appropriate loading, empty, success, validation, disabled, and error states.
- Make controls behave predictably and provide clear feedback for user actions.
- Avoid unnecessary decoration, visual clutter, excessive cards, gradients, and inconsistent component styles.
- Do not add decorative badges, pills, chips, or tag-shaped labels. Use them only for meaningful dynamic statuses, alerts, counts, or genuine selectable interactions.
- During UI reviews, identify and remove decorative UI that does not provide clear user value.
- Follow platform conventions and respect reduced-motion preferences where applicable.
- Never wrap ordinary metadata, technology names, database types, framework names, feature descriptions, or informational labels in a colored or rounded container.
- Render this information as plain text, subtle secondary text, or within an existing content row.
- Use badges only for meaningful dynamic statuses, alerts, counts, or other cases where visual emphasis provides clear user value.
- Use interactive chips only for genuine user interactions, such as selectable filters.
- Do not introduce new badge variants or badge components merely to decorate a screen.
- Apply this rule consistently across all screens, components, dialogs, cards, and future features.
- When uncertain, choose plain text over a badge.
- During UI reviews, actively identify and remove unnecessary badges from both new and existing interfaces.

## 8. Error Handling, Logging & Reliability

- Handle errors explicitly and consistently; never silently swallow meaningful failures.
- Show users actionable, appropriately worded messages without exposing sensitive details.
- Use the project's logging conventions and structured logging where supported.
- Do not log credentials, tokens, personal data, or sensitive request payloads.
- Handle asynchronous work, cancellation, retries, timeouts, and resource cleanup appropriately.
- Avoid retry behavior that can duplicate non-idempotent operations.
- Preserve data integrity and provide safe recovery behavior where relevant.

## 9. Dependencies & Configuration

- Prefer existing dependencies and native platform capabilities.
- Add a dependency only when its value justifies its maintenance, security, licensing, and bundle/runtime costs.
- Check compatibility, maintenance status, license, and known security concerns before adding or upgrading packages.
- Use the repository's existing package manager and lockfile. Do not mix package managers.
- Avoid unnecessary dependency upgrades and do not modify lockfiles by hand.
- Keep configuration consistent across development, test, and production environments.
- Document new environment variables and provide safe example values; never commit actual secrets.

## 10. Performance

- Avoid unnecessary rendering, computation, network requests, database queries, and large dependencies.
- Use pagination, caching, batching, and lazy loading when appropriate to actual usage and architecture.
- Consider performance on the project's target devices and network conditions.
- Measure or identify a concrete bottleneck before introducing complex optimization.
- Prioritize correctness and maintainability; avoid premature optimization.

## 11. Documentation

- Keep documentation accurate and synchronized with implemented behavior.
- Update setup, usage, configuration, architecture, and troubleshooting documentation when affected.
- Document non-obvious decisions, important constraints, and operational procedures.
- Never describe planned or placeholder functionality as implemented.
- Keep Markdown documentation in `docs/`, except essential root-level files required by tools or platforms, including `README.md` and `AGENTS.md` `CHANGELOG.md` etc..
- Use the repository's existing documentation naming and organization conventions where possible.

## 12. Definition of Done

A task is complete when, as applicable:

- The requested behavior is implemented and matches the requirements.
- Existing functionality and data compatibility are preserved.
- Relevant tests and quality checks have been run and their results are known.
- Security, accessibility, responsive behavior, and error states have been considered.
- Related documentation and configuration examples are updated.
- No unintended files, temporary code, debug output, or unrelated changes remain.
- Remaining limitations, assumptions, and follow-up work are clearly disclosed.

## 13. Agent Completion Report

At the end of implementation work, provide a concise report containing:

- **Implemented:** What changed and why.
- **Files:** The important files created or modified.
- **Verification:** Checks actually run and their results.
- **Limitations:** Known issues, checks not run, and remaining work.
- **Git:** Local commit hash/message if a commit was created; explicitly state that no remote push or publication was performed.
<!-- ag-init:rule:engineering-standards:end -->

<!-- ag-init:rule:language:start -->
INPUT_LANGUAGE is english
OUTPUT_LANGUAGE is english
<!-- ag-init:rule:language:end -->
