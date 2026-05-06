package diffs

import (
	"html/template"
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

type SymbolChange struct {
	Package  string
	Kind     string
	Scope    string
	Name     string
	Old      string
	New      string
	Status   string
	TypeBody *APITypeBody
}

type StructFieldChange struct {
	Owner  string
	Name   string
	Type   string
	Status string
}

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
	ChangedPackages int
	Total           int
}

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
	Rows       []compactRowData
	Body       template.HTML
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

var reportTmpl = template.Must(
	template.New("report").
		Funcs(template.FuncMap{
			"not": func(v interface{}) bool {
				xs, ok := v.([]pkgSectionData)
				return ok && len(xs) == 0
			},
		}).
		Parse(reportTemplates),
)

func (d *APIDiff) HTML(meta ReportMetadata) string {
	if meta.Now.IsZero() {
		meta.Now = time.Now()
	}

	changes := d.collectSymbolChanges()
	packages := summarizeByPackage(changes)
	summary := summarizeOverall(packages)

	data := reportData{
		CSS: template.HTML(reportCSS),
		Meta: reportMetaData{
			Repo:        defaultString(meta.Repo, "repository"),
			OldRef:      defaultString(meta.OldRef, "old"),
			NewRef:      defaultString(meta.NewRef, "new"),
			GeneratedAt: meta.Now.Format("2006-01-02 15:04:05"),
		},
		Verdict:          buildVerdict(summary),
		Summary:          summary,
		BreakingPackages: buildSidebarEntries(packages),
		Packages:         buildPackageSections(changes),
	}

	var sb strings.Builder
	if err := reportTmpl.ExecuteTemplate(&sb, "report", data); err != nil {
		return "<!-- template error: " + err.Error() + " -->"
	}
	return sb.String()
}

func buildVerdict(summary overallSummary) verdictData {
	if summary.Breaking == 0 {
		return verdictData{
			Class: "verdict-ok",
			Icon:  "circle-check",
			Text:  "Compatible API changes only - no breaking changes detected.",
		}
	}
	return verdictData{
		Class: "verdict-breaking",
		Icon:  "alert-triangle",
		Text:  "Breaking API changes detected - review before release.",
	}
}

func buildSidebarEntries(packages []PackageChangeSummary) []sidebarEntry {
	out := make([]sidebarEntry, 0, len(packages))
	for _, p := range packages {
		if p.Breaking == 0 {
			continue
		}
		out = append(out, sidebarEntry{
			AnchorID:  anchorID(p.Package),
			Package:   p.Package,
			ShortName: shortPackage(p.Package),
			Breaking:  p.Breaking,
		})
	}
	return out
}

func buildPackageSections(changes []SymbolChange) []pkgSectionData {
	byPkg := groupChangesByPackage(changes)
	pkgs := sortedKeys(byPkg)

	out := make([]pkgSectionData, 0, len(pkgs))
	for _, pkg := range pkgs {
		pkgChanges := byPkg[pkg]
		sortChanges(pkgChanges)
		out = append(out, pkgSectionData{
			AnchorID:       anchorID(pkg),
			Package:        pkg,
			ShortName:      shortPackage(pkg),
			IsBreaking:     hasBreakingChanges(pkgChanges),
			StatusSections: buildStatusSections(pkgChanges),
		})
	}
	return out
}

func buildStatusSections(changes []SymbolChange) []statusSectionData {
	sections := make([]statusSectionData, 0, 3)
	for _, spec := range []struct {
		Status string
		Title  string
		Accent string
	}{
		{"changed", "Changed signatures", "changed"},
		{"removed", "Removed API", "removed"},
		{"added", "Added API", "added"},
	} {
		sectionChanges := filterByStatus(changes, spec.Status)
		if len(sectionChanges) == 0 {
			continue
		}
		sections = append(
			sections,
			buildStatusSection(spec.Title, spec.Accent, spec.Status, sectionChanges),
		)
	}
	return sections
}

