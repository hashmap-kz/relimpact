package diffs

import (
	"html/template"
	"regexp"
	"sort"
	"strings"
	"time"
)

// -----------------------------------------------------------------------
// Public types
// -----------------------------------------------------------------------

// ReportMetadata holds context for rendering the HTML report header.
type ReportMetadata struct {
	Repo   string
	OldRef string
	NewRef string
	Now    time.Time
}

// SymbolChange is a processed, display-ready record of a single public symbol
// that was added, removed, or had its signature changed between two refs.
type SymbolChange struct {
	Package  string
	Kind     string
	Scope    string
	Name     string
	Old      string
	New      string
	Status   string
	TypeBody *APITypeBody // non-nil for whole-type adds/removes with struct or interface bodies
}

// StructFieldChange holds a parsed field belonging to a specific struct owner,
// ready for grouped diff rendering.
type StructFieldChange struct {
	Owner  string
	Name   string
	Type   string
	Status string
}

// PackageChangeSummary aggregates change counts for a single package.
type PackageChangeSummary struct {
	Package  string
	Breaking int
	Changed  int
	Removed  int
	Added    int
	Total    int
}

type overallSummary struct {
	Breaking        int
	Changed         int
	Removed         int
	Added           int
	PackagesAdded   int
	PackagesRemoved int
	ChangedPackages int
	Total           int
}

// -----------------------------------------------------------------------
// Template data types
// -----------------------------------------------------------------------

type reportData struct {
	CSS              template.HTML
	Meta             reportMetaData
	Verdict          verdictData
	Summary          overallSummary
	BreakingPackages []sidebarEntry
	Packages         []pkgSectionData
}

type reportMetaData struct {
	Repo        string
	OldRef      string
	NewRef      string
	GeneratedAt string
}

type verdictData struct {
	Class string
	Icon  string
	Text  string
}

type sidebarEntry struct {
	AnchorID  string
	Package   string
	ShortName string
	Breaking  int
}

type pkgSectionData struct {
	AnchorID       string
	Package        string
	ShortName      string
	IsBreaking     bool
	StatusSections []statusSectionData
}

type statusSectionData struct {
	Title            string
	AccentClass      string
	ChangeCards      []changeCardData
	KindGroups       []kindGroupData
	TypeDefBlocks    []typeDefBlockData
	StructFieldDiffs []structFieldDiffData
}

type changeCardData struct {
	KindLabel   string
	Name        string
	UnifiedDiff template.HTML
}

type kindGroupData struct {
	KindLabel  string
	CountClass string
	Count      int
	Rows       []compactRowData // used for Func/Method — inline signature per row
	Body       template.HTML    // used for Var/Const/Type — rendered as a code block
}

type compactRowData struct {
	PrefixClass string
	Prefix      string
	Code        string
	Type        string
}

type structFieldDiffData struct {
	Title      string
	CountClass string
	Count      int
	Body       template.HTML
}

type typeDefBlockData struct {
	KindLabel  string
	CountClass string
	Count      int
	Body       template.HTML
}

// -----------------------------------------------------------------------
// Entry point
// -----------------------------------------------------------------------

var reportTmpl = template.Must(
	template.New("report").
		Funcs(template.FuncMap{
			"not": func(v interface{}) bool {
				switch x := v.(type) {
				case []pkgSectionData:
					return len(x) == 0
				}
				return false
			},
		}).
		Parse(reportTemplates),
)

func (d *APIDiff) HTML(meta ReportMetadata) string {
	if meta.Now.IsZero() {
		meta.Now = time.Now()
	}

	changes := d.collectSymbolChanges()
	pkgSummaries := summarizeByPackage(changes)
	summary := summarizeOverall(changes, len(pkgSummaries))

	verdict := verdictData{
		Class: "verdict-ok",
		Icon:  "circle-check",
		Text:  "Compatible API changes only — no breaking changes detected.",
	}
	if summary.Breaking > 0 {
		verdict = verdictData{
			Class: "verdict-breaking",
			Icon:  "alert-triangle",
			Text:  "Breaking API changes detected — review before release.",
		}
	}

	data := reportData{
		CSS: template.HTML(reportCSS),
		Meta: reportMetaData{
			Repo:        defaultString(meta.Repo, "repository"),
			OldRef:      defaultString(meta.OldRef, "old"),
			NewRef:      defaultString(meta.NewRef, "new"),
			GeneratedAt: meta.Now.Format("2006-01-02 15:04:05"),
		},
		Verdict:          verdict,
		Summary:          summary,
		BreakingPackages: buildSidebarEntries(pkgSummaries),
		Packages:         buildPackageSections(changes),
	}

	var sb strings.Builder
	if err := reportTmpl.ExecuteTemplate(&sb, "report", data); err != nil {
		return "<!-- template error: " + err.Error() + " -->"
	}
	return sb.String()
}

// -----------------------------------------------------------------------
// Sidebar
// -----------------------------------------------------------------------

func buildSidebarEntries(pkgs []PackageChangeSummary) []sidebarEntry {
	var out []sidebarEntry
	for _, p := range pkgs {
		if p.Breaking > 0 {
			out = append(out, sidebarEntry{
				AnchorID:  anchorID(p.Package),
				Package:   p.Package,
				ShortName: shortPackage(p.Package),
				Breaking:  p.Breaking,
			})
		}
	}
	return out
}

