package swagger

import (
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"

	"github.com/sagoo-cloud/nexframe/nf/swagger/swaggerFiles"
	"github.com/swaggo/swag"
)

// WrapHandler wraps swaggerFiles.Handler and returns http.HandlerFunc.
var WrapHandler = Handler()

// Config stores httpSwagger configuration variables.
type Config struct {
	// The url pointing to API definition (normally swagger.json or swagger.yaml). Default is `doc.json`.
	URL                      string
	DocExpansion             string
	DomID                    string
	InstanceName             string
	BeforeScript             template.JS
	AfterScript              template.JS
	Plugins                  []template.JS
	UIConfig                 map[template.JS]template.JS
	DeepLinking              bool
	PersistAuthorization     bool
	Layout                   SwaggerLayout
	DefaultModelsExpandDepth ModelsExpandDepthType
	TemplateContent          string
}

// URL presents the url pointing to API definition (normally swagger.json or swagger.yaml).
func URL(url string) func(*Config) {
	return func(c *Config) {
		c.URL = url
	}
}
func TemplateContent(templateContent string) func(*Config) {
	return func(c *Config) {
		c.TemplateContent = templateContent
	}
}

// DeepLinking true, false.
func DeepLinking(deepLinking bool) func(*Config) {
	return func(c *Config) {
		c.DeepLinking = deepLinking
	}
}

// DocExpansion list, full, none.
func DocExpansion(docExpansion string) func(*Config) {
	return func(c *Config) {
		c.DocExpansion = docExpansion
	}
}

// DomID #swagger-ui.
func DomID(domID string) func(*Config) {
	return func(c *Config) {
		c.DomID = domID
	}
}

// InstanceName set the instance name that was used to generate the swagger documents
// Defaults to swag.Name ("swagger").
func InstanceName(name string) func(*Config) {
	return func(c *Config) {
		c.InstanceName = name
	}
}

// PersistAuthorization Persist authorization information over browser close/refresh.
// Defaults to false.
func PersistAuthorization(persistAuthorization bool) func(*Config) {
	return func(c *Config) {
		c.PersistAuthorization = persistAuthorization
	}
}

// Plugins specifies additional plugins to load into Swagger UI.
func Plugins(plugins []string) func(*Config) {
	return func(c *Config) {
		vs := make([]template.JS, len(plugins))
		for i, v := range plugins {
			vs[i] = template.JS(v)
		}
		c.Plugins = vs
	}
}

// UIConfig specifies additional SwaggerUIBundle config object properties.
func UIConfig(props map[string]string) func(*Config) {
	return func(c *Config) {
		vs := make(map[template.JS]template.JS, len(props))
		for k, v := range props {
			vs[template.JS(k)] = template.JS(v)
		}
		c.UIConfig = vs
	}
}

// BeforeScript holds JavaScript to be run right before the Swagger UI object is created.
func BeforeScript(js string) func(*Config) {
	return func(c *Config) {
		c.BeforeScript = template.JS(js)
	}
}

// AfterScript holds JavaScript to be run right after the Swagger UI object is created
// and set on the window.
func AfterScript(js string) func(*Config) {
	return func(c *Config) {
		c.AfterScript = template.JS(js)
	}
}

type SwaggerLayout string

const (
	BaseLayout       SwaggerLayout = "BaseLayout"
	StandaloneLayout SwaggerLayout = "StandaloneLayout"
)

// Define Layout options are BaseLayout or StandaloneLayout
func Layout(layout SwaggerLayout) func(*Config) {
	return func(c *Config) {
		c.Layout = layout
	}
}

type ModelsExpandDepthType int

const (
	ShowModel ModelsExpandDepthType = 1
	HideModel ModelsExpandDepthType = -1
)

// DefaultModelsExpandDepth presents the model of response and request.
// set the default expansion depth for models
func DefaultModelsExpandDepth(defaultModelsExpandDepth ModelsExpandDepthType) func(*Config) {
	return func(c *Config) {
		c.DefaultModelsExpandDepth = defaultModelsExpandDepth
	}
}

func newConfig(configFns ...func(*Config)) *Config {
	config := Config{
		URL:                      "doc.json",
		DocExpansion:             "list",
		DomID:                    "swagger-ui",
		InstanceName:             "swagger",
		DeepLinking:              true,
		PersistAuthorization:     false,
		Layout:                   StandaloneLayout,
		DefaultModelsExpandDepth: ShowModel,
	}

	for _, fn := range configFns {
		fn(&config)
	}

	if config.InstanceName == "" {
		config.InstanceName = swag.Name
	}

	return &config
}

