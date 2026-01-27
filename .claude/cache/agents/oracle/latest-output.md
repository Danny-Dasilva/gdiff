# Research Report: Git Diff Tool Name Availability
Generated: 2026-01-27

## Summary

Most of the proposed names (rift, sift, flux, prism, glint, drift) are already taken by established projects. The "go-" prefixed variants and diff-go combinations are also largely occupied. However, several names remain viable with modifications, and there are creative alternatives that combine diff concepts with Go-related naming.

## Questions Answered

### Q1: Which names are already taken?

**Answer:** All six proposed names have significant existing projects:

| Name | Status | Taken By |
|------|--------|----------|
| **rift** | TAKEN | Multiple: Microsoft RIFT (Rust malware analysis), PipeRift/rift (programming language), openRIFT (file downloader), macOS tiling WM |
| **sift** | TAKEN | svent/sift - popular grep alternative written in Go (900+ stars), SIFT forensics workstation, multiple SIFT algorithm implementations |
| **flux** | TAKEN | FluxCD - major CNCF GitOps project (5k+ stars), Black Forest Labs FLUX (AI image generation) |
| **prism** | TAKEN | Stoplight Prism (API mocking), PrismLibrary (XAML framework), Primer Prism (color tool) |
| **glint** | TAKEN | brigand/glint (conventional commits tool), glint.info (Git GUI), typed-ember/glint (TypeScript tooling) |
| **drift** | TAKEN | simolus3/drift (Dart persistence library), snyk/driftctl (infrastructure drift), convidev-tud/driftool (git repository drift analysis) |

**Source:** GitHub searches and web search results
**Confidence:** High

### Q2: Which "go-" prefixed names are taken?

**Answer:**

| Name | Status | Taken By |
|------|--------|----------|
| **go-rift** | Partially taken | paralin/go-rift-api (League of Legends API bindings) |
| **go-sift** | TAKEN | bhoriuchi/go-siftjs (MongoDB query port) |
| **godiff** | TAKEN | Multiple: spcau/godiff, viant/godiff, craftslab/godiff, seletskiy/godiff |
| **diffgo** | Available | No matching repository found |
| **giff** | TAKEN | bahdotsh/giff - terminal Git diff viewer (direct competitor!) |
| **digo** | TAKEN | Multiple DI libraries: cone/digo, werbenhu/digo, dynport/digo |

**Source:** GitHub and pkg.go.dev searches
**Confidence:** High

### Q3: What Git diff TUI tools exist in Go?

**Answer:** The Go ecosystem has several relevant projects:

| Tool | Description |
|------|-------------|
| **lazygit** | The most popular Git TUI in Go (37k+ stars) |
| **gitti** | Lightweight terminal UI for git operations |
| **go-git-tui** | Staging and diff TUI inspired by lazygit |
| **giff** | Terminal-based Git diff viewer with interactive rebase |

**Source:** GitHub and pkg.go.dev
**Confidence:** High

### Q4: Which names are potentially safe to use?

**Answer:**

| Name | Assessment |
|------|------------|
| **diffgo** | Appears available - no matching repos found |
| **go-rift** | Risky - related API project exists |
| **rift** (standalone) | Crowded namespace, but no dominant git tool |

**Confidence:** Medium - always verify before final selection

## Detailed Findings

### Finding 1: "giff" is a Direct Competitor
**Source:** https://github.com/bahdotsh/giff
**Key Points:**
- Terminal-based Git diff viewer with interactive rebase capabilities
- Written in Rust, installable via cargo
- Features: j/k navigation, unified/side-by-side toggle, rebase mode
- This is almost exactly what you're building

### Finding 2: Established Diff Tool Naming Patterns
**Source:** https://github.com/dandavison/delta, difftastic, diff-so-fancy
**Key Points:**
- Single short words: delta, meld
- Descriptive compounds: difftastic, vimdiff, icdiff
- Most successful tools have 4-7 character names
- "delta" is the most popular modern diff pager (25k+ stars)

### Finding 3: Go TUI Ecosystem
**Source:** https://github.com/charmbracelet/bubbletea
**Key Points:**
- Bubbletea powers 10,000+ applications
- lazygit is the gold standard for Git TUIs
- Charm ecosystem (lipgloss, bubbles) provides styling primitives

## Recommendations

### For This Codebase

Given the competitive landscape, here are 5 creative name suggestions:

1. **diffy** - Playful, short, memorable (verify availability)
2. **gdif** - Go + diff, 4 chars, punchy
3. **viewdiff** or **vdiff** - Descriptive, clear purpose
4. **splitdiff** - If focusing on side-by-side view
5. **diffo** - diff + Go suffix, unique

### Alternative Creative Combinations

| Name | Rationale |
|------|-----------|
| **gode** | Go + delta-inspired |
| **delt** | Delta-like but shorter |
| **seam** | Where two things meet (diff context) |
| **span** | Covers the diff range |
| **hunk** | Git terminology for diff sections |
| **patch** | Classic but may be taken |

### Names to Avoid

- **giff** - direct competitor exists
- **godiff** - multiple projects exist
- **flux** - major CNCF project
- **delta** - extremely popular diff tool

## Sources
1. [svent/sift](https://github.com/svent/sift) - grep alternative in Go
2. [fluxcd/flux2](https://github.com/fluxcd/flux2) - CNCF GitOps project
3. [stoplightio/prism](https://github.com/stoplightio/prism) - API mocking tool
4. [brigand/glint](https://github.com/brigand/glint) - conventional commits tool
5. [simolus3/drift](https://github.com/simolus3/drift) - Dart persistence library
6. [bahdotsh/giff](https://github.com/bahdotsh/giff) - terminal Git diff viewer
7. [dandavison/delta](https://github.com/dandavison/delta) - syntax-highlighting pager
8. [jesseduffield/lazygit](https://github.com/jesseduffield/lazygit) - terminal UI for git
9. [rothgar/awesome-tuis](https://github.com/rothgar/awesome-tuis) - TUI project list
10. [gohyuhan/gitti](https://github.com/gohyuhan/gitti) - lightweight Git TUI

## Open Questions
- Is "diffy" taken? (would need separate verification)
- What about compound names like "godelta" or "go-delta"?
- Consider two-word names that are still typeable quickly
