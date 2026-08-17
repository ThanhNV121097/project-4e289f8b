# Design System — Task board for one person

> Source of truth: approved `index.html`.
> Every value below is extracted from it. Changing a value here without changing approved design is defect.

Last updated: 2026-08-17

## 1. Foundations

### 1.1 Color

Semantic tokens. Name by job, never by hue.

| Token | Value | Used for |
|---|---|---|
| `--color-bg` | `#F8FAFC` | Page background, skeleton shimmer |
| `--color-surface` | `#FFFFFF` | Cards, panels, modal, form controls, nav menu |
| `--color-column-bg` | `#F1F5F9` | Board columns, secondary hover |
| `--color-muted-surface` | `#EEF2FF` | Navigation hover |
| `--color-info-surface` | `#EFF6FF` | Persistence notice |
| `--color-warning-surface` | `#FFFBEB` | Doing status pill |
| `--color-success-surface` | `#ECFDF5` | Done status pill |
| `--color-border` | `#E5E7EB` | Default border, dividers |
| `--color-border-strong` | `#CBD5E1` | Task hover border, dashed empty border |
| `--color-info-border` | `#BFDBFE` | Notice border |
| `--color-warning-border` | `#FDE68A` | Doing pill border |
| `--color-success-border` | `#BBF7D0` | Done pill border |
| `--color-error-border` | `#FCA5A5` | Invalid input border |
| `--color-text` | `#0F172A` | Body text, toast background |
| `--color-label` | `#334155` | Field labels |
| `--color-text-muted` | `#64748B` | Secondary text, captions, todo status dot |
| `--color-muted-icon` | `#94A3B8` | Empty-state icon |
| `--color-primary` | `#2563EB` | Primary action, brand mark stroke, status marker |
| `--color-primary-hover` | `#1D4ED8` | Primary action hover |
| `--color-primary-link-hover` | `#1E40AF` | Nav hover text |
| `--color-primary-text` | `#FFFFFF` | Text on primary and toast |
| `--color-info-text` | `#1E3A8A` | Notice text |
| `--color-success` | `#059669` | Done status dot |
| `--color-success-text` | `#065F46` | Done status pill text |
| `--color-warning` | `#D97706` | Doing status dot |
| `--color-warning-text` | `#92400E` | Doing status pill text |
| `--color-muted-status-text` | `#475569` | Todo status pill text |
| `--color-danger` | `#DC2626` | Delete action, validation message |
| `--color-focus` | `#93C5FD` | Focus ring |
| `--color-backdrop` | `rgba(15,23,42,.42)` | Modal backdrop |
| `--color-topbar-bg` | `rgba(248,250,252,.92)` | Sticky header background |

#### Contrast audit

Every text-on-background pair actually used. Body text ≥ 4.5:1, large text ≥ 3:1, UI borders ≥ 3:1.

| Foreground | Background | Ratio | Passes |
|---|---|---|---|
| `--color-text` `#0F172A` | `--color-bg` `#F8FAFC` | `16.6:1` | AA |
| `--color-text` `#0F172A` | `--color-surface` `#FFFFFF` | `17.8:1` | AA |
| `--color-text-muted` `#64748B` | `--color-bg` `#F8FAFC` | `4.5:1` | AA |
| `--color-text-muted` `#64748B` | `--color-surface` `#FFFFFF` | `4.8:1` | AA |
| `--color-label` `#334155` | `--color-surface` `#FFFFFF` | `10.4:1` | AA |
| `--color-primary-text` `#FFFFFF` | `--color-primary` `#2563EB` | `5.2:1` | AA |
| `--color-primary-text` `#FFFFFF` | `--color-primary-hover` `#1D4ED8` | `6.7:1` | AA |
| `--color-primary-link-hover` `#1E40AF` | `--color-muted-surface` `#EEF2FF` | `8.1:1` | AA |
| `--color-info-text` `#1E3A8A` | `--color-info-surface` `#EFF6FF` | `10.6:1` | AA |
| `--color-warning-text` `#92400E` | `--color-warning-surface` `#FFFBEB` | `7.4:1` | AA |
| `--color-success-text` `#065F46` | `--color-success-surface` `#ECFDF5` | `7.0:1` | AA |
| `--color-muted-status-text` `#475569` | `--color-bg` `#F8FAFC` | `7.1:1` | AA |
| `--color-danger` `#DC2626` | `--color-surface` `#FFFFFF` | `4.8:1` | AA |
| `--color-primary-text` `#FFFFFF` | `--color-text` `#0F172A` | `17.8:1` | AA |
| `--color-muted-icon` `#94A3B8` | `--color-surface` `#FFFFFF` | `2.6:1` | FAIL for text; icon-only decorative use |
| `--color-border` `#E5E7EB` | `--color-surface` `#FFFFFF` | `1.2:1` | FAIL for UI boundary contrast |
| `--color-border-strong` `#CBD5E1` | `--color-surface` `#FFFFFF` | `1.5:1` | FAIL for UI boundary contrast |

