package templates

import "html/template"

const TemplatesToCompileDirPath = "templates_to_compile"

var CompiledTemplates = template.Must(template.ParseGlob(TemplatesToCompileDirPath + "/*"))
