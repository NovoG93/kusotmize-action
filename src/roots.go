package main

import "strings"

func mapRootsToRepoRootRelative(workingDir string, roots []string) []string {
	wd := normalizeRepoRelativeDir(workingDir)
	out := make([]string, 0, len(roots))
	for _, r := range roots {
		root := normalizeRepoRelativeDir(r)
		if root == "." {
			root = ""
		}

		// Root "" means the base directory itself.
		if root == "" {
			out = append(out, wd)
			continue
		}
		if wd == "." {
			out = append(out, root)
			continue
		}
		out = append(out, wd+"/"+root)
	}
	return out
}

func normalizeRepoRelativeDir(dir string) string {
	d := strings.TrimSpace(dir)
	if d == "" {
		return "."
	}
	d = strings.ReplaceAll(d, "\\", "/")
	d = strings.TrimPrefix(d, "./")
	d = strings.Trim(d, "/")
	if d == "" || d == "." {
		return "."
	}
	return d
}
