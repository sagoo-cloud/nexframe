package nf

func (f *APIFramework) SetIndexFiles(indexFiles []string) {
	f.config.IndexFiles = indexFiles
}

func (f *APIFramework) GetIndexFiles() []string {
	return f.config.IndexFiles
}

func (f *APIFramework) SetIndexFolder(enabled bool) {
	f.config.IndexFolder = enabled
}

func (f *APIFramework) SetFileServerEnabled(enabled bool) {
	f.config.FileServerEnabled = enabled
}

// SetRewrite sets rewrites for static URI for server.
func (f *APIFramework) SetRewrite(uri string, rewrite string) {
	f.config.Rewrites[uri] = rewrite
}

// SetRewriteMap sets the rewritten map for server.
func (f *APIFramework) SetRewriteMap(rewrites map[string]string) {
	for k, v := range rewrites {
		f.config.Rewrites[k] = v
	}
}

// SetRouteOverWrite sets the RouteOverWrite for server.
func (f *APIFramework) SetRouteOverWrite(enabled bool) {
	f.config.RouteOverWrite = enabled
}

// SetSwaggerPath sets the SwaggerPath for server.
func (f *APIFramework) SetSwaggerPath(path string) {
	f.config.SwaggerPath = path
}

// SetSwaggerUITemplate sets the Swagger template for server.
func (f *APIFramework) SetSwaggerUITemplate(swaggerUITemplate string) {
	f.config.SwaggerUITemplate = swaggerUITemplate
}

// SetOpenApiPath sets the OpenApiPath for server.
func (f *APIFramework) SetOpenApiPath(path string) {
	f.config.OpenApiPath = path
}