// -----------------------------------------------------------------------
// Package sections
// -----------------------------------------------------------------------

func buildPackageSections(changes []SymbolChange) []pkgSectionData {
	byPkg := make(map[string][]SymbolChange)
	for _, ch := range changes {
		byPkg[ch.Package] = append(byPkg[ch.Package], ch)
	}
	pkgs := make([]string, 0, len(byPkg))
	for pkg := range byPkg {
		pkgs = append(pkgs, pkg)
	}
	sort.Strings(pkgs)

	var out []pkgSectionData
	for _, pkg := range pkgs {
		pkgChanges := byPkg[pkg]
		sort.Slice(pkgChanges, func(i, j int) bool {
			if statusRank(pkgChanges[i].Status) != statusRank(pkgChanges[j].Status) {
				return statusRank(pkgChanges[i].Status) < statusRank(pkgChanges[j].Status)
			}
			if kindRank(pkgChanges[i].Kind) != kindRank(pkgChanges[j].Kind) {
				return kindRank(pkgChanges[i].Kind) < kindRank(pkgChanges[j].Kind)
			}
			return pkgChanges[i].Name < pkgChanges[j].Name
		})

		isBreaking := false
		for _, ch := range pkgChanges {
			if ch.Status == "changed" || ch.Status == "removed" {
				isBreaking = true
				break
			}
		}

		out = append(out, pkgSectionData{
			AnchorID:       anchorID(pkg),
			Package:        pkg,
			ShortName:      shortPackage(pkg),
			IsBreaking:     isBreaking,
			StatusSections: buildStatusSections(pkgChanges),
		})
	}
	return out
}

func buildStatusSections(changes []SymbolChange) []statusSectionData {
	var sections []statusSectionData
	specs := []struct {
		status      string
		title       string
		accentClass string
	}{
		{"changed", "Changed signatures", "changed"},
		{"removed", "Removed API", "removed"},
		{"added", "Added API", "added"},
	}
	for _, spec := range specs {
		var group []SymbolChange
		for _, ch := range changes {
			if ch.Status == spec.status {
				group = append(group, ch)
			}
		}
		if len(group) == 0 {
			continue
		}
		sections = append(sections, buildStatusSection(spec.title, spec.accentClass, spec.status, group))
	}
	return sections
}

func buildStatusSection(title, accentClass, status string, changes []SymbolChange) statusSectionData {
	s := statusSectionData{
		Title:       title,
		AccentClass: accentClass,
	}

	if status == "changed" {
		for _, ch := range changes {
			s.ChangeCards = append(s.ChangeCards, buildChangeCard(ch))
		}
		return s
	}

	// Group by kind in canonical order.
	byKind := make(map[string][]SymbolChange)
	for _, ch := range changes {
		byKind[ch.Kind] = append(byKind[ch.Kind], ch)
	}

	for _, kind := range apiKindOrder() {
		xs := byKind[kind]
		if len(xs) == 0 {
			continue
		}
		sort.Slice(xs, func(i, j int) bool { return xs[i].Name < xs[j].Name })

		switch kind {
		case "Field":
			s.StructFieldDiffs = append(s.StructFieldDiffs, buildStructFieldDiffs(xs, status)...)
		case "Type":
			var withBody, withoutBody []SymbolChange
			for _, ch := range xs {
				if ch.TypeBody != nil && (ch.TypeBody.Kind == "struct" || ch.TypeBody.Kind == "interface") {
					withBody = append(withBody, ch)
				} else {
					withoutBody = append(withoutBody, ch)
				}
			}
			for _, ch := range withBody {
				s.TypeDefBlocks = append(s.TypeDefBlocks, buildTypeDefBlock(ch, status))
			}
			if len(withoutBody) > 0 {
				s.KindGroups = append(s.KindGroups, buildScalarTypeGroup(withoutBody, status))
			}
		case "Var":
			s.KindGroups = append(s.KindGroups, buildVarGroup(xs, status))
		case "Const":
			s.KindGroups = append(s.KindGroups, buildConstGroup(xs, status))
		default:
			s.KindGroups = append(s.KindGroups, buildKindGroup(kind, xs, status))
		}
	}
	return s
}

// -----------------------------------------------------------------------
// Change cards (changed signatures)
// -----------------------------------------------------------------------

func buildChangeCard(ch SymbolChange) changeCardData {
	oldFmt := formatAPIValue(ch.Kind, ch.Scope, ch.Old)
	newFmt := formatAPIValue(ch.Kind, ch.Scope, ch.New)
	return changeCardData{
		KindLabel:   ch.Kind + " signature",
		Name:        ch.Name,
		UnifiedDiff: buildUnifiedDiff(oldFmt, newFmt),
	}
}

