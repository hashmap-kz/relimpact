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

type APIReportMeta struct {
	Repo   string
	OldRef string
	NewRef string
	Now    time.Time
}

type apiChange struct {
	Package string
	Kind    string
	Scope   string
	Name    string
	Old     string
	New     string
	Status  string
}

type apiFieldChange struct {
	Owner  string
	Name   string
	Type   string
	Status string
}

type apiPackageSummary struct {
	Package  string
	Breaking int
	Changed  int
	Removed  int
	Added    int
	Total    int
}

type apiSummary struct {
	Breaking        int
	Changed         int
	Removed         int
	Added           int
	PackagesAdded   int
	PackagesRemoved int
	ChangedPackages int
	Total           int
}

func (d *APIDiff) HTML(meta APIReportMeta) string {
	if meta.Now.IsZero() {
		meta.Now = time.Now()
	}

	changes := d.apiChanges()
	pkgSummaries := summarizeAPIPackages(changes)
	summary := summarizeAPI(changes, len(pkgSummaries))
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
	writeAPISidebar(&sb, pkgSummaries)
	sb.WriteString("<main>\n")

	xFprintf(&sb, "<header class=\"hero\">\n")
	xFprintf(&sb, "<div class=\"eyebrow\">relimpact api</div>\n")
	xFprintf(&sb, "<h1>API Compatibility Report</h1>\n")
	xFprintf(&sb, "<p class=\"sub\">%s · %s → %s · generated %s</p>\n",
		esc(defaultString(meta.Repo, "repository")),
		esc(defaultString(meta.OldRef, "old")),
		esc(defaultString(meta.NewRef, "new")),
		esc(meta.Now.Format("2006-01-02 15:04:05")),
	)
	xFprintf(&sb, "</header>\n")

	xFprintf(&sb, "<section class=\"verdict %s\">\n", verdictClass)
	xFprintf(&sb, "<div class=\"verdict-title\">%s</div>\n", esc(verdictTitle))
	xFprintf(&sb, "<div class=\"verdict-text\">%s</div>\n", esc(verdictText))
	sb.WriteString("</section>\n")

	writeAPISummaryCards(&sb, summary)
	writeAPIChanges(&sb, changes)

	sb.WriteString("</main>\n</div>\n</body>\n</html>\n")
	return sb.String()
}

