package mtemplate

import "html/template"

func MainTemplate() *template.Template {
	//tmpl := template.Must(template.ParseFiles("../../internal/mtemplate/layout.html"))
	tmpl := template.Must(template.New("html-tmpl").Parse("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n    <meta charset=\"UTF-8\">\n    <title>Title</title>\n</head>\n<body>\n<h1>Метрики</h1>\n\n<h2>Gauge</h2>\n<ul>\n    {{range $name, $val := .Gauge}}\n    <li>{{$name}} - {{$val}}</li>`\n    {{end}}\n</ul>\n\n<h2>Counter</h2>\n<ul>\n    {{range $name, $val := .Counter}}\n    <li>{{$name}} - {{$val}}</li>`\n    {{end}}\n</ul>\n\n</body>\n</html>"))
	return tmpl
}
