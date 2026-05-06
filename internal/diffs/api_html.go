package diffs

import (
	"fmt"
	"html"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"
)

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
	Package string
	Kind    string
	Scope   string
	Name    string
	Old     string
	New     string
	Status  string
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

func (d *APIDiff) HTML(meta ReportMetadata) string {
	if meta.Now.IsZero() {
		meta.Now = time.Now()
	}

	changes := d.collectSymbolChanges()
	pkgSummaries := summarizeByPackage(changes)
	summary := summarizeOverall(changes, len(pkgSummaries))
	verdictClass := "verdict-ok"
	verdictTitle := "Compatible API changes only"
	verdictText := "No breaking API changes were detected."
	if summary.Breaking > 0 {
		verdictClass = "verdict-breaking"
		verdictTitle = "Breaking API changes detected"
		verdictText = "Review changed and removed public symbols before release."
	}

	var sb strings.Builder
	sb.WriteString("<!doctype html>\n<html lang=\"en\">\n<head>\n")
	sb.WriteString("<meta charset=\"utf-8\">\n")
	sb.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	sb.WriteString("<title>API Compatibility Report</title>\n")
	sb.WriteString(apiReportCSS())
	sb.WriteString("</head>\n<body>\n")
	sb.WriteString("<div class=\"layout\">\n")
	renderSidebar(&sb, pkgSummaries)
	sb.WriteString("<main>\n")

	hfprintf(&sb, "<header class=\"hero\">\n")
	hfprintf(&sb, "<div class=\"eyebrow\">relimpact api</div>\n")
	hfprintf(&sb, "<h1>API Compatibility Report</h1>\n")
	hfprintf(&sb, "<p class=\"sub\">%s · %s → %s · generated %s</p>\n",
		esc(defaultString(meta.Repo, "repository")),
		esc(defaultString(meta.OldRef, "old")),
		esc(defaultString(meta.NewRef, "new")),
		esc(meta.Now.Format("2006-01-02 15:04:05")),
	)
	hfprintf(&sb, "</header>\n")

	hfprintf(&sb, "<section class=\"verdict %s\">\n", verdictClass)
	hfprintf(&sb, "<div class=\"verdict-title\">%s</div>\n", esc(verdictTitle))
	hfprintf(&sb, "<div class=\"verdict-text\">%s</div>\n", esc(verdictText))
	sb.WriteString("</section>\n")

	renderSummaryCards(&sb, summary)

	if summary.Breaking > 0 {
		renderBreakingChangeSummary(&sb, changes)
	}

	renderAllPackageChanges(&sb, changes)

	sb.WriteString("</main>\n</div>\n</body>\n</html>\n")
	return sb.String()
}

// collectSymbolChanges builds the full flat list of SymbolChanges from the raw
// APIDiff, merging paired add+remove entries into "changed" records.
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
				Package: x.Path,
				Kind:    cleanKind,
				Scope:   scope,
				Name:    displayAPIName(cleanKind, scope, name),
				Old:     value,
				New:     value,
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

// mergeIntoChangedSymbols pairs up add+remove entries for the same symbol into
// a single "changed" record, leaving unpaired entries as-is.
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

func renderSidebar(w io.Writer, pkgs []PackageChangeSummary) {
	hfprintf(w, "<aside>\n")
	hfprintf(w, "<div class=\"brand\">relimpact</div>\n")
	hfprintf(w, "<div class=\"nav-title\">Packages</div>\n")
	if len(pkgs) == 0 {
		hfprintf(w, "<div class=\"empty-nav\">No API changes</div>\n")
		hfprintf(w, "</aside>\n")
		return
	}
	for _, pkg := range pkgs {
		navClass := "nav-item"
		badge := fmt.Sprintf("%d", pkg.Total)
		if pkg.Breaking > 0 {
			navClass = "nav-item nav-breaking"
			badge = fmt.Sprintf("%d breaking", pkg.Breaking)
		} else if pkg.Added > 0 {
			navClass = "nav-item nav-added"
		}
		hfprintf(w, "<a class=\"%s\" href=\"#%s\"><span title=\"%s\">%s</span><strong>%s</strong></a>\n",
			esc(navClass), esc(anchorID(pkg.Package)), esc(pkg.Package), esc(shortPackage(pkg.Package)), esc(badge))
	}
	hfprintf(w, "</aside>\n")
}

