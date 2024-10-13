package contracts

const (
	supportedHttpMethods = "GET,PUT,POST,DELETE,PATCH,HEAD,CONNECT,OPTIONS,TRACE"
	defaultMethod        = "ALL"
	CtxKeyForRequest     = "NfHttpRequestObject"
	DomainInfoCode       = "DomainInfoCode"
)

type DomainInfo struct {
	FullDomain  string
	SubDomain   string
	SecondLevel string
	TopLevel    string
}