// buildUnifiedDiff returns safe HTML for a unified diff block.
// Removed lines are red, added lines are green, shared context is muted.
func buildUnifiedDiff(oldSig, newSig string) template.HTML {
	oldLines := strings.Split(oldSig, "\n")
	newLines := strings.Split(newSig, "\n")

	oldSet := make(map[string]bool, len(oldLines))
	newSet := make(map[string]bool, len(newLines))
	for _, l := range oldLines {
		oldSet[l] = true
	}
	for _, l := range newLines {
		newSet[l] = true
	}

	var b strings.Builder
	for _, line := range oldLines {
		if !newSet[line] {
			b.WriteString(`<span class="diff-removed">− `)
			b.WriteString(template.HTMLEscapeString(line))
			b.WriteString("\n</span>")
		}
	}
	for _, line := range oldLines {
		if newSet[line] {
			b.WriteString(`<span class="diff-context">  `)
			b.WriteString(template.HTMLEscapeString(line))
			b.WriteString("\n</span>")
		}
	}
	for _, line := range newLines {
		if !oldSet[line] {
			b.WriteString(`<span class="diff-added">+ `)
			b.WriteString(template.HTMLEscapeString(line))
			b.WriteString("\n</span>")
		}
	}
	return template.HTML(b.String())
}

// -----------------------------------------------------------------------
// Compact kind groups
// -----------------------------------------------------------------------

func buildKindGroup(kind string, changes []SymbolChange, status string) kindGroupData {
	countClass := "add"
	prefix := "+"
	prefixClass := "add"
	if status == "removed" {
		countClass = "rem"
		prefix = "−"
		prefixClass = "rem"
	}

	rows := make([]compactRowData, 0, len(changes))
	for _, ch := range changes {
		code, typ := compactRowParts(ch)
		rows = append(rows, compactRowData{
			PrefixClass: prefixClass,
			Prefix:      prefix,
			Code:        code,
			Type:        typ,
		})
	}

	return kindGroupData{
		KindLabel:  pluralKind(kind),
		CountClass: countClass,
		Count:      len(changes),
		Rows:       rows,
	}
}

// buildScalarTypeGroup renders named scalar types (type Foo string, type ID int, etc.)
// as a code block — more context than a bare name.
func buildScalarTypeGroup(changes []SymbolChange, status string) kindGroupData {
	prefix, spanClass, countClass := diffPrefixClasses(status)
	var b strings.Builder
	for _, ch := range changes {
		sig := abbreviateImportPaths(strings.TrimSpace(firstNonEmpty(ch.New, ch.Old)))
		// sig is just the type name; formatAPIValue prepends "type "
		line := formatAPIValue("Type", "", sig)
		b.WriteString(`<span class="` + spanClass + `">`)
		b.WriteString(prefix + " " + template.HTMLEscapeString(line) + "</span>")
	}
	return kindGroupData{
		KindLabel:  pluralKind("Type"),
		CountClass: countClass,
		Count:      len(changes),
		Body:       template.HTML(b.String()),
	}
}

// buildVarGroup renders package-level variables as a code block.
func buildVarGroup(changes []SymbolChange, status string) kindGroupData {
	prefix, spanClass, countClass := diffPrefixClasses(status)
	var b strings.Builder
	for _, ch := range changes {
		sig := abbreviateImportPaths(strings.TrimSpace(firstNonEmpty(ch.New, ch.Old)))
		parts := strings.Fields(sig)
		var line string
		if len(parts) >= 2 {
			name := parts[0]
			typ := strings.TrimSpace(strings.TrimPrefix(sig, name))
			line = "var " + name + " " + typ
		} else {
			line = "var " + sig
		}
		b.WriteString(`<span class="` + spanClass + `">` + prefix + " " + template.HTMLEscapeString(line) + "\n</span>")
	}
	return kindGroupData{
		KindLabel:  pluralKind("Var"),
		CountClass: countClass,
		Count:      len(changes),
		Body:       template.HTML(b.String()),
	}
}

// parseConstSignature splits a const signature "Name Type = value" into parts.
// isUntypedBasic reports whether a Go type string is an untyped basic kind
// (e.g. "untyped string", "untyped int"). For these, the type is noise in the
// report — the value is what matters.
func isUntypedBasic(typ string) bool {
	return strings.HasPrefix(typ, "untyped ")
}

// parseConstSignature splits "Name Type = value" into its parts and decides
// whether to surface the type in the display.
func parseConstSignature(sig string) (name, typ, value string) {
	sig = abbreviateImportPaths(strings.TrimSpace(sig))
	if idx := strings.Index(sig, " = "); idx >= 0 {
		value = strings.TrimSpace(sig[idx+3:])
		sig = sig[:idx]
	}
	parts := strings.Fields(sig)
	if len(parts) == 0 {
		return "", "", value
	}
	name = parts[0]
	if len(parts) >= 2 {
		typ = strings.TrimSpace(strings.TrimPrefix(sig, name))
		// Drop untyped basic types — they add no information.
		if isUntypedBasic(typ) {
			typ = ""
		}
	}
	return name, typ, value
}