func renderSummaryCards(w io.Writer, s overallSummary) {
	hfprintf(w, "<section class=\"cards\">\n")
	renderCard(w, "Breaking", s.Breaking)
	renderCard(w, "Changed", s.Changed)
	renderCard(w, "Removed", s.Removed)
	renderCard(w, "Added", s.Added)
	renderCard(w, "Packages", s.ChangedPackages)
	hfprintf(w, "</section>\n")
}

func renderCard(w io.Writer, label string, value int) {
	hfprintf(w, "<div class=\"card\"><div class=\"num\">%d</div><div class=\"label\">%s</div></div>\n", value, esc(label))
}

// renderBreakingChangeSummary renders a compact cross-package list of all
// breaking changes (changed + removed) at the top of the report body so
// reviewers can grasp the full scope before reading per-package details.
func renderBreakingChangeSummary(w io.Writer, changes []SymbolChange) {
	var breaking []SymbolChange
	for _, ch := range changes {
		if ch.Status == "changed" || ch.Status == "removed" {
			breaking = append(breaking, ch)
		}
	}
	if len(breaking) == 0 {
		return
	}

	hfprintf(w, "<section class=\"breaking-summary\">\n")
	hfprintf(w, "<div class=\"breaking-summary-title\">Breaking changes at a glance</div>\n")
	hfprintf(w, "<table class=\"breaking-table\">\n")
	hfprintf(w, "<thead><tr><th>Symbol</th><th>Kind</th><th>Package</th><th>Change</th></tr></thead>\n")
	hfprintf(w, "<tbody>\n")
	for _, ch := range breaking {
		statusLabel := "Removed"
		statusCls := "badge badge-breaking"
		if ch.Status == "changed" {
			statusLabel = "Changed"
			statusCls = "badge badge-changed"
		}
		hfprintf(w, "<tr><td><code>%s</code></td><td>%s</td><td><code class=\"pkg-path\">%s</code></td><td><span class=\"%s\">%s</span></td></tr>\n",
			esc(ch.Name), esc(ch.Kind), esc(shortPackage(ch.Package)), esc(statusCls), esc(statusLabel))
	}
	hfprintf(w, "</tbody>\n</table>\n</section>\n")
}

func renderAllPackageChanges(w io.Writer, changes []SymbolChange) {
	if len(changes) == 0 {
		hfprintf(w, "<section class=\"empty\">No public API changes detected.</section>\n")
		return
	}

	byPkg := make(map[string][]SymbolChange)
	for _, ch := range changes {
		byPkg[ch.Package] = append(byPkg[ch.Package], ch)
	}
	pkgs := make([]string, 0, len(byPkg))
	for pkg := range byPkg {
		pkgs = append(pkgs, pkg)
	}
	sort.Strings(pkgs)

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

		s := summarizeByPackage(pkgChanges)[0]
		hfprintf(w, "<section class=\"pkg\" id=\"%s\">\n", esc(anchorID(pkg)))
		hfprintf(w, "<div class=\"pkg-head\"><div><h2>%s</h2><div class=\"pkg-full\">%s</div></div><div class=\"pkg-counts\">", esc(shortPackage(pkg)), esc(pkg))
		renderInlineCount(w, "breaking", s.Breaking)
		renderInlineCount(w, "changed", s.Changed)
		renderInlineCount(w, "removed", s.Removed)
		renderInlineCount(w, "added", s.Added)
		hfprintf(w, "</div></div>\n")

		renderStatusSection(w, "Changed signatures", pkgChanges, "changed")
		renderStatusSection(w, "Removed API", pkgChanges, "removed")
		renderStatusSection(w, "Added API", pkgChanges, "added")
		hfprintf(w, "</section>\n")
	}
}

func renderInlineCount(w io.Writer, label string, value int) {
	if value == 0 {
		return
	}
	cls := "count-neutral"
	switch label {
	case "breaking", "removed":
		cls = "count-breaking"
	case "added":
		cls = "count-added"
	case "changed":
		cls = "count-changed"
	}
	hfprintf(w, "<span class=\"%s\">%d %s</span>", cls, value, esc(label))
}