func (d *APIDiff) apiChanges() []apiChange {
	var changes []apiChange

	addPkgChanges := func(status string, xs []string) {
		for _, x := range xs {
			ch := apiChange{
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

	addChanges := func(status, kind string, xs []APIDiffRes) {
		for _, x := range xs {
			cleanKind := cleanAPILabel(kind, x.Label)
			scope := apiScope(x.Label)
			name := apiSymbolName(x.X)
			value := x.X

			ch := apiChange{
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

	return compactAPIChanges(changes)
}

func compactAPIChanges(raw []apiChange) []apiChange {
	added := make(map[string]apiChange)
	removed := make(map[string]apiChange)
	var out []apiChange

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
		out = append(out, apiChange{
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

func (ch apiChange) changeKey() string {
	return ch.Package + "\x00" + ch.Kind + "\x00" + ch.Scope + "\x00" + apiSymbolName(firstNonEmpty(ch.New, ch.Old))
}

func summarizeAPI(changes []apiChange, changedPackages int) apiSummary {
	var s apiSummary
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

func summarizeAPIPackages(changes []apiChange) []apiPackageSummary {
	byPkg := make(map[string]*apiPackageSummary)
	for _, ch := range changes {
		pkg := ch.Package
		if pkg == "" {
			pkg = "unknown"
		}
		s, ok := byPkg[pkg]
		if !ok {
			s = &apiPackageSummary{Package: pkg}
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
	out := make([]apiPackageSummary, 0, len(byPkg))
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

func writeAPISidebar(w io.Writer, pkgs []apiPackageSummary) {
	xFprintf(w, "<aside>\n")
	xFprintf(w, "<div class=\"brand\">relimpact</div>\n")
	xFprintf(w, "<div class=\"nav-title\">Packages</div>\n")
	if len(pkgs) == 0 {
		xFprintf(w, "<div class=\"empty-nav\">No API changes</div>\n")
		xFprintf(w, "</aside>\n")
		return
	}
	for _, pkg := range pkgs {
		badge := fmt.Sprintf("%d", pkg.Total)
		if pkg.Breaking > 0 {
			badge = fmt.Sprintf("%d breaking", pkg.Breaking)
		}
		xFprintf(w, "<a class=\"nav-item\" href=\"#%s\"><span title=\"%s\">%s</span><strong>%s</strong></a>\n",
			esc(anchorID(pkg.Package)), esc(pkg.Package), esc(shortPackage(pkg.Package)), esc(badge))
	}
	xFprintf(w, "</aside>\n")
}

func writeAPISummaryCards(w io.Writer, s apiSummary) {
	xFprintf(w, "<section class=\"cards\">\n")
	writeAPICard(w, "Breaking", s.Breaking)
	writeAPICard(w, "Changed", s.Changed)
	writeAPICard(w, "Removed", s.Removed)
	writeAPICard(w, "Added", s.Added)
	writeAPICard(w, "Packages", s.ChangedPackages)
	xFprintf(w, "</section>\n")
}

func writeAPICard(w io.Writer, label string, value int) {
	xFprintf(w, "<div class=\"card\"><div class=\"num\">%d</div><div class=\"label\">%s</div></div>\n", value, esc(label))
}

func writeAPIChanges(w io.Writer, changes []apiChange) {
	if len(changes) == 0 {
		xFprintf(w, "<section class=\"empty\">No public API changes detected.</section>\n")
		return
	}

	byPkg := make(map[string][]apiChange)
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

		s := summarizeAPIPackages(pkgChanges)[0]
		xFprintf(w, "<section class=\"pkg\" id=\"%s\">\n", esc(anchorID(pkg)))
		xFprintf(w, "<div class=\"pkg-head\"><div><h2>%s</h2><div class=\"pkg-full\">%s</div></div><div class=\"pkg-counts\">", esc(shortPackage(pkg)), esc(pkg))
		writeTinyCount(w, "breaking", s.Breaking)
		writeTinyCount(w, "changed", s.Changed)
		writeTinyCount(w, "removed", s.Removed)
		writeTinyCount(w, "added", s.Added)
		xFprintf(w, "</div></div>\n")

		writeStatusGroup(w, "Changed signatures", pkgChanges, "changed")
		writeStatusGroup(w, "Removed API", pkgChanges, "removed")
		writeStatusGroup(w, "Added API", pkgChanges, "added")
		xFprintf(w, "</section>\n")
	}
}

func writeTinyCount(w io.Writer, label string, value int) {
	if value == 0 {
		return
	}
	xFprintf(w, "<span>%d %s</span>", value, esc(label))
}

func writeStatusGroup(w io.Writer, title string, changes []apiChange, status string) {
	var group []apiChange
	for _, ch := range changes {
		if ch.Status == status {
			group = append(group, ch)
		}
	}
	if len(group) == 0 {
		return
	}

	xFprintf(w, "<div class=\"change-group\"><div class=\"group-title\">%s</div>\n", esc(title))
	if status == "changed" {
		for _, ch := range group {
			writeAPIChangeCard(w, ch)
		}
	} else {
		writeCompactAPIGroup(w, group, status)
	}
	xFprintf(w, "</div>\n")
}

func writeCompactAPIGroup(w io.Writer, changes []apiChange, status string) {
	byKind := make(map[string][]apiChange)
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
			writeStructFieldDiffs(w, xs, status)
			continue
		}
		writeCompactKindGroup(w, kind, xs, status)
	}
}

func writeCompactKindGroup(w io.Writer, kind string, changes []apiChange, status string) {
	badgeClass := "compact-added"
	if status == "removed" {
		badgeClass = "compact-removed"
	}

	xFprintf(w, "<section class=\"compact-kind\">\n")
	xFprintf(w, "<div class=\"compact-kind-head\"><span>%s</span><strong>%d</strong></div>\n", esc(pluralKind(kind)), len(changes))
	xFprintf(w, "<div class=\"compact-list %s\">\n", esc(badgeClass))
	for _, ch := range changes {
		writeCompactChange(w, ch)
	}
	xFprintf(w, "</div>\n")
	xFprintf(w, "</section>\n")
}

func writeCompactChange(w io.Writer, ch apiChange) {
	xFprintf(w, "<div class=\"compact-row\">\n")

	switch ch.Kind {
	case "Func":
		xFprintf(w, "<code>%s</code>", esc(compactFunc(ch)))
	case "Method":
		xFprintf(w, "<code>%s</code>", esc(compactMethod(ch)))
	case "Var", "Const":
		name, typ := compactTypedSymbol(ch)
		xFprintf(w, "<code>%s</code>", esc(name))
		if typ != "" {
			xFprintf(w, "<span class=\"compact-type\">%s</span>", esc(typ))
		}
	default:
		xFprintf(w, "<code>%s</code>", esc(compactDisplayValue(ch, "")))
	}

	xFprintf(w, "\n</div>\n")
}

func writeStructFieldDiffs(w io.Writer, changes []apiChange, status string) {
	grouped := groupFieldChanges(changes, status)
	if len(grouped) == 0 {
		writeCompactKindGroup(w, "Field", changes, status)
		return
	}

	owners := make([]string, 0, len(grouped))
	for owner := range grouped {
		owners = append(owners, owner)
	}
	sort.Strings(owners)

	xFprintf(w, "<section class=\"struct-diff-group\">\n")
	xFprintf(w, "<div class=\"compact-kind-head\"><span>Struct fields</span><strong>%d</strong></div>\n", len(changes))
	for _, owner := range owners {
		writeStructFieldDiff(w, owner, grouped[owner], status)
	}
	xFprintf(w, "</section>\n")
}

func writeStructFieldDiff(w io.Writer, owner string, fields []apiFieldChange, status string) {
	xFprintf(w, "<article class=\"struct-diff\">\n")
	xFprintf(w, "<div class=\"struct-diff-title\">type %s struct</div>\n", esc(owner))
	xFprintf(w, "<pre class=\"struct-pre\">%s</pre>\n", esc(formatStructFieldDiff(owner, fields, status)))
	xFprintf(w, "</article>\n")
}

func writeAPIChangeCard(w io.Writer, ch apiChange) {
	badgeClass := "badge-added"
	badgeText := "Added"
	subtitle := ch.Kind
	switch ch.Status {
	case "changed":
		badgeClass = "badge-changed"
		badgeText = "Changed"
		subtitle = ch.Kind + " signature"
	case "removed":
		badgeClass = "badge-breaking"
		badgeText = "Removed"
	}

	xFprintf(w, "<article class=\"change\">\n")
	xFprintf(w, "<div class=\"change-head\"><div><div class=\"kind\">%s</div><div class=\"symbol\">%s</div></div><span class=\"badge %s\">%s</span></div>\n",
		esc(subtitle), esc(ch.Name), esc(badgeClass), esc(badgeText))

	switch ch.Status {
	case "changed":
		writeSignatureBlock(w, "Before", formatAPIValue(ch.Kind, ch.Scope, ch.Old))
		writeSignatureBlock(w, "After", formatAPIValue(ch.Kind, ch.Scope, ch.New))
	case "removed":
		writeSignatureBlock(w, "Removed", formatAPIValue(ch.Kind, ch.Scope, ch.Old))
	default:
		writeSignatureBlock(w, "Added", formatAPIValue(ch.Kind, ch.Scope, ch.New))
	}
	xFprintf(w, "</article>\n")
}

func writeSignatureBlock(w io.Writer, label, value string) {
	xFprintf(w, "<div class=\"sig-label\">%s</div><pre>%s</pre>\n", esc(label), esc(value))
}

func formatStructFieldDiff(owner string, fields []apiFieldChange, status string) string {
	width := 0
	for _, f := range fields {
		if len(f.Name) > width {
			width = len(f.Name)
		}
	}

	prefix := "+"
	if status == "removed" {
		prefix = "-"
	}

	var b strings.Builder
	b.WriteString("type ")
	b.WriteString(owner)
	b.WriteString(" struct {\n")
	for _, f := range fields {
		b.WriteString(prefix)
		b.WriteString("   ")
		b.WriteString(f.Name)
		padding := width - len(f.Name) + 1
		if padding < 1 {
			padding = 1
		}
		b.WriteString(strings.Repeat(" ", padding))
		b.WriteString(f.Type)
		b.WriteString("\n")
	}
	b.WriteString("}")
	return b.String()
}

func fieldChangeFromAPIChange(ch apiChange) (apiFieldChange, bool) {
	raw := shortenGoTypes(strings.TrimSpace(firstNonEmpty(ch.New, ch.Old)))
	if raw == "" {
		return apiFieldChange{}, false
	}

	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return apiFieldChange{}, false
	}

	fullName := parts[0]
	fieldType := strings.TrimSpace(strings.TrimPrefix(raw, fullName))
	dot := strings.LastIndex(fullName, ".")
	if dot < 0 || dot+1 >= len(fullName) {
		if ch.Scope == "" || ch.Name == "" {
			return apiFieldChange{}, false
		}
		return apiFieldChange{Owner: ch.Scope, Name: ch.Name, Type: fieldType, Status: ch.Status}, true
	}

	return apiFieldChange{Owner: fullName[:dot], Name: fullName[dot+1:], Type: fieldType, Status: ch.Status}, true
}

func groupFieldChanges(changes []apiChange, status string) map[string][]apiFieldChange {
	out := make(map[string][]apiFieldChange)
	for _, ch := range changes {
		if ch.Status != status || ch.Kind != "Field" {
			continue
		}
		f, ok := fieldChangeFromAPIChange(ch)
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
	raw = shortenGoTypes(strings.TrimSpace(raw))
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
		funcHead += "(" + shortenGoTypes(scope) + ") "
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
	name, typ := compactField(apiChange{Kind: "Field", Scope: scope, Name: apiSymbolName(raw), New: raw})
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

var goTypePathRE = regexp.MustCompile(`([A-Za-z0-9_./~-]+/[A-Za-z0-9_./~-]+)\.([A-Za-z_][A-Za-z0-9_]*)`)

func shortenGoTypes(s string) string {
	return goTypePathRE.ReplaceAllStringFunc(s, func(match string) string {
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

func compactDisplayValue(ch apiChange, _ string) string {
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

func compactField(ch apiChange) (name string, typ string) {
	raw := shortenGoTypes(strings.TrimSpace(firstNonEmpty(ch.New, ch.Old)))
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

func compactTypedSymbol(ch apiChange) (name string, typ string) {
	raw := shortenGoTypes(strings.TrimSpace(firstNonEmpty(ch.New, ch.Old)))
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return ch.Name, ""
	}
	name = parts[0]
	typ = strings.TrimSpace(strings.TrimPrefix(raw, name))
	return name, typ
}

func compactFunc(ch apiChange) string {
	raw := shortenGoTypes(strings.TrimSpace(firstNonEmpty(ch.New, ch.Old)))
	name := apiSymbolName(raw)
	if name == "" || name == "unknown" {
		name = ch.Name
	}
	params, results := compactCallableParts(raw)
	if params == "" && results == "" {
		return name + "(...)"
	}
	return name + "(" + params + ")" + results
}

func compactMethod(ch apiChange) string {
	raw := shortenGoTypes(strings.TrimSpace(firstNonEmpty(ch.New, ch.Old)))
	name := apiSymbolName(raw)
	if name == "" || name == "unknown" {
		name = ch.Name
	}
	fullName := name
	if ch.Scope != "" && !strings.Contains(name, ".") {
		fullName = ch.Scope + "." + name
	}
	params, results := compactCallableParts(raw)
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

func compactCallableParts(raw string) (params string, results string) {
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
  --green: #087443;
  --green-bg: #ecfdf3;
  --amber: #b45309;
  --amber-bg: #fffbeb;
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
.nav-item span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.nav-item strong {
  color: var(--muted);
  font-size: 12px;
  white-space: nowrap;
}
.empty-nav { color: var(--muted); font-size: 14px; }
.verdict {
  margin-top: 26px;
  border-radius: 18px;
  padding: 18px 20px;
  border: 1px solid var(--border);
  background: var(--panel);
  box-shadow: var(--shadow);
}
.verdict-breaking { background: var(--red-bg); border-color: #fecaca; color: #7f1d1d; }
.verdict-ok { background: var(--green-bg); border-color: #bbf7d0; color: #064e3b; }
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
.pkg-counts span {
  border: 1px solid var(--border);
  background: var(--panel);
  border-radius: 999px;
  padding: 5px 8px;
  color: var(--muted);
  font-size: 12px;
  font-weight: 750;
}
.change-group { margin-top: 18px; }
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
.sig-label {
  color: var(--muted);
  font-size: 11px;
  font-weight: 850;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  margin: 12px 0 6px;
}
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
.compact-kind-head strong { color: var(--muted); font-size: 12px; font-weight: 850; }
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
.struct-diff-title {
  font-size: 13px;
  font-weight: 850;
  margin-bottom: 8px;
  color: var(--text);
}
.struct-pre { margin: 0; padding: 14px; }
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