// buildConstGroup renders constants as a code block, grouping consts that share
// the same named type into a single const ( ... ) block — the common iota-enum
// pattern. Untyped consts are rendered as individual "const Name = value" lines.
func buildConstGroup(changes []SymbolChange, status string) kindGroupData {
	prefix, spanClass, countClass := diffPrefixClasses(status)

	type constEntry struct{ name, typ, value string }
	var entries []constEntry
	for _, ch := range changes {
		name, typ, value := parseConstSignature(firstNonEmpty(ch.New, ch.Old))
		entries = append(entries, constEntry{name, typ, value})
	}

	// Only group by named types (typ != ""), and only when more than one const
	// shares that type.
	typCount := make(map[string]int)
	for _, e := range entries {
		if e.typ != "" {
			typCount[e.typ]++
		}
	}

	var b strings.Builder
	rendered := make(map[string]bool)

	// Gather grouped types in order of first appearance.
	var groupedTypes []string
	seenType := make(map[string]bool)
	for _, e := range entries {
		if e.typ != "" && typCount[e.typ] > 1 && !seenType[e.typ] {
			groupedTypes = append(groupedTypes, e.typ)
			seenType[e.typ] = true
		}
	}
	for _, typ := range groupedTypes {
		b.WriteString(`<span class="diff-context">const (` + "\n</span>")
		for _, e := range entries {
			if e.typ != typ {
				continue
			}
			rendered[e.name] = true
			line := "    " + e.name + " " + typ
			if e.value != "" {
				line += " = " + e.value
			}
			b.WriteString(`<span class="` + spanClass + `">` + prefix + " " + template.HTMLEscapeString(line) + "\n</span>")
		}
		b.WriteString(`<span class="diff-context">)` + "\n</span>")
	}

	// Remaining entries: named-type singletons and all untyped consts.
	for _, e := range entries {
		if rendered[e.name] {
			continue
		}
		line := "const " + e.name
		if e.typ != "" {
			line += " " + e.typ
		}
		if e.value != "" {
			line += " = " + e.value
		}
		b.WriteString(`<span class="` + spanClass + `">` + prefix + " " + template.HTMLEscapeString(line) + "\n</span>")
	}

	return kindGroupData{
		KindLabel:  pluralKind("Const"),
		CountClass: countClass,
		Count:      len(changes),
		Body:       template.HTML(b.String()),
	}
}

// diffPrefixClasses returns the diff prefix string, span CSS class, and count
// badge CSS class for a given status ("added" or "removed").
func diffPrefixClasses(status string) (prefix, spanClass, countClass string) {
	if status == "removed" {
		return "−", "diff-removed", "rem"
	}
	return "+", "diff-added", "add"
}

// compactRowParts returns the code string and optional type annotation for a
// compact list row, depending on the symbol kind.
func compactRowParts(ch SymbolChange) (code, typ string) {
	switch ch.Kind {
	case "Func":
		return compactFunc(ch), ""
	case "Method":
		return compactMethod(ch), ""
	case "Var", "Const":
		return compactTypedSymbol(ch)
	default:
		return formatCompactSymbol(ch, ""), ""
	}
}

// -----------------------------------------------------------------------
// Struct field diffs
// -----------------------------------------------------------------------

func buildStructFieldDiffs(changes []SymbolChange, status string) []structFieldDiffData {
	grouped := groupFieldsByOwner(changes, status)
	if len(grouped) == 0 {
		// Fallback: no owner info, render as a plain kind group.
		kg := buildKindGroup("Field", changes, status)
		return []structFieldDiffData{{
			Title:      kg.KindLabel,
			CountClass: kg.CountClass,
			Count:      kg.Count,
			Body:       buildFallbackFieldBody(changes, status),
		}}
	}

	owners := make([]string, 0, len(grouped))
	for owner := range grouped {
		owners = append(owners, owner)
	}
	sort.Strings(owners)

	countClass := "add"
	if status == "removed" {
		countClass = "rem"
	}

	var out []structFieldDiffData
	for _, owner := range owners {
		fields := grouped[owner]
		out = append(out, structFieldDiffData{
			Title:      "type " + owner + " struct",
			CountClass: countClass,
			Count:      len(fields),
			Body:       buildStructFieldDiffBody(fields, status),
		})
	}
	return out
}

func buildStructFieldDiffBody(fields []StructFieldChange, status string) template.HTML {
	prefix := "+"
	spanClass := "diff-added"
	if status == "removed" {
		prefix = "-"
		spanClass = "diff-removed"
	}

	width := 0
	for _, f := range fields {
		if len(f.Name) > width {
			width = len(f.Name)
		}
	}

	var b strings.Builder
	b.WriteString(`<span class="diff-context">  {` + "\n</span>")
	for _, f := range fields {
		pad := width - len(f.Name) + 1
		if pad < 1 {
			pad = 1
		}
		b.WriteString(`<span class="` + spanClass + `">`)
		b.WriteString(prefix + "   ")
		b.WriteString(template.HTMLEscapeString(f.Name))
		b.WriteString(strings.Repeat(" ", pad))
		b.WriteString(template.HTMLEscapeString(f.Type))
		b.WriteString("\n</span>")
	}
	b.WriteString(`<span class="diff-context">  }` + "</span>")
	return template.HTML(b.String())
}

func buildFallbackFieldBody(changes []SymbolChange, status string) template.HTML {
	prefix := "+"
	spanClass := "diff-added"
	if status == "removed" {
		prefix = "-"
		spanClass = "diff-removed"
	}
	var b strings.Builder
	for _, ch := range changes {
		raw := abbreviateImportPaths(strings.TrimSpace(firstNonEmpty(ch.New, ch.Old)))
		b.WriteString(`<span class="` + spanClass + `">`)
		b.WriteString(prefix + "   ")
		b.WriteString(template.HTMLEscapeString(raw))
		b.WriteString("\n</span>")
	}
	return template.HTML(b.String())
}