### 1.2 Spacing

Base unit: `2px`. Most spacing uses 4px multiples; `6px`, `10px`, `14px`, `18px`, `20px`, `22px`, `28px`, `56px` also appear in approved design.

| Token | Value |
|---|---|
| `--space-0` | `0` |
| `--space-1` | `3px` |
| `--space-2` | `4px` |
| `--space-3` | `5px` |
| `--space-4` | `6px` |
| `--space-5` | `7px` |
| `--space-6` | `8px` |
| `--space-7` | `10px` |
| `--space-8` | `12px` |
| `--space-9` | `14px` |
| `--space-10` | `16px` |
| `--space-11` | `18px` |
| `--space-12` | `20px` |
| `--space-13` | `22px` |
| `--space-14` | `24px` |
| `--space-15` | `26px` |
| `--space-16` | `28px` |
| `--space-17` | `32px` |
| `--space-18` | `56px` |

### 1.3 Typography

Font families, as loaded by CSS stack only:

- Body: `Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif`
- Headings: same as body
- Mono: none used

| Token | Size | Line height | Weight | Used for |
|---|---|---|---|---|
| `--text-xs` | `12px` | normal | `400`, `700` | Count pill, meta, labels, validation |
| `--text-sm` | `13px` | `1.45` or normal | `400`, `700` | Eyebrow, task description |
| `--text-base-sm` | `14px` | normal | `400`, `650` | Nav, notice, panel helper, empty state, buttons |
| `--text-base` | `15px` | `1.3` | browser default bold | Task card h3 |
| `--text-lg` | `18px` | normal | browser default bold | Panel h2, modal h2 |
| `--text-2xl` | `24px` | normal | `750` | State totals |
| `--text-3xl` | `32px` | `1.1` | browser default bold | Page h1 |

Heading levels used in order: page `h1`, form/modal `h2`, task card `h3`.

### 1.4 Radius, border, shadow, motion

| Token | Value | Used for |
|---|---|---|
| `--radius-sm` | `8px` | Brand mark, nav buttons, menu button, mini buttons |
| `--radius-md` | `10px` | Primary/secondary buttons, inputs |
| `--radius-lg` | `12px` | Notice, nav menu, toast |
| `--radius-xl` | `14px` | State cards, task cards, empty states, loading/error panels |
| `--radius-2xl` | `16px` | Board columns, create panel |
| `--radius-modal` | `18px` | Modal |
| `--radius-full` | `999px` | Count pills, status pills, skeleton lines |
| `--border-width` | `1px` | Default border |
| `--focus-width` | `3px` | Keyboard focus outline |
| `--shadow-sm` | `0 1px 1px rgba(15,23,42,.03)` | State cards |
| `--shadow-card` | `0 1px 2px rgba(15,23,42,.04)` | Task cards resting |
| `--shadow-md` | `0 1px 2px rgba(15,23,42,.06), 0 8px 24px rgba(15,23,42,.06)` | Primary button, panels, hover task, toast, mobile nav |
| `--shadow-lg` | `0 24px 80px rgba(15,23,42,.25)` | Modal |
| `--duration-fast` | `.14s` | Modal pop animation |
| `--duration-base` | `.16s` | Button/task hover transitions |
| `--duration-toast` | `.18s` | Toast show/hide |
| `--duration-loading` | `1s` | Skeleton shimmer |
| `--easing` | `ease-out` for modal pop; default CSS easing otherwise | Transitions and animation |

