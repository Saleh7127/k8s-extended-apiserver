package k8s_extended_apiserver

const (
	RemoteUser   = "X-Remote-User"
	ClientCertCn = "Client-Cert-CN"
	SysAnonymous = "system:anonymous"
	ExApiAdd     = "127.0.0.2:8443"
	ApiAdd       = "127.0.0.1:8443"
	ReqHeader    = "requestheader"
	ExApiPath    = "/extended-apiserver/{resource}"
	ApiPath      = "/api-server/{resource}"
	ExApiIp      = "127.0.0.2"
	ApiIp        = "127.0.0.1"
	ExApiServer  = "extended-apiserver"
	ApiServer    = "api-server"
)