func buildStatusSection(
	title, accentClass, status string,
	changes []SymbolChange,
) statusSectionData {
	section := statusSectionData{Title: title, AccentClass: accentClass}
	if status == "changed" {
		for _, ch := range changes {
			section.ChangeCards = append(section.ChangeCards, buildChangeCard(ch))
		}
		return section
	}

	byKind := groupChangesByKind(changes)
	for _, kind := range apiKindOrder() {
		kindChanges := byKind[kind]
		if len(kindChanges) == 0 {
			continue
		}
		sortChanges(kindChanges)
		addKindToSection(&section, kind, kindChanges, status)
	}
	return section
}

func addKindToSection(
	section *statusSectionData,
	kind string,
	changes []SymbolChange,
	status string,
) {
	switch kind {
	case "Field":
		section.StructFieldDiffs = append(
			section.StructFieldDiffs,
			buildStructFieldDiffs(changes, status)...)
	case "Type":
		blocks, rows := splitTypes(changes)
		for _, ch := range blocks {
			section.TypeDefBlocks = append(section.TypeDefBlocks, buildTypeDefBlock(ch, status))
		}
		if len(rows) > 0 {
			section.KindGroups = append(section.KindGroups, buildScalarTypeGroup(rows, status))
		}
	case "Var":
		section.KindGroups = append(section.KindGroups, buildVarGroup(changes, status))
	case "Const":
		section.KindGroups = append(section.KindGroups, buildConstGroup(changes, status))
	default:
		section.KindGroups = append(section.KindGroups, buildKindGroup(kind, changes, status))
	}
}

func splitTypes(changes []SymbolChange) (withBody, withoutBody []SymbolChange) {
	for _, ch := range changes {
		if ch.TypeBody != nil && (ch.TypeBody.Kind == "struct" || ch.TypeBody.Kind == "interface") {
			withBody = append(withBody, ch)
		} else {
			withoutBody = append(withoutBody, ch)
		}
	}
	return withBody, withoutBody
}

func buildChangeCard(ch SymbolChange) changeCardData {
	return changeCardData{
		KindLabel: ch.Kind + " signature",
		Name:      ch.Name,
		UnifiedDiff: buildUnifiedDiff(
			formatAPIValue(ch.Kind, ch.Scope, ch.Old),
			formatAPIValue(ch.Kind, ch.Scope, ch.New),
		),
	}
}

func buildUnifiedDiff(oldSig, newSig string) template.HTML {
	oldLines := strings.Split(oldSig, "\n")
	newLines := strings.Split(newSig, "\n")
	oldSet := lineSet(oldLines)
	newSet := lineSet(newLines)

	var b strings.Builder
	for _, line := range oldLines {
		if !newSet[line] {
			writeDiffLine(&b, "diff-removed", "-", line)
		}
	}
	for _, line := range oldLines {
		if newSet[line] {
			writeDiffLine(&b, "diff-context", " ", line)
		}
	}
	for _, line := range newLines {
		if !oldSet[line] {
			writeDiffLine(&b, "diff-added", "+", line)
		}
	}
	return template.HTML(b.String())
}

func lineSet(lines []string) map[string]bool {
	out := make(map[string]bool, len(lines))
	for _, line := range lines {
		out[line] = true
	}
	return out
}

func writeDiffLine(b *strings.Builder, class, prefix, line string) {
	b.WriteString(`<span class="` + class + `">`)
	b.WriteString(prefix + " ")
	b.WriteString(template.HTMLEscapeString(line))
	b.WriteString("\n</span>")
}