Motion currently has no `prefers-reduced-motion: reduce` override.

### 1.5 Layout and breakpoints

| Name | Min width | Container | Columns | Gutter |
|---|---|---|---|---|
| `base` | `0` | `100%` with `20px` side padding | 1 column below `900px` | `12px` form/state, `14px` board |
| `md` | `900px` | `1180px` max | 3 board columns, 3 state cards, 5 form columns | `12px` form/state, `14px` board |

Z-index scale, extracted from CSS:

| Layer | Value |
|---|---|
| Base | `0` |
| Sticky header | `10` |
| Modal backdrop | `20` |
| Toast | `30` |
| Dropdown | Uses sticky header stacking context, no explicit value |
| Modal | Inherits modal backdrop layer |

## 2. Components

One subsection per reusable component. Every component lists all states.

### 2.1 Topbar

**Purpose** — Persistent page identity and in-page navigation. Use once per screen; do not use for account or workspace controls because scope excludes accounts and multiple boards.

**Anatomy** — `[brand mark] [product name] [menu button on mobile] [Board link] [Create task link] [States link]`.

**Variants**

| Variant | Tokens | When to use |
|---|---|---|
| Sticky desktop | `--color-topbar-bg`, `--color-border`, `--space-9`, `--space-12` | Width above `900px` |
| Sticky mobile | `--color-topbar-bg`, `--color-surface`, `--shadow-md`, `--radius-lg` | Width at or below `900px` |

**Sizes**

| Size | Height | Padding | Text token |
|---|---|---|---|
| Default | Content-driven | `14px 20px` | `--text-base-sm` |
| Brand mark | `28px` | none | icon stroke |

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | Sticky translucent surface with bottom border | `--color-topbar-bg`, `--color-border` |
| Hover | Nav item gets pale primary surface and darker primary text | `--color-muted-surface`, `--color-primary-link-hover` |
| Focus (keyboard) | 3px focus outline with 2px offset | `--color-focus`, `--focus-width` |
| Active / pressed | No distinct pressed style in approved design | none |
| Disabled | Not used; do not render disabled nav links | none |
| Loading | Not used; topbar remains stable while board loads | none |
| Error | Not used; errors render in error-state panel | none |
| Empty | Not applicable; brand and navigation always present | none |

**Accessibility** — `nav` uses `aria-label="Page navigation"`. Mobile menu button uses `aria-expanded` and `aria-controls`. Links are keyboard reachable with visible focus. Menu button hit target is close to 44px through padding and text; icon brand is decorative.

### 2.2 Button

**Purpose** — Trigger task actions. Use primary for main create/save actions, secondary for neutral actions, ghost for low-emphasis header/modal actions, mini for card-level actions, danger mini for delete.

**Anatomy** — `[label]`.

**Variants**

| Variant | Tokens | When to use |
|---|---|---|
| Primary | `--color-primary`, `--color-primary-hover`, `--color-primary-text`, `--radius-md`, `--shadow-md` | Create task, add task, save changes |
| Secondary | `--color-surface`, `--color-text`, `--color-border`, `--radius-md` | Cancel, retry, neutral secondary action |
| Ghost | transparent background, `--color-text-muted`, `--radius-sm` | Close modal, low-emphasis nav-adjacent action |
| Mini | `--color-surface`, `--color-text`, `--color-border`, `--radius-sm` | Task-card actions |
| Danger mini | `--color-danger` text on mini button | Delete task |

**Sizes**

