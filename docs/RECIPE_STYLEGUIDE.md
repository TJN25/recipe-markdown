
# Recipe Markdown Style Guide

This vault uses a consistent Markdown structure for recipe files. The goal is to normalize formatting without changing recipe content.

??? note "Core Rules"

    - Do not add recipe content that is not already in the file.
    - Do not remove recipe content, notes, commentary, ratings, or history.
    - Do not invent metadata, sections, components, or phases.
    - Normalize Markdown structure only when the structure is already clearly present.

??? note "File Shape"

    When the content exists in the file, use this order:

    1. Existing frontmatter
    2. `# Title`
    3. Metadata block
    4. Major sections as `##`
    5. Existing subsections/components as `###`

    Do not add missing parts just to match this order.

??? note "Titles"

    - Use a single `# Title`.
    - Remove decorative bold/italic formatting from titles.
    - Keep the existing title wording.
    - If a file has no clear title use the file name (capitalise and swap - for space)

??? note "Metadata"

    When servings/makes/prep/cook/total metadata already exists, normalize it to bold label/value lines:

    - `**Serves:** 2`
    - `**Makes:** 300-350ml`
    - `**Prep:** 15 minutes`
    - `**Cook:** 25 minutes`
    - `**Total:** 40 minutes`

    Rules:

    - Standardize metadata when it exists.
    - Do not add missing metadata.
    - Keep metadata as inline bold label/value lines, not headings.
    - If timing text exists without a standard label, preserve the text unless the intended label is obvious.

??? note "Headings"

    Use Markdown headings for existing section labels.

    Major sections become `##`:

    - Ingredients
    - Method
    - Instructions
    - Notes
    - Storage
    - Tips
    - Timeline
    - To Serve
    - Baking
    - Preparation
    - Cooking

    Existing grouped subsections/components become `###`:

    - Sauce
    - Filling
    - Per serving
    - Stage 1 (Chili Prep)
    - Main Components
    - Prep Phase
    - Cook Phase
    - Serve

    Rules:

    - Convert heading-like bold/italic labels into Markdown headings.
    - Do not keep bold/italic as the main section-heading format.
    - Do not promote ordinary sentences into headings.

??? note "Lists"

    - Ingredient lists use `-`.
    - Method steps use numbered lists when the content is clearly sequential.
    - Supporting lists under a subsection may stay as `-` when the original content is a grouped checklist or grouped item list rather than a step-by-step sequence.
    - Normalize mixed bullet symbols like `•`, `◦`, and `▪` to `-`.
    - Keep existing nested content only when it is part of the recipe structure already present in the file.

??? note "Spacing"

    - Use a blank line after frontmatter.
    - Use a blank line after headings.
    - Use a blank line between metadata lines and the next section.
    - Do not leave random empty lines between every list item.
    - Remove empty headings and stray Markdown artifacts.

??? note "Safe Fixes"

    These are always safe when clearly applicable:

    - `##` by itself becomes `## Method` when it clearly introduces method steps.
    - `**Ingredients:**` becomes `## Ingredients`.
    - `*Stage 1 (Chili Prep)*` becomes `### Stage 1 (Chili Prep)`.
    - `*Serves 2*` becomes `**Serves:** 2`.
    - Bullet markers normalize to `-`.

??? note "Non-Goals"

    - No rewriting for tone, grammar, or clarity.
    - No standardizing ingredient wording beyond Markdown structure.
    - No forced conversion of every file into the same complexity.
    - No automatic assumptions about ambiguous content.