// -----------------------------------------------------------------------
// Type definition blocks (whole struct / interface added or removed)
// -----------------------------------------------------------------------

func buildTypeDefBlock(ch SymbolChange, status string) typeDefBlockData {
	body := ch.TypeBody
	prefix := "+"
	spanClass := "diff-added"
	countClass := "add"
	if status == "removed" {
		prefix = "-"
		spanClass = "diff-removed"
		countClass = "rem"
	}

	var b strings.Builder

	switch body.Kind {
	case "struct":
		b.WriteString(`<span class="diff-context">type `)
		b.WriteString(template.HTMLEscapeString(ch.Name))
		b.WriteString(" struct {\n</span>")

		// Align field name column.
		width := 0
		for _, f := range body.Fields {
			parts := strings.Fields(abbreviateImportPaths(f))
			if len(parts) > 0 && len(parts[0]) > width {
				width = len(parts[0])
			}
		}
		for _, f := range body.Fields {
			f = abbreviateImportPaths(f)
			parts := strings.Fields(f)
			if len(parts) == 0 {
				continue
			}
			name := parts[0]
			typ := strings.TrimSpace(strings.TrimPrefix(f, name))
			pad := width - len(name) + 1
			if pad < 1 {
				pad = 1
			}
			b.WriteString(`<span class="` + spanClass + `">`)
			b.WriteString(prefix + "   ")
			b.WriteString(template.HTMLEscapeString(name))
			b.WriteString(strings.Repeat(" ", pad))
			b.WriteString(template.HTMLEscapeString(typ))
			b.WriteString("\n</span>")
		}
		b.WriteString(`<span class="diff-context">}</span>`)

	case "interface":
		b.WriteString(`<span class="diff-context">type `)
		b.WriteString(template.HTMLEscapeString(ch.Name))
		b.WriteString(" interface {\n</span>")
		for _, m := range body.Methods {
			sig := abbreviateImportPaths(m)
			b.WriteString(`<span class="` + spanClass + `">`)
			b.WriteString(prefix + "   ")
			b.WriteString(template.HTMLEscapeString(sig))
			b.WriteString("\n</span>")
		}
		b.WriteString(`<span class="diff-context">}</span>`)
	}

	kindLabel := body.Kind // "struct" or "interface"
	memberCount := len(body.Fields) + len(body.Methods)

	return typeDefBlockData{
		KindLabel:  "type " + ch.Name + " " + kindLabel,
		CountClass: countClass,
		Count:      memberCount,
		Body:       template.HTML(b.String()),
	}
}

// -----------------------------------------------------------------------
// Data pipeline (collect → merge → summarize)
// -----------------------------------------------------------------------

func (d *APIDiff) collectSymbolChanges() []SymbolChange {
	var changes []SymbolChange

	addPkgChanges := func(status string, xs []string) {
		for _, x := range xs {
			ch := SymbolChange{
				Package: x,
				Kind:    "Package",
				Name:    shortPackage(x),
				Old:     x,
				New:     x,
				Status:  status,
			}
			if status == "added" {
				ch.Old = ""
			} else {
				ch.New = ""
			}
			changes = append(changes, ch)
		}
	}

	addChanges := func(status, kind string, xs []DiffItem) {
		for _, x := range xs {
			cleanKind := cleanAPILabel(kind, x.Label)
			scope := apiScope(x.Label)
			name := apiSymbolName(x.Signature)
			value := x.Signature

			ch := SymbolChange{
				Package:  x.Path,
				Kind:     cleanKind,
				Scope:    scope,
				Name:     displayAPIName(cleanKind, scope, name),
				Old:      value,
				New:      value,
				Status:   status,
				TypeBody: x.TypeBody,
			}
			if status == "added" {
				ch.Old = ""
			} else {
				ch.New = ""
			}
			changes = append(changes, ch)
		}
	}

	addPkgChanges("added", d.PackagesAdded)
	addPkgChanges("removed", d.PackagesRemoved)
	addChanges("added", "Func", d.FuncsAdded)
	addChanges("removed", "Func", d.FuncsRemoved)
	addChanges("added", "Var", d.VarsAdded)
	addChanges("removed", "Var", d.VarsRemoved)
	addChanges("added", "Const", d.ConstsAdded)
	addChanges("removed", "Const", d.ConstsRemoved)
	addChanges("added", "Type", d.TypesAdded)
	addChanges("removed", "Type", d.TypesRemoved)
	addChanges("added", "Field", d.FieldsAdded)
	addChanges("removed", "Field", d.FieldsRemoved)
	addChanges("added", "Method", d.MethodsAdded)
	addChanges("removed", "Method", d.MethodsRemoved)

	return mergeIntoChangedSymbols(changes)
}