| Size | Height | Padding | Text token |
|---|---|---|---|
| Default | Content-driven, about `42px` | `10px 14px` | `--text-base-sm` |
| Ghost/nav | Content-driven, about `36px` | `8px 10px` | `--text-base-sm` |
| Mini | Content-driven, about `30px` | `6px 8px` | `--text-xs` |

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | Variant fill, border, radius | variant tokens |
| Hover | Primary darkens and moves up `-1px`; secondary/mini gets muted background; ghost/nav gets primary-tinted surface | `--color-primary-hover`, `--color-column-bg`, `--color-muted-surface` |
| Focus (keyboard) | 3px focus outline with 2px offset | `--color-focus`, `--focus-width` |
| Active / pressed | No distinct pressed style in approved design | none |
| Disabled | Not shown in approved design; if needed, keep non-interactive with muted text and no transform | `--color-text-muted`, `--color-border` |
| Loading | Not shown on buttons; loading appears as page-level loading panel | none |
| Error | Danger variant uses danger text; validation errors render near fields | `--color-danger` |
| Empty | Not applicable to button itself | none |

**Accessibility** — Use native `button`. Minimum hit target should be 44×44px where possible; mini buttons are below target in approved design and must only be used inside spacious card actions. Icon-only buttons are not present. Focus outline must remain visible.

### 2.3 Hero

**Purpose** — State one-screen task-board purpose and expose primary create action. Use once at top of screen.

**Anatomy** — `[eyebrow] [h1] [description] [primary action] [secondary action]`.

**Variants**

| Variant | Tokens | When to use |
|---|---|---|
| Desktop split | `--space-14`, `--space-12`, `--text-3xl` | Width above `900px` |
| Mobile stacked | `--space-10`, `--text-3xl` | Width at or below `900px` |

**Sizes**

| Size | Height | Padding | Text token |
|---|---|---|---|
| Default | Content-driven | bottom `20px` | `--text-3xl`, `--text-sm`, browser body text |

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | Left-aligned product copy, actions aligned end on desktop | `--color-text`, `--color-text-muted`, `--color-primary` |
| Hover | Only contained buttons change | Button tokens |
| Focus (keyboard) | Button focus only | `--color-focus` |
| Active / pressed | No hero-level pressed style | none |
| Disabled | Not applicable | none |
| Loading | Hero remains visible while loading panel appears below | none |
| Error | Hero remains visible while error panel appears below | none |
| Empty | Not applicable | none |

**Accessibility** — Section labelled by page `h1`. Copy must stay factual and scope-bound: one board, one table, one screen, no accounts.

### 2.4 Notice

**Purpose** — Highlight persistence rule or other system rule. Do not use as marketing banner.

**Anatomy** — `[status dot] [bold label] [message]`.

**Variants**

| Variant | Tokens | When to use |
|---|---|---|
| Info | `--color-info-surface`, `--color-info-border`, `--color-info-text`, `--color-primary` | Persistence/API rule |

**Sizes**

| Size | Height | Padding | Text token |
|---|---|---|---|
| Default | Content-driven | `12px 14px` | `--text-base-sm` |

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | Blue-tinted surface, border, dot | notice tokens |
| Hover | None; not interactive | none |
| Focus (keyboard) | None; not focusable | none |
| Active / pressed | None | none |
| Disabled | Not applicable | none |
| Loading | Not used | none |
| Error | Use ErrorState component instead | none |
| Empty | Do not render empty notice | none |

**Accessibility** — Plain content, no role needed unless message becomes time-sensitive. Dot is decorative.

### 2.5 LoadingState

**Purpose** — Show API fetch in progress without showing stale browser data.

**Anatomy** — `[strong loading text] [skeleton line] [skeleton line]`.

**Variants**

| Variant | Tokens | When to use |
|---|---|---|
| Page loading panel | `--color-surface`, `--color-border`, `--radius-xl`, `--color-bg`, `--duration-loading` | Initial task fetch and retry fetch |

**Sizes**

| Size | Height | Padding | Text token |
|---|---|---|---|
| Default | Content-driven | `16px` | browser body text |
| Skeleton line | `14px` | none | none |

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | Hidden with `display:none` until needed | none |
| Hover | None | none |
| Focus (keyboard) | None; not interactive | none |
| Active / pressed | None | none |
| Disabled | Not applicable | none |
| Loading | Panel displays; board opacity reduces to `.35`; skeleton shimmer runs | `--color-surface`, `--color-border`, `--duration-loading` |
| Error | Hide and show ErrorState instead | none |
| Empty | Not applicable | none |

