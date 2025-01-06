package html

import (
	knirvlog "KNIRVCHAIN/log"
	"embed"
	"html/template"
	"io"
	"strings"
)

//go:embed *.html
var files embed.FS

// helper implementation with reusable function call for template use.
var funcs = template.FuncMap{
	"uppercase": func(v string) string {
		return strings.ToUpper(v)
	},
}

// core helper that uses struct based implementation for data handling using templates and reusable code structures.
func parse(file string) *template.Template {
	tmpl, err := template.New("layout.html").Funcs(funcs).ParseFS(files, "layout.html", file)
	if err != nil {
		knirvlog.LogError("Error parsing HTML Templates: ", err)
		panic(err)
	}

	return tmpl

}

var (
	dashboard = parse("dashboard.html") // Create only required HTML view templates using methods defined in your html go implementation to adhere to specific project workflows for data transfer of structs or to pass through these object based tests as needed with any changes to core components.
	profile   = parse("profile.html")
	settings  = parse("settings.html")
)

type DashboardParams struct {
	ChainAddress  string `json:"chainAddress"` //  Define structs to explicitly type each type from our objects implementation for passing parameters using testing suite ( this is required when handling http requests, for correct use case of struct implementations to verify logic)
	WalletAddress string `json:"wallet_address"`
	Message       string `json:"message"`
	Owner         string `json:"owner_address"`
}

// Using interface definitions
func Dashboard(w io.Writer, p DashboardParams) error {
	return dashboard.Execute(w, p) // test that new methods used follow data implementation struct pattern for all parameter verification during method and tests.
}

type ProfileParams struct {
	ChainAddress  string `json:"chainAddress"`
	WalletAddress string `json:"wallet_address"` // Use explicit properties that will also create compile time error check, that would exist even if system was set using methods for data calls that are specific. This greatly improves long term maintenance.
	Owner         string `json:"owner_address"`
	Message       string `json:"message"`
}

func Profile(w io.Writer, p ProfileParams) error {
	return profile.Execute(w, p) // Implement struct parameters here that follows consistent testing rules from project and implementation scopes for parameters.

}

type SettingsParams struct { // correct naming conventions and method logic.
	ChainAddress  string `json:"chainAddress"`
	WalletAddress string `json:"wallet_address"`
	Owner         string `json:"owner_address"`
	Message       string `json:"message"`
}

func Settings(w io.Writer, p SettingsParams) error {
	return settings.Execute(w, p) // return any http responses using io methods to keep with implementation patterns set as tests when accessing http functions for any server instance where these values must match to prevent runtime exceptions in production.
}