func buildKindGroup(kind string, changes []SymbolChange, status string) kindGroupData {
	prefix, prefixClass, countClass := compactPrefix(status)
	rows := make([]compactRowData, 0, len(changes))
	for _, ch := range changes {
		code, typ := compactRowParts(ch)
		rows = append(
			rows,
			compactRowData{PrefixClass: prefixClass, Prefix: prefix, Code: code, Type: typ},
		)
	}
	return kindGroupData{
		KindLabel:  pluralKind(kind),
		CountClass: countClass,
		Count:      len(changes),
		Rows:       rows,
	}
}

func buildScalarTypeGroup(changes []SymbolChange, status string) kindGroupData {
	return buildKindGroup("Type", changes, status)
}

func buildVarGroup(changes []SymbolChange, status string) kindGroupData {
	return buildKindGroup("Var", changes, status)
}

func buildConstGroup(changes []SymbolChange, status string) kindGroupData {
	prefix, prefixClass, countClass := compactPrefix(status)
	rows := make([]compactRowData, 0, len(changes))
	for _, ch := range changes {
		name, _, value := parseConstSignature(firstNonEmpty(ch.New, ch.Old))
		code := name
		if value != "" {
			code += " = " + value
		}
		rows = append(rows, compactRowData{PrefixClass: prefixClass, Prefix: prefix, Code: code})
	}
	return kindGroupData{
		KindLabel:  pluralKind("Const"),
		CountClass: countClass,
		Count:      len(changes),
		Rows:       rows,
	}
}

func compactPrefix(status string) (prefix, prefixClass, countClass string) {
	if status == "removed" {
		return "-", "rem", "rem"
	}
	return "+", "add", "add"
}

func compactRowParts(ch SymbolChange) (code, typ string) {
	switch ch.Kind {
	case "Func":
		return compactFunc(ch), ""
	case "Method":
		return compactMethod(ch), ""
	case "Var":
		return compactTypedSymbol(ch)
	case "Const":
		name, _, value := parseConstSignature(firstNonEmpty(ch.New, ch.Old))
		if value != "" {
			return name + " = " + value, ""
		}
		return name, ""
	case "Field":
		return compactField(ch)
	default:
		return ch.Name, ""
	}
}

func buildStructFieldDiffs(changes []SymbolChange, status string) []structFieldDiffData {
	byOwner := groupFieldsByOwner(changes)
	owners := sortedKeys(byOwner)
	if len(owners) == 0 {
		return []structFieldDiffData{{
			Title:      pluralKind("Field"),
			CountClass: countClass(status),
			Count:      len(changes),
			Body:       buildFallbackFieldBody(changes, status),
		}}
	}

	out := make([]structFieldDiffData, 0, len(owners))
	for _, owner := range owners {
		fields := byOwner[owner]
		out = append(out, structFieldDiffData{
			Title:      "type " + owner + " struct",
			CountClass: countClass(status),
			Count:      len(fields),
			Body:       buildStructFieldDiffBody(fields, status),
		})
	}
	return out
}

func groupFieldsByOwner(changes []SymbolChange) map[string][]StructFieldChange {
	out := make(map[string][]StructFieldChange)
	for _, ch := range changes {
		field, ok := parseFieldChange(ch)
		if !ok {
			continue
		}
		out[field.Owner] = append(out[field.Owner], field)
	}
	for owner := range out {
		sort.Slice(
			out[owner],
			func(i, j int) bool { return out[owner][i].Name < out[owner][j].Name },
		)
	}
	return out
}

func parseFieldChange(ch SymbolChange) (StructFieldChange, bool) {
	raw := abbreviateImportPaths(strings.TrimSpace(firstNonEmpty(ch.New, ch.Old)))
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return StructFieldChange{}, false
	}
	fullName := parts[0]
	fieldType := strings.TrimSpace(strings.TrimPrefix(raw, fullName))
	if owner, name, ok := strings.Cut(fullName, "."); ok && name != "" {
		return StructFieldChange{Owner: owner, Name: name, Type: fieldType, Status: ch.Status}, true
	}
	if ch.Scope == "" || ch.Name == "" {
		return StructFieldChange{}, false
	}
	return StructFieldChange{
		Owner:  ch.Scope,
		Name:   ch.Name,
		Type:   fieldType,
		Status: ch.Status,
	}, true
}