**Accessibility** — Uses `aria-live="polite"`. Loading copy must say source is API.

### 2.6 ErrorState

**Purpose** — Explain API load failure and offer retry; never show stale task data as substitute.

**Anatomy** — `[strong error title] [helper text] [retry button]`.

**Variants**

| Variant | Tokens | When to use |
|---|---|---|
| API load error | `--color-surface`, `--color-border`, `--radius-xl`, `--color-text-muted` | Failed task fetch |

**Sizes**

| Size | Height | Padding | Text token |
|---|---|---|---|
| Default | Content-driven | `16px` | browser body text and `--text-base-sm` helper |

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | Hidden with `display:none` | none |
| Hover | Retry button hover only | Button tokens |
| Focus (keyboard) | Retry button focus | `--color-focus` |
| Active / pressed | No panel active state | none |
| Disabled | Retry button disabled state not shown in approved design | none |
| Loading | Hide while LoadingState shows | none |
| Error | Panel displays with retry action | `--color-surface`, `--color-border` |
| Empty | Not applicable | none |

**Accessibility** — `role="alert"`. Retry is native button. Message must be specific: cannot load tasks, retry avoids stale data.

### 2.7 StateSummaryCard

**Purpose** — Show totals for Todo, Doing, Done above board.

**Anatomy** — `[status label] [count]`.

**Variants**

| Variant | Tokens | When to use |
|---|---|---|
| Summary card | `--color-surface`, `--color-border`, `--radius-xl`, `--shadow-sm` | Task status totals |

**Sizes**

| Size | Height | Padding | Text token |
|---|---|---|---|
| Default | Content-driven | `14px` | `--text-sm`, `--text-2xl` |

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | White surface, small muted label, large count | `--color-surface`, `--color-text-muted`, `--color-text` |
| Hover | None; not interactive | none |
| Focus (keyboard) | None; not focusable | none |
| Active / pressed | None | none |
| Disabled | Not applicable | none |
| Loading | Count can remain previous value while LoadingState visible in mockup | none |
| Error | Count can remain previous value while ErrorState visible in mockup | none |
| Empty | Count shows `0` | `--color-text` |

**Accessibility** — Group in section with `aria-label="Task totals"`. Counts must update with board.

### 2.8 BoardColumn

**Purpose** — Group tasks by status. Use exactly three statuses: todo, doing, done.

**Anatomy** — `[column header: status dot + title + count] [task list or empty state]`.

**Variants**

| Variant | Tokens | When to use |
|---|---|---|
| Todo | `--color-text-muted` dot | Tasks not started |
| Doing | `--color-warning` dot | Tasks in progress |
| Done | `--color-success` dot | Finished tasks |

**Sizes**

| Size | Height | Padding | Text token |
|---|---|---|---|
| Desktop column | min `360px` | header `14px 14px 10px`, cards `12px` | browser body text, `--text-xs` count |
| Mobile column | min `360px` | same | same |

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | Muted column surface with border and rounded corners | `--color-column-bg`, `--color-border`, `--radius-2xl` |
| Hover | None at column level | none |
| Focus (keyboard) | No column focus; child controls focus | `--color-focus` |
| Active / pressed | None | none |
| Disabled | Not applicable | none |
| Loading | Board opacity reduces while loading | opacity `.35` |
| Error | Board can remain visible behind error panel in mockup | none |
| Empty | Renders EmptyState inside cards area | EmptyState tokens |

**Accessibility** — Board section uses `aria-label="Task board"`. Columns use visible headings through text. Do not add drag-only movement; move buttons keep keyboard path.

### 2.9 TaskCard

**Purpose** — Show one task and allow edit, move, delete. Use only for persisted task entity.

**Anatomy** — `[title] [description or No description] [due date] [status pill] [Move left] [Move right] [Edit] [Delete]`.

**Variants**

| Variant | Tokens | When to use |
|---|---|---|
| Todo task | `--color-muted-status-text`, `--color-bg`, `--color-border` | Status `todo` |
| Doing task | `--color-warning-text`, `--color-warning-surface`, `--color-warning-border` | Status `doing` |
| Done task | `--color-success-text`, `--color-success-surface`, `--color-success-border` | Status `done` |

