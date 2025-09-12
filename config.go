package typesense_prometheus_exporter

type Config struct {
	LogLevel    int    `env:"LOG_LEVEL" envDefault:"0"`
	ApiKey      string `env:"TYPESENSE_API_KEY,required"`
	Host        string `env:"TYPESENSE_HOST,required"`
	ApiPort     uint   `env:"TYPESENSE_PORT" envDefault:"8108"`
	MetricsPort uint   `env:"METRICS_PORT" envDefault:"8908"`
	Protocol    string `env:"TYPESENSE_PROTOCOL" envDefault:"http"`
	Cluster     string `env:"TYPESENSE_CLUSTER,required"`
}

const LandingPageTemplate = `
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>Typesense Prometheus Exporter</title>
				<style>
					* {
						margin: 0;
						padding: 0;
						box-sizing: border-box;
					}
					html, body {
						height: 100%;
						display: flex;
						align-items: center;
						justify-content: center;
						background-color: black;
						color: white;
						font-family: Palanquin,-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,"Noto Sans","Liberation Sans",sans-serif,"Apple Color Emoji","Segoe UI Emoji","Segoe UI Symbol","Noto Color Emoji";
					}
					.container {
						text-align: center;
					}
					img {
						margin-top: 20px;
						max-width: 200px;
						height: auto;
						margin-bottom: 20px;
					}
					a {
						text-decoration: none;
						color: #00bcd4;
						font-size: 18px;
					}
					a:hover {
						text-decoration: underline;
					}
				</style>
			</head>
			<body>
				<div class="container">
					<img src="https://prometheus.io/_next/static/media/prometheus-logo.7aa022e5.svg" alt="Prometheus Logo"/><br/>
					<img src="https://typesense.org/_nuxt/img/typesense_logo_white.0f9fb0a.svg" alt="Typesense Logo"/>
					<p><a href="/metrics">Go to Metrics & Stats</a></p>
				</div>
			</body>
			</html>
		`