func renderStatusSection(w io.Writer, title string, changes []SymbolChange, status string) {
	var group []SymbolChange
	for _, ch := range changes {
		if ch.Status == status {
			group = append(group, ch)
		}
	}
	if len(group) == 0 {
		return
	}

	accentClass := "group-accent-added"
	if status == "changed" {
		accentClass = "group-accent-changed"
	} else if status == "removed" {
		accentClass = "group-accent-removed"
	}

	hfprintf(w, "<div class=\"change-group %s\"><div class=\"group-title\">%s</div>\n", accentClass, esc(title))
	if status == "changed" {
		for _, ch := range group {
			renderSignatureChangeCard(w, ch)
		}
	} else {
		renderKindGroups(w, group, status)
	}
	hfprintf(w, "</div>\n")
}

func renderKindGroups(w io.Writer, changes []SymbolChange, status string) {
	byKind := make(map[string][]SymbolChange)
	for _, ch := range changes {
		byKind[ch.Kind] = append(byKind[ch.Kind], ch)
	}

	for _, kind := range apiKindOrder() {
		xs := byKind[kind]
		if len(xs) == 0 {
			continue
		}

		sort.Slice(xs, func(i, j int) bool {
			return xs[i].Name < xs[j].Name
		})

		if kind == "Field" {
			renderStructFieldDiffs(w, xs, status)
			continue
		}
		renderKindGroup(w, kind, xs, status)
	}
}

func renderKindGroup(w io.Writer, kind string, changes []SymbolChange, status string) {
	badgeClass := "compact-added"
	if status == "removed" {
		badgeClass = "compact-removed"
	}

	hfprintf(w, "<section class=\"compact-kind\">\n")
	hfprintf(w, "<div class=\"compact-kind-head\"><span>%s</span><strong class=\"%s\">%d</strong></div>\n",
		esc(pluralKind(kind)), esc(badgeClass+"-count"), len(changes))
	hfprintf(w, "<div class=\"compact-list %s\">\n", esc(badgeClass))
	for _, ch := range changes {
		renderCompactSymbolRow(w, ch)
	}
	hfprintf(w, "</div>\n")
	hfprintf(w, "</section>\n")
}

func renderCompactSymbolRow(w io.Writer, ch SymbolChange) {
	hfprintf(w, "<div class=\"compact-row\">\n")

	switch ch.Kind {
	case "Func":
		hfprintf(w, "<code>%s</code>", esc(compactFunc(ch)))
	case "Method":
		hfprintf(w, "<code>%s</code>", esc(compactMethod(ch)))
	case "Var", "Const":
		name, typ := compactTypedSymbol(ch)
		hfprintf(w, "<code>%s</code>", esc(name))
		if typ != "" {
			hfprintf(w, "<span class=\"compact-type\">%s</span>", esc(typ))
		}
	default:
		hfprintf(w, "<code>%s</code>", esc(formatCompactSymbol(ch, "")))
	}

	hfprintf(w, "\n</div>\n")
}

func renderStructFieldDiffs(w io.Writer, changes []SymbolChange, status string) {
	grouped := groupFieldsByOwner(changes, status)
	if len(grouped) == 0 {
		renderKindGroup(w, "Field", changes, status)
		return
	}

	owners := make([]string, 0, len(grouped))
	for owner := range grouped {
		owners = append(owners, owner)
	}
	sort.Strings(owners)

	hfprintf(w, "<section class=\"struct-diff-group\">\n")
	hfprintf(w, "<div class=\"compact-kind-head\"><span>Struct fields</span><strong>%d</strong></div>\n", len(changes))
	for _, owner := range owners {
		renderStructFieldDiff(w, owner, grouped[owner], status)
	}
	hfprintf(w, "</section>\n")
}

func renderStructFieldDiff(w io.Writer, owner string, fields []StructFieldChange, status string) {
	hfprintf(w, "<article class=\"struct-diff\">\n")
	hfprintf(w, "<div class=\"struct-diff-title\">type %s struct</div>\n", esc(owner))
	hfprintf(w, "<pre class=\"struct-pre\">%s</pre>\n", formatStructFieldDiff(owner, fields, status))
	hfprintf(w, "</article>\n")
}

