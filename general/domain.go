package general

import (
	"context"
	"errors"
	"github.com/sagoo-cloud/nexframe/contracts"
	"github.com/sagoo-cloud/nexframe/nf"
	"net/http"
)

// GetDomainInfo 获取访问的域名信息
func GetDomainInfo(ctx context.Context) (*contracts.DomainInfo, error) {
	domainInfo, ok := ctx.Value(contracts.DomainInfoCode).(*contracts.DomainInfo)
	if !ok {
		return nil, errors.New("domain info not found in context")
	}
	return domainInfo, nil
}

// RequestFromCtx retrieves and returns the Request object from context.
func RequestFromCtx(ctx context.Context) *http.Request {
	return nf.RequestFromCtx(ctx)
}
