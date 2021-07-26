package testformatter

// TemplatePackageDownloads is the default template for package downloads.
const TemplatePackageDownloads = `{{- $failed := .Failed -}}
{{- with .Packages }}::group::{{ if $failed }}\033[0;31m❌{{ else }}\033[0;34m📥{{ end }} Dependency downloads\033[0m
{{- range . }}   {{ if .Failed }}\033[0;31m❌{{ else }}📦{{ end }} {{ .Package }} {{ .Version }}\033[0m
{{ with .Reason }}     {{ . }}
{{ end -}}{{ end -}}
::endgroup::
{{ end }}`

// PackageDownloads is the context for TemplatePackageDownloads.
type PackageDownloads struct {
	// Packages is a list of packages
	Packages []PackageDownload
	// Failed indicates that one or more package downloads failed.
	Failed bool
}

// PackageDownload is a single download of a package.
type PackageDownload struct {
	// Package is the name of the package being downloaded.
	Package string
	// Version is the version of the package being downloaded
	Version string
	// Failed indicates that the download failed.
	Failed bool
	// Reason is the reason text of the download failure.
	Reason string
}