func buildStructFieldDiffBody(fields []StructFieldChange, status string) template.HTML {
	prefix, class := diffPrefix(status)
	width := maxFieldNameWidth(fields)
	var b strings.Builder
	b.WriteString(`<span class="diff-context">  {` + "\n</span>")
	for _, f := range fields {
		writeDiffLine(&b, class, prefix, alignField(f.Name, f.Type, width))
	}
	b.WriteString(`<span class="diff-context">  }</span>`)
	return template.HTML(b.String())
}

func buildFallbackFieldBody(changes []SymbolChange, status string) template.HTML {
	prefix, class := diffPrefix(status)
	var b strings.Builder
	for _, ch := range changes {
		writeDiffLine(
			&b,
			class,
			prefix,
			abbreviateImportPaths(strings.TrimSpace(firstNonEmpty(ch.New, ch.Old))),
		)
	}
	return template.HTML(b.String())
}

func buildTypeDefBlock(ch SymbolChange, status string) typeDefBlockData {
	body := ch.TypeBody
	count := len(body.Fields) + len(body.Methods)
	return typeDefBlockData{
		KindLabel:  "type " + ch.Name + " " + body.Kind,
		CountClass: countClass(status),
		Count:      count,
		Body:       buildTypeBody(ch.Name, body, status),
	}
}

func buildTypeBody(name string, body *APITypeBody, status string) template.HTML {
	prefix, class := diffPrefix(status)
	var b strings.Builder
	switch body.Kind {
	case "struct":
		b.WriteString(
			`<span class="diff-context">type ` + template.HTMLEscapeString(
				name,
			) + " struct {\n</span>",
		)
		width := maxRawFieldNameWidth(body.Fields)
		for _, raw := range body.Fields {
			fieldName, fieldType := splitNameType(abbreviateImportPaths(raw))
			writeDiffLine(&b, class, prefix, alignField(fieldName, fieldType, width))
		}
		b.WriteString(`<span class="diff-context">}</span>`)
	case "interface":
		b.WriteString(
			`<span class="diff-context">type ` + template.HTMLEscapeString(
				name,
			) + " interface {\n</span>",
		)
		for _, method := range body.Methods {
			writeDiffLine(&b, class, prefix, abbreviateImportPaths(method))
		}
		b.WriteString(`<span class="diff-context">}</span>`)
	}
	return template.HTML(b.String())
}

func diffPrefix(status string) (prefix, class string) {
	if status == "removed" {
		return "-", "diff-removed"
	}
	return "+", "diff-added"
}

func countClass(status string) string {
	if status == "removed" {
		return "rem"
	}
	return "add"
}

func maxFieldNameWidth(fields []StructFieldChange) int {
	width := 0
	for _, f := range fields {
		if len(f.Name) > width {
			width = len(f.Name)
		}
	}
	return width
}

func maxRawFieldNameWidth(fields []string) int {
	width := 0
	for _, raw := range fields {
		name, _ := splitNameType(abbreviateImportPaths(raw))
		if len(name) > width {
			width = len(name)
		}
	}
	return width
}

func alignField(name, typ string, width int) string {
	if typ == "" {
		return name
	}
	pad := width - len(name) + 1
	if pad < 1 {
		pad = 1
	}
	return name + strings.Repeat(" ", pad) + typ
}

func splitNameType(raw string) (name, typ string) {
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return "", ""
	}
	name = parts[0]
	typ = strings.TrimSpace(strings.TrimPrefix(raw, name))
	return name, typ
}