func mergeIntoChangedSymbols(raw []SymbolChange) []SymbolChange {
	added := make(map[string]SymbolChange)
	removed := make(map[string]SymbolChange)
	var out []SymbolChange

	for _, ch := range raw {
		key := ch.changeKey()
		switch ch.Status {
		case "added":
			added[key] = ch
		case "removed":
			removed[key] = ch
		}
	}

	usedAdded := make(map[string]bool)
	usedRemoved := make(map[string]bool)
	for key, old := range removed {
		newer, ok := added[key]
		if !ok || old.Kind == "Package" {
			continue
		}
		out = append(out, SymbolChange{
			Package: old.Package,
			Kind:    old.Kind,
			Scope:   old.Scope,
			Name:    old.Name,
			Old:     old.Old,
			New:     newer.New,
			Status:  "changed",
		})
		usedAdded[key] = true
		usedRemoved[key] = true
	}

	for _, ch := range raw {
		key := ch.changeKey()
		if usedAdded[key] || usedRemoved[key] {
			continue
		}
		out = append(out, ch)
	}

	sort.Slice(out, func(i, j int) bool {
		if statusRank(out[i].Status) != statusRank(out[j].Status) {
			return statusRank(out[i].Status) < statusRank(out[j].Status)
		}
		if out[i].Package != out[j].Package {
			return out[i].Package < out[j].Package
		}
		if kindRank(out[i].Kind) != kindRank(out[j].Kind) {
			return kindRank(out[i].Kind) < kindRank(out[j].Kind)
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func (ch SymbolChange) changeKey() string {
	return ch.Package + "\x00" + ch.Kind + "\x00" + ch.Scope + "\x00" + apiSymbolName(firstNonEmpty(ch.New, ch.Old))
}

func summarizeOverall(changes []SymbolChange, changedPackages int) overallSummary {
	var s overallSummary
	s.ChangedPackages = changedPackages
	for _, ch := range changes {
		s.Total++
		switch ch.Status {
		case "changed":
			s.Changed++
			s.Breaking++
		case "removed":
			s.Removed++
			s.Breaking++
			if ch.Kind == "Package" {
				s.PackagesRemoved++
			}
		case "added":
			s.Added++
			if ch.Kind == "Package" {
				s.PackagesAdded++
			}
		}
	}
	return s
}

func summarizeByPackage(changes []SymbolChange) []PackageChangeSummary {
	byPkg := make(map[string]*PackageChangeSummary)
	for _, ch := range changes {
		pkg := ch.Package
		if pkg == "" {
			pkg = "unknown"
		}
		s, ok := byPkg[pkg]
		if !ok {
			s = &PackageChangeSummary{Package: pkg}
			byPkg[pkg] = s
		}
		s.Total++
		switch ch.Status {
		case "changed":
			s.Changed++
			s.Breaking++
		case "removed":
			s.Removed++
			s.Breaking++
		case "added":
			s.Added++
		}
	}
	out := make([]PackageChangeSummary, 0, len(byPkg))
	for _, s := range byPkg {
		out = append(out, *s)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Breaking != out[j].Breaking {
			return out[i].Breaking > out[j].Breaking
		}
		if out[i].Total != out[j].Total {
			return out[i].Total > out[j].Total
		}
		return out[i].Package < out[j].Package
	})
	return out
}

// -----------------------------------------------------------------------
// Field parsing helpers
// -----------------------------------------------------------------------

func parseFieldChange(ch SymbolChange) (StructFieldChange, bool) {
	raw := abbreviateImportPaths(strings.TrimSpace(firstNonEmpty(ch.New, ch.Old)))
	if raw == "" {
		return StructFieldChange{}, false
	}
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return StructFieldChange{}, false
	}
	fullName := parts[0]
	fieldType := strings.TrimSpace(strings.TrimPrefix(raw, fullName))
	dot := strings.LastIndex(fullName, ".")
	if dot < 0 || dot+1 >= len(fullName) {
		if ch.Scope == "" || ch.Name == "" {
			return StructFieldChange{}, false
		}
		return StructFieldChange{Owner: ch.Scope, Name: ch.Name, Type: fieldType, Status: ch.Status}, true
	}
	return StructFieldChange{Owner: fullName[:dot], Name: fullName[dot+1:], Type: fieldType, Status: ch.Status}, true
}

func groupFieldsByOwner(changes []SymbolChange, status string) map[string][]StructFieldChange {
	out := make(map[string][]StructFieldChange)
	for _, ch := range changes {
		if ch.Status != status || ch.Kind != "Field" {
			continue
		}
		f, ok := parseFieldChange(ch)
		if !ok {
			continue
		}
		out[f.Owner] = append(out[f.Owner], f)
	}
	for owner := range out {
		sort.Slice(out[owner], func(i, j int) bool {
			return out[owner][i].Name < out[owner][j].Name
		})
	}
	return out
}

// -----------------------------------------------------------------------
// Signature formatting
// -----------------------------------------------------------------------

func formatAPIValue(kind, scope, raw string) string {
	raw = abbreviateImportPaths(strings.TrimSpace(raw))
	name := apiSymbolName(raw)
	switch kind {
	case "Func":
		return formatCallable("func", "", name, raw)
	case "Method":
		return formatCallable("func", scope, name, raw)
	case "Field":
		return formatField(scope, raw)
	case "Var":
		return "var " + raw
	case "Const":
		return "const " + raw
	case "Type":
		return "type " + raw
	case "Package":
		return raw
	default:
		return raw
	}
}

func formatCallable(prefix, scope, name, raw string) string {
	idx := strings.Index(raw, "(")
	if idx < 0 {
		return prefix + " " + raw
	}
	sig := raw[idx:]
	params, rest, ok := splitFirstParen(sig)
	if !ok {
		return prefix + " " + raw
	}
	results := ""
	rest = strings.TrimSpace(rest)
	if strings.HasPrefix(rest, "->") {
		resRaw := strings.TrimSpace(strings.TrimPrefix(rest, "->"))
		if strings.HasPrefix(resRaw, "(") {
			res, _, ok := splitFirstParen(resRaw)
			if ok {
				resParts := splitTopLevel(res)
				switch len(resParts) {
				case 0:
				case 1:
					results = " " + resParts[0]
				default:
					results = " (" + strings.Join(resParts, ", ") + ")"
				}
			}
		}
	}

	parts := splitTopLevel(params)
	funcHead := prefix + " "
	if scope != "" {
		funcHead += "(" + abbreviateImportPaths(scope) + ") "
	}
	funcHead += name

	oneLine := funcHead + "(" + strings.Join(parts, ", ") + ")" + results
	if len(oneLine) <= 96 && len(parts) <= 1 {
		return oneLine
	}
	var b strings.Builder
	b.WriteString(funcHead)
	b.WriteString("(\n")
	for _, p := range parts {
		if p == "" {
			continue
		}
		b.WriteString("    ")
		b.WriteString(p)
		b.WriteString(",\n")
	}
	b.WriteString(")")
	b.WriteString(results)
	return b.String()
}

func formatField(scope, raw string) string {
	name, typ := compactField(SymbolChange{Kind: "Field", Scope: scope, Name: apiSymbolName(raw), New: raw})
	if typ == "" {
		return name
	}
	return name + " " + typ
}

func splitFirstParen(s string) (inside, rest string, ok bool) {
	if !strings.HasPrefix(s, "(") {
		return "", s, false
	}
	depth := 0
	for i, r := range s {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return s[1:i], s[i+1:], true
			}
		}
	}
	return "", s, false
}

