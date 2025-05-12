package fstore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"go.opentelemetry.io/otel/trace"

	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/telemetry/b7ssemconv"
)

func (f *FStore) Sync(ctx context.Context, haltOnError bool) error {

	functions, err := f.store.RetrieveFunctions(ctx)
	if err != nil {
		return fmt.Errorf("could not retrieve functions: %w", err)
	}

	var (
		multierr *multierror.Error
		total    int
	)

	for _, function := range functions {
		err := f.sync(ctx, function)
		if err != nil {
			// Add CID info to error to know what erred.
			wrappedErr := fmt.Errorf("could not sync function (cid: %s): %w", function.CID, err)
			if haltOnError {
				return wrappedErr
			}

			multierr = multierror.Append(multierr, wrappedErr)
			continue
		}

		total++
	}

	f.functionCount.Do(func() {
		f.metrics.IncrCounter(functionsInstalledMetric, float32(total))
	})

	return multierr.ErrorOrNil()
}

// Sync will verify that the function identified by `cid` is still found on the local filesystem.
// If the function archive of function files are missing, they will be recreated.
func (f *FStore) sync(ctx context.Context, fn bls.FunctionRecord) error {

	ctx, span := f.tracer.Start(ctx, spanSync, trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(b7ssemconv.FunctionCID.String(fn.CID)))
	defer span.End()

	// Read the function directly from storage - we don't want to update the timestamp
	// since this is a 'maintenance' access.

	f.log.Debug().
		Str("cid", fn.CID).
		Str("archive", fn.Archive).
		Str("files", fn.Files).
		Msg("checking function installation")

	haveArchive, haveFiles, err := f.checkFunctionFiles(fn)
	if err != nil {
		return fmt.Errorf("could not verify function cache: %w", err)
	}

	// If both archive and files are there - we're done.
	if haveArchive && haveFiles {
		f.log.Debug().Str("cid", fn.CID).Msg("function files found, done")
		return nil
	}

	f.log.Debug().
		Bool("have_archive", haveArchive).
		Bool("have_files", haveFiles).
		Str("cid", fn.CID).
		Msg("function installation missing files")

	// If we don't have the archive - redownload it.
	if !haveArchive {
		path, err := f.download(ctx, fn.CID, fn.Manifest)
		if err != nil {
			return fmt.Errorf("could not download the function archive (cid: %v): %w", fn.CID, err)
		}

		// Update path in case it changed.
		fn.Archive = f.cleanPath(path)
	}

	// If we don't have the files OR if we redownloaded the archive - recreate the files.
	if !haveFiles || !haveArchive {

		archivePath := filepath.Join(f.workdir, fn.Archive)
		files := filepath.Join(f.workdir, fn.CID)

		f.log.Info().
			Str("archive", archivePath).
			Str("fn_archive", fn.Archive).
			Msg("archive path to use")

		err = f.unpackArchive(archivePath, files)
		if err != nil {
			return fmt.Errorf("could not unpack gzip archive (cid: %v, file: %s): %w", fn.CID, fn.Archive, err)
		}

		fn.Files = files
	}

	// Save the updated function record.
	err = f.saveFunction(ctx, fn)
	if err != nil {
		return fmt.Errorf("could not save function (cid: %v): %w", fn.CID, err)
	}

	return nil
}

// checkFunctionFiles checks if the files required by the function are found on local storage.
// It returns two booleans indicating presence of the archive file, the unpacked files, and a potential error.
func (f *FStore) checkFunctionFiles(fn bls.FunctionRecord) (bool, bool, error) {

	// Check if the archive is found.
	archiveFound := true

	apath := filepath.Join(f.workdir, fn.Archive)
	_, err := os.Stat(apath)
	if err != nil && os.IsNotExist(err) {
		archiveFound = false
	} else if err != nil {
		return false, false, fmt.Errorf("could not stat function archive: %w", err)
	}

	// NOTE: We could check that it's a regular file (plus checksum), but lets not go overboard for now.

	// Check if the files are found.
	filesFound := true

	fpath := filepath.Join(f.workdir, fn.Files)
	_, err = os.Stat(fpath)
	if err != nil && os.IsNotExist(err) {
		filesFound = false
	} else if err != nil {
		return false, false, fmt.Errorf("could not stat function files: %w", err)
	}

	return archiveFound, filesFound, nil
}
