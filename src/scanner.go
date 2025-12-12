package main

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

func findKustomizationFiles(root string) ([]string, error) {
	return findKustomizationFilesWithExclusions(root, []string{".git"})
}

func findKustomizationFilesWithExclusions(root string, excludedDirs []string) ([]string, error) {
	excludedBase := make(map[string]struct{}, len(excludedDirs))
	excludedRel := make(map[string]struct{}, len(excludedDirs))
	for _, e := range excludedDirs {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		clean := filepath.Clean(e)
		b := filepath.Base(clean)
		// Only basename-skip .git (and similar) to avoid accidentally skipping
		// arbitrary directories that share the same basename as config.OutputDir.
		if b == ".git" {
			excludedBase[b] = struct{}{}
		}
		rel := filepath.ToSlash(clean)
		rel = strings.TrimPrefix(rel, "./")
		rel = strings.Trim(rel, "/")
		if rel != "" && rel != "." {
			excludedRel[rel] = struct{}{}
		}
	}

	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			base := filepath.Base(path)
			if _, ok := excludedBase[base]; ok {
				return fs.SkipDir
			}
			rel := relDir(root, path)
			if rel != "" {
				if _, ok := excludedRel[rel]; ok {
					return fs.SkipDir
				}
			}
			return nil
		}
		base := filepath.Base(path)
		if base == "kustomization.yaml" || base == "kustomization.yml" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Ensure stable ordering
	sort.Strings(files)
	return files, nil
}

func findRootKustomizations(root string) ([]string, error) {
	files, err := findKustomizationFiles(root)
	if err != nil {
		return nil, err
	}
	// Reuse the same normalization/dedupe logic
	return kustomizationDirsFromFiles(files, root), nil
}

func dedupeTopLevelDirs(paths []string) []string {
	if len(paths) == 0 {
		return paths
	}
	sort.Strings(paths)
	var keep []string
	for _, p := range paths {
		p = strings.TrimSuffix(p, "/")
		skip := false
		for _, k := range keep {
			if p == k || strings.HasPrefix(p, k+"/") {
				skip = true
				break
			}
		}
		if !skip {
			keep = append(keep, p)
		}
	}
	return keep
}

// relDir returns a normalized, slash-separated, trimmed directory path relative to base.
func relDir(base, dir string) string {
	rel, err := filepath.Rel(base, dir)
	if err != nil {
		rel = dir
	}
	rel = filepath.ToSlash(rel)
	rel = strings.TrimPrefix(rel, "./")
	rel = strings.Trim(rel, "/")
	if rel == "." {
		rel = ""
	}
	return rel
}

// kustomizationDirsFromFiles converts kustomization file paths to unique, sorted directories.
func kustomizationDirsFromFiles(files []string, base string) []string {
	dirs := make([]string, 0, len(files))
	for _, f := range files {
		d := filepath.Dir(f)
		dirs = append(dirs, relDir(base, d))
	}
	return uniqueSorted(dirs)
}

func uniqueSorted(in []string) []string {
	m := make(map[string]struct{}, len(in))
	for _, s := range in {
		if s == "." {
			s = ""
		}
		m[s] = struct{}{}
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
