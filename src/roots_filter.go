package main

import "strings"

func selectRootsForChangedFiles(roots []string, changedFiles []string) []string {
	if len(roots) == 0 || len(changedFiles) == 0 {
		return []string{}
	}

	selected := make(map[string]bool, len(roots))

	for _, f := range changedFiles {
		file := normalizeRepoRelativePath(f)
		best := ""
		bestLen := -1

		for _, r := range roots {
			root := normalizeRepoRelativeDir(r)
			if !rootPrefixesFile(root, file) {
				continue
			}
			l := len(root)
			if root == "." {
				l = 0
			}
			if l > bestLen {
				bestLen = l
				best = root
			}
		}

		if best != "" {
			selected[best] = true
		}
	}

	out := make([]string, 0, len(selected))
	for _, r := range roots {
		root := normalizeRepoRelativeDir(r)
		if selected[root] {
			out = append(out, root)
		}
	}
	return out
}

func rootPrefixesFile(root, file string) bool {
	root = normalizeRepoRelativeDir(root)
	file = normalizeRepoRelativePath(file)

	if root == "." {
		return true
	}
	if file == "" || file == "." {
		return root == "."
	}
	return file == root || strings.HasPrefix(file, root+"/")
}

func normalizeRepoRelativePath(path string) string {
	p := strings.TrimSpace(path)
	p = strings.ReplaceAll(p, "\\", "/")
	p = strings.TrimPrefix(p, "./")
	p = strings.TrimPrefix(p, "/")
	p = strings.Trim(p, "/")
	return p
}
