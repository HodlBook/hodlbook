package handler

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

type Renderer struct {
	templatesDir string
	templates    map[string]*template.Template
	mu           sync.RWMutex
	funcs        template.FuncMap
}

func NewRenderer(templatesDir string) *Renderer {
	return &Renderer{
		templatesDir: templatesDir,
		templates:    make(map[string]*template.Template),
		funcs:        defaultFuncs(),
	}
}

func defaultFuncs() template.FuncMap {
	return template.FuncMap{
		"formatCurrency": func(value float64) string {
			return formatCurrency(value, "USD")
		},
		"formatPercent": func(value float64) string {
			return formatPercent(value)
		},
		"isPositive": func(value float64) bool {
			return value >= 0
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"mul": func(a, b float64) float64 {
			return a * b
		},
		"div": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
	}
}

func formatCurrency(value float64, currency string) string {
	symbol := "$"
	switch currency {
	case "EUR":
		symbol = "€"
	case "GBP":
		symbol = "£"
	}
	if value < 0 {
		return "-" + symbol + formatNumber(-value)
	}
	return symbol + formatNumber(value)
}

func formatPrice(value float64) string {
	if value == 0 {
		return "$0"
	}
	abs := value
	sign := ""
	if value < 0 {
		abs = -value
		sign = "-"
	}
	if abs >= 1 {
		return sign + "$" + formatFloat(abs)
	}
	return sign + "$" + fmt.Sprintf("%.2f", abs)
}

func formatExchangeRate(value float64) string {
	if value >= 1 {
		return formatFloat(value)
	}
	if value == 0 {
		return "0"
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func formatNumber(value float64) string {
	if value >= 1000000 {
		return intToStr(int(value/1000)) + "K"
	}
	if value >= 1000 {
		return intToStr(int(value))
	}
	return formatFloat(value)
}

func formatFloat(value float64) string {
	if value == float64(int(value)) {
		return template.HTMLEscapeString(floatToStr(value, 0))
	}
	return template.HTMLEscapeString(floatToStr(value, 2))
}

func floatToStr(value float64, decimals int) string {
	format := "%." + string(rune('0'+decimals)) + "f"
	return sprintf(format, value)
}

func sprintf(format string, value float64) string {
	return template.HTMLEscapeString(sprintfRaw(format, value))
}

func sprintfRaw(format string, value float64) string {
	switch format {
	case "%.0f":
		return intToStr(int(value))
	case "%.2f":
		return floatToStr2(value)
	}
	return floatToStr2(value)
}

func intToStr(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func floatToStr2(v float64) string {
	intPart := int(v)
	fracPart := int((v - float64(intPart)) * 100)
	if fracPart < 0 {
		fracPart = -fracPart
	}
	result := intToStr(intPart) + "."
	if fracPart < 10 {
		result += "0"
	}
	result += intToStr(fracPart)
	return result
}

func formatPercent(value float64) string {
	sign := ""
	if value > 0 {
		sign = "+"
	}
	return sign + floatToStr2(value) + "%"
}

func (r *Renderer) loadTemplate(name string) (*template.Template, error) {
	r.mu.RLock()
	if tmpl, ok := r.templates[name]; ok {
		r.mu.RUnlock()
		return tmpl, nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if tmpl, ok := r.templates[name]; ok {
		return tmpl, nil
	}

	layoutPath := filepath.Join(r.templatesDir, "layouts", "base.html")
	pagePath := filepath.Join(r.templatesDir, "pages", name+".html")
	componentsGlob := filepath.Join(r.templatesDir, "partials", "components", "*.html")

	tmpl, err := template.New("").Funcs(r.funcs).ParseGlob(componentsGlob)
	if err != nil {
		tmpl = template.New("").Funcs(r.funcs)
	}

	tmpl, err = tmpl.ParseFiles(layoutPath, pagePath)
	if err != nil {
		return nil, err
	}

	r.templates[name] = tmpl
	return tmpl, nil
}

func (r *Renderer) Render(w io.Writer, name string, data any) error {
	tmpl, err := r.loadTemplate(name)
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, "base", data)
}

func (r *Renderer) RenderPartial(w io.Writer, name string, data any) error {
	partialPath := filepath.Join(r.templatesDir, "partials", name+".html")
	tmpl, err := template.New("").Funcs(r.funcs).ParseFiles(partialPath)
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, filepath.Base(partialPath), data)
}

func (r *Renderer) HTML(c *gin.Context, code int, name string, data any) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(code)
	if err := r.Render(c.Writer, name, data); err != nil {
		c.String(500, "Template error: %v", err)
	}
}

func (r *Renderer) Partial(c *gin.Context, code int, name string, data any) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(code)
	if err := r.RenderPartial(c.Writer, name, data); err != nil {
		c.String(500, "Template error: %v", err)
	}
}