// renderSignatureChangeCard renders a full before/after card for a changed
// symbol using a unified diff view with colored prefix lines.
func renderSignatureChangeCard(w io.Writer, ch SymbolChange) {
	subtitle := ch.Kind + " signature"

	hfprintf(w, "<article class=\"change\">\n")
	hfprintf(w, "<div class=\"change-head\"><div><div class=\"kind\">%s</div><div class=\"symbol\">%s</div></div><span class=\"badge badge-changed\">Changed</span></div>\n",
		esc(subtitle), esc(ch.Name))

	oldFmt := formatAPIValue(ch.Kind, ch.Scope, ch.Old)
	newFmt := formatAPIValue(ch.Kind, ch.Scope, ch.New)

	hfprintf(w, "<div class=\"sig-diff\">\n")
	renderUnifiedDiff(w, oldFmt, newFmt)
	hfprintf(w, "</div>\n")

	hfprintf(w, "</article>\n")
}

// renderUnifiedDiff renders old and new signature strings as a unified diff
// block: removed lines in red with "−", added lines in green with "+", and
// shared context lines in muted blue-gray.
func renderUnifiedDiff(w io.Writer, oldSig, newSig string) {
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

	hfprintf(w, "<pre class=\"sig-unified\">")
	for _, line := range oldLines {
		if !newSet[line] {
			hfprintf(w, "<span class=\"diff-removed\">− %s\n</span>", esc(line))
		}
	}
	for _, line := range oldLines {
		if newSet[line] {
			hfprintf(w, "<span class=\"diff-context\">  %s\n</span>", esc(line))
		}
	}
	for _, line := range newLines {
		if !oldSet[line] {
			hfprintf(w, "<span class=\"diff-added\">+ %s\n</span>", esc(line))
		}
	}
	hfprintf(w, "</pre>\n")
}

