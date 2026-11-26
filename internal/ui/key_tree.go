package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Ryequ/redis-cockpit/internal/redis/adapters"
)

const keyTreeSeparator = ":"

type keyTreeNode struct {
	name     string
	path     string
	keyIndex int
	children map[string]*keyTreeNode
	order    []string
}

type keyTreeRow struct {
	depth      int
	label      string
	path       string
	keyIndex   int
	expandable bool
	expanded   bool
}

func buildKeyTree(keys []adapters.KeyInfo) *keyTreeNode {
	root := &keyTreeNode{keyIndex: -1, children: map[string]*keyTreeNode{}, order: []string{}}
	for index, key := range keys {
		node := root
		path := ""
		for _, part := range strings.Split(key.Name, keyTreeSeparator) {
			if part == "" {
				continue
			}
			if path == "" {
				path = part
			} else {
				path += keyTreeSeparator + part
			}
			child := node.children[part]
			if child == nil {
				child = &keyTreeNode{name: part, path: path, keyIndex: -1, children: map[string]*keyTreeNode{}, order: []string{}}
				node.children[part] = child
				node.order = append(node.order, part)
			}
			node = child
		}
		node.keyIndex = index
	}
	return root
}

func (n *keyTreeNode) isExpanded(expanded map[string]bool) bool {
	return expanded[n.path]
}

func (n *keyTreeNode) branchPaths() []string {
	var paths []string
	for _, child := range n.children {
		if len(child.children) > 0 {
			paths = append(paths, child.path)
		}
		paths = append(paths, child.branchPaths()...)
	}
	return paths
}

func (n *keyTreeNode) visibleRows(expanded map[string]bool) []keyTreeRow {
	return n.visibleRowsAtDepth(expanded, 0)
}

func (n *keyTreeNode) visibleRowsAtDepth(expanded map[string]bool, depth int) []keyTreeRow {
	names := append([]string{}, n.order...)
	sort.SliceStable(names, func(i, j int) bool {
		left, right := n.children[names[i]], n.children[names[j]]
		leftBranch, rightBranch := len(left.children) > 0, len(right.children) > 0
		if leftBranch != rightBranch {
			return leftBranch
		}
		return names[i] < names[j]
	})

	var rows []keyTreeRow
	for _, name := range names {
		child := n.children[name]
		row := keyTreeRow{
			depth:      depth,
			label:      name,
			path:       child.path,
			keyIndex:   child.keyIndex,
			expandable: len(child.children) > 0,
			expanded:   child.isExpanded(expanded),
		}
		rows = append(rows, row)
		if row.expandable && row.expanded {
			rows = append(rows, child.visibleRowsAtDepth(expanded, depth+1)...)
		}
	}
	return rows
}

func allKeyTreeBranchPaths(keys []adapters.KeyInfo) map[string]bool {
	paths := map[string]bool{}
	for _, path := range buildKeyTree(keys).branchPaths() {
		paths[path] = true
	}
	return paths
}

func keyViewToggleLabel(m Model) string {
	if m.keyTreeMode {
		return "[Flat]"
	}
	return "[Tree]"
}

func (m Model) visibleKeyRows() []keyTreeRow {
	if !m.keyTreeMode {
		rows := make([]keyTreeRow, 0, len(m.keys))
		for i, key := range m.keys {
			rows = append(rows, keyTreeRow{
				label:    key.Name,
				path:     key.Name,
				keyIndex: i,
			})
		}
		return rows
	}
	return buildKeyTree(m.keys).visibleRows(m.keyTreeExpanded)
}

func (m Model) keyListHeight() int {
	reserved := headerRows + footerRows + panelVerticalOverhead + 10
	return maxInt(m.height-reserved, 4)
}

func (m *Model) ensureKeyVisible() {
	rows := m.visibleKeyRows()
	rowIndex := -1
	for i, row := range rows {
		if row.keyIndex == m.selectedKey {
			rowIndex = i
			break
		}
	}
	if rowIndex < 0 {
		return
	}
	height := m.keyListHeight()
	maxScroll := maxInt(len(rows)-height, 0)
	if m.keyScroll < 0 {
		m.keyScroll = 0
	}
	if m.keyScroll > maxScroll {
		m.keyScroll = maxScroll
	}
	if rowIndex < m.keyScroll {
		m.keyScroll = rowIndex
		return
	}
	if rowIndex >= m.keyScroll+height {
		m.keyScroll = rowIndex - height + 1
	}
}

func (m *Model) scrollKeyList(delta int) {
	rows := m.visibleKeyRows()
	maxScroll := maxInt(len(rows)-m.keyListHeight(), 0)
	m.keyScroll += delta
	if m.keyScroll < 0 {
		m.keyScroll = 0
	}
	if m.keyScroll > maxScroll {
		m.keyScroll = maxScroll
	}
}

func (m *Model) toggleKeyTreeNode(path string) {
	if m.keyTreeExpanded == nil {
		m.keyTreeExpanded = map[string]bool{}
	}
	m.keyTreeExpanded[path] = !m.keyTreeExpanded[path]
}

func parentKeyTreePath(key string) string {
	parts := strings.Split(key, keyTreeSeparator)
	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], keyTreeSeparator)
}

func (m *Model) expandSelectedKeyBranch() {
	if !m.keyTreeMode || m.selectedKey < 0 || m.selectedKey >= len(m.keys) {
		return
	}
	path := parentKeyTreePath(m.keys[m.selectedKey].Name)
	if path == "" {
		return
	}
	if m.keyTreeExpanded == nil {
		m.keyTreeExpanded = map[string]bool{}
	}
	m.keyTreeExpanded[path] = true
}

func (m *Model) collapseSelectedKeyBranch() {
	if !m.keyTreeMode || m.selectedKey < 0 || m.selectedKey >= len(m.keys) {
		return
	}
	path := parentKeyTreePath(m.keys[m.selectedKey].Name)
	if path == "" {
		return
	}
	if m.keyTreeExpanded == nil {
		m.keyTreeExpanded = map[string]bool{}
	}
	m.keyTreeExpanded[path] = false
}

func (m Model) renderKeyRows() string {
	rows := m.visibleKeyRows()
	height := m.keyListHeight()
	start := m.keyScroll
	if start < 0 {
		start = 0
	}
	if start > maxInt(len(rows)-height, 0) {
		start = maxInt(len(rows)-height, 0)
	}
	end := start + height
	if end > len(rows) {
		end = len(rows)
	}

	lines := make([]string, 0, end-start+2)
	if start > 0 {
		lines = append(lines, subtitleStyle.Render("  ▲ more above"))
	}
	for _, row := range rows[start:end] {
		prefix := strings.Repeat("  ", row.depth)
		if row.expandable {
			marker := "▸"
			if row.expanded {
				marker = "▾"
			}
			lines = append(lines, fmt.Sprintf("%s%s %s/", prefix, marker, row.label))
			continue
		}
		marker := "  "
		if row.keyIndex == m.selectedKey {
			marker = okStyle.Render("▸ ")
		}
		key := m.keys[row.keyIndex]
		lines = append(lines, fmt.Sprintf("%s%s%s  [%s]  ttl=%s", prefix, marker, row.label, key.Type, ttlLabel(key.TTL)))
	}
	if end < len(rows) {
		lines = append(lines, subtitleStyle.Render("  ▼ more below"))
	}
	return strings.Join(lines, "\n")
}