func (d *APIDiff) collectSymbolChanges() []SymbolChange {
	var changes []SymbolChange
	add := func(status, kind string, xs []DiffItem) {
		for _, x := range xs {
			changes = append(changes, newSymbolChange(status, kind, x))
		}
	}

	add("added", "Func", d.FuncsAdded)
	add("removed", "Func", d.FuncsRemoved)
	add("added", "Var", d.VarsAdded)
	add("removed", "Var", d.VarsRemoved)
	add("added", "Const", d.ConstsAdded)
	add("removed", "Const", d.ConstsRemoved)
	add("added", "Type", d.TypesAdded)
	add("removed", "Type", d.TypesRemoved)
	add("added", "Field", d.FieldsAdded)
	add("removed", "Field", d.FieldsRemoved)
	add("added", "Method", d.MethodsAdded)
	add("removed", "Method", d.MethodsRemoved)

	return mergeIntoChangedSymbols(changes)
}

func newSymbolChange(status, fallbackKind string, item DiffItem) SymbolChange {
	kind := cleanAPILabel(fallbackKind, item.Label)
	scope := apiScope(item.Label)
	name := apiSymbolName(item.Signature)
	change := SymbolChange{
		Package:  item.Path,
		Kind:     kind,
		Scope:    scope,
		Name:     displayAPIName(kind, scope, name),
		Old:      item.Signature,
		New:      item.Signature,
		Status:   status,
		TypeBody: item.TypeBody,
	}
	if status == "added" {
		change.Old = ""
	} else {
		change.New = ""
	}
	return change
}

func mergeIntoChangedSymbols(raw []SymbolChange) []SymbolChange {
	added := make(map[string]SymbolChange)
	removed := make(map[string]SymbolChange)
	for _, ch := range raw {
		switch ch.Status {
		case "added":
			added[ch.changeKey()] = ch
		case "removed":
			removed[ch.changeKey()] = ch
		}
	}

	used := make(map[string]bool)
	out := make([]SymbolChange, 0, len(raw))
	for key, old := range removed {
		newer, ok := added[key]
		if !ok {
			continue
		}
		out = append(
			out,
			SymbolChange{
				Package: old.Package,
				Kind:    old.Kind,
				Scope:   old.Scope,
				Name:    old.Name,
				Old:     old.Old,
				New:     newer.New,
				Status:  "changed",
			},
		)
		used[key] = true
	}
	for _, ch := range raw {
		if used[ch.changeKey()] {
			continue
		}
		out = append(out, ch)
	}
	sortChanges(out)
	return out
}

func (ch SymbolChange) changeKey() string {
	return ch.Package + "\x00" + ch.Kind + "\x00" + ch.Scope + "\x00" + apiSymbolName(
		firstNonEmpty(ch.New, ch.Old),
	)
}

func summarizeOverall(packages []PackageChangeSummary) overallSummary {
	var s overallSummary
	s.ChangedPackages = len(packages)
	for _, p := range packages {
		s.Breaking += p.Breaking
		s.Changed += p.Changed
		s.Removed += p.Removed
		s.Added += p.Added
		s.Total += p.Total
	}
	return s
}