// formatStructFieldDiff renders the annotated struct body for a group of
// added or removed fields, with HTML-colored prefix markers.
func formatStructFieldDiff(owner string, fields []StructFieldChange, status string) string {
	width := 0
	for _, f := range fields {
		if len(f.Name) > width {
			width = len(f.Name)
		}
	}

	prefix := "+"
	spanClass := "diff-added"
	if status == "removed" {
		prefix = "-"
		spanClass = "diff-removed"
	}

	var b strings.Builder
	b.WriteString(`<span class="diff-context">type `)
	b.WriteString(html.EscapeString(owner))
	b.WriteString(" struct {\n</span>")
	for _, f := range fields {
		padding := width - len(f.Name) + 1
		if padding < 1 {
			padding = 1
		}
		b.WriteString(`<span class="`)
		b.WriteString(spanClass)
		b.WriteString(`">`)
		b.WriteString(prefix)
		b.WriteString("   ")
		b.WriteString(html.EscapeString(f.Name))
		b.WriteString(strings.Repeat(" ", padding))
		b.WriteString(html.EscapeString(f.Type))
		b.WriteString("\n</span>")
	}
	b.WriteString(`<span class="diff-context">}</span>`)
	return b.String()
}

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
					results = ""
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
	start := 0
	parenDepth := 0
	bracketDepth := 0
	braceDepth := 0
	for i, r := range s {
		switch r {
		case '(':
			parenDepth++
		case ')':
			parenDepth--
		case '[':
			bracketDepth++
		case ']':
			bracketDepth--
		case '{':
			braceDepth++
		case '}':
			braceDepth--
		case ',':
			if parenDepth == 0 && bracketDepth == 0 && braceDepth == 0 {
				out = append(out, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
	}
	out = append(out, strings.TrimSpace(s[start:]))
	return out
}

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

// formatCompactSymbol returns the short display string for a symbol in a
// compact (add/remove) list row.
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

// extractCallableSignatureParts parses the parameter list and return types
// from a raw callable signature string.
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

func apiReportCSS() string {
	return `<style>
:root {
  --bg: #f6f7fb;
  --panel: #ffffff;
  --text: #172033;
  --muted: #667085;
  --border: #e6e8ef;
  --soft: #f8fafc;
  --red: #b42318;
  --red-bg: #fff1f1;
  --red-border: #fecaca;
  --green: #087443;
  --green-bg: #ecfdf3;
  --green-border: #bbf7d0;
  --amber: #b45309;
  --amber-bg: #fffbeb;
  --amber-border: #fde68a;
  --shadow: 0 12px 28px rgba(15, 23, 42, 0.06);
}
* { box-sizing: border-box; }
body {
  margin: 0;
  background: var(--bg);
  color: var(--text);
  font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}
.layout {
  display: grid;
  grid-template-columns: 292px minmax(0, 1fr);
  min-height: 100vh;
}
aside {
  position: sticky;
  top: 0;
  height: 100vh;
  overflow: auto;
  padding: 24px;
  background: var(--panel);
  border-right: 1px solid var(--border);
}
main {
  width: min(1160px, 100%);
  padding: 36px 44px 64px;
}
.hero { margin-bottom: 0; }
.eyebrow {
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.12em;
  font-weight: 850;
  font-size: 12px;
  margin-bottom: 10px;
}
h1 {
  margin: 0;
  font-size: 34px;
  line-height: 1.1;
  letter-spacing: -0.045em;
}
.sub {
  margin: 10px 0 0;
  color: var(--muted);
  font-size: 14px;
}
.brand {
  font-weight: 850;
  letter-spacing: -0.04em;
  font-size: 22px;
  margin-bottom: 28px;
}
.nav-title {
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-weight: 850;
  font-size: 12px;
  margin-bottom: 10px;
}
.nav-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 14px;
  padding: 10px 0;
  text-decoration: none;
  color: var(--text);
  border-bottom: 1px solid #f1f3f7;
  font-size: 13px;
}
.nav-item span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.nav-item strong { color: var(--muted); font-size: 12px; white-space: nowrap; }
.nav-breaking strong { color: var(--red); font-weight: 850; }
.nav-breaking::before {
  content: "";
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--red);
  margin-right: 6px;
  flex-shrink: 0;
}
.nav-added strong { color: var(--green); }
.empty-nav { color: var(--muted); font-size: 14px; }
.verdict {
  margin-top: 26px;
  border-radius: 18px;
  padding: 18px 20px;
  border: 1px solid var(--border);
  background: var(--panel);
  box-shadow: var(--shadow);
}
.verdict-breaking { background: var(--red-bg); border-color: var(--red-border); color: #7f1d1d; }
.verdict-ok { background: var(--green-bg); border-color: var(--green-border); color: #064e3b; }
.verdict-title { font-size: 17px; font-weight: 850; }
.verdict-text { margin-top: 4px; font-size: 14px; }
.cards {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 12px;
  margin: 24px 0 34px;
}
.card {
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 18px;
  padding: 16px;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
}
.num { font-size: 32px; line-height: 1; font-weight: 850; letter-spacing: -0.05em; }
.label { color: var(--muted); font-size: 13px; margin-top: 8px; }
.breaking-summary {
  background: var(--red-bg);
  border: 1px solid var(--red-border);
  border-radius: 18px;
  padding: 18px 20px;
  margin-bottom: 34px;
  box-shadow: var(--shadow);
}
.breaking-summary-title {
  font-size: 14px;
  font-weight: 850;
  color: var(--red);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  margin-bottom: 14px;
}
.breaking-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.breaking-table th {
  text-align: left;
  color: var(--muted);
  font-size: 11px;
  font-weight: 850;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  padding: 0 12px 10px 0;
  border-bottom: 1px solid var(--red-border);
}
.breaking-table td {
  padding: 8px 12px 8px 0;
  border-bottom: 1px solid rgba(254,202,202,0.5);
  vertical-align: middle;
}
.breaking-table tr:last-child td { border-bottom: none; }
.breaking-table code { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-size: 13px; }
.pkg-path { color: var(--muted); font-size: 12px; }
.pkg { margin-top: 34px; scroll-margin-top: 24px; }
.pkg-head {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  align-items: flex-end;
  margin-bottom: 18px;
}
.pkg h2 { margin: 0; font-size: 23px; letter-spacing: -0.03em; }
.pkg-full {
  margin-top: 5px;
  color: var(--muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
}
.pkg-counts { display: flex; flex-wrap: wrap; gap: 6px; justify-content: flex-end; }
.pkg-counts span { border-radius: 999px; padding: 5px 8px; font-size: 12px; font-weight: 750; }
.count-neutral { border: 1px solid var(--border); background: var(--panel); color: var(--muted); }
.count-breaking { background: var(--red-bg); color: var(--red); border: 1px solid var(--red-border); }
.count-changed { background: var(--amber-bg); color: var(--amber); border: 1px solid var(--amber-border); }
.count-added { background: var(--green-bg); color: var(--green); border: 1px solid var(--green-border); }
.change-group { margin-top: 18px; border-left: 3px solid transparent; padding-left: 12px; }
.group-accent-changed { border-left-color: var(--amber); }
.group-accent-removed { border-left-color: var(--red); }
.group-accent-added   { border-left-color: var(--green); }
.group-title {
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-weight: 850;
  font-size: 12px;
  margin: 18px 0 10px;
}
.change {
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 18px;
  padding: 18px;
  margin-bottom: 12px;
  box-shadow: var(--shadow);
}
.change-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
  margin-bottom: 14px;
}
.kind {
  color: var(--muted);
  font-size: 12px;
  font-weight: 850;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  margin-bottom: 5px;
}
.symbol { font-size: 16px; font-weight: 850; letter-spacing: -0.02em; }
.badge {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 6px 9px;
  font-size: 11px;
  font-weight: 850;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  white-space: nowrap;
}
.badge-breaking { background: var(--red-bg); color: var(--red); }
.badge-added { background: var(--green-bg); color: var(--green); }
.badge-changed { background: var(--amber-bg); color: var(--amber); }
.sig-diff { margin-top: 6px; }
.sig-unified {
  margin: 0;
  padding: 14px;
  background: #0b1020;
  border-radius: 14px;
  overflow-x: auto;
  white-space: pre;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, "Liberation Mono", monospace;
  font-size: 13px;
  line-height: 1.6;
}
.diff-removed { color: #ff8080; background: rgba(180,35,24,0.18); display: block; border-radius: 3px; }
.diff-added   { color: #6ee7b7; background: rgba(8,116,67,0.2);   display: block; border-radius: 3px; }
.diff-context { color: #8899bb; display: block; }
pre {
  margin: 0;
  padding: 14px;
  background: #0b1020;
  color: #e8eefc;
  border-radius: 14px;
  overflow-x: auto;
  white-space: pre;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, "Liberation Mono", monospace;
  font-size: 13px;
  line-height: 1.55;
}
.compact-kind,
.struct-diff-group {
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 18px;
  margin-bottom: 12px;
  overflow: hidden;
  box-shadow: var(--shadow);
}
.compact-kind-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 13px 16px;
  background: var(--soft);
  border-bottom: 1px solid var(--border);
}
.compact-kind-head span { font-size: 13px; font-weight: 850; letter-spacing: 0.02em; }
.compact-kind-head strong { font-size: 12px; font-weight: 850; border-radius: 999px; padding: 2px 8px; }
.compact-added-count   { background: var(--green-bg); color: var(--green); }
.compact-removed-count { background: var(--red-bg);   color: var(--red); }
.compact-list { padding: 8px; }
.compact-row {
  display: grid;
  grid-template-columns: minmax(220px, 1fr) auto;
  align-items: baseline;
  gap: 16px;
  padding: 9px 10px;
  border-radius: 12px;
}
.compact-row + .compact-row { margin-top: 2px; }
.compact-row:hover { background: #f8fafc; }
.compact-row code,
.struct-diff-title {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, "Liberation Mono", monospace;
}
.compact-row code { font-size: 13px; color: var(--text); word-break: break-word; }
.compact-type {
  color: var(--muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, "Liberation Mono", monospace;
  font-size: 12px;
  text-align: right;
  white-space: nowrap;
}
.compact-added .compact-row code::before {
  content: "+";
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  margin-right: 8px;
  color: var(--green);
  font-weight: 900;
}
.compact-removed .compact-row code::before {
  content: "−";
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  margin-right: 8px;
  color: var(--red);
  font-weight: 900;
}
.struct-diff { padding: 14px 16px 16px; }
.struct-diff + .struct-diff { border-top: 1px solid var(--border); }
.struct-diff-title { font-size: 13px; font-weight: 850; margin-bottom: 8px; color: var(--text); }
.struct-pre { margin: 0; padding: 14px; background: #0b1020; border-radius: 14px; }
.empty {
  margin-top: 28px;
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 18px;
  padding: 24px;
  color: var(--muted);
}
@media (max-width: 900px) {
  .layout { display: block; }
  aside { position: static; width: auto; height: auto; }
  main { padding: 28px 20px 44px; }
  .cards { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .pkg-head { display: block; }
  .pkg-counts { justify-content: flex-start; margin-top: 10px; }
  .compact-row { grid-template-columns: 1fr; gap: 6px; }
  .compact-type { text-align: left; white-space: normal; }
}
</style>
`
}

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

func esc(s string) string {
	return html.EscapeString(s)
}

func defaultString(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}
