package templates

import "html/template"

const TemplatesToCompileDirPath = "templates_to_compile"

var CompiledHomeTemplate = template.Must(template.ParseFiles(TemplatesToCompileDirPath + "/tpl-home.html"))
var CompiledRoomTemplate = template.Must(template.ParseFiles(TemplatesToCompileDirPath+"/tpl-room.html", TemplatesToCompileDirPath+"/tpl-frag-about-terms-of-service.html"))
var CompiledAboutTemplate = template.Must(template.ParseFiles(TemplatesToCompileDirPath+"/tpl-about.html", TemplatesToCompileDirPath+"/tpl-frag-about-body.html"))
var CompiledTermsOfServiceTemplate = template.Must(template.ParseFiles(TemplatesToCompileDirPath+"/tpl-about.html", TemplatesToCompileDirPath+"/tpl-frag-about-terms-of-service.html"))
var CompiledPrivacyPolicyTemplate = template.Must(template.ParseFiles(TemplatesToCompileDirPath+"/tpl-about.html", TemplatesToCompileDirPath+"/tpl-frag-about-privacy-policy.html"))
var CompiledConnectionMethodsTemplate = template.Must(template.ParseFiles(TemplatesToCompileDirPath+"/tpl-about.html", TemplatesToCompileDirPath+"/tpl-frag-about-connection-methods.html"))
var CompiledUniversalAccessTemplate = template.Must(template.ParseFiles(TemplatesToCompileDirPath+"/tpl-about.html", TemplatesToCompileDirPath+"/tpl-frag-about-universal-access.html"))

var CompiledRoomCtrlPageProxyTemplate = template.Must(template.ParseFiles(TemplatesToCompileDirPath + "/tpl-control-page-proxy.html"))
