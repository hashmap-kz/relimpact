package diffs

// reportTemplates holds every HTML template used by the report renderer.
// Templates are named and call each other; the root is "report".
const reportTemplates = `
{{define "report"}}<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>API Compatibility Report</title>
{{.CSS}}
</head>
<body>
<div class="layout">
  {{template "sidebar" .Sidebar}}
  <main>
    <header class="hero">
      <div class="eyebrow">relimpact api</div>
      <h1>API compatibility report</h1>
      <p class="sub">{{.Meta.Repo}} · {{.Meta.OldRef}} -> {{.Meta.NewRef}} · generated {{.Meta.GeneratedAt}}</p>
    </header>

    <div class="verdict {{.Verdict.Class}}">
      <i class="ti ti-{{.Verdict.Icon}}" aria-hidden="true"></i>
      {{.Verdict.Text}}
    </div>

    {{template "cards" .Summary}}

    {{if .Breaking}}
      <section class="page-block" id="breaking-changes">
        <div class="page-block-head">
          <div>
            <div class="page-block-kicker removed">Breaking</div>
            <h2>Breaking changes</h2>
          </div>
          <p>Changed and removed public API. Review these before release.</p>
        </div>
        {{range .Breaking}}
          {{template "pkg-section" .}}
        {{end}}
      </section>
    {{end}}

    {{if .Added}}
      <section class="page-block" id="new-api">
        <div class="page-block-head">
          <div>
            <div class="page-block-kicker added">New API</div>
            <h2>New features added</h2>
          </div>
          <p>Compatible public additions. Useful for release notes.</p>
        </div>
        {{range .Added}}
          {{template "pkg-section" .}}
        {{end}}
      </section>
    {{end}}

    {{if and (not .Breaking) (not .Added)}}
      <section class="empty">No public API changes detected.</section>
    {{end}}
  </main>
</div>
</body>
</html>
{{end}}


{{define "sidebar"}}
<aside>
  <div class="brand">rel<span class="brand-accent">impact</span></div>
  {{if .Breaking}}
    <a class="nav-top nav-top-breaking" href="#breaking-changes">Breaking changes</a>
    <div class="nav-section-label nav-label-breaking">Breaking</div>
    {{range .Breaking}}
      <a class="nav-item nav-breaking" href="#{{.AnchorID}}" title="{{.Package}}">
        <span>{{.ShortName}}</span>
        {{if gt .Count 1}}<strong>{{.Count}}</strong>{{end}}
      </a>
    {{end}}
  {{else}}
    <div class="empty-nav">No breaking changes</div>
  {{end}}

  {{if .Added}}
    <a class="nav-top nav-top-added" href="#new-api">New API</a>
    <div class="nav-section-label nav-label-added">Additions</div>
    {{range .Added}}
      <a class="nav-item nav-added" href="#{{.AnchorID}}" title="{{.Package}}">
        <span>{{.ShortName}}</span>
        {{if gt .Count 1}}<strong>{{.Count}}</strong>{{end}}
      </a>
    {{end}}
  {{end}}
</aside>
{{end}}


{{define "cards"}}
<section class="cards">
  <div class="card red"><div class="num">{{.Breaking}}</div><div class="lbl">Breaking</div></div>
  <div class="card amber"><div class="num">{{.Changed}}</div><div class="lbl">Changed</div></div>
  <div class="card red"><div class="num">{{.Removed}}</div><div class="lbl">Removed</div></div>
  <div class="card green"><div class="num">{{.Added}}</div><div class="lbl">Added</div></div>
  <div class="card"><div class="num">{{.ChangedPackages}}</div><div class="lbl">Packages</div></div>
</section>
{{end}}


{{define "pkg-section"}}
<section class="pkg{{if .IsBreaking}} breaking{{end}}" id="{{.AnchorID}}">
  <div class="pkg-head">
    <div class="pkg-name">{{.ShortName}}</div>
    <div class="pkg-path">{{.Package}}</div>
  </div>
  {{range .StatusSections}}
    {{template "status-section" .}}
  {{end}}
</section>
{{end}}


{{define "status-section"}}
<div class="group-label {{.AccentClass}}">{{.Title}}</div>
{{range .ChangeCards}}
  {{template "change-card" .}}
{{end}}
{{range .KindGroups}}
  {{template "kind-group" .}}
{{end}}
{{range .TypeDefBlocks}}
  {{template "type-def-block" .}}
{{end}}
{{range .StructFieldDiffs}}
  {{template "struct-field-diff" .}}
{{end}}
{{end}}


{{define "change-card"}}
<article class="change">
  <div class="change-head">
    <div>
      <div class="change-kind">{{.KindLabel}}</div>
      <div class="change-sym">{{.Name}}</div>
    </div>
    <span class="badge badge-changed">Changed</span>
  </div>
  <pre class="sig-unified">{{.UnifiedDiff}}</pre>
</article>
{{end}}


{{define "kind-group"}}
<div class="compact">
  <div class="compact-head">
    <span>{{.KindLabel}}</span>
    <span class="compact-cnt {{.CountClass}}">{{.Count}}</span>
  </div>
  {{if .Body}}
    <pre class="struct-pre">{{.Body}}</pre>
  {{else}}
    {{range .Rows}}
      <div class="compact-row">
        <span class="pfx {{.PrefixClass}}">{{.Prefix}}</span>
        <code>{{.Code}}</code>
        {{if .Type}}<span class="ctype">{{.Type}}</span>{{end}}
      </div>
    {{end}}
  {{end}}
</div>
{{end}}


{{define "struct-field-diff"}}
<div class="compact">
  <div class="compact-head">
    <span>{{.Title}}</span>
    <span class="compact-cnt {{.CountClass}}">{{.Count}}</span>
  </div>
  <pre class="struct-pre">{{.Body}}</pre>
</div>
{{end}}


{{define "type-def-block"}}
<div class="compact">
  <div class="compact-head">
    <span>{{.KindLabel}}</span>
    <span class="compact-cnt {{.CountClass}}">{{.Count}}</span>
  </div>
  <pre class="struct-pre">{{.Body}}</pre>
</div>
{{end}}
`

