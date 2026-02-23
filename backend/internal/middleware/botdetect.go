package middleware

import (
	"log/slog"
	"net/http"
	"strings"
)

var botPatterns = []string{
	"SafeLinks",
	"Mimecast",
	"Barracuda",
	"Slackbot",
	"Slack-ImgProxy",
	"Discordbot",
	"WhatsApp",
	"TelegramBot",
	"facebookexternalhit",
	"Facebot",
	"Twitterbot",
	"LinkedInBot",
	"Googlebot",
	"bingbot",
	"YandexBot",
	"DuckDuckBot",
	"PetalBot",
	"AhrefsBot",
	"SemrushBot",
	"MJ12bot",
	"DotBot",
	"Embedly",
	"Iframely",
	"outbrain",
	"Feedfetcher",
	"MetaInspector",
	"LinkPreview",
	"URL2PNG",
	"urlscan",
	"NetcraftSurveyAgent",
}

const botDummyPage = `<!DOCTYPE html>
<html><head><title>SharePwd</title></head>
<body style="font-family:sans-serif;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0;background:#09090b;color:#fafafa">
<div style="text-align:center;max-width:400px">
<h1>SharePwd</h1>
<p>This link contains a shared secret. Open it in your browser to view it.</p>
</div></body></html>`

func BotDetect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.UserAgent()
		if ua == "" {
			next.ServeHTTP(w, r)
			return
		}

		uaLower := strings.ToLower(ua)
		for _, pattern := range botPatterns {
			if strings.Contains(uaLower, strings.ToLower(pattern)) {
				slog.Info("bot detected",
					"pattern", pattern,
					"user_agent", ua,
					"path", r.URL.Path,
					"ip", r.RemoteAddr,
				)
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(botDummyPage))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