// Handler wraps `http.Handler` into `http.HandlerFunc`.
func Handler(configFns ...func(*Config)) http.HandlerFunc {

	config := newConfig(configFns...)
	var templateContent = indexTempl
	if config.TemplateContent != "" {
		templateContent = config.TemplateContent
	}

	// create a template with name
	index, _ := template.New("swagger_index.html").Parse(templateContent)

	re := regexp.MustCompile(`^(.*/)([^?].*)?[?|.]*$`)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

			return
		}

		matches := re.FindStringSubmatch(r.RequestURI)

		path := matches[2]

		switch filepath.Ext(path) {
		case ".html":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")

		case ".css":
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		case ".js":
			w.Header().Set("Content-Type", "application/javascript")
		case ".png":
			w.Header().Set("Content-Type", "image/png")
		case ".json":
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		}

		switch path {
		case "index.html":
			_ = index.Execute(w, config)
		case "doc.json":
			doc, err := swag.ReadDoc(config.InstanceName)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

				return
			}

			_, _ = w.Write([]byte(doc))
		case "":
			http.Redirect(w, r, matches[1]+"/"+"index.html", http.StatusMovedPermanently)
		default:
			var err error
			r.URL, err = url.Parse(matches[2])
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

				return
			}
			http.FileServer(http.FS(swaggerFiles.FS)).ServeHTTP(w, r)
		}
	}
}

