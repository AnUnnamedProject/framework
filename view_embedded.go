// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

// EmbedShortcodes defines the views shortcodes.
func (s *View) EmbedShortcodes() {
	// s.New("meta").Parse(`{{range $key, $value := .FrameworkMeta}}<meta name="{{$key}}" content="{{$value}}">{{end}}`)
}

// EmbedTemplates defines the views templates (e.g. google analytics).
func (s *View) EmbedTemplates() {
	_, _ = s.New("google_analytics").Parse(`{{ with .GoogleAnalytics }}
		<script>
		window.ga=window.ga||function(){(ga.q=ga.q||[]).push(arguments)};ga.l=+new Date;
		ga('create', '{{ . }}', 'auto');
		ga('send', 'pageview');
		</script>
		<script async src='//www.google-analytics.com/analytics.js'></script>
		{{ end }}`)
}

// SetGoogleAnalytics sets the google analytics id (UA-XXXXXXX-X)
// you can get the analytics code on views:
// {{ template "google_analytics" . }}
func SetGoogleAnalytics(code string) {
	App.sharedData["GoogleAnalytics"] = code
}