**Sizes**

| Size | Height | Padding | Text token |
|---|---|---|---|
| Default | Content-driven | `12px` | `--text-base`, `--text-sm`, `--text-xs` |

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | White card, border, small shadow | `--color-surface`, `--color-border`, `--radius-xl`, `--shadow-card` |
| Hover | Moves up `-1px`, stronger shadow, stronger border | `--shadow-md`, `--color-border-strong`, `--duration-base` |
| Focus (keyboard) | Child buttons receive focus outline | `--color-focus` |
| Active / pressed | No card pressed state; action buttons handle press | none |
| Disabled | Not shown; if API action pending, disable child buttons only | Button disabled tokens |
| Loading | Not shown per-card; page loading panel handles fetch | none |
| Error | Not shown per-card; page error panel handles fetch, field errors handled in forms | none |
| Empty | Not rendered; EmptyState replaces missing list | EmptyState tokens |

**Accessibility** — Render as article or list item. Action buttons need visible text. Move buttons must no-op or be disabled at first/last status in implementation to avoid dead controls.

### 2.10 StatusPill

**Purpose** — Label task status or persistence note compactly.

**Anatomy** — `[status label]`.

**Variants**

| Variant | Tokens | When to use |
|---|---|---|
| Todo | `--color-muted-status-text`, `--color-bg`, `--color-border` | Todo tasks |
| Doing | `--color-warning-text`, `--color-warning-surface`, `--color-warning-border` | Doing tasks |
| Done | `--color-success-text`, `--color-success-surface`, `--color-success-border` | Done tasks |
| Neutral info | `--color-muted-status-text`, `--color-bg`, `--color-border` | “Saved through API” note |

**Sizes**

| Size | Height | Padding | Text token |
|---|---|---|---|
| Default | Content-driven | `3px 8px` | `--text-xs` |
| Count | Content-driven | `3px 7px` | `--text-xs` |

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | Full radius bordered pill | variant tokens, `--radius-full` |
| Hover | None; not interactive | none |
| Focus (keyboard) | None; not focusable | none |
| Active / pressed | None | none |
| Disabled | Not applicable | none |
| Loading | Count pill can show existing number while loading | none |
| Error | Not used for errors | none |
| Empty | Count pill shows `0` | `--color-text-muted`, `--color-surface` |

**Accessibility** — Text label is visible; do not encode status by color alone.

### 2.11 EmptyState

**Purpose** — Explain when a status column has no tasks.

**Anatomy** — `[decorative outline icon] [message]`.

**Variants**

| Variant | Tokens | When to use |
|---|---|---|
| Column empty | `--color-border-strong`, `--color-muted-icon`, `--color-text-muted`, `--radius-xl` | No tasks in a status |

**Sizes**

| Size | Height | Padding | Text token |
|---|---|---|---|
| Default | Content-driven | `22px` | `--text-base-sm` |
| Icon | `28px` | none | none |

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | Dashed border, translucent white surface, centered muted text | `--color-border-strong`, `--color-text-muted` |
| Hover | None | none |
| Focus (keyboard) | None; not focusable | none |
| Active / pressed | None | none |
| Disabled | Not applicable | none |
| Loading | Hidden/replaced while list loading if no data loaded | LoadingState tokens |
| Error | Hidden/replaced by ErrorState if load fails before data | ErrorState tokens |
| Empty | Shows “No {status} tasks.” | `--color-text-muted` |

**Accessibility** — Icon uses `aria-hidden="true"`. Message names missing status. Empty area is never blank.

### 2.12 CreateTaskPanel

**Purpose** — Create one task with title, optional description, optional due date, status.

**Anatomy** — `[panel header: title + helper + API pill] [title field] [description field] [due date field] [status select] [submit button]`.

**Variants**

| Variant | Tokens | When to use |
|---|---|---|
| Create panel | `--color-surface`, `--color-border`, `--radius-2xl`, `--shadow-md` | Bottom create form |