// const indexTempl = "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n    <meta charset=\"UTF-8\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n  <link rel=\"icon\" href=\"/swagger/favicon.ico\" />  <title>API Documentation</title>\n    <style>\n        body {\n            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;\n            margin: 0;\n            padding: 0;\n            display: flex;\n            height: 100vh;\n            color: #3b4151;\n            background-color: #fafafa;\n        }\n        .sidebar {\n            display: flex;\n            flex-direction: column;\n            width: 260px;\n            background-color: #fafafa;\n            border-right: 1px solid #e0e0e0;\n        }\n        .logo {\n            padding: 20px;\n            background-color: #f0f0f0;\n            display: flex;\n            align-items: center;\n            justify-content: flex-start;\n        }\n        .logo svg {\n            width: 40px;\n            height: 40px;\n            margin-right: 10px;\n        }\n        .logo-title {\n            font-size: 20px;\n            font-weight: bold;\n            color: #3b4151;\n        }\n        .search-container {\n            padding: 10px;\n            border-bottom: 1px solid #e0e0e0;\n        }\n        .search-input {\n            width: calc(100% - 16px);\n            padding: 8px;\n            border: 1px solid #ccc;\n            border-radius: 4px;\n            font-size: 14px;\n            box-sizing: border-box;\n        }\n        .nav-menu {\n            flex-grow: 1;\n            overflow-y: auto;\n            padding: 0;\n            margin: 0;\n            list-style-type: none;\n        }\n        .nav-menu li {\n            border-bottom: 1px solid #e0e0e0;\n        }\n        .nav-menu .tag {\n            display: block;\n            padding: 10px 20px;\n            cursor: pointer;\n            transition: background-color 0.3s;\n        }\n        .nav-menu .tag:hover {\n            background-color: #e0e0e0;\n        }\n        .api-list {\n            list-style-type: none;\n            padding-left: 20px;\n            display: none;\n        }\n        .api-list.active {\n            display: block;\n        }\n        .api-item a {\n            display: block;\n            padding: 5px 10px;\n            text-decoration: none;\n            color: #3b4151;\n            transition: background-color 0.3s;\n        }\n        .api-item a:hover {\n            background-color: #e0e0e0;\n        }\n        .method {\n            font-size: 0.8em;\n            font-weight: bold;\n            padding: 2px 5px;\n            border-radius: 3px;\n            margin-right: 5px;\n        }\n        .get { color: #00aa13; }\n        .post { color: #9b708b; }\n        .put { color: #c5862b; }\n        .delete { color: #d41f1c; }\n        .main-content {\n            flex: 1;\n            overflow-y: auto;\n            padding: 20px;\n            background-color: white;\n        }\n        .endpoint {\n            margin-bottom: 30px;\n        }\n        .endpoint-header {\n            display: flex;\n            align-items: center;\n            margin-bottom: 20px;\n            padding-bottom: 10px;\n            border-bottom: 1px solid #e0e0e0;\n        }\n        .endpoint-path {\n            font-family: monospace;\n            font-size: 1.2em;\n            margin-left: 10px;\n        }\n        h2, h3 {\n            color: #3b4151;\n        }\n        .params-table {\n            width: 100%;\n            border-collapse: collapse;\n            margin-top: 20px;\n        }\n        .params-table th, .params-table td {\n            border: 1px solid #e0e0e0;\n            padding: 10px;\n            text-align: left;\n        }\n        .params-table th {\n            background-color: #f0f0f0;\n        }\n        .required {\n            color: #d41f1c;\n        }\n        .response-code {\n            font-weight: bold;\n            margin-right: 10px;\n        }\n        .response-container {\n            background-color: #f0f0f0;\n            padding: 15px;\n            border-radius: 4px;\n            margin-top: 20px;\n        }\n        pre {\n            white-space: pre-wrap;\n            word-wrap: break-word;\n        }\n        .try-it-out {\n            margin-top: 20px;\n            border: 1px solid #e0e0e0;\n            border-radius: 4px;\n            padding: 15px;\n        }\n        .try-it-out h3 {\n            margin-top: 0;\n        }\n        .try-it-out form {\n            display: flex;\n            flex-direction: column;\n        }\n        .try-it-out label {\n            margin-top: 10px;\n        }\n        .try-it-out input, .try-it-out select {\n            margin-top: 5px;\n            padding: 5px;\n            border: 1px solid #ccc;\n            border-radius: 3px;\n        }\n        .try-it-out button {\n            margin-top: 15px;\n            padding: 10px;\n            background-color: #4a90e2;\n            color: white;\n            border: none;\n            border-radius: 3px;\n            cursor: pointer;\n        }\n        .try-it-out button:hover {\n            background-color: #3a7cbd;\n        }\n        .response-area {\n            margin-top: 20px;\n            background-color: #f5f5f5;\n            border: 1px solid #e0e0e0;\n            border-radius: 4px;\n            padding: 15px;\n        }\n        .response-area pre {\n            margin: 0;\n            white-space: pre-wrap;\n        }\n        .hidden {\n            display: none !important;\n        }\n    </style>\n</head>\n<body>\n    <div class=\"sidebar\">\n        <div class=\"logo\">\n            <svg xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"0 0 130.36 134.32\"><defs><style>.cls-1{fill:#f2f8fe;}.cls-1,.cls-2{stroke:#fff;stroke-miterlimit:10;}.cls-2{fill:#a3d4f6;opacity:0.5;}.cls-3{fill:#34a853;}.cls-4{fill:#4285f4;}.cls-5{fill:#fbbb04;}.cls-6{fill:#ea4236;}</style></defs><g id=\"图层_2\" data-name=\"图层 2\"><g id=\"图层_1-2\" data-name=\"图层 1\"><polygon class=\"cls-1\" points=\"83.57 31.19 97.09 37.5 98.72 44.59 83.19 36.94 83.57 31.19\"/><path class=\"cls-2\" d=\"M107.57,44.78c5.11,4,9.14,8.44,9.42,15.23-2.46-2.36-4.92-4.72-7.24-7.17Z\"/><polygon class=\"cls-1\" points=\"41.44 99.02 67.34 106.76 61.49 121.24 39.89 114.64 41.44 99.02\"/><path class=\"cls-2\" d=\"M25.2,32.46c2.37-1,4.84-1.28,7.12-2.57-3.14,4.4-7.14,7.61-12.31,8.78Z\"/><path class=\"cls-2\" d=\"M32.38,30c3.2-.63,6.31-1.45,9.41-2.27C38,32.89,32.59,35,26.59,36.19Z\"/><path class=\"cls-2\" d=\"M41.8,27.73c3.92-.37,8.14-.88,12.07-1.23-4.14,6.28-11.37,6.44-18.12,7.14Z\"/><polygon class=\"cls-2\" points=\"53.91 26.47 67.56 27.85 65.1 32.42 48.52 31.84 53.91 26.47\"/><polygon class=\"cls-1\" points=\"70.26 31.65 81.52 32.42 83.57 31.19 83.19 36.94 65.1 32.42 67.22 29.16 70.26 31.65\"/><polygon class=\"cls-2\" points=\"97.09 37.5 107.57 44.78 109.75 52.84 98.72 44.59 97.09 37.5\"/><path class=\"cls-2\" d=\"M115.09,51.74c4.6,3.74,6.38,8.55,6.75,14.16-1.64-1.89-3.27-3.78-4.86-5.84Z\"/><path class=\"cls-2\" d=\"M20,38.67c2.19-.79,4.34-1.64,6.49-2.48-2.91,4.47-5.95,9.38-11.64,10.16Z\"/><path class=\"cls-2\" d=\"M26.59,36.19c3-.82,6.18-1.86,9.22-2.46-4.11,5.33-8.16,10.45-15.29,10.64Z\"/><polygon class=\"cls-2\" points=\"35.75 33.64 48.52 31.84 41.62 40.22 28.86 42.11 35.75 33.64\"/><polygon class=\"cls-1\" points=\"48.52 31.84 65.1 32.42 60.34 40.7 41.62 40.22 48.52 31.84\"/><polygon class=\"cls-1\" points=\"65.1 32.42 83.19 36.94 82.26 46.14 60.34 40.7 65.1 32.42\"/><polygon class=\"cls-1\" points=\"83.19 36.94 98.72 44.59 100.14 55.14 82.26 46.14 83.19 36.94\"/><polygon class=\"cls-1\" points=\"98.72 44.59 109.75 52.84 111.4 63.73 100.14 55.14 98.72 44.59\"/><polygon class=\"cls-1\" points=\"109.79 52.84 117 60.03 118.13 70.33 111.44 63.69 109.79 52.84\"/><path class=\"cls-1\" d=\"M117,60.06c4.68,3.88,5.55,9.39,5.35,15.33-1.19-1.77-3-3.21-4.21-5Z\"/><path class=\"cls-2\" d=\"M14.86,46.35l5.55-2c-2.91,4.46-3.77,11.9-10.18,11.13Z\"/><polygon class=\"cls-2\" points=\"20.52 44.37 28.86 42.11 21.9 53.82 14.85 54.7 20.52 44.37\"/><polygon class=\"cls-2\" points=\"28.86 42.11 41.62 40.22 33.67 53.28 21.9 53.82 28.86 42.11\"/><polygon class=\"cls-1\" points=\"41.62 40.22 60.34 40.7 53.64 54.91 33.67 53.28 41.62 40.22\"/><polygon class=\"cls-1\" points=\"60.34 40.7 82.26 46.14 79.84 61.57 53.64 54.91 60.34 40.7\"/><polygon class=\"cls-1\" points=\"82.26 46.14 100.14 55.14 100.12 70.67 79.84 61.57 82.26 46.14\"/><polygon class=\"cls-1\" points=\"100.14 55.14 111.4 63.73 111.45 77.78 100.12 70.67 100.14 55.14\"/><polygon class=\"cls-1\" points=\"111.4 63.73 118.12 70.39 117.76 82.58 111.45 77.78 111.4 63.73\"/><path class=\"cls-1\" d=\"M118.13,70.33C123.64,74,122,80,121.62,85.62a45.87,45.87,0,0,1-3.84-3.15Z\"/><polygon class=\"cls-2\" points=\"10.23 55.47 14.85 54.7 10.47 66.71 6.65 65.66 10.23 55.47\"/><polygon class=\"cls-2\" points=\"14.85 54.7 21.9 53.82 16.39 68.35 10.47 66.71 14.85 54.7\"/><polygon class=\"cls-1\" points=\"21.9 53.82 33.67 53.28 26.76 71.22 16.39 68.35 21.9 53.82\"/><polygon class=\"cls-1\" points=\"33.67 53.28 53.64 54.91 46.18 76.59 26.76 71.22 33.67 53.28\"/><polygon class=\"cls-1\" points=\"53.64 54.91 79.84 61.57 74.53 84.42 46.18 76.59 53.64 54.91\"/><polygon class=\"cls-1\" points=\"79.84 61.57 100.12 70.67 96.61 90.53 74.53 84.42 79.84 61.57\"/><polygon class=\"cls-1\" points=\"100.12 70.67 111.45 77.78 108.61 93.85 96.61 90.53 100.12 70.67\"/><polygon class=\"cls-1\" points=\"111.45 77.78 117.76 82.58 115.28 95.69 108.61 93.85 111.45 77.78\"/><polygon class=\"cls-1\" points=\"117.78 82.47 121.69 85.75 120.38 92.24 118.66 94.08 117.52 96.15 115.31 95.53 117.78 82.47\"/><polygon class=\"cls-2\" points=\"6.65 65.66 10.47 66.71 8.06 79.27 4.49 76.23 6.65 65.66\"/><polygon class=\"cls-2\" points=\"10.47 66.71 16.39 68.35 13.65 83.65 8.06 79.27 10.47 66.71\"/><polygon class=\"cls-1\" points=\"16.39 68.35 26.76 71.22 23.47 90.16 13.65 83.65 16.39 68.35\"/><polygon class=\"cls-1\" points=\"26.76 71.22 46.18 76.59 41.44 99.02 23.47 90.16 26.76 71.22\"/><polygon class=\"cls-1\" points=\"46.18 76.59 74.53 84.42 67.34 106.76 41.44 99.02 46.18 76.59\"/><polygon class=\"cls-1\" points=\"74.53 84.42 96.61 90.53 89.42 109.37 67.34 106.76 74.53 84.42\"/><polygon class=\"cls-1\" points=\"96.61 90.53 108.61 93.85 102.79 109.09 89.42 109.37 96.61 90.53\"/><polygon class=\"cls-1\" points=\"108.61 93.85 115.28 95.69 110.67 108.22 102.79 109.09 108.61 93.85\"/><path class=\"cls-2\" d=\"M4.49,76.23c5.59,2.88,2.79,9.53,3,14.86C6.24,89.55,5,88,3.77,86.43Z\"/><polygon class=\"cls-2\" points=\"8.06 79.27 13.65 83.65 13.61 97.26 7.62 91.04 8.06 79.27\"/><polygon class=\"cls-1\" points=\"13.65 83.65 23.47 90.16 23.59 105.44 13.61 97.26 13.65 83.65\"/><polygon class=\"cls-1\" points=\"23.47 90.16 41.44 99.02 39.89 114.64 23.59 105.44 23.47 90.16\"/><polygon class=\"cls-1\" points=\"67.34 106.76 89.42 109.37 81.46 122.71 61.49 121.24 67.34 106.76\"/><polygon class=\"cls-2\" points=\"89.42 109.37 102.79 109.09 95.53 121.12 81.46 122.71 89.42 109.37\"/><path class=\"cls-2\" d=\"M3.77,86.43c4.47,3.67,4.55,9.45,4.75,14.78-1.41-1.83-1.82-3.95-3.27-5.83Z\"/><path class=\"cls-2\" d=\"M7.62,91c6.28,3.75,6.84,10.37,7.63,17.06-2.24-2.3-4.42-4.82-6.63-6.92Z\"/><polygon class=\"cls-1\" points=\"13.61 97.26 23.59 105.44 25.2 116.18 15.17 108.08 13.61 97.26\"/><polygon class=\"cls-1\" points=\"23.59 105.44 39.89 114.64 39.73 124.19 25.2 116.18 23.59 105.44\"/><polygon class=\"cls-2\" points=\"39.89 114.64 61.49 121.24 57.57 129.61 39.73 124.19 39.89 114.64\"/><polygon class=\"cls-2\" points=\"61.49 121.24 81.46 122.71 74.82 131.03 57.57 129.61 61.49 121.24\"/><path class=\"cls-3\" d=\"M78.14,18.65c5.25-10,19.3-23.69,38.6-16.78C113.52,4.36,97.41,24.23,78.14,18.65Z\"/><path class=\"cls-4\" d=\"M124.07,45.87c-8.8-12.75-24.5-22.56-47.62-14.33,7.16,2.9,14.49,5.78,21.34,10.51a11.65,11.65,0,1,0,16.44,16.06c7.47,11.36,11.43,27.41-1.27,39C134.3,86.33,133.44,60.3,124.07,45.87Z\"/><path class=\"cls-5\" d=\"M81,110.57a11.69,11.69,0,1,0-21.75,4.28c-16.7,1.65-34.07,1.43-44.7-9.93-2.83,6.33,18.22,33.32,54.66,28.92,34.66-2.6,56.85-41.66,59.69-55.7C117.12,98.15,96.8,105.78,81,110.57Z\"/><path class=\"cls-6\" d=\"M82.54,34c-21.12-17.12-57-11.84-73,11.22-9.46,13.65-18.95,40.25,9.15,70C8.44,81.73,16,67.69,25.79,57.56A13.37,13.37,0,1,0,45.12,39.44C60.45,29.73,78.15,33.77,82.54,34Z\"/></g></g></svg>\n            <span class=\"logo-title\">SagooIoT V3 API</span>\n        </div>\n        <div class=\"search-container\">\n            <input type=\"text\" class=\"search-input\" placeholder=\"Search APIs...\" id=\"api-search\">\n        </div>\n        <ul class=\"nav-menu\" id=\"nav-menu\">\n            <!-- 动态生成的导航菜单将在这里 -->\n        </ul>\n    </div>\n    <div class=\"main-content\" id=\"main-content\">\n        <!-- 动态生成的API内容将在这里 -->\n    </div>\n    <script>\n        let swaggerSpec = null;\n        async function loadSwaggerSpec() {\n            try {\n                const response = await fetch('/swagger/doc.json');\n                if (!response.ok) {\n                    throw new Error(`HTTP error! status: ${response.status}`);\n                }\n                swaggerSpec = await response.json();\n                initializeApiDocs();\n            } catch (error) {\n                console.error('Failed to load API specification:', error);\n                document.getElementById('main-content').innerHTML = `\n            <div class=\"error-message\">\n                <h2>Error Loading API Specification</h2>\n                <p>There was a problem loading the API documentation. Please try again later or contact support.</p>\n                <p>Error details: ${error.message}</p>\n            </div>\n        `;\n            }\n        }\n\n        function initializeApiDocs() {\n            generateNavMenu();\n            setupSearch();\n            setupNavigation();\n        }\n\n        document.addEventListener('DOMContentLoaded', loadSwaggerSpec);\n\n        function generateNavMenu() {\n            const navMenu = document.getElementById('nav-menu');\n            navMenu.innerHTML = ”; // 清空现有内容\n            const tags = {};\n\n            for (const path in swaggerSpec.paths) {\n                for (const method in swaggerSpec.paths[path]) {\n                    const api = swaggerSpec.paths[path][method];\n                    api.tags.forEach(tag => {\n                        if (!tags[tag]) {\n                            tags[tag] = [];\n                        }\n                        tags[tag].push({path, method, summary: api.summary});\n                    });\n                }\n            }\n\n            for (const tag in tags) {\n                const li = document.createElement('li');\n                const tagSpan = document.createElement('span');\n                tagSpan.textContent = tag;\n                tagSpan.classList.add('tag');\n                li.appendChild(tagSpan);\n\n                const ul = document.createElement('ul');\n                ul.classList.add('api-list');\n\n                tags[tag].forEach(api => {\n                    const apiLi = document.createElement('li');\n                    apiLi.classList.add('api-item');\n                    const apiA = document.createElement('a');\n                    const methodSpan = document.createElement('span');\n                    methodSpan.textContent = api.method.toUpperCase();\n                    methodSpan.classList.add('method', api.method);\n                    apiA.appendChild(methodSpan);\n                    apiA.appendChild(document.createTextNode(api.summary));\n                    apiA.setAttribute('data-path', api.path);\n                    apiA.setAttribute('data-method', api.method);\n                    apiLi.appendChild(apiA);\n                    ul.appendChild(apiLi);\n                });\n\n                li.appendChild(ul);\n                navMenu.appendChild(li);\n            }\n        }\n\n        function setupNavigation() {\n            document.getElementById('nav-menu').addEventListener('click', function(e) {\n                if (e.target.classList.contains('tag')) {\n                    const apiList = e.target.nextElementSibling;\n                    apiList.classList.toggle('active');\n                } else if (e.target.closest('.api-item')) {\n                    const apiLink = e.target.closest('.api-item').querySelector('a');\n                    const path = apiLink.getAttribute('data-path');\n                    const method = apiLink.getAttribute('data-method');\n                    generateApiContent(path, method);\n                    e.preventDefault();\n                }\n            });\n        }\n\n        function setupSearch() {\n            const searchInput = document.getElementById('api-search');\n            searchInput.addEventListener('input', function() {\n                const searchTerm = this.value.toLowerCase().trim();\n                const apiItems = document.querySelectorAll('.api-item');\n                const tags = document.querySelectorAll('.tag');\n\n                if (searchTerm === ”) {\n                    // 搜索框为空，恢复所有项目到原始状态\n                    apiItems.forEach(item => {\n                        item.classList.remove('hidden');\n                    });\n                    tags.forEach(tag => {\n                        tag.classList.remove('hidden');\n                        tag.nextElementSibling.classList.remove('active');\n                    });\n                } else {\n                    // 执行搜索\n                    apiItems.forEach(item => {\n                        const apiText = item.textContent.toLowerCase();\n                        if (apiText.includes(searchTerm)) {\n                            item.classList.remove('hidden');\n                            item.closest('.api-list').classList.remove('hidden');\n                            item.closest('.api-list').classList.add('active');\n                            item.closest('li').querySelector('.tag').classList.remove('hidden');\n                        } else {\n                            item.classList.add('hidden');\n                        }\n                    });\n\n                    tags.forEach(tag => {\n                        const visibleApis = tag.nextElementSibling.querySelectorAll('.api-item:not(.hidden)');\n                        if (visibleApis.length === 0) {\n                            tag.classList.add('hidden');\n                            tag.nextElementSibling.classList.remove('active');\n                        } else {\n                            tag.classList.remove('hidden');\n                        }\n                    });\n                }\n            });\n        }\n\n        document.addEventListener('DOMContentLoaded', loadSwaggerSpec);\n\n        function generateApiContent(path, method) {\n            const mainContent = document.getElementById('main-content');\n            mainContent.innerHTML = ”;\n\n            const api = swaggerSpec.paths[path][method];\n            const endpoint = document.createElement('div');\n            endpoint.classList.add('endpoint');\n\n            const header = document.createElement('div');\n            header.classList.add('endpoint-header');\n            const methodSpan = document.createElement('span');\n            methodSpan.classList.add('method', method);\n            methodSpan.textContent = method.toUpperCase();\n            const pathSpan = document.createElement('span');\n            pathSpan.classList.add('endpoint-path');\n            pathSpan.textContent = path;\n            header.appendChild(methodSpan);\n            header.appendChild(pathSpan);\n\n            const title = document.createElement('h2');\n            title.textContent = api.summary;\n\n            const description = document.createElement('p');\n            description.textContent = api.description;\n\n            endpoint.appendChild(header);\n            endpoint.appendChild(title);\n            endpoint.appendChild(description);\n\n            if (api.parameters && api.parameters.length > 0) {\n                const paramsTitle = document.createElement('h3');\n                paramsTitle.textContent = 'Parameters';\n                endpoint.appendChild(paramsTitle);\n\n                const paramsTable = document.createElement('table');\n                paramsTable.classList.add('params-table');\n                const thead = document.createElement('thead');\n                thead.innerHTML = '<tr><th>Name</th><th>Type</th><th>Description</th><th>Required</th></tr>';\n                paramsTable.appendChild(thead);\n\n                const tbody = document.createElement('tbody');\n                api.parameters.forEach(param => {\n                    const tr = document.createElement('tr');\n                    tr.innerHTML = `\n                        <td>${param.name}</td>\n                        <td>${param.type || '-'}</td>\n                        <td>${param.description || '-'}</td>\n                        <td>${param.required ? '<span class=\"required\">Yes</span>' : 'No'}</td>\n                    `;\n                    tbody.appendChild(tr);\n                });\n                paramsTable.appendChild(tbody);\n                endpoint.appendChild(paramsTable);\n            }\n\n            const responsesTitle = document.createElement('h3');\n            responsesTitle.textContent = 'Responses';\n            endpoint.appendChild(responsesTitle);\n\n            for (const statusCode in api.responses) {\n                const responseContainer = document.createElement('div');\n                responseContainer.classList.add('response-container');\n\n                const statusCodeSpan = document.createElement('span');\n                statusCodeSpan.classList.add('response-code');\n                statusCodeSpan.textContent = statusCode;\n\n                const responseDescription = document.createElement('span');\n                responseDescription.textContent = api.responses[statusCode].description;\n\n                responseContainer.appendChild(statusCodeSpan);\n                responseContainer.appendChild(responseDescription);\n\n                // 添加响应示例\n                if (api.responses[statusCode].example) {\n                    const exampleResponse = document.createElement('div');\n                    exampleResponse.classList.add('example-response');\n                    const exampleTitle = document.createElement('h4');\n                    exampleTitle.textContent = 'Example Response:';\n                    const examplePre = document.createElement('pre');\n                    examplePre.textContent = JSON.stringify(api.responses[statusCode].example, null, 2);\n                    exampleResponse.appendChild(exampleTitle);\n                    exampleResponse.appendChild(examplePre);\n                    responseContainer.appendChild(exampleResponse);\n                }\n\n                endpoint.appendChild(responseContainer);\n            }\n\n            // 添加 \"Try it out\" 功能\n            const tryItOut = document.createElement('div');\n            tryItOut.classList.add('try-it-out');\n            const tryItOutTitle = document.createElement('h3');\n            tryItOutTitle.textContent = 'Try it out';\n            tryItOut.appendChild(tryItOutTitle);\n\n            const form = document.createElement('form');\n            api.parameters.forEach(param => {\n                const label = document.createElement('label');\n                label.textContent = `${param.name}${param.required ? ' *' : ”}:`;\n                const input = document.createElement('input');\n                input.type = param.type === 'number' ? 'number' : 'text';\n                input.name = param.name;\n                input.required = param.required;\n                input.placeholder = param.description;\n                form.appendChild(label);\n                form.appendChild(input);\n            });\n\n            const submitButton = document.createElement('button');\n            submitButton.type = 'submit';\n            submitButton.textContent = 'Send Request';\n            form.appendChild(submitButton);\n\n            const responseArea = document.createElement('div');\n            responseArea.classList.add('response-area');\n            responseArea.style.display = 'none';\n            const responseTitle = document.createElement('h4');\n            responseTitle.textContent = 'Response';\n            const responsePre = document.createElement('pre');\n            responseArea.appendChild(responseTitle);\n            responseArea.appendChild(responsePre);\n\n            form.addEventListener('submit', function(e) {\n                e.preventDefault();\n                const formData = new FormData(form);\n                const params = Object.fromEntries(formData);\n\n                // 执行实际的API调用\n                executeApiCall(path, method, params)\n                    .then(response => {\n                        responsePre.textContent = JSON.stringify(response, null, 2);\n                        responseArea.style.display = 'block';\n                    })\n                    .catch(error => {\n                        responsePre.textContent = `Error: ${error.message}`;\n                        responsePre.classList.add('error-message');\n                        responseArea.style.display = 'block';\n                    });\n            });\n\n            tryItOut.appendChild(form);\n            tryItOut.appendChild(responseArea);\n            endpoint.appendChild(tryItOut);\n\n            mainContent.appendChild(endpoint);\n        }\n\n        // API调用的函数\n        async function executeApiCall(path, method, params) {\n            // 这里应该是您的API基础URL\n            const baseUrl = window.location.origin;  // 请替换为实际的API基础URL\n            const url = new URL(path, baseUrl);\n\n            // 对于GET请求，将参数添加到URL中\n            if (method.toLowerCase() === 'get') {\n                Object.keys(params).forEach(key => url.searchParams.append(key, params[key]));\n            }\n\n            const options = {\n                method: method.toUpperCase(),\n                headers: {\n                    'Content-Type': 'application/json',\n                    // 如果需要认证，在这里添加认证头\n                    // 'Authorization': 'Bearer YOUR_TOKEN_HERE'\n                },\n            };\n\n            // 对于非GET请求，将参数放在请求体中\n            if (method.toLowerCase() !== 'get') {\n                options.body = JSON.stringify(params);\n            }\n\n            const response = await fetch(url, options);\n\n            if (!response.ok) {\n                throw new Error(`HTTP error! status: ${response.status}`);\n            }\n\n            return await response.json();\n        }\n\n        document.addEventListener('DOMContentLoaded', function() {\n            generateNavMenu();\n            setupSearch();\n\n            document.getElementById('nav-menu').addEventListener('click', function(e) {\n                if (e.target.classList.contains('tag')) {\n                    const apiList = e.target.nextElementSibling;\n                    apiList.classList.toggle('active');\n                } else if (e.target.closest('.api-item')) {\n                    const apiLink = e.target.closest('.api-item').querySelector('a');\n                    const path = apiLink.getAttribute('data-path');\n                    const method = apiLink.getAttribute('data-method');\n                    generateApiContent(path, method);\n                    e.preventDefault();\n                }\n            });\n\n        });\n    </script>\n</body>\n</html>"
const indexTempl = "<!-- HTML for static distribution bundle build -->\n<!DOCTYPE html>\n<html lang=\"en\">\n  <head>\n    <meta charset=\"UTF-8\">\n    <title>Swagger UI</title>\n    <link rel=\"stylesheet\" type=\"text/css\" href=\"./swagger-ui.css\" />\n    <link rel=\"stylesheet\" type=\"text/css\" href=\"index.css\" />\n    <link rel=\"icon\" type=\"image/png\" href=\"./favicon-32x32.png\" sizes=\"32x32\" />\n    <link rel=\"icon\" type=\"image/png\" href=\"./favicon-16x16.png\" sizes=\"16x16\" />\n  </head>\n\n  <body>\n    <div id=\"swagger-ui\"></div>\n    <script src=\"./swagger-ui-bundle.js\" charset=\"UTF-8\"> </script>\n    <script src=\"./swagger-ui-standalone-preset.js\" charset=\"UTF-8\"> </script>\n    <script src=\"./swagger-initializer.js\" charset=\"UTF-8\"> </script>\n  </body>\n</html>\n"