func summarizeByPackage(changes []SymbolChange) []PackageChangeSummary {
	byPkg := make(map[string]*PackageChangeSummary)
	for _, ch := range changes {
		pkg := defaultString(ch.Package, "unknown")
		summary := byPkg[pkg]
		if summary == nil {
			summary = &PackageChangeSummary{Package: pkg}
			byPkg[pkg] = summary
		}
		summary.Total++
		switch ch.Status {
		case "changed":
			summary.Changed++
			summary.Breaking++
		case "removed":
			summary.Removed++
			summary.Breaking++
		case "added":
			summary.Added++
		}
	}

	out := make([]PackageChangeSummary, 0, len(byPkg))
	for _, summary := range byPkg {
		out = append(out, *summary)
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

func groupChangesByPackage(changes []SymbolChange) map[string][]SymbolChange {
	out := make(map[string][]SymbolChange)
	for _, ch := range changes {
		out[ch.Package] = append(out[ch.Package], ch)
	}
	return out
}

func groupChangesByKind(changes []SymbolChange) map[string][]SymbolChange {
	out := make(map[string][]SymbolChange)
	for _, ch := range changes {
		out[ch.Kind] = append(out[ch.Kind], ch)
	}
	return out
}

func filterByStatus(changes []SymbolChange, status string) []SymbolChange {
	out := make([]SymbolChange, 0, len(changes))
	for _, ch := range changes {
		if ch.Status == status {
			out = append(out, ch)
		}
	}
	return out
}

func hasBreakingChanges(changes []SymbolChange) bool {
	for _, ch := range changes {
		if ch.Status == "changed" || ch.Status == "removed" {
			return true
		}
	}
	return false
}

func sortChanges(changes []SymbolChange) {
	sort.Slice(changes, func(i, j int) bool {
		if statusRank(changes[i].Status) != statusRank(changes[j].Status) {
			return statusRank(changes[i].Status) < statusRank(changes[j].Status)
		}
		if changes[i].Package != changes[j].Package {
			return changes[i].Package < changes[j].Package
		}
		if kindRank(changes[i].Kind) != kindRank(changes[j].Kind) {
			return kindRank(changes[i].Kind) < kindRank(changes[j].Kind)
		}
		return changes[i].Name < changes[j].Name
	})
}

func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func parseConstSignature(sig string) (name, typ, value string) {
	sig = abbreviateImportPaths(strings.TrimSpace(sig))
	if idx := strings.Index(sig, " = "); idx >= 0 {
		value = strings.TrimSpace(sig[idx+3:])
		sig = sig[:idx]
	}
	name, typ = splitNameType(sig)
	if strings.HasPrefix(typ, "untyped ") {
		typ = ""
	}
	return name, typ, value
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
	default:
		return raw
	}
}

func formatCallable(prefix, scope, name, raw string) string {
	idx := strings.Index(raw, "(")
	if idx < 0 {
		return prefix + " " + raw
	}
	params, rest, ok := splitFirstParen(raw[idx:])
	if !ok {
		return prefix + " " + raw
	}
	result := formatResults(rest)
	parts := splitTopLevel(params)
	head := prefix + " "
	if scope != "" {
		head += "(" + abbreviateImportPaths(scope) + ") "
	}
	head += name

	oneLine := head + "(" + strings.Join(parts, ", ") + ")" + result
	if len(oneLine) <= 96 && len(parts) <= 1 {
		return oneLine
	}

	var b strings.Builder
	b.WriteString(head + "(\n")
	for _, part := range parts {
		if part != "" {
			b.WriteString("    " + part + ",\n")
		}
	}
	b.WriteString(")" + result)
	return b.String()
}

func formatResults(rest string) string {
	rest = strings.TrimSpace(rest)
	if !strings.HasPrefix(rest, "->") {
		return ""
	}
	resRaw := strings.TrimSpace(strings.TrimPrefix(rest, "->"))
	if !strings.HasPrefix(resRaw, "(") {
		return " " + resRaw
	}
	resultText, _, ok := splitFirstParen(resRaw)
	if !ok {
		return ""
	}
	parts := splitTopLevel(resultText)
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return " " + parts[0]
	default:
		return " (" + strings.Join(parts, ", ") + ")"
	}
}

func formatField(scope, raw string) string {
	name, typ := compactField(
		SymbolChange{Kind: "Field", Scope: scope, Name: apiSymbolName(raw), New: raw},
	)
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
	return append(out, strings.TrimSpace(s[start:]))
}

func compactField(ch SymbolChange) (name string, typ string) {
	raw := abbreviateImportPaths(strings.TrimSpace(firstNonEmpty(ch.New, ch.Old)))
	fieldName, fieldType := splitNameType(raw)
	if strings.Contains(fieldName, ".") {
		return fieldName, fieldType
	}
	return compactScopedName(ch.Scope, fieldName), fieldType
}

