package bundle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

// EsbuildRunner compiles the bundler into the provider. It never invokes
// Node.js, npm or an external executable, or loads JavaScript plugin configs.
type EsbuildRunner struct{}

func (EsbuildRunner) Run(ctx context.Context, request RunRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	build, err := api.Context(api.BuildOptions{
		AbsWorkingDir: request.WorkingDirectory,
		EntryPoints:   []string{request.Entrypoint},
		Outfile:       request.OutputFile,
		Bundle:        true,
		Platform:      api.PlatformNode,
		Format:        api.FormatCommonJS,
		LogLevel:      api.LogLevelSilent,
		// Inspect outputs before writing: the ZIP contract is exactly index.js.
		// In particular, never silently discard a generated CSS or asset file.
		Write: false,
	})
	if err != nil {
		return formatBuildErrors(err.Errors)
	}
	defer build.Dispose()
	cancelled := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		build.Cancel()
		close(cancelled)
	})
	defer func() {
		// Don't dispose the build while a cancellation callback still uses it.
		if !stop() {
			<-cancelled
		}
	}()

	result := build.Rebuild()
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(result.Errors) > 0 {
		return formatBuildErrors(result.Errors)
	}
	if len(result.OutputFiles) != 1 || filepath.Clean(result.OutputFiles[0].Path) != filepath.Clean(request.OutputFile) {
		return errors.New("esbuild must produce a single JavaScript bundle; CSS and additional assets are not supported")
	}
	if err := os.WriteFile(request.OutputFile, result.OutputFiles[0].Contents, 0o644); err != nil {
		return fmt.Errorf("write esbuild bundle: %w", err)
	}
	return nil
}

func formatBuildErrors(messages []api.Message) error {
	formatted := api.FormatMessages(messages, api.FormatMessagesOptions{Kind: api.ErrorMessage, Color: false})
	return fmt.Errorf("esbuild: %s", strings.TrimSpace(strings.Join(formatted, "\n")))
}
