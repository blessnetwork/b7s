package fstore

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/blocklessnetworking/b7s/models/blockless"
)

// Install will download and install function identified by the manifest/CID.
func (h *FStore) Install(address string, cid string) error {

	h.log.Debug().
		Str("cid", cid).
		Str("address", address).
		Msg("installing function")

	// Retrieve function manifest from the given address.
	var manifest blockless.FunctionManifest
	err := h.getJSON(address, &manifest)
	if err != nil {
		return fmt.Errorf("could not retrieve manifest: %w", err)
	}

	// If the runtime URL is specified, use it to fill in the deployment info.
	if manifest.Runtime.URL != "" {
		err = updateDeploymentInfo(&manifest, address)
		if err != nil {
			return fmt.Errorf("could not update deployment info: %w", err)
		}
	}

	// Download the function identified by the manifest.
	functionPath, err := h.download(cid, manifest)
	if err != nil {
		return fmt.Errorf("could not download function: %w", err)
	}

	out := filepath.Join(h.workdir, cid)

	// Unpack the .tar.gz archive.
	// TODO: Would be good to know the content of the .tar.gz archive.
	// We're unpacking the archive here and storing the path to the .tar.gz in the DB.
	err = h.unpackArchive(functionPath, out)
	if err != nil {
		return fmt.Errorf("could not unpack gzip archive (file: %s): %w", functionPath, err)
	}

	manifest.Deployment.File = functionPath

	// Store the function record.
	fn := functionRecord{
		CID:      cid,
		URL:      address,
		Manifest: manifest,
		Archive:  functionPath,
		Files:    out,
	}
	err = h.saveFunction(fn)
	if err != nil {
		h.log.Error().
			Err(err).
			Str("cid", cid).
			Msg("could not save function record")
	}

	h.log.Debug().
		Str("cid", cid).
		Str("address", address).
		Msg("installed function")

	return nil
}

// Installed checks if the function with the given CID is installed.
func (h *FStore) Installed(cid string) (bool, error) {

	fn, err := h.getFunction(cid)
	if err != nil && errors.Is(err, blockless.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("could not get function from store: %w", err)
	}

	haveArchive, haveFiles, err := h.checkFunctionFiles(*fn)
	if err != nil {
		return false, fmt.Errorf("could not verify function cache: %w", err)
	}

	// If we don't have all files found, treat it as not installed.
	if !haveArchive || !haveFiles {
		return false, nil
	}

	// We have the function in the database and all files - we're good.
	return true, nil
}
