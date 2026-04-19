package bengine

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archives"
)

func ExtractArchiveFromReader(ctx context.Context, dest string, fileName string, r io.Reader) error {
	// Detect archive filetype
	format, reader, err := archives.Identify(ctx, fileName, r)
	if err != nil {
		return err
	}

	ex, ok := format.(archives.Extractor)
	if !ok {
		return fmt.Errorf("file not extractable: %w", err)
	}

	destAbs, err := filepath.Abs(dest)
	if err != nil {
		return fmt.Errorf("could not pull absolute path from dest %q: %w", dest, err)
	}

	err = os.MkdirAll(destAbs, 0o755)
	if err != nil {
		return fmt.Errorf("could not create directory of archive: %w", err)
	}

	err = ex.Extract(ctx, reader, func(ctx context.Context, info archives.FileInfo) error {
		if info.NameInArchive == "" || info.NameInArchive == "." {
			return nil
		}

		targetFileName := filepath.Clean(info.NameInArchive)
		target := filepath.Join(destAbs, targetFileName)

		abs, err := filepath.Abs(target)
		if err != nil {
			return fmt.Errorf("target path of %s is not absolute: %w", target, err)
		}

		// Avoid traversal attacks
		if !strings.HasPrefix(abs, destAbs+string(os.PathSeparator)) && abs != destAbs {
			return fmt.Errorf("unsafe path in archive: %q %q", info.NameInArchive, abs)
		}

		ifo, err := info.Stat()
		if err != nil {
			return fmt.Errorf("could not stat file %s: %w", info.NameInArchive, err)
		}

		if ifo.IsDir() {
			return os.MkdirAll(abs, 0o755)
		}

		// Ensure parent dir exists
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			return fmt.Errorf("parent directory of `%s` is not present: %w", info.NameInArchive, err)
		}

		// Opening file inside of archive
		rc, err := info.Open()
		if err != nil {
			return fmt.Errorf("open file %s failed: %w", info.NameInArchive, err)
		}

		defer rc.Close()

		// Opening destination file
		out, err := os.OpenFile(abs, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, ifo.Mode().Perm())
		if err != nil {
			return fmt.Errorf("could not open destination file `%q`: %w", abs, err)
		}

		defer out.Close()

		if _, err := io.Copy(out, rc); err != nil {
			return fmt.Errorf("could not write file on %q: %w", abs, err)
		}

		return nil
	})

	return err
}