func compactTypedSymbol(ch SymbolChange) (name string, typ string) {
	return splitNameType(abbreviateImportPaths(strings.TrimSpace(firstNonEmpty(ch.New, ch.Old))))
}

func compactFunc(ch SymbolChange) string {
	return compactCallable(ch, false)
}

func compactMethod(ch SymbolChange) string {
	return compactCallable(ch, true)
}

func compactCallable(ch SymbolChange, includeScope bool) string {
	raw := abbreviateImportPaths(strings.TrimSpace(firstNonEmpty(ch.New, ch.Old)))
	name := apiSymbolName(raw)
	if name == "" || name == "unknown" {
		name = ch.Name
	}
	if includeScope && ch.Scope != "" && !strings.Contains(name, ".") {
		name = ch.Scope + "." + name
	}
	params, results := extractCallableSignatureParts(raw)
	if params == "" && results == "" {
		return name + "(...)"
	}
	return name + "(" + params + ")" + results
}

func compactScopedName(scope, name string) string {
	name = strings.TrimSpace(name)
	scope = strings.TrimSpace(scope)
	if scope == "" || name == "" || strings.HasPrefix(name, scope+".") {
		return name
	}
	return scope + "." + name
}

func extractCallableSignatureParts(raw string) (params string, results string) {
	idx := strings.Index(raw, "(")
	if idx < 0 {
		return "", ""
	}
	paramText, rest, ok := splitFirstParen(raw[idx:])
	if !ok {
		return "", ""
	}
	return compactParamList(splitTopLevel(paramText)), formatResults(rest)
}

func compactParamList(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	const maxInline = 72
	joined := strings.Join(parts, ", ")
	if len(joined) <= maxInline {
		return joined
	}
	short := make([]string, 0, len(parts))
	for _, part := range parts {
		fields := strings.Fields(strings.TrimSpace(part))
		if len(fields) >= 2 {
			short = append(short, fields[len(fields)-1])
		} else if len(fields) == 1 {
			short = append(short, fields[0])
		}
	}
	joined = strings.Join(short, ", ")
	if len(joined) <= maxInline {
		return joined
	}
	return "..."
}

var importPathRE = regexp.MustCompile(
	`([A-Za-z0-9_./~-]+/[A-Za-z0-9_./~-]+)\.([A-Za-z_][A-Za-z0-9_]*)`,
)

func abbreviateImportPaths(s string) string {
	return importPathRE.ReplaceAllStringFunc(s, func(match string) string {
		idx := strings.LastIndex(match, ".")
		if idx < 0 {
			return match
		}
		path := match[:idx]
		ident := match[idx+1:]
		if slash := strings.LastIndex(path, "/"); slash >= 0 {
			path = path[slash+1:]
		}
		return path + "." + ident
	})
}

func cleanAPILabel(fallback, label string) string {
	label = strings.ReplaceAll(strings.TrimSpace(label), "`", "")
	if label == "" {
		return fallback
	}
	if strings.Contains(label, "Fields") {
		return "Field"
	}
	if strings.Contains(label, "Methods") {
		return "Method"
	}
	return strings.TrimSuffix(label, "s")
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
	name, _ := splitNameType(x)
	if dot := strings.LastIndex(name, "."); dot >= 0 && dot+1 < len(name) {
		return name[dot+1:]
	}
	if name != "" {
		return name
	}
	return x
}

func displayAPIName(kind, scope, name string) string {
	if kind == "Field" || kind == "Method" {
		return compactScopedName(scope, name)
	}
	return name
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
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
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
	for i, candidate := range apiKindOrder() {
		if kind == candidate {
			return i
		}
	}
	return 99
}

func apiKindOrder() []string {
	return []string{"Type", "Func", "Method", "Field", "Var", "Const"}
}

func pluralKind(kind string) string {
	switch kind {
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