**Sizes**

| Size | Height | Padding | Text token |
|---|---|---|---|
| Header | Content-driven | `16px 18px` | `--text-lg`, `--text-base-sm` |
| Form desktop | Content-driven | `18px`, grid gap `12px` | field tokens |
| Form mobile | Content-driven | `18px`, one column | field tokens |

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | Raised white panel with divided header | `--color-surface`, `--color-border`, `--shadow-md` |
| Hover | Submit button hover only | Button tokens |
| Focus (keyboard) | Form controls and submit show focus ring | `--color-focus` |
| Active / pressed | Submit button active not distinct | none |
| Disabled | Not shown; use disabled button during API submit if needed | Button disabled tokens |
| Loading | Not shown for submit; page loading used for initial fetch only | none |
| Error | Invalid field border and message | `--color-error-border`, `--color-danger` |
| Empty | Empty form fields with placeholders | placeholder copy |

**Accessibility** — Form uses labels for every input. Title is required and validates before submit. Due date uses native date input. Status uses native select. Error message appears next to invalid title.

### 2.13 Field

**Purpose** — Capture task title, description, due date, and status.

**Anatomy** — `[label] [input | textarea | select] [error message?]`.

**Variants**

| Variant | Tokens | When to use |
|---|---|---|
| Text input | `--color-surface`, `--color-border`, `--radius-md` | Title |
| Textarea | same | Description |
| Date input | same | Due date |
| Select | same | Status |
| Invalid | `--color-error-border`, `--color-danger` | Required title missing |

**Sizes**

| Size | Height | Padding | Text token |
|---|---|---|---|
| Default input/select | Content-driven, about `42px` | `10px` | browser body text |
| Textarea | min `42px` | `10px` | browser body text |
| Label | Content-driven | bottom margin `6px` | `--text-xs` |
| Error | Content-driven | top margin `5px` | `--text-xs` |

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | White field, default border | `--color-surface`, `--color-border`, `--radius-md` |
| Hover | No hover style in approved design | none |
| Focus (keyboard) | 3px blue focus outline with offset | `--color-focus`, `--focus-width` |
| Active / pressed | Native text cursor/select behavior | none |
| Disabled | Not shown; use muted border/text if needed | `--color-text-muted`, `--color-border` |
| Loading | Not shown | none |
| Error | Border turns light red; error text shown | `--color-error-border`, `--color-danger` |
| Empty | Placeholder describes expected input | placeholder copy |

**Accessibility** — Every control has `label for`. Required title error text must be linked with `aria-describedby` in implementation. Do not remove native date/select semantics.

### 2.14 EditModal

**Purpose** — Edit title, description, due date, and status without leaving one-screen board.

**Anatomy** — `[backdrop] [dialog] [header title + close] [hidden id] [fields] [footer cancel + save]`.

**Variants**

| Variant | Tokens | When to use |
|---|---|---|
| Edit task modal | `--color-backdrop`, `--color-surface`, `--radius-modal`, `--shadow-lg` | Editing existing task |

**Sizes**

| Size | Height | Padding | Text token |
|---|---|---|---|
| Dialog | `min(560px, 100%)` width | header `16px 18px`, form `18px` | `--text-lg`, field tokens |
| Backdrop | viewport | `18px` | none |

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | Hidden with `display:none` | none |
| Hover | Close/cancel/save buttons hover | Button tokens |
| Focus (keyboard) | Focus moves to title input; controls show focus ring | `--color-focus` |
| Active / pressed | Button active not distinct | none |
| Disabled | Not shown; disable save during API request if needed | Button disabled tokens |
| Loading | Not shown; use disabled save if needed in implementation | none |
| Error | Invalid title field shows field error | Field invalid tokens |
| Empty | Fields can be empty except title; description/due date optional | Field tokens |

**Accessibility** — Dialog uses `role="dialog"`, `aria-modal="true"`, `aria-labelledby`. Escape closes modal. Backdrop click closes. Implementation should trap focus inside modal while open and restore focus to opener on close.

### 2.15 Toast

**Purpose** — Confirm REST API actions without blocking workflow.