func splitTopLevel(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var out []string
	start, paren, bracket, brace := 0, 0, 0, 0
	for i, r := range s {
		switch r {
		case '(':
			paren++
		case ')':
			paren--
		case '[':
			bracket++
		case ']':
			bracket--
		case '{':
			brace++
		case '}':
			brace--
		case ',':
			if paren == 0 && bracket == 0 && brace == 0 {
				out = append(out, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
	}
	out = append(out, strings.TrimSpace(s[start:]))
	return out
}

// -----------------------------------------------------------------------
// Compact display helpers
// -----------------------------------------------------------------------

func formatCompactSymbol(ch SymbolChange, _ string) string {
	switch ch.Kind {
	case "Type", "Package":
		return ch.Name
	case "Var", "Const":
		name, _ := compactTypedSymbol(ch)
		return name
	case "Field":
		name, _ := compactField(ch)
		return name
	case "Method":
		return compactMethod(ch)
	case "Func":
		return compactFunc(ch)
	default:
		return ch.Name
	}
}

func compactField(ch SymbolChange) (name string, typ string) {
	raw := abbreviateImportPaths(strings.TrimSpace(firstNonEmpty(ch.New, ch.Old)))
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return compactScopedName(ch.Scope, ch.Name), ""
	}
	fieldName := parts[0]
	fieldType := strings.TrimSpace(strings.TrimPrefix(raw, fieldName))
	if strings.Contains(fieldName, ".") {
		return fieldName, fieldType
	}
	return compactScopedName(ch.Scope, fieldName), fieldType
}

func compactTypedSymbol(ch SymbolChange) (name string, typ string) {
	raw := abbreviateImportPaths(strings.TrimSpace(firstNonEmpty(ch.New, ch.Old)))
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return ch.Name, ""
	}
	name = parts[0]
	typ = strings.TrimSpace(strings.TrimPrefix(raw, name))
	return name, typ
}

func compactFunc(ch SymbolChange) string {
	raw := abbreviateImportPaths(strings.TrimSpace(firstNonEmpty(ch.New, ch.Old)))
	name := apiSymbolName(raw)
	if name == "" || name == "unknown" {
		name = ch.Name
	}
	params, results := extractCallableSignatureParts(raw)
	if params == "" && results == "" {
		return name + "(...)"
	}
	return name + "(" + params + ")" + results
}

func compactMethod(ch SymbolChange) string {
	raw := abbreviateImportPaths(strings.TrimSpace(firstNonEmpty(ch.New, ch.Old)))
	name := apiSymbolName(raw)
	if name == "" || name == "unknown" {
		name = ch.Name
	}
	fullName := name
	if ch.Scope != "" && !strings.Contains(name, ".") {
		fullName = ch.Scope + "." + name
	}
	params, results := extractCallableSignatureParts(raw)
	if params == "" && results == "" {
		return fullName + "(...)"
	}
	return fullName + "(" + params + ")" + results
}

func compactScopedName(scope, name string) string {
	name = strings.TrimSpace(name)
	scope = strings.TrimSpace(scope)
	if scope == "" || name == "" {
		return name
	}
	if strings.HasPrefix(name, scope+".") {
		return name
	}
	return scope + "." + name
}

func extractCallableSignatureParts(raw string) (params string, results string) {
	idx := strings.Index(raw, "(")
	if idx < 0 {
		return "", ""
	}
	sig := raw[idx:]
	paramText, rest, ok := splitFirstParen(sig)
	if !ok {
		return "", ""
	}
	params = compactParamList(splitTopLevel(paramText))
	rest = strings.TrimSpace(rest)
	if !strings.HasPrefix(rest, "->") {
		return params, ""
	}
	resRaw := strings.TrimSpace(strings.TrimPrefix(rest, "->"))
	if !strings.HasPrefix(resRaw, "(") {
		return params, " " + resRaw
	}
	resultText, _, ok := splitFirstParen(resRaw)
	if !ok {
		return params, ""
	}
	resultParts := splitTopLevel(resultText)
	switch len(resultParts) {
	case 0:
		return params, ""
	case 1:
		return params, " " + resultParts[0]
	default:
		return params, " (" + strings.Join(resultParts, ", ") + ")"
	}
}

func compactParamList(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	const maxInline = 72
	s := strings.Join(parts, ", ")
	if len(s) <= maxInline {
		return s
	}
	var short []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		fields := strings.Fields(p)
		if len(fields) >= 2 {
			short = append(short, fields[len(fields)-1])
		} else {
			short = append(short, p)
		}
	}
	s = strings.Join(short, ", ")
	if len(s) <= maxInline {
		return s
	}
	return "..."
}