const reportCSS = `<style>
:root {
  --bg: #f4f5f8;
  --panel: #ffffff;
  --text: #172033;
  --muted: #667085;
  --border: #e3e5ec;
  --soft: #fafafa;
  --red: #991b1b;
  --red-bg: #fff1f1;
  --red-border: #fecaca;
  --green: #166534;
  --green-bg: #f0fdf4;
  --green-border: #bbf7d0;
  --amber: #92400e;
  --amber-bg: #fffbeb;
  --amber-border: #fde68a;
}
* { box-sizing: border-box; margin: 0; padding: 0; }
body {
  background: var(--bg);
  color: var(--text);
  font-family: Inter, ui-sans-serif, system-ui, -apple-system, sans-serif;
  font-size: 13px;
  line-height: 1.5;
}
.layout { display: flex; min-height: 100vh; }
aside {
  flex-shrink: 0;
  position: sticky;
  top: 0;
  height: 100vh;
  overflow: auto;
  padding: 16px;
  background: var(--panel);
  border-right: 0.5px solid var(--border);
}
main { flex: 1; min-width: 0; max-width: 860px; padding: 28px 36px 56px; }
.brand { font-weight: 800; font-size: 16px; letter-spacing: -0.04em; margin-bottom: 20px; }
.brand-accent { color: var(--red); }
.nav-section-label {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  margin: 12px 0 4px;
}
.nav-label-breaking { color: var(--red); }
.nav-label-added { color: var(--green); }
.nav-top {
  display: block;
  padding: 7px 8px;
  margin: 14px 0 6px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 750;
  text-decoration: none;
}
.nav-top-breaking { background: var(--red-bg); color: var(--red); }
.nav-top-added { background: var(--green-bg); color: var(--green); }
.nav-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 5px 0 5px 8px;
  border-left: 2px solid var(--border);
  margin-bottom: 2px;
  text-decoration: none;
  color: var(--text);
  font-size: 12px;
  border-radius: 0 4px 4px 0;
}
.nav-breaking { border-left-color: var(--red-border); }
.nav-added { border-left-color: var(--green-border); }
.nav-breaking:hover { background: var(--red-bg); border-left-color: var(--red); }
.nav-added:hover { background: var(--green-bg); border-left-color: var(--green); }
.nav-item span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: ui-monospace, monospace;
}
.nav-item strong { font-size: 11px; margin-left: 6px; white-space: nowrap; }
.nav-breaking strong { color: var(--red); }
.nav-added strong { color: var(--green); }
.empty-nav { color: var(--muted); font-size: 12px; }
.hero { margin-bottom: 0; }
.eyebrow {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.12em;
  color: var(--muted);
  margin-bottom: 6px;
}
h1 { font-size: 22px; font-weight: 800; letter-spacing: -0.03em; line-height: 1.1; }
.sub { font-size: 11px; color: var(--muted); margin-top: 4px; font-family: ui-monospace, monospace; }
.verdict {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  margin: 14px 0 20px;
  padding: 7px 12px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
}
.verdict i { font-size: 14px; }
.verdict-breaking { background: var(--red-bg); border: 0.5px solid var(--red-border); color: #7f1d1d; }
.verdict-ok { background: var(--green-bg); border: 0.5px solid var(--green-border); color: #14532d; }
.cards { display: flex; gap: 8px; margin-bottom: 28px; flex-wrap: wrap; }
.card { background: var(--panel); border: 0.5px solid var(--border); border-radius: 6px; padding: 10px 14px; min-width: 72px; }
.card.red .num { color: var(--red); }
.card.amber .num { color: var(--amber); }
.card.green .num { color: var(--green); }
.num { font-size: 22px; font-weight: 800; letter-spacing: -0.05em; line-height: 1; font-family: ui-monospace, monospace; }
.lbl { font-size: 11px; color: var(--muted); margin-top: 3px; }
.page-block { margin-top: 26px; }
.page-block + .page-block { margin-top: 42px; }
.page-block-head {
  display: flex;
  justify-content: space-between;
  gap: 24px;
  align-items: flex-end;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border);
  margin-bottom: 8px;
}
.page-block-kicker {
  font-size: 10px;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  margin-bottom: 3px;
}
.page-block-head h2 { font-size: 18px; line-height: 1.2; letter-spacing: -0.03em; }
.page-block-head p { max-width: 420px; color: var(--muted); font-size: 12px; text-align: right; }
.pkg { padding: 28px 0; border-top: 1px solid var(--border); }
.page-block .pkg:first-of-type { border-top: none; padding-top: 12px; }
.pkg-head { padding-bottom: 8px; border-bottom: 0.5px solid var(--border); margin-bottom: 10px; }
.pkg-name { font-size: 13px; font-weight: 700; font-family: ui-monospace, monospace; }
.pkg-path { font-size: 11px; color: var(--muted); font-family: ui-monospace, monospace; margin-top: 1px; }
.group-label {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  margin: 12px 0 6px;
  display: flex;
  align-items: center;
  gap: 5px;
}
.group-label::before { content: ''; display: block; width: 3px; height: 10px; border-radius: 2px; background: currentColor; flex-shrink: 0; }
.changed { color: var(--amber); }
.removed { color: var(--red); }
.added { color: var(--green); }
.change { background: var(--soft); border: 0.5px solid var(--border); border-radius: 6px; padding: 10px 12px; margin-bottom: 6px; }
.change-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.change-kind { font-size: 10px; color: var(--muted); font-weight: 600; text-transform: uppercase; letter-spacing: 0.06em; margin-bottom: 2px; }
.change-sym { font-size: 13px; font-weight: 700; font-family: ui-monospace, monospace; }
.badge { border-radius: 4px; padding: 3px 7px; font-size: 10px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; white-space: nowrap; }
.badge-changed { background: var(--amber-bg); color: var(--amber); }
.sig-unified,
.struct-pre {
  background: #0d1117;
  border-radius: 6px;
  padding: 10px 12px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  line-height: 1.55;
  overflow-x: auto;
  white-space: pre;
}
.diff-removed { color: #fca5a5; background: rgba(239,68,68,0.12); display: block; border-radius: 2px; padding: 0 2px; }
.diff-added { color: #86efac; background: rgba(34,197,94,0.10); display: block; border-radius: 2px; padding: 0 2px; }
.diff-context { color: #4a5568; display: block; padding: 0 2px; }
.compact { background: var(--soft); border: 0.5px solid var(--border); border-radius: 6px; overflow: hidden; margin-bottom: 6px; }
.compact-head { display: flex; justify-content: space-between; align-items: center; padding: 7px 12px; border-bottom: 0.5px solid var(--border); font-size: 11px; font-weight: 700; color: var(--text); }
.compact-cnt { font-size: 10px; font-weight: 700; border-radius: 3px; padding: 1px 6px; }
.compact-cnt.add { background: var(--green-bg); color: var(--green); }
.compact-cnt.rem { background: var(--red-bg); color: var(--red); }
.compact-row { display: flex; align-items: baseline; gap: 6px; padding: 5px 12px; border-bottom: 0.5px solid #f1f3f7; font-family: ui-monospace, monospace; font-size: 12px; }
.compact-row:last-child { border-bottom: none; }
.pfx { font-weight: 700; width: 12px; flex-shrink: 0; }
.pfx.add { color: var(--green); }
.pfx.rem { color: var(--red); }
code { font-family: ui-monospace, monospace; font-size: 12px; }
.ctype { color: var(--muted); font-size: 11px; margin-left: auto; white-space: nowrap; }
.empty { background: var(--panel); border: 0.5px solid var(--border); border-radius: 6px; padding: 20px; color: var(--muted); margin-top: 20px; }
@media (max-width: 860px) {
  .layout { display: block; }
  aside { position: static; height: auto; }
  main { padding: 20px 16px 40px; }
  .cards { display: grid; grid-template-columns: repeat(3, 1fr); }
  .page-block-head { display: block; }
  .page-block-head p { text-align: left; margin-top: 6px; }
}
</style>`
