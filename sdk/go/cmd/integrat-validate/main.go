// CLI-инструмент для валидации integrat.yaml спецификаций.
//
// Использование:
//
//	integrat-validate [--offline] <file.yaml> [file2.yaml ...]
//	integrat-validate ./integrat.yaml
//	integrat-validate --offline /path/to/integrat.yaml
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/plagness/Integrat/sdk/go/internal/validator"
)

func main() {
	offline := flag.Bool("offline", false, "Пропустить проверку доступности base_url")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Использование: %s [--offline] <file.yaml> [file2.yaml ...]\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "Валидирует integrat.yaml спецификации плагинов.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	files := flag.Args()
	if len(files) == 0 {
		// По умолчанию ищем integrat.yaml в текущей директории
		if _, err := os.Stat("integrat.yaml"); err == nil {
			files = []string{"integrat.yaml"}
		} else {
			flag.Usage()
			os.Exit(1)
		}
	}

	hasErrors := false
	for _, path := range files {
		ok := validateFile(path, *offline)
		if !ok {
			hasErrors = true
		}
	}

	if hasErrors {
		os.Exit(1)
	}
}

func validateFile(path string, offline bool) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s: %v\n", path, err)
		return false
	}

	spec, result := validator.ValidateBytes(data)
	if spec == nil {
		// Ошибка парсинга
		fmt.Printf("✗ %s\n", path)
		for _, e := range result.Errors {
			fmt.Printf("  ✗ %s\n", e)
		}
		return false
	}

	// Вывод результатов
	fmt.Printf("─ %s (%s)\n", path, spec.Plugin.Slug)

	if result.OK() {
		fmt.Printf("  ✓ Schema valid\n")
	} else {
		for _, e := range result.Errors {
			fmt.Printf("  ✗ %s\n", e)
		}
	}

	for _, w := range result.Warnings {
		fmt.Printf("  ⚠ %s\n", w)
	}

	if result.OK() {
		fmt.Printf("  ✓ Plugin slug: %s (format OK)\n", spec.Plugin.Slug)
		fmt.Printf("  ✓ %d endpoints, all slugs unique\n", len(spec.Endpoints))

		schemaCount := 0
		for _, ep := range spec.Endpoints {
			if ep.ParamsSchema.Kind != 0 {
				schemaCount++
			}
		}
		if schemaCount > 0 {
			fmt.Printf("  ✓ params_schema valid JSON Schema (%d/%d endpoints)\n", schemaCount, len(spec.Endpoints))
		}

		if len(spec.ConfigFields) > 0 {
			fmt.Printf("  ✓ %d config_fields\n", len(spec.ConfigFields))
		}

		if !offline {
			// Проверяем base_url только если не env var
			base := spec.Provider.BaseURL
			if len(base) > 0 && base[0] != '$' {
				fmt.Printf("  ⊘ base_url reachability: skipped (используйте --offline для явного пропуска)\n")
			} else {
				fmt.Printf("  ⊘ base_url: %s (env var, проверка пропущена)\n", base)
			}
		} else {
			fmt.Printf("  ⊘ base_url reachability: пропущено (--offline)\n")
		}
	}

	return result.OK()
}