// -----------------------------------------------------------------------
// Import path abbreviation
// -----------------------------------------------------------------------

var importPathRE = regexp.MustCompile(`([A-Za-z0-9_./~-]+/[A-Za-z0-9_./~-]+)\.([A-Za-z_][A-Za-z0-9_]*)`)

// abbreviateImportPaths shortens fully-qualified import paths in type strings
// to just the last path segment, e.g. "github.com/foo/bar.Baz" → "bar.Baz".
func abbreviateImportPaths(s string) string {
	return importPathRE.ReplaceAllStringFunc(s, func(match string) string {
		idx := strings.LastIndex(match, ".")
		if idx < 0 {
			return match
		}
		path := match[:idx]
		ident := match[idx+1:]
		pkg := path
		if slash := strings.LastIndex(pkg, "/"); slash >= 0 {
			pkg = pkg[slash+1:]
		}
		return pkg + "." + ident
	})
}

// -----------------------------------------------------------------------
// Label / name utilities
// -----------------------------------------------------------------------

func cleanAPILabel(fallback, label string) string {
	if strings.TrimSpace(label) == "" {
		return fallback
	}
	label = strings.ReplaceAll(label, "`", "")
	if strings.Contains(label, "Fields") {
		return "Field"
	}
	if strings.Contains(label, "Methods") {
		return "Method"
	}
	if strings.HasSuffix(label, "s") {
		label = strings.TrimSuffix(label, "s")
	}
	return label
}

func apiScope(label string) string {
	start := strings.Index(label, "`")
	if start < 0 {
		return ""
	}
	end := strings.Index(label[start+1:], "`")
	if end < 0 {
		return ""
	}
	return label[start+1 : start+1+end]
}

func apiSymbolName(x string) string {
	x = strings.TrimSpace(x)
	if x == "" {
		return "unknown"
	}
	if idx := strings.Index(x, "("); idx > 0 {
		return x[:idx]
	}
	if idx := strings.IndexByte(x, ' '); idx > 0 {
		name := x[:idx]
		if dot := strings.LastIndex(name, "."); dot >= 0 && dot+1 < len(name) {
			return name[dot+1:]
		}
		return name
	}
	if dot := strings.LastIndex(x, "."); dot >= 0 && dot+1 < len(x) {
		return x[dot+1:]
	}
	return x
}

func displayAPIName(kind, scope, name string) string {
	switch kind {
	case "Field", "Method":
		return compactScopedName(scope, name)
	default:
		return name
	}
}

func shortPackage(pkg string) string {
	parts := strings.Split(pkg, "/")
	if len(parts) <= 2 {
		return pkg
	}
	return strings.Join(parts[len(parts)-2:], "/")
}

func anchorID(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

func statusRank(status string) int {
	switch status {
	case "changed":
		return 0
	case "removed":
		return 1
	case "added":
		return 2
	default:
		return 3
	}
}

func kindRank(kind string) int {
	switch kind {
	case "Package":
		return 0
	case "Type":
		return 1
	case "Func":
		return 2
	case "Method":
		return 3
	case "Field":
		return 4
	case "Var":
		return 5
	case "Const":
		return 6
	default:
		return 7
	}
}

func apiKindOrder() []string {
	return []string{"Package", "Type", "Func", "Method", "Field", "Var", "Const"}
}

func pluralKind(kind string) string {
	switch kind {
	case "Package":
		return "Packages"
	case "Type":
		return "Types"
	case "Func":
		return "Functions"
	case "Method":
		return "Methods"
	case "Field":
		return "Fields"
	case "Var":
		return "Variables"
	case "Const":
		return "Constants"
	default:
		return kind
	}
}

func firstNonEmpty(xs ...string) string {
	for _, x := range xs {
		if strings.TrimSpace(x) != "" {
			return x
		}
	}
	return ""
}

func defaultString(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}
