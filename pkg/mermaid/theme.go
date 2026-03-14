package mermaid

type themeVariables struct {
	PrimaryColor       string `json:"primaryColor,omitempty"`
	PrimaryTextColor   string `json:"primaryTextColor,omitempty"`
	PrimaryBorderColor string `json:"primaryBorderColor,omitempty"`
	LineColor          string `json:"lineColor,omitempty"`
}

var darkThemeVariables = themeVariables{
	PrimaryColor:       "#1f2020",
	PrimaryTextColor:   "lightgrey",
	PrimaryBorderColor: "#ccc",
	LineColor:          "#ccc",
}

var lightThemeVariables = themeVariables{
	PrimaryColor:       "#ECECFF",
	PrimaryTextColor:   "black",
	PrimaryBorderColor: "hsl(259.6261682243, 59.7765363128%, 87.9019607843%)",
	LineColor:          "hsl(259.6261682243, 59.7765363128%, 87.9019607843%)",
}