**Anatomy** — `[message]`.

**Variants**

| Variant | Tokens | When to use |
|---|---|---|
| Status toast | `--color-text`, `--color-primary-text`, `--radius-lg`, `--shadow-md`, `--duration-toast` | Created, updated, deleted, moved, loaded confirmations |

**Sizes**

| Size | Height | Padding | Text token |
|---|---|---|---|
| Default | Content-driven | `12px 14px` | browser body text |

**States**

| State | Visual change | Tokens |
|---|---|---|
| Default | Fixed bottom-right, hidden, translated down `20px`, opacity `0` | `--color-text`, `--color-primary-text` |
| Hover | None | none |
| Focus (keyboard) | None; not focusable | none |
| Active / pressed | None | none |
| Disabled | Not applicable | none |
| Loading | Not used for loading | none |
| Error | Not used for blocking errors; use ErrorState | none |
| Empty | Do not show empty toast | none |
| Shown | Moves to resting position, opacity `1`, auto-hides after `2200ms` | `--duration-toast` |

**Accessibility** — `role="status"` and `aria-live="polite"`. Keep messages short and action-specific.

## 3. Content and formatting

- Voice and tone: restrained, factual work-tool copy; no marketing hype.
- Date format: native `input type="date"` stores ISO `YYYY-MM-DD`; displayed dates use browser locale with short month, numeric day, numeric year, e.g. `Aug 20, 2026` in English locale.
- Time format: none used.
- Number format: plain integer counts for status totals.
- Currency format: none used.
- Capitalization: sentence case for headings, labels, buttons, and messages; status labels capitalized as `Todo`, `Doing`, `Done`.
- Empty-state wording pattern: `No {status} tasks.`
- Error-message wording pattern: short problem sentence, then action if needed. Example: `Could not load tasks.` plus retry explanation.
- Validation wording: `Title is required.`
- Persistence wording: action confirmations end with `through REST API` when confirming saved data flow.

## 4. Known deviations

Places where approved design does not follow its own rules or anti-patterns in `references/ai-defaults.md`. Record, do not silently fix.

| Where | Deviation | Why it stands | Follow-up |
|---|---|---|---|
| `html` | `scroll-behavior:smooth` has no `prefers-reduced-motion: reduce` override | Approved mockup includes smooth scroll and no reduced-motion block | Add reduced-motion override during frontend implementation unless stakeholder rejects |
| Skeleton shimmer and modal pop | Animations have no reduced-motion override | Approved mockup includes `shine 1s infinite` and `pop .14s ease-out` | Add reduced-motion override during frontend implementation unless stakeholder rejects |
| Borders | Default and strong borders fail 3:1 UI boundary contrast on white | Approved mockup uses quiet low-contrast borders for restrained work-tool look | Review in accessibility pass if boundaries look unclear |
| Mini buttons | Approximate hit target is below 44×44px | Approved mockup uses dense card actions | Increase padding or provide alternate mobile controls during implementation if touch usability fails |
| Spacing scale | Uses several 2px-step values, not clean 4/8/12/16 scale only | Approved mockup uses compact tuning values | Keep extracted values; avoid adding new spacing values |
| `#EEF2FF` nav hover | Pale indigo-tinted hover resembles common AI default hue family | Approved mockup otherwise uses flat, restrained blue accents and no gradients | Keep as hover only; do not expand indigo palette |
| Mockup-only controls | “Simulate reload from API” button and “Persistence rule” notice appear in approved HTML, but project memory excludes them from build | Stakeholder/team decision supersedes mockup for implementation scope | Dev must not build these controls; browser reload remains persistence test |

AI default checks avoided by approved design:

- No decorative gradients.
- No purple/violet primary palette.
- No emoji iconography.
- No generic marketing feature grid; layout follows task board content.
- No filler lorem ipsum; task examples are product-relevant.
- Empty, loading, and error states are present.
- Focus states are visible through `:focus-visible` outline.
- No text over images.

## 5. Change log

| Date | Change | Design PR |
|---|---|---|
| 2026-08-17 | Initial design system extracted from approved `index.html` | This PR |
